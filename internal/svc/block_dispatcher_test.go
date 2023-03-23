package svc

import (
	"crypto/ecdsa"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/rpc/contracts"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"math/big"
	"testing"
	"time"
)

const (
	withdrawalFrequency     = 10
	withdrawalConfirmations = 1
)

var (
	projectOwner     = types.Address{Address: common.HexToAddress("0x7E618Ee2D08fcb730f3fd8C3F4e7C7Fd1A166ABD")}
	projectRecipient = types.Address{Address: common.HexToAddress("0xf54d3639B783B3d77cce0A6c2BcE042a201F8614")}
	projectUri       = "test-uri"
	projectContracts = []types.Address{
		{Address: common.HexToAddress("0x8d2CfE86E5bc0D1a99f7848CE96B86AbCe72413F")},
		{Address: common.HexToAddress("0x4B763D578273F5704cF5D57cF4A46452Ef5Cd659")},
	}
)

type DispatcherTestSuite struct {
	suite.Suite
	testChain *TestChain
	testDb    *db.TestDatabase
	testRepo  *repository.Repository
	// gas monetization contract and sessions
	gasMonetizationAddr    common.Address
	gasMonetization        *contracts.GasMonetization
	adminSession           *contracts.GasMonetizationSession
	funderSession          *contracts.GasMonetizationSession
	projectsManagerSession *contracts.GasMonetizationSession
	projectOwnerSession    *contracts.GasMonetizationSession
	// sfc mock for obtaining the current epoch and session
	sfcMockAddr common.Address
	sfcMock     *contracts.SfcMock
	sfcSession  *contracts.SfcMockSession
	// current epoch
	currentEpoch uint64
	// blkDispatcher is the block dispatcher to test
	blkDispatcher blkDispatcher
}

func TestDispatcherTestSuite(t *testing.T) {
	suite.Run(t, new(DispatcherTestSuite))
}

// SetupSuite sets up the test suite and initializes the block dispatcher
func (s *DispatcherTestSuite) SetupSuite() {
	testLogger := logger.New(log.Writer(), "test", logging.ERROR)
	// initialize dependencies
	s.testChain = SetupTestChain(testLogger)
	s.testDb = db.SetupTestDatabase(testLogger)
	s.testRepo = repository.NewWithInstances(s.testDb.Db, s.testChain.Rpc, testLogger)
	// initialize block dispatcher
	s.blkDispatcher = blkDispatcher{
		service: service{
			repo: s.testRepo,
			log:  testLogger,
			mgr:  New(s.testRepo, testLogger),
		},
	}
	s.blkDispatcher.init()
	// make channel for receiving blocks
	s.blkDispatcher.inBlock = make(chan *types.Block)
	// make channel for receiving dispatched block ids
	s.blkDispatcher.outDispatched = make(chan uint64)
	s.blkDispatcher.run()
}

// SetupTest sets up the test
func (s *DispatcherTestSuite) SetupTest() {
	s.initializeSfc()
	s.initializeSfcSession()
	s.initializeGasMonetization()
	s.initializeGasMonetizationSessions()
	s.initializeGasMonetizationRoles()
	// shift epoch by one by beginning of the test
	s.shiftEpochs(s.currentEpoch + 1)
}

// TearDownSuite tears down the test suite
func (s *DispatcherTestSuite) TearDownSuite() {
	s.blkDispatcher.close()
	s.testChain.TearDown()
	s.testDb.TearDown()
}

// TestShiftEpochs tests the shiftEpochs function
func (s *DispatcherTestSuite) TestShiftEpochs() {
	initialEpoch, err := s.sfcSession.CurrentEpoch()
	assert.Nil(s.T(), err)
	s.shiftEpochs(50)
	assert.EqualValues(s.T(), initialEpoch.Uint64()+50, s.currentEpoch)
}

