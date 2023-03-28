package svc

import (
	"context"
	"crypto/ecdsa"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/db"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/rpc"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/rpc/contracts"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/types"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	eth "github.com/ethereum/go-ethereum/core/types"
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
	testRpc   *rpc.Rpc
	testDb    *db.TestDatabase
	testRepo  *repository.Repository
	// gas monetization contract and sessions
	gasMonetizationAddr    common.Address
	gasMonetization        *contracts.GasMonetization
	adminSession           *contracts.GasMonetizationSession
	funderSession          *contracts.GasMonetizationSession
	projectsManagerSession *contracts.GasMonetizationSession
	projectOwnerSession    *contracts.GasMonetizationSession
	dataProviderSession    *contracts.GasMonetizationSession
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
	s.testRpc = s.testChain.SetupTestRpc(s.gasMonetizationAddr, testLogger)
	s.testDb = db.SetupTestDatabase(testLogger)
	s.testRepo = repository.NewWithInstances(s.testDb.Db, s.testRpc, testLogger)
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
	// set data provider session, because contract is re-deployed
	// on every test, so the session is lost
	s.testRpc.SetDataProviderSession(s.dataProviderSession)
	// migrate tables to ensure they are empty
	err := s.testDb.Migrate()
	assert.Nil(s.T(), err)
	s.blkDispatcher.init()
	// shift epoch by one by beginning of the test
	s.shiftEpochs(s.currentEpoch + 1)
}

