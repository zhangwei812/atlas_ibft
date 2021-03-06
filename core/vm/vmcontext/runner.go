package vmcontext

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/mapprotocol/atlas/core/state"
	"github.com/mapprotocol/atlas/core/types"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
	"math/big"
)

// VMAddress is the address the VM uses to make internal calls to contracts
var VMAddress = params.ZeroAddress

// evmRunnerContext defines methods required to create an EVMRunner
type evmRunnerContext interface {
	chainContext

	// GetVMConfig returns the node's vm configuration
	GetVMConfig() *vm.Config

	CurrentHeader() *types.Header

	State() (*state.StateDB, error)
}

func NewEVMRunner(chain evmRunnerContext, header *types.Header, state vm.StateDB) vm.EVMRunner {

	return &evmRunner{
		state: state,
		newEVM: func(from common.Address) *vm.EVM {
			// The EVM Context requires a msg, but the actual field values don't really matter for this case.
			// Putting in zero values for gas price and tx fee recipient
			context := New(from, common.Big0, header, chain, nil)
			return vm.NewEVM(context, vm.TxContext{}, state, chain.Config(), *chain.GetVMConfig())
		},
	}
}

type evmRunner struct {
	newEVM func(from common.Address) *vm.EVM
	state  vm.StateDB

	dontMeterGas bool
}

func (ev *evmRunner) Execute(recipient common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, err error) {
	evm := ev.newEVM(VMAddress)
	if ev.dontMeterGas {
		//evm.StopGasMetering()
	}
	ret, _, err = evm.Call(vm.AccountRef(evm.Origin), recipient, input, gas, value)
	return ret, err
}

func (ev *evmRunner) ExecuteFrom(sender, recipient common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, err error) {
	evm := ev.newEVM(sender)
	if ev.dontMeterGas {
		//evm.StopGasMetering()
	}
	ret, _, err = evm.Call(vm.AccountRef(sender), recipient, input, gas, value)
	return ret, err
}

func (ev *evmRunner) Query(recipient common.Address, input []byte, gas uint64) (ret []byte, err error) {
	evm := ev.newEVM(VMAddress)
	if ev.dontMeterGas {
		evm.StopGasMetering()
	}
	ret, _, err = evm.StaticCall(vm.AccountRef(evm.Origin), recipient, input, gas)
	return ret, err
}

func (ev *evmRunner) StopGasMetering() {
	ev.dontMeterGas = true
}

func (ev *evmRunner) StartGasMetering() {
	ev.dontMeterGas = false
}

// GetStateDB implements Backend.GetStateDB
func (ev *evmRunner) GetStateDB() vm.StateDB {
	return ev.state
}