// TestAddProject tests the addProject function
func (s *DispatcherTestSuite) TestAddProject() {
	s.setupTestProject()
	// assert project was added
	pq := s.testRepo.ProjectQuery()
	project, err := pq.WhereOwner(&projectOwner).GetFirstOrFail()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), uint64(1), project.ProjectId)
	assert.EqualValues(s.T(), projectOwner.Hex(), project.OwnerAddress.Hex())
	assert.EqualValues(s.T(), projectRecipient.Hex(), project.ReceiverAddress.Hex())
	assert.Nil(s.T(), project.LastWithdrawalEpoch)
	assert.EqualValues(s.T(), s.currentEpoch, project.ActiveFromEpoch)
	assert.Nil(s.T(), project.ActiveToEpoch)
	// assert contracts were added
	pcq := s.testRepo.ProjectContractQuery()
	pcl, err := pcq.WhereProjectId(project.ProjectId).GetAll()
	assert.Nil(s.T(), err)
	assert.Len(s.T(), projectContracts, 2)
	for i, pc := range pcl {
		assert.EqualValues(s.T(), project.Id, pc.ProjectId)
		assert.True(s.T(), pc.Enabled)
		assert.EqualValues(s.T(), projectContracts[i].Address.Hex(), pc.Address.Hex())
	}
}

// TestProjectSuspensionAndActivation tests the project suspension and activation
func (s *DispatcherTestSuite) TestProjectSuspensionAndActivation() {
	s.setupTestProject()
	// shift epochs
	s.shiftEpochs(10)
	// get project
	pq := s.testRepo.ProjectQuery()
	project, err := pq.WhereOwner(&projectOwner).GetFirstOrFail()
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), project.ActiveToEpoch)
	// suspend project
	_, err = s.projectsManagerSession.SuspendProject(new(big.Int).SetUint64(project.ProjectId))
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// fetch project again and assert it was suspended
	project, err = pq.GetFirstOrFail()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), s.currentEpoch, *project.ActiveToEpoch)
	// shift epochs
	s.shiftEpochs(10)
	// re-enable project
	_, err = s.projectsManagerSession.EnableProject(new(big.Int).SetUint64(project.ProjectId))
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// fetch project again and assert it was re-enabled
	project, err = pq.GetFirstOrFail()
	assert.Nil(s.T(), err)
	assert.Nil(s.T(), project.ActiveToEpoch)
	assert.EqualValues(s.T(), s.currentEpoch, project.ActiveFromEpoch)
}

// TestAddContract tests the add contract functionality
func (s *DispatcherTestSuite) TestAddContract() {
	s.setupTestProject()
	contractAddr := types.Address{Address: common.HexToAddress("0xFA76958E85faB71A8B9e4AAAE59D55a1f54f2B42")}
	// assert the contract does not exist
	pcq := s.testRepo.ProjectContractQuery()
	_, err := pcq.WhereAddress(&contractAddr).GetFirstOrFail()
	assert.NotNil(s.T(), err)
	// add contract
	_, err = s.projectsManagerSession.AddProjectContract(new(big.Int).SetUint64(1), contractAddr.Address)
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// get project
	pq := s.testRepo.ProjectQuery()
	project, err := pq.WhereOwner(&projectOwner).GetFirstOrFail()
	assert.Nil(s.T(), err)
	// assert the contract exists
	c, err := pcq.GetFirstOrFail()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), project.Id, c.ProjectId)
	assert.EqualValues(s.T(), contractAddr.Hex(), c.Address.Hex())
	assert.True(s.T(), c.Enabled)
}

// TestAddContract tests the remove contract functionality
func (s *DispatcherTestSuite) TestRemoveContract() {
	s.setupTestProject()
	// assert the contract does exist
	pcq := s.testRepo.ProjectContractQuery()
	_, err := pcq.WhereAddress(&projectContracts[0]).GetFirstOrFail()
	assert.Nil(s.T(), err)
	// remove contract
	_, err = s.projectsManagerSession.RemoveProjectContract(new(big.Int).SetUint64(1), projectContracts[0].Address)
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// assert the contract does not exist
	_, err = pcq.GetFirstOrFail()
	assert.NotNil(s.T(), err)
	// assert the other contract still exists
	pcq = s.testRepo.ProjectContractQuery()
	_, err = pcq.WhereAddress(&projectContracts[1]).GetFirstOrFail()
	assert.Nil(s.T(), err)
}

