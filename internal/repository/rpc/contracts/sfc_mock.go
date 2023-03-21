// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// SfcMockMetaData contains all meta data concerning the SfcMock contract.
var SfcMockMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"currentEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"}],\"name\":\"setEpoch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040526000805534801561001457600080fd5b5060ac806100236000396000f3fe6080604052348015600f57600080fd5b506004361060325760003560e01c80630ceb2cef14603757806376671808146049575b600080fd5b60476042366004605e565b600055565b005b60005460405190815260200160405180910390f35b600060208284031215606f57600080fd5b503591905056fea2646970667358221220b4fed90ece33e12aaf5821edb9fe025069b633787d8220e02dbcc2b835c8362a64736f6c63430008130033",
}

// SfcMockABI is the input ABI used to generate the binding from.
// Deprecated: Use SfcMockMetaData.ABI instead.
var SfcMockABI = SfcMockMetaData.ABI

// SfcMockBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SfcMockMetaData.Bin instead.
var SfcMockBin = SfcMockMetaData.Bin

// DeploySfcMock deploys a new Ethereum contract, binding an instance of SfcMock to it.
func DeploySfcMock(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SfcMock, error) {
	parsed, err := SfcMockMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SfcMockBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SfcMock{SfcMockCaller: SfcMockCaller{contract: contract}, SfcMockTransactor: SfcMockTransactor{contract: contract}, SfcMockFilterer: SfcMockFilterer{contract: contract}}, nil
}

// SfcMock is an auto generated Go binding around an Ethereum contract.
type SfcMock struct {
	SfcMockCaller     // Read-only binding to the contract
	SfcMockTransactor // Write-only binding to the contract
	SfcMockFilterer   // Log filterer for contract events
}

// SfcMockCaller is an auto generated read-only Go binding around an Ethereum contract.
type SfcMockCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SfcMockTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SfcMockTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SfcMockFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SfcMockFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SfcMockSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SfcMockSession struct {
	Contract     *SfcMock          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SfcMockCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SfcMockCallerSession struct {
	Contract *SfcMockCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// SfcMockTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SfcMockTransactorSession struct {
	Contract     *SfcMockTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// SfcMockRaw is an auto generated low-level Go binding around an Ethereum contract.
type SfcMockRaw struct {
	Contract *SfcMock // Generic contract binding to access the raw methods on
}

// SfcMockCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SfcMockCallerRaw struct {
	Contract *SfcMockCaller // Generic read-only contract binding to access the raw methods on
}

// SfcMockTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SfcMockTransactorRaw struct {
	Contract *SfcMockTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSfcMock creates a new instance of SfcMock, bound to a specific deployed contract.
func NewSfcMock(address common.Address, backend bind.ContractBackend) (*SfcMock, error) {
	contract, err := bindSfcMock(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SfcMock{SfcMockCaller: SfcMockCaller{contract: contract}, SfcMockTransactor: SfcMockTransactor{contract: contract}, SfcMockFilterer: SfcMockFilterer{contract: contract}}, nil
}

// NewSfcMockCaller creates a new read-only instance of SfcMock, bound to a specific deployed contract.
func NewSfcMockCaller(address common.Address, caller bind.ContractCaller) (*SfcMockCaller, error) {
	contract, err := bindSfcMock(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SfcMockCaller{contract: contract}, nil
}

// NewSfcMockTransactor creates a new write-only instance of SfcMock, bound to a specific deployed contract.
func NewSfcMockTransactor(address common.Address, transactor bind.ContractTransactor) (*SfcMockTransactor, error) {
	contract, err := bindSfcMock(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SfcMockTransactor{contract: contract}, nil
}

// NewSfcMockFilterer creates a new log filterer instance of SfcMock, bound to a specific deployed contract.
func NewSfcMockFilterer(address common.Address, filterer bind.ContractFilterer) (*SfcMockFilterer, error) {
	contract, err := bindSfcMock(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SfcMockFilterer{contract: contract}, nil
}

// bindSfcMock binds a generic wrapper to an already deployed contract.
func bindSfcMock(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SfcMockMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SfcMock *SfcMockRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SfcMock.Contract.SfcMockCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SfcMock *SfcMockRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SfcMock.Contract.SfcMockTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SfcMock *SfcMockRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SfcMock.Contract.SfcMockTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SfcMock *SfcMockCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SfcMock.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SfcMock *SfcMockTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SfcMock.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SfcMock *SfcMockTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SfcMock.Contract.contract.Transact(opts, method, params...)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256)
func (_SfcMock *SfcMockCaller) CurrentEpoch(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SfcMock.contract.Call(opts, &out, "currentEpoch")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256)
func (_SfcMock *SfcMockSession) CurrentEpoch() (*big.Int, error) {
	return _SfcMock.Contract.CurrentEpoch(&_SfcMock.CallOpts)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256)
func (_SfcMock *SfcMockCallerSession) CurrentEpoch() (*big.Int, error) {
	return _SfcMock.Contract.CurrentEpoch(&_SfcMock.CallOpts)
}

// SetEpoch is a paid mutator transaction binding the contract method 0x0ceb2cef.
//
// Solidity: function setEpoch(uint256 epoch) returns()
func (_SfcMock *SfcMockTransactor) SetEpoch(opts *bind.TransactOpts, epoch *big.Int) (*types.Transaction, error) {
	return _SfcMock.contract.Transact(opts, "setEpoch", epoch)
}

// SetEpoch is a paid mutator transaction binding the contract method 0x0ceb2cef.
//
// Solidity: function setEpoch(uint256 epoch) returns()
func (_SfcMock *SfcMockSession) SetEpoch(epoch *big.Int) (*types.Transaction, error) {
	return _SfcMock.Contract.SetEpoch(&_SfcMock.TransactOpts, epoch)
}

// SetEpoch is a paid mutator transaction binding the contract method 0x0ceb2cef.
//
// Solidity: function setEpoch(uint256 epoch) returns()
func (_SfcMock *SfcMockTransactorSession) SetEpoch(epoch *big.Int) (*types.Transaction, error) {
	return _SfcMock.Contract.SetEpoch(&_SfcMock.TransactOpts, epoch)
}
