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
	err := s.addProject(
		common.HexToAddress("0x9DCA89E400E8aAD271a92E6c4e2BaE86646Cb6C9"),
		common.HexToAddress("0xf54d3639B783B3d77cce0A6c2BcE042a201F8614"),
		"test-uri",
		[]common.Address{
			common.HexToAddress("0x8d2CfE86E5bc0D1a99f7848CE96B86AbCe72413F"),
			common.HexToAddress("0x4B763D578273F5704cF5D57cF4A46452Ef5Cd659"),
		},
	)
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// assert project was added
	pq := s.testRepo.ProjectQuery()
	pq.WhereOwner(&types.Address{Address: common.HexToAddress("0x9DCA89E400E8aAD271a92E6c4e2BaE86646Cb6C9")})
	project, err := pq.GetFirstOrFail()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), uint64(1), project.ProjectId)
	assert.EqualValues(s.T(), "0x9DCA89E400E8aAD271a92E6c4e2BaE86646Cb6C9", project.OwnerAddress.Hex())
	assert.EqualValues(s.T(), "0xf54d3639B783B3d77cce0A6c2BcE042a201F8614", project.ReceiverAddress.Hex())
	assert.Nil(s.T(), project.LastWithdrawalEpoch)
	assert.EqualValues(s.T(), s.currentEpoch, project.ActiveFromEpoch)
	assert.Nil(s.T(), project.ActiveToEpoch)
	// assert contracts were added
	pcq := s.testRepo.ProjectContractQuery()
	pcq.WhereProjectId(project.ProjectId)
	projectContracts, err := pcq.GetAll()
	assert.Nil(s.T(), err)
	assert.Len(s.T(), projectContracts, 2)
	for i, pc := range projectContracts {
		assert.EqualValues(s.T(), project.Id, pc.ProjectId)
		assert.True(s.T(), pc.Enabled)
		assert.EqualValues(
			s.T(),
			[]string{
				"0x8d2CfE86E5bc0D1a99f7848CE96B86AbCe72413F",
				"0x4B763D578273F5704cF5D57cF4A46452Ef5Cd659",
			}[i],
			pc.Address.Hex())
	}
}

// TestProjectSuspensionAndActivation tests the project suspension and activation
func (s *DispatcherTestSuite) TestProjectSuspensionAndActivation() {
	err := s.addProject(
		common.HexToAddress("0x9DCA89E400E8aAD271a92E6c4e2BaE86646Cb6C9"),
		common.HexToAddress("0xf54d3639B783B3d77cce0A6c2BcE042a201F8614"),
		"test-uri",
		[]common.Address{
			common.HexToAddress("0x8d2CfE86E5bc0D1a99f7848CE96B86AbCe72413F"),
			common.HexToAddress("0x4B763D578273F5704cF5D57cF4A46452Ef5Cd659"),
		},
	)
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// shift epochs
	s.shiftEpochs(10)
	// get project
	pq := s.testRepo.ProjectQuery()
	pq.WhereOwner(&types.Address{Address: common.HexToAddress("0x9DCA89E400E8aAD271a92E6c4e2BaE86646Cb6C9")})
	project, err := pq.GetFirstOrFail()
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
	err := s.addProject(
		common.HexToAddress("0x9DCA89E400E8aAD271a92E6c4e2BaE86646Cb6C9"),
		common.HexToAddress("0xf54d3639B783B3d77cce0A6c2BcE042a201F8614"),
		"test-uri",
		[]common.Address{
			common.HexToAddress("0x8d2CfE86E5bc0D1a99f7848CE96B86AbCe72413F"),
			common.HexToAddress("0x4B763D578273F5704cF5D57cF4A46452Ef5Cd659"),
		},
	)
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// assert the contract does not exist
	pcq := s.testRepo.ProjectContractQuery()
	pcq.WhereAddress(&types.Address{Address: common.HexToAddress("0xFA76958E85faB71A8B9e4AAAE59D55a1f54f2B42")})
	_, err = pcq.GetFirstOrFail()
	assert.NotNil(s.T(), err)
	// add contract
	_, err = s.projectsManagerSession.AddProjectContract(
		new(big.Int).SetUint64(1),
		common.HexToAddress("0xFA76958E85faB71A8B9e4AAAE59D55a1f54f2B42"),
	)
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// get project
	pq := s.testRepo.ProjectQuery()
	pq.WhereOwner(&types.Address{Address: common.HexToAddress("0x9DCA89E400E8aAD271a92E6c4e2BaE86646Cb6C9")})
	project, err := pq.GetFirstOrFail()
	assert.Nil(s.T(), err)
	// assert the contract exists
	c, err := pcq.GetFirstOrFail()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), project.Id, c.ProjectId)
	assert.EqualValues(s.T(), common.HexToAddress("0xFA76958E85faB71A8B9e4AAAE59D55a1f54f2B42"), c.Address.Address)
	assert.True(s.T(), c.Enabled)
}