// TestUpdateUri tests the update uri functionality
func (s *DispatcherTestSuite) TestUpdateUri() {
	s.setupTestProject()
	// assert the contract does exist
	pq := s.testRepo.ProjectQuery()
	_, err := pq.WhereOwner(&projectOwner).GetFirstOrFail()
	assert.Nil(s.T(), err)
	// remove contract
	_, err = s.projectsManagerSession.UpdateProjectMetadataUri(
		new(big.Int).SetUint64(1),
		"new-uri",
	)
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// TODO: assert metadata changed
}

// TestUpdateRewardsRecipient tests the update rewards recipient functionality
func (s *DispatcherTestSuite) TestUpdateRewardsRecipient() {
	s.setupTestProject()
	// fetch project
	pq := s.testRepo.ProjectQuery()
	project, err := pq.WhereOwner(&projectOwner).GetFirstOrFail()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), projectRecipient.Hex(), project.ReceiverAddress.Hex())
	// update recipient
	addr := types.Address{Address: common.HexToAddress("0xFA76958E85faB71A8B9e4AAAE59D55a1f54f2B42")}
	_, err = s.projectOwnerSession.UpdateProjectRewardsRecipient(new(big.Int).SetUint64(1), addr.Address)
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// assert recipient updated
	project, err = pq.GetFirstOrFail()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), addr.Hex(), project.ReceiverAddress.Hex())
}

// TestUpdateOwner tests the update owner functionality
func (s *DispatcherTestSuite) TestUpdateOwner() {
	s.setupTestProject()
	// fetch project
	pq := s.testRepo.ProjectQuery()
	project, err := pq.WhereOwner(&projectOwner).GetFirstOrFail()
	id := project.Id
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), projectRecipient.Hex(), project.ReceiverAddress.Hex())
	// update recipient
	addr := types.Address{Address: common.HexToAddress("0xFA76958E85faB71A8B9e4AAAE59D55a1f54f2B42")}
	_, err = s.projectsManagerSession.UpdateProjectOwner(new(big.Int).SetUint64(1), addr.Address)
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// assert owner updated
	pq = s.testRepo.ProjectQuery()
	project, err = pq.WhereOwner(&addr).GetFirstOrFail()
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), project)
	assert.EqualValues(s.T(), id, project.Id)
}

// TestUpdateOwner tests the update owner functionality
func (s *DispatcherTestSuite) TestWithdrawalRequest() {
	s.setupTestProject()
	s.shiftEpochs(100)
	// fund the contract
	s.fundContract(new(big.Int).SetUint64(5_000))
	// request withdrawal
	_, err := s.projectOwnerSession.RequestWithdrawal(new(big.Int).SetUint64(1))
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// fetch project
	pq := s.testRepo.ProjectQuery()
	project, err := pq.WhereOwner(&projectOwner).GetFirstOrFail()
	assert.Nil(s.T(), err)
	// assert withdrawal request exists
	wrq := s.testRepo.WithdrawalRequestQuery()
	wr, err := wrq.WhereProjectId(project.Id).GetFirstOrFail()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), project.Id, wr.ProjectId)
	assert.EqualValues(s.T(), s.currentEpoch, wr.Epoch)
}

