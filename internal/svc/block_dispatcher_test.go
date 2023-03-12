package svc_test

import (
	"crypto/ecdsa"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/rpc/contracts"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/svc"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"math/big"
	"testing"
)

const (
	withdrawalFrequency     = 10
	withdrawalConfirmations = 1
	withdrawalDeviation     = 0
)

type DispatcherTestSuite struct {
	suite.Suite
	testChain              *svc.TestChain
	gasMonetizationAddr    common.Address
	gasMonetization        *contracts.GasMonetization
	adminSession           *contracts.GasMonetizationSession
	funderSession          *contracts.GasMonetizationSession
	projectsManagerSession *contracts.GasMonetizationSession

	// initialBlockId is the block id before the test starts
	initialBlockId *big.Int
}

func TestDispatcherTestSuite(t *testing.T) {
	suite.Run(t, new(DispatcherTestSuite))
}

func (s *DispatcherTestSuite) SetupSuite() {
	s.testChain = svc.SetupTestChain(logger.New(log.Writer(), "test", logging.ERROR))
}

func (s *DispatcherTestSuite) SetupTest() {
	s.initializeContract()
	s.initializeSessions()
	s.initializeRoles()
	s.initialBlockId = s.getCurrentBlockId()
}

func (s *DispatcherTestSuite) TearDownSuite() {
	s.testChain.TearDown()
}

// TestAddFunds tests the addFunds function
func (s *DispatcherTestSuite) TestAddFunds() {
	err := addFunds(s.funderSession, big.NewInt(50_000))
	assert.Nil(s.T(), err)
	itr, err := s.gasMonetization.FilterFundsAdded(&bind.FilterOpts{
		Start: s.initialBlockId.Uint64(),
	}, []common.Address{s.testChain.FunderAcc.Address})
	assert.Nil(s.T(), err)
	assert.True(s.T(), itr.Next())
	assert.NotNil(s.T(), itr.Event)
	assert.Equal(s.T(), s.testChain.FunderAcc.Address, itr.Event.Funder)
	assert.EqualValues(s.T(), 50_000, itr.Event.Amount.Uint64())
}

// TestBlockDispatcher tests the block dispatcher
func (s *DispatcherTestSuite) TestBlockDispatcher() {

}

// initializeContract deploys the contract to the test chain
func (s *DispatcherTestSuite) initializeContract() {
	auth, err := bind.NewKeyedTransactorWithChainID(s.testChain.AdminAcc.PrivateKey, big.NewInt(svc.TestChainId))
	if err != nil {
		s.T().Fatal(err)
	}
	address, _, _, err := contracts.DeployGasMonetization(
		auth, s.testChain.RawRpc, big.NewInt(withdrawalFrequency), big.NewInt(withdrawalConfirmations),
		big.NewInt(withdrawalDeviation),
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

// initializeSessions initializes sessions for the test accounts
func (s *DispatcherTestSuite) initializeSessions() {
	s.adminSession = initializeSession(s.T(), s.gasMonetization, s.testChain.AdminAcc.PrivateKey)
	s.funderSession = initializeSession(s.T(), s.gasMonetization, s.testChain.FunderAcc.PrivateKey)
	s.projectsManagerSession = initializeSession(s.T(), s.gasMonetization, s.testChain.ProjectsManagerAcc.PrivateKey)
}

// initializeSession initializes a session for the given account
func initializeSession(
	t *testing.T,
	gasMonetization *contracts.GasMonetization,
	key *ecdsa.PrivateKey) *contracts.GasMonetizationSession {
	auth, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(svc.TestChainId))
	if err != nil {
		t.Fatal(err)
	}
	return &contracts.GasMonetizationSession{
		Contract: gasMonetization,
		CallOpts: bind.CallOpts{},
		TransactOpts: bind.TransactOpts{
			From:     auth.From,
			Signer:   auth.Signer,
			GasLimit: svc.TestChainGasLimit,
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

// initializeRoles assigns roles to the test accounts
func (s *DispatcherTestSuite) initializeRoles() {
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

func addProject(session *contracts.GasMonetizationSession, projectAddress common.Address) error {
	_, err := session.AddProject(
		// Owner address is also contract address to simplify testing
		projectAddress, "test-uri", []common.Address{projectAddress})
	if err != nil {
		return err
	}
	return nil
}