// TestAddContract tests the remove contract functionality
func (s *DispatcherTestSuite) TestRemoveContract() {
	err := s.addProject(
		common.HexToAddress("0x9DCA89E400E8aAD271a92E6c4e2BaE86646Cb6C9"),
		common.HexToAddress("0xf54d3639B783B3d77cce0A6c2BcE042a201F8614"),
		"test-uri",
		[]common.Address{
			common.HexToAddress("0x8d2CfE86E5bc0D1a99f7848CE96B86AbCe72413F"),
			common.HexToAddress("0x4B763D578273F5704cF5D57cF4A46452Ef5Cd659"),
		},
	)
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// assert the contract does exist
	pcq := s.testRepo.ProjectContractQuery()
	pcq.WhereAddress(&types.Address{Address: common.HexToAddress("0x8d2CfE86E5bc0D1a99f7848CE96B86AbCe72413F")})
	_, err = pcq.GetFirstOrFail()
	assert.Nil(s.T(), err)
	// remove contract
	_, err = s.projectsManagerSession.RemoveProjectContract(
		new(big.Int).SetUint64(1),
		common.HexToAddress("0x8d2CfE86E5bc0D1a99f7848CE96B86AbCe72413F"),
	)
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// assert the contract does not exist
	_, err = pcq.GetFirstOrFail()
	assert.NotNil(s.T(), err)
}

// initializeSfc deploys the sfc mock contract to the test chain
func (s *DispatcherTestSuite) initializeSfc() {
	auth, err := bind.NewKeyedTransactorWithChainID(s.testChain.AdminAcc.PrivateKey, big.NewInt(TestChainId))
	if err != nil {
		s.T().Fatal(err)
	}
	address, _, _, err := contracts.DeploySfcMock(auth, s.testChain.RawRpc)
	if err != nil {
		s.T().Fatal(err)
	}
	sfcMock, err := contracts.NewSfcMock(address, s.testChain.RawRpc)
	if err != nil {
		s.T().Fatal(err)
	}
	s.sfcMock = sfcMock
	s.sfcMockAddr = address
}

// initializeGasMonetization deploys the contract to the test chain
func (s *DispatcherTestSuite) initializeGasMonetization() {
	auth, err := bind.NewKeyedTransactorWithChainID(s.testChain.AdminAcc.PrivateKey, big.NewInt(TestChainId))
	if err != nil {
		s.T().Fatal(err)
	}
	address, _, _, err := contracts.DeployGasMonetization(
		auth, s.testChain.RawRpc, s.sfcMockAddr, big.NewInt(withdrawalFrequency), big.NewInt(withdrawalConfirmations),
	)
	if err != nil {
		s.T().Fatal(err)
	}
	gasMonetization, err := contracts.NewGasMonetization(address, s.testChain.RawRpc)
	if err != nil {
		s.T().Fatal(err)
	}
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
}

// initializeSfcSession initializes a session for sfc mock
func (s *DispatcherTestSuite) initializeSfcSession() {
	auth, err := bind.NewKeyedTransactorWithChainID(s.testChain.AdminAcc.PrivateKey, big.NewInt(TestChainId))
	if err != nil {
		s.T().Fatal(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}
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
	if err != nil {
		s.T().Fatal(err)
	}
	return block.ToInt()
}

// initializeGasMonetizationRoles assigns roles to the test accounts
func (s *DispatcherTestSuite) initializeGasMonetizationRoles() {
	funderRole, err := s.gasMonetization.FUNDERROLE(nil)
	if err != nil {
		s.T().Fatal(err)
	}
	_, err = s.adminSession.GrantRole(funderRole, s.testChain.FunderAcc.Address)
	if err != nil {
		s.T().Fatal(err)
	}
	projectsManagerRole, err := s.gasMonetization.PROJECTSMANAGERROLE(nil)
	if err != nil {
		s.T().Fatal(err)
	}
	_, err = s.adminSession.GrantRole(projectsManagerRole, s.testChain.ProjectsManagerAcc.Address)
	if err != nil {
		s.T().Fatal(err)
	}
}

func addFunds(session *contracts.GasMonetizationSession, amount *big.Int) error {
	// set amount to send
	session.TransactOpts.Value = amount
	_, err := session.AddFunds()
	if err != nil {
		return err
	}
	// reset amount to send
	session.TransactOpts.Value = nil
	return nil
}

// shiftEpochs shifts the current epoch by the given number of epochs
func (s *DispatcherTestSuite) shiftEpochs(epochs uint64) {
	s.currentEpoch += epochs
	_, err := s.sfcSession.SetEpoch(big.NewInt(int64(s.currentEpoch)))
	if err != nil {
		s.T().Fatal("failed to shift epoch: ", err)
	}
}

// addProject adds a project to gas monetization contract
func (s *DispatcherTestSuite) addProject(
	owner common.Address,
	recipient common.Address,
	metadataUri string,
	contracts []common.Address,
) error {
	_, err := s.projectsManagerSession.AddProject(owner, recipient, metadataUri, contracts)
	if err != nil {
		return err
	}
	return nil
}

// getLatestBlock returns the latest block
func (s *DispatcherTestSuite) getLatestBlock() *types.Block {
	block, err := s.testRepo.BlockByNumber(nil)
	if err != nil {
		s.T().Fatal(err)
	}
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