// initializeSfc deploys the sfc mock contract to the test chain
func (s *DispatcherTestSuite) initializeSfc() {
	auth, err := bind.NewKeyedTransactorWithChainID(s.testChain.AdminAcc.PrivateKey, big.NewInt(TestChainId))
	assert.Nil(s.T(), err)
	address, _, _, err := contracts.DeploySfcMock(auth, s.testChain.RawRpc)
	assert.Nil(s.T(), err)
	sfcMock, err := contracts.NewSfcMock(address, s.testChain.RawRpc)
	assert.Nil(s.T(), err)
	s.sfcMock = sfcMock
	s.sfcMockAddr = address
}

// initializeGasMonetization deploys the contract to the test chain
func (s *DispatcherTestSuite) initializeGasMonetization() {
	auth, err := bind.NewKeyedTransactorWithChainID(s.testChain.AdminAcc.PrivateKey, big.NewInt(TestChainId))
	assert.Nil(s.T(), err)
	address, _, _, err := contracts.DeployGasMonetization(
		auth, s.testChain.RawRpc, s.sfcMockAddr, big.NewInt(withdrawalFrequency), big.NewInt(withdrawalConfirmations),
	)
	assert.Nil(s.T(), err)
	gasMonetization, err := contracts.NewGasMonetization(address, s.testChain.RawRpc)
	assert.Nil(s.T(), err)
	s.gasMonetization = gasMonetization
	s.gasMonetizationAddr = address
}

//
//func TestLastProcessedBlock(t *testing.T) {
//	privateKey, err := crypto.HexToECDSA("bb39aa88008bc6260ff9ebc816178c47a01c44efe55810ea1f271c00f5878812")
//	if err != nil {
//		log.Fatal(err)
//	}
//	publicKey := privateKey.Public()
//	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
//	if !ok {
//		log.Fatal("error casting public key to ECDSA")
//	}
//	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
//	nonce, err := testChain.RawRpc.PendingNonceAt(context.Background(), fromAddress)
//	if err != nil {
//		log.Fatal(err)
//	}
//	value := big.NewInt(1000000000000000000)
//	gasLimit := uint64(21000)
//	gasPrice := big.NewInt(6721975)
//	toAddress := common.HexToAddress(svc.TestAccount2)
//	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
//	chainID, err := testChain.RawRpc.NetworkID(context.Background())
//	if err != nil {
//		log.Fatal(err)
//	}
//	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
//	if err != nil {
//		log.Fatal(err)
//	}
//	err = testChain.RawRpc.SendTransaction(context.Background(), signedTx)
//	if err != nil {
//		log.Fatal(err)
//	}
//	t.Logf("tx sent: %s", signedTx.Hash().Hex())
//
//	blockHeigh, err := testChain.BlockHeight()
//	if err != nil {
//		t.Fatal(err)
//	}
//	t.Logf("block height: %s", blockHeigh.String())
//}

// initializeGasMonetizationSessions initializes sessions for the test accounts
func (s *DispatcherTestSuite) initializeGasMonetizationSessions() {
	s.adminSession = initializeGasMonetizationSession(s.T(), s.gasMonetization, s.testChain.AdminAcc.PrivateKey)
	s.funderSession = initializeGasMonetizationSession(s.T(), s.gasMonetization, s.testChain.FunderAcc.PrivateKey)
	s.projectsManagerSession = initializeGasMonetizationSession(s.T(), s.gasMonetization, s.testChain.ProjectsManagerAcc.PrivateKey)
	s.projectOwnerSession = initializeGasMonetizationSession(s.T(), s.gasMonetization, s.testChain.ProjectOwnerAcc.PrivateKey)
}

// initializeSfcSession initializes a session for sfc mock
func (s *DispatcherTestSuite) initializeSfcSession() {
	auth, err := bind.NewKeyedTransactorWithChainID(s.testChain.AdminAcc.PrivateKey, big.NewInt(TestChainId))
	assert.Nil(s.T(), err)
	s.sfcSession = &contracts.SfcMockSession{
		Contract: s.sfcMock,
		CallOpts: bind.CallOpts{},
		TransactOpts: bind.TransactOpts{
			From:     auth.From,
			Signer:   auth.Signer,
			GasLimit: TestChainGasLimit,
		},
	}
}