// TearDownTest tears down the test
func (s *DispatcherTestSuite) TearDownTest() {
	// drop all tables on teardown
	err := s.testDb.Drop()
	assert.Nil(s.T(), err)
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
	pcl, err := pcq.WhereProjectId(project.Id).GetAll()
	assert.Nil(s.T(), err)
	assert.Len(s.T(), projectContracts, 2)
	for i, pc := range pcl {
		assert.EqualValues(s.T(), project.Id, pc.ProjectId)
		assert.True(s.T(), pc.Approved)
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
	assert.True(s.T(), c.Approved)
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
func (s *DispatcherTestSuite) TestCollectRelatedTransactions() {
	s.setupTestProject()
	// send 10 related transactions
	for i := 0; i < 10; i++ {
		s.sendTransaction(s.testChain.FunderAcc, projectContracts[0].Address, big.NewInt(1_000))
		s.processBlock(s.getLatestBlock())
	}
	tq := s.testRepo.TransactionQuery()
	transactions, err := tq.GetAll()
	assert.Nil(s.T(), err)
	assert.Len(s.T(), transactions, 10)
	// send 10 unrelated transactions
	for i := 0; i < 10; i++ {
		s.sendTransaction(s.testChain.FunderAcc, s.testChain.ProjectOwnerAcc.Address, big.NewInt(1_000))
		s.processBlock(s.getLatestBlock())
	}
	transactions, err = tq.GetAll()
	assert.Nil(s.T(), err)
	assert.Len(s.T(), transactions, 10)
}

// TestAddingContractWillCollectTransactions tests that adding a contract will collect transactions
func (s *DispatcherTestSuite) TestAddingContractWillCollectTransactions() {
	s.setupTestProject()
	addr := common.HexToAddress("0x91567C6F4B31cd51dFE6ADE09579d43240187FF1")
	// send 10 transactions
	for i := 0; i < 10; i++ {
		s.sendTransaction(s.testChain.FunderAcc, addr, big.NewInt(1_000))
		s.processBlock(s.getLatestBlock())
	}
	// verify that no transactions were collected
	tq := s.testRepo.TransactionQuery()
	transactions, err := tq.GetAll()
	assert.Nil(s.T(), err)
	assert.Len(s.T(), transactions, 0)
	// add new contract
	_, err = s.projectsManagerSession.AddProjectContract(new(big.Int).SetUint64(1), addr)
	assert.Nil(s.T(), err)
	s.processBlock(s.getLatestBlock())
	// send 10 transactions
	for i := 0; i < 10; i++ {
		s.sendTransaction(s.testChain.FunderAcc, addr, big.NewInt(1_000))
		s.processBlock(s.getLatestBlock())
	}
	// verify that transactions were collected
	transactions, err = tq.GetAll()
	assert.Nil(s.T(), err)
	assert.Len(s.T(), transactions, 10)
}

// TestAddingContractForSuspendedProjectWontCollectTransactions tests that adding a contract for a suspended
// project won't collect transactions
func (s *DispatcherTestSuite) TestAddingContractForSuspendedProjectWontCollectTransactions() {
	s.setupTestProject()
	// suspend project
	_, err := s.projectsManagerSession.SuspendProject(new(big.Int).SetUint64(1))
	assert.Nil(s.T(), err)
	s.processBlock(s.getLatestBlock())
	// add new contract
	addr := common.HexToAddress("0x91567C6F4B31cd51dFE6ADE09579d43240187FF1")
	_, err = s.projectsManagerSession.AddProjectContract(new(big.Int).SetUint64(1), addr)
	assert.Nil(s.T(), err)
	s.processBlock(s.getLatestBlock())
	// send 10 transactions
	for i := 0; i < 10; i++ {
		s.sendTransaction(s.testChain.FunderAcc, addr, big.NewInt(1_000))
		s.processBlock(s.getLatestBlock())
	}
	// verify that no transactions were collected
	tq := s.testRepo.TransactionQuery()
	transactions, err := tq.GetAll()
	assert.Nil(s.T(), err)
	assert.Len(s.T(), transactions, 0)
	// enable project
	_, err = s.projectsManagerSession.EnableProject(new(big.Int).SetUint64(1))
	assert.Nil(s.T(), err)
	s.processBlock(s.getLatestBlock())
	// send 10 transactions
	for i := 0; i < 10; i++ {
		s.sendTransaction(s.testChain.FunderAcc, addr, big.NewInt(1_000))
		s.processBlock(s.getLatestBlock())
	}
	// verify that transactions were collected
	transactions, err = tq.GetAll()
	assert.Nil(s.T(), err)
	assert.Len(s.T(), transactions, 10)
}

// TestAddingContractWillCollectTransactions tests that adding a contract will collect transactions
func (s *DispatcherTestSuite) TestDeletingContractWontCollectTransactions() {
	s.setupTestProject()
	// send 10 transactions
	for i := 0; i < 10; i++ {
		s.sendTransaction(s.testChain.FunderAcc, projectContracts[0].Address, big.NewInt(1_000))
		s.processBlock(s.getLatestBlock())
	}
	// verify that transactions were collected
	tq := s.testRepo.TransactionQuery()
	transactions, err := tq.GetAll()
	assert.Nil(s.T(), err)
	assert.Len(s.T(), transactions, 10)
	// remove contract
	_, err = s.projectsManagerSession.RemoveProjectContract(new(big.Int).SetUint64(1), projectContracts[0].Address)
	assert.Nil(s.T(), err)
	s.processBlock(s.getLatestBlock())
	// send 10 transactions
	for i := 0; i < 10; i++ {
		s.sendTransaction(s.testChain.FunderAcc, projectContracts[0].Address, big.NewInt(1_000))
		s.processBlock(s.getLatestBlock())
	}
	// verify that no transactions were collected
	transactions, err = tq.GetAll()
	assert.Nil(s.T(), err)
	assert.Len(s.T(), transactions, 10)
}

// TestAmountToWithdrawIsCalculatedCorrectly tests that amount to withdraw is calculated correctly
func (s *DispatcherTestSuite) TestAmountToWithdrawIsCalculatedCorrectly() {
	s.setupTestProject()
	s.sendTransaction(s.testChain.FunderAcc, projectContracts[0].Address, big.NewInt(1_000))
	s.processBlock(s.getLatestBlock())
	tq := s.testRepo.TransactionQuery()
	transaction, err := tq.GetFirst()
	assert.Nil(s.T(), err)
	// calculate expected amount
	a := new(big.Int).Mul(new(big.Int).SetUint64(TestChainGasPrice), new(big.Int).SetUint64(uint64(*transaction.GasUsed)))
	b := new(big.Int).Mul(a, big.NewInt(rewardsPercentage))
	expectedAmount := new(big.Int).Div(b, big.NewInt(100))
	assert.EqualValues(s.T(), expectedAmount, transaction.RewardToClaim.ToInt())
	// send 10 transactions
	for i := 0; i < 10; i++ {
		s.sendTransaction(s.testChain.FunderAcc, projectContracts[0].Address, big.NewInt(1_000))
		s.processBlock(s.getLatestBlock())
	}
	// shift epoch and send transaction to trigger rewards calculation
	s.shiftEpochs(10)
	s.sendTransaction(s.testChain.FunderAcc, projectContracts[0].Address, big.NewInt(1_000))
	s.processBlock(s.getLatestBlock())
	// we sent 11 transactions in total, expected amount should be 11 * expectedAmount
	// the 12th transaction is the one that triggers rewards calculation, but it's not
	// included in the rewards calculation for previous epoch
	totalAmount, err := s.testRepo.TotalAmountCollected()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), new(big.Int).Mul(expectedAmount, big.NewInt(11)), totalAmount)
	// verify that amount is also updated on project
	pq := s.testRepo.ProjectQuery()
	project, err := pq.WhereOwner(&projectOwner).GetFirstOrFail()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), new(big.Int).Mul(expectedAmount, big.NewInt(11)), project.CollectedRewards.ToInt())
	assert.EqualValues(s.T(), new(big.Int).Mul(expectedAmount, big.NewInt(11)), project.RewardsToClaim.ToInt())
}

