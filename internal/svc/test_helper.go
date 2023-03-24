package svc

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/logger"
	"github.com/Mike-CZ/ftm-gas-monetization/internal/repository/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	client "github.com/ethereum/go-ethereum/ethclient"
	"github.com/testcontainers/testcontainers-go/wait"
	"strconv"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

const (
	TestChainId       = 1337
	TestChainGasLimit = 10000000000
	TestChainGasPrice = 6721975
)

type testAccount struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
}

type TestChain struct {
	RawRpc             *client.Client
	AdminAcc           *testAccount
	FunderAcc          *testAccount
	ProjectsManagerAcc *testAccount
	ProjectOwnerAcc    *testAccount
	*rpc.Rpc
	container testcontainers.Container
}

func SetupTestChain(logger *logger.AppLogger) *TestChain {
	// setup db container
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	container, port, err := createContainer(ctx, logger)
	if err != nil {
		logger.Fatal("failed to setup test", err)
	}
	c, err := client.Dial(fmt.Sprintf("http://localhost:%s", port))
	if err != nil {
		logger.Fatal("failed to connect to ganache", err)
		return nil
	}
	return &TestChain{
		Rpc:                rpc.New(fmt.Sprintf("http://localhost:%s", port), logger),
		container:          container,
		RawRpc:             c,
		AdminAcc:           initializeTestAccount("bb39aa88008bc6260ff9ebc816178c47a01c44efe55810ea1f271c00f5878812"),
		FunderAcc:          initializeTestAccount("29c8b4ff78e41dafd561f5cd4a90103faf20a5b509a4b6281947b8fcdcfa8f71"),
		ProjectsManagerAcc: initializeTestAccount("460503be96e3b97c2d6fb737bef83d89df42e4a36adef2e8fb4f0976b70d1b2a"),
		ProjectOwnerAcc:    initializeTestAccount("1516a467486cd4340e5f0e8193eea05c9106bb0dce26a03047580c25c9191f93"),
	}
}

func (tch *TestChain) TearDown() {
	tch.RawRpc.Close()
	_ = tch.container.Terminate(context.Background())
}

// createContainer creates a test container for postgres database
func createContainer(ctx context.Context, logger *logger.AppLogger) (testcontainers.Container, string, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "trufflesuite/ganache-cli:v6.12.2",
			ExposedPorts: []string{"8545/tcp"},
			Env:          nil,
			WaitingFor:   wait.ForListeningPort("8545/tcp"),
			Cmd: []string{
				"--chainId", strconv.Itoa(TestChainId),
				"--gasLimit", strconv.Itoa(TestChainGasLimit),
				"--gasPrice", "6721975",
				"--account", "0xbb39aa88008bc6260ff9ebc816178c47a01c44efe55810ea1f271c00f5878812,100000000000000000000",
				"--account", "0x29c8b4ff78e41dafd561f5cd4a90103faf20a5b509a4b6281947b8fcdcfa8f71,100000000000000000000",
				"--account", "0x460503be96e3b97c2d6fb737bef83d89df42e4a36adef2e8fb4f0976b70d1b2a,100000000000000000000",
				"--account", "0x1516a467486cd4340e5f0e8193eea05c9106bb0dce26a03047580c25c9191f93,100000000000000000000",
				"--account", "0x904d5dea0bdffb09d78a81c15f0b3b893f504679eb8cd1de585309cad58e6285,100000000000000000000",
				"--account", "0x41cc3b5e94e73b9135ff13a62d885f6ddb996a29b70a35eb22cf665494d224c3,100000000000000000000",
				"--account", "0x8aa6aebfb06fe75b9ccd82a9c8f0eaf5b0201c7bff1b67225be881e659149434,100000000000000000000",
				"--account", "0x15364f287dfff78d5cb13c103bc47b27170f2b857c6a301fe644ddbe730ef94c,100000000000000000000",
				"--account", "0x2af7b50b4cb680afb9d312e41038008133ab55339136e355ddc54c361afdb40b,100000000000000000000",
				"--account", "0x118f2ba03f3d9be6e810fd129b009a29deb419cc3614ade7f251e4a42ad79378,100000000000000000000",
			},
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return container, "", fmt.Errorf("failed to start container: %v", err)
	}
	p, err := container.MappedPort(ctx, "8545/tcp")
	if err != nil {
		return container, "", fmt.Errorf("failed to get container external port: %v", err)
	}
	logger.Infof("ganache container ready and running at port: ", p.Port())
	// wait for the chain to be ready
	time.Sleep(time.Second)
	return container, p.Port(), nil
}

// initializeTestAccount initializes a test account
func initializeTestAccount(privateKey string) *testAccount {
	acc := &testAccount{}
	acc.PrivateKey, _ = crypto.HexToECDSA(privateKey)
	acc.Address = crypto.PubkeyToAddress(acc.PrivateKey.PublicKey)
	return acc
}