// initializeGasMonetizationSession initializes a session for the given account
func initializeGasMonetizationSession(
	t *testing.T,
	gasMonetization *contracts.GasMonetization,
	key *ecdsa.PrivateKey) *contracts.GasMonetizationSession {
	auth, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(TestChainId))
	assert.Nil(t, err)
	return &contracts.GasMonetizationSession{
		Contract: gasMonetization,
		CallOpts: bind.CallOpts{},
		TransactOpts: bind.TransactOpts{
			From:     auth.From,
			Signer:   auth.Signer,
			GasLimit: TestChainGasLimit,
		},
	}
}

// getCurrentBlockId returns the current block id
func (s *DispatcherTestSuite) getCurrentBlockId() *big.Int {
	block, err := s.testChain.BlockHeight()
	assert.Nil(s.T(), err)
	return block.ToInt()
}

// initializeGasMonetizationRoles assigns roles to the test accounts
func (s *DispatcherTestSuite) initializeGasMonetizationRoles() {
	funderRole, err := s.gasMonetization.FUNDERROLE(nil)
	assert.Nil(s.T(), err)
	_, err = s.adminSession.GrantRole(funderRole, s.testChain.FunderAcc.Address)
	assert.Nil(s.T(), err)
	projectsManagerRole, err := s.gasMonetization.PROJECTSMANAGERROLE(nil)
	assert.Nil(s.T(), err)
	_, err = s.adminSession.GrantRole(projectsManagerRole, s.testChain.ProjectsManagerAcc.Address)
	assert.Nil(s.T(), err)
}

// shiftEpochs shifts the current epoch by the given number of epochs
func (s *DispatcherTestSuite) shiftEpochs(epochs uint64) {
	s.currentEpoch += epochs
	_, err := s.sfcSession.SetEpoch(big.NewInt(int64(s.currentEpoch)))
	assert.Nil(s.T(), err)
}

// addProject adds a project to gas monetization contract
func (s *DispatcherTestSuite) addProject(
	owner common.Address,
	recipient common.Address,
	metadataUri string,
	contracts []common.Address,
) {
	_, err := s.projectsManagerSession.AddProject(owner, recipient, metadataUri, contracts)
	assert.Nil(s.T(), err)
}

// fundContract funds the gas monetization contract
func (s *DispatcherTestSuite) fundContract(amount *big.Int) {
	// set amount to send
	s.funderSession.TransactOpts.Value = amount
	_, err := s.funderSession.AddFunds()
	assert.Nil(s.T(), err)
	// reset amount to send
	s.funderSession.TransactOpts.Value = nil
}

// addProject adds a project to gas monetization contract
func (s *DispatcherTestSuite) setupTestProject() {
	var contractAddresses []common.Address
	for _, addr := range projectContracts {
		contractAddresses = append(contractAddresses, addr.Address)
	}
	s.addProject(projectOwner.Address, projectRecipient.Address, projectUri, contractAddresses)
	// process the latest block
	s.processBlock(s.getLatestBlock())
}

// getLatestBlock returns the latest block
func (s *DispatcherTestSuite) getLatestBlock() *types.Block {
	block, err := s.testRepo.BlockByNumber(nil)
	assert.Nil(s.T(), err)
	return block
}

// processBlock processes the given block by sending it to the dispatcher
func (s *DispatcherTestSuite) processBlock(blk *types.Block) {
	// inject epoch number into block, because it is not set by the test chain
	blk.Epoch = hexutil.Uint64(s.currentEpoch)
	// send block to dispatcher
	s.blkDispatcher.inBlock <- blk
	// wait for block to be processed
	select {
	case id, ok := <-s.blkDispatcher.outDispatched:
		if !ok || id != uint64(blk.Number) {
			s.T().Fatal("block not processed")
		}
	case <-time.After(5 * time.Second):
		s.T().Fatal("block not processed")
	}
}