// TestNumberOfTransactionsIsStoredCorrectly tests that number of transactions is stored correctly
func (s *DispatcherTestSuite) TestNumberOfTransactionsIsStoredCorrectly() {
	s.setupTestProject()
	// send 10 transactions
	for i := 0; i < 10; i++ {
		s.sendTransaction(s.testChain.FunderAcc, projectContracts[0].Address, big.NewInt(1_000))
		s.processBlock(s.getLatestBlock())
	}
	// shift epoch and send transaction to trigger rewards calculation
	s.shiftEpochs(10)
	s.sendTransaction(s.testChain.FunderAcc, projectContracts[0].Address, big.NewInt(1_000))
	s.processBlock(s.getLatestBlock())
	// 10 transactions should be stored, the 11th transaction is the one that triggers rewards calculation,
	// but it's not included in the rewards calculation for previous epoch
	totalCount, err := s.testRepo.TotalTransactionsCount()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 10, totalCount)
	// verify that number of transactions is also updated on project
	pq := s.testRepo.ProjectQuery()
	project, err := pq.WhereOwner(&projectOwner).GetFirstOrFail()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), 10, project.TransactionsCount)
}

// TestWithdrawal tests that withdrawal works correctly
func (s *DispatcherTestSuite) TestWithdrawal() {
	s.setupTestProject()
	// fund the contract with required amount - 21_000 stands for transaction cost
	requiredAmount := 10 * TestChainGasPrice * 21_000 * rewardsPercentage / 100
	s.fundContract(new(big.Int).SetUint64(uint64(requiredAmount)))
	// send 10 transactions
	for i := 0; i < 10; i++ {
		s.sendTransaction(s.testChain.FunderAcc, projectContracts[0].Address, big.NewInt(1_000))
		s.processBlock(s.getLatestBlock())
	}
	// shift epoch to trigger rewards calculation
	s.shiftEpochs(withdrawalFrequency)
	// request withdrawal
	_, err := s.projectOwnerSession.RequestWithdrawal(new(big.Int).SetUint64(1))
	assert.Nil(s.T(), err)
	// process the latest block
	s.processBlock(s.getLatestBlock())
	// process next block that contains event about withdrawal
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
	// assert that withdrawal was executed
	totalClaimed, err := s.testRepo.TotalAmountClaimed()
	assert.Nil(s.T(), err)
	assert.EqualValues(s.T(), project.ClaimedRewards.ToInt(), totalClaimed)
	assert.EqualValues(s.T(), 0, project.RewardsToClaim.ToInt().Uint64())
}

//// TestClaimingRewards tests that rewards can be claimed
//func (s *DispatcherTestSuite) TestProjectWithdrawal() {
//	s.setupTestProject()
//	// send 10 transactions
//	for i := 0; i < 10; i++ {
//		s.sendTransaction(s.testChain.FunderAcc, projectContracts[0].Address, big.NewInt(1_000))
//		s.processBlock(s.getLatestBlock())
//	}
//}

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

// sendTransaction sends a transaction from the given account to the given address
func (s *DispatcherTestSuite) sendTransaction(from *testAccount, to common.Address, value *big.Int) {
	nonce, err := s.testChain.RawRpc.PendingNonceAt(context.Background(), from.Address)
	assert.Nil(s.T(), err)
	tx := eth.NewTx(&eth.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(TestChainGasPrice),
		Gas:      TestChainGasLimit,
		To:       &to,
		Value:    value,
	})
	signedTx, err := eth.SignTx(tx, eth.NewEIP155Signer(big.NewInt(TestChainId)), from.PrivateKey)
	assert.Nil(s.T(), err)
	err = s.testChain.RawRpc.SendTransaction(context.Background(), signedTx)
	assert.Nil(s.T(), err)
}

// initializeGasMonetizationSessions initializes sessions for the test accounts
func (s *DispatcherTestSuite) initializeGasMonetizationSessions() {
	s.adminSession = initializeGasMonetizationSession(s.T(), s.gasMonetization, s.testChain.AdminAcc.PrivateKey)
	s.funderSession = initializeGasMonetizationSession(s.T(), s.gasMonetization, s.testChain.FunderAcc.PrivateKey)
	s.projectsManagerSession = initializeGasMonetizationSession(s.T(), s.gasMonetization, s.testChain.ProjectsManagerAcc.PrivateKey)
	s.projectOwnerSession = initializeGasMonetizationSession(s.T(), s.gasMonetization, s.testChain.ProjectOwnerAcc.PrivateKey)
	s.dataProviderSession = initializeGasMonetizationSession(s.T(), s.gasMonetization, s.testChain.DataProviderAcc.PrivateKey)
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
	dataProviderRole, err := s.gasMonetization.REWARDSDATAPROVIDERROLE(nil)
	assert.Nil(s.T(), err)
	_, err = s.adminSession.GrantRole(dataProviderRole, s.testChain.DataProviderAcc.Address)
}

// shiftEpochs shifts the current epoch by the given number of epochs
func (s *DispatcherTestSuite) shiftEpochs(epochs uint64) {
	s.currentEpoch += epochs
	_, err := s.sfcSession.SetEpoch(big.NewInt(int64(s.currentEpoch)))
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
	contractAddresses := utils.Map(projectContracts, func(c *types.Address) common.Address { return c.Address })
	_, err := s.projectsManagerSession.AddProject(
		projectOwner.Address, projectRecipient.Address, projectUri, contractAddresses)
	assert.Nil(s.T(), err)
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
