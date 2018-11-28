package parity

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tenderly/tenderly-trace/ethereum"
	"github.com/tenderly/tenderly-trace/ethereum/core/vm"
	"math/big"
)

// Core Types

type Header struct {
	HNumber *ethereum.Number `json:"number"`
}

func (h *Header) Number() *ethereum.Number {
	return h.HNumber
}

type Block struct {
	ValuesTransactions []*Transaction `json:"transactions"`
}

func (b *Block) Transactions() []ethereum.Transaction {
	if b.ValuesTransactions == nil {
		return []ethereum.Transaction{}
	}

	traces := make([]ethereum.Transaction, len(b.ValuesTransactions))
	for k, v := range b.ValuesTransactions {
		traces[k] = v
	}

	return traces
}

type BlockHeader struct {
	ValueNumber     *ethereum.Number `json:"number"`
	ValueBlockHash  *common.Hash     `json:"hash"`
	ValueParentHash *common.Hash     `json:"parentHash"`
	ValueTime       *hexutil.Big     `json:"timestamp"`
	ValueDifficulty *hexutil.Big     `json:"difficulty"`
	ValueGasLimit   *hexutil.Big     `json:"gasLimit"`
	ValueGasPrice   *hexutil.Big     `json:"gasPrice"`
	ValueCoinbase   *common.Address  `json:"miner"`
}

func (b *BlockHeader) Number() *ethereum.Number {
	return b.ValueNumber
}

func (b *BlockHeader) Hash() *common.Hash {
	return b.ValueBlockHash
}

func (b *BlockHeader) ParentHash() *common.Hash {
	return b.ValueParentHash
}

func (b *BlockHeader) Time() *hexutil.Big {
	return b.ValueTime
}

func (b *BlockHeader) Difficulty() *hexutil.Big {
	return b.ValueDifficulty
}

func (b *BlockHeader) GasLimit() *hexutil.Big {
	return b.ValueGasLimit
}

func (b *BlockHeader) GasPrice() *hexutil.Big {
	return b.ValueGasPrice
}

func (b *BlockHeader) Coinbase() *common.Address {
	return b.ValueCoinbase
}

type Transaction struct {
	ValueHash        *common.Hash    `json:"hash"`
	ValueFrom        *common.Address `json:"from"`
	ValueTo          *common.Address `json:"to"`
	ValueInput       hexutil.Bytes   `json:"input"`
	ValueValue       *hexutil.Big    `json:"value"`
	ValueGas         *hexutil.Big    `json:"gas"`
	ValueGasPrice    *hexutil.Big    `json:"gasPrice"`
	ValueBlockNumber *hexutil.Big    `json:"blockNumber"`
	ValueBlockHash   *common.Hash    `json:"blockHash"`
}

func (t *Transaction) Hash() *common.Hash {
	return t.ValueHash
}

func (t *Transaction) From() *common.Address {
	return t.ValueFrom
}

func (t *Transaction) To() *common.Address {
	return t.ValueTo
}

func (t *Transaction) Input() hexutil.Bytes {
	return t.ValueInput
}

func (t *Transaction) Value() *hexutil.Big {
	return t.ValueValue
}

func (t *Transaction) Gas() *hexutil.Big {
	return t.ValueGas
}

func (t *Transaction) GasPrice() *hexutil.Big {
	return t.ValueGasPrice
}

func (t *Transaction) BlockNumber() *big.Int {
	return t.ValueBlockNumber.ToInt()
}

func (t *Transaction) BlockHash() *common.Hash {
	return t.ValueBlockHash
}

type Log struct {
	ValueAddress             string   `json:"address"`
	ValueBlockHash           string   `json:"blockHash"`
	ValueBlockNumber         string   `json:"blockNumber"`
	ValueData                string   `json:"data"`
	ValueLogIndex            string   `json:"logIndex"`
	ValueRemoved             bool     `json:"removed"`
	ValueTopics              []string `json:"topics"`
	ValueTransactionHash     string   `json:"transactionHash"`
	ValueTransactionIndex    string   `json:"transactionIndex"`
	ValueTransactionLogIndex string   `json:"transactionLogIndex"`
	ValueType                string   `json:"type"`
}

func (l *Log) Data() string {
	return l.ValueData
}

func (l *Log) Topics() []string {
	return l.ValueTopics
}

type TransactionReceipt struct {
	TTransactionHash  string `json:"transactionHash"`
	TTransactionIndex string `json:"transactionIndex"`
	TBlockHash        string `json:"blockHash"`
	TBlockNumber      string `json:"blockNumber"`

	TGasUsed           *hexutil.Big    `json:"gasUsed"`
	TCumulativeGasUsed *hexutil.Big    `json:"cumulativeGasUsed"`
	TContractAddress   *common.Address `json:"contractAddress"`

	TStatus    string  `json:"status"` // Can be null, if null do a check anyways. 0x0 fail, 0x1 success
	TLogs      []*Log  `json:"logs"`
	TLogsBloom string  `json:"logsBloom"`
	TRoot      *string `json:"root"`
}

func (t *TransactionReceipt) SetStatus(trace string) {
	t.TStatus = "0x0 " + trace
}

func (t *TransactionReceipt) Hash() string {
	return t.TTransactionHash
}

func (t *TransactionReceipt) GasUsed() *hexutil.Big {
	return t.TGasUsed
}

func (t *TransactionReceipt) CumulativeGasUsed() *hexutil.Big {
	return t.TCumulativeGasUsed
}

func (t *TransactionReceipt) ContractAddress() *common.Address {
	return t.TContractAddress
}

func (t *TransactionReceipt) Status() string {
	return t.TStatus
}

func (t *TransactionReceipt) Logs() []ethereum.Log {
	var logs []ethereum.Log

	for _, log := range t.TLogs {
		logs = append(logs, log)
	}

	return logs
}

type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

type VersionInfo struct {
	Hash    string  `json:"hash"`
	Track   string  `json:"track"`
	Version Version `json:"version"`
}

// States Types

type Mem struct {
	Data hexutil.Bytes `json:"data"`
	Off  uint64        `json:"off"`
}

type Ex struct {
	Mem  Mem      `json:"mem"`
	Push []string `json:"push"`
	Used uint64   `json:"used"`
}

type VmState struct {
	ValuePc      uint64             `json:"pc"`
	ValueOp      string             `json:"op"`
	ValueEx      Ex                 `json:"ex"`
	ValueSub     *VmTrace           `json:"sub"`
	ValueGas     uint64             `json:"gas"`
	ValueGasCost int64              `json:"cost"`
	ValueDepth   int                `json:"depth"`
	ValueError   json.RawMessage    `json:"error,omitempty"`
	ValueStack   *vm.Stack          `json:"stack,omitempty"`
	ValueMemory  *vm.Memory         `json:"memory,omitempty"`
	ValueStorage *map[string]string `json:"storage,omitempty"`
	Terminating  bool
}

func (pvs *VmState) Pc() uint64 {
	return pvs.ValuePc
}

func (pvs *VmState) Depth() int {
	return pvs.ValueDepth + 1
}

func (pvs *VmState) Op() string {
	return "Not implemented"
}

func (pvs *VmState) Stack() *vm.Stack {
	return pvs.ValueStack
}

func (pvs *VmState) Memory() *vm.Memory {
	return pvs.ValueMemory
}

type TraceResult struct {
	VmTrace   *VmTrace `json:"vmTrace"`
	CallTrace []*Trace `json:"traceSchema"`
}

type VmTrace struct {
	Logs []*VmState    `json:"ops"`
	Code hexutil.Bytes `json:"code"`
}

func (tr *TraceResult) States() []ethereum.EvmState {
	if tr.VmTrace == nil {
		return []ethereum.EvmState{}
	}

	traces := make([]ethereum.EvmState, len(tr.VmTrace.Logs))
	for k, v := range tr.VmTrace.Logs {
		traces[k] = v
	}

	return traces
}

func (tr *TraceResult) Traces() []ethereum.Trace {
	if tr.VmTrace == nil {
		return []ethereum.Trace{}
	}

	traces := make([]ethereum.Trace, len(tr.CallTrace))
	for k, v := range tr.CallTrace {
		traces[k] = v
	}

	return traces
}

func (tr *TraceResult) ProcessTrace(tx ethereum.Transaction) {
	if tr.VmTrace == nil {
		return
	}

	tr.VmTrace.Logs = Walk(tr.VmTrace, tx)
}

func Walk(vmt *VmTrace, tx ethereum.Transaction) []*VmState {
	var traces []*VmState

	vmt.Logs[0].ValueOp = ethereum.OpCode(vmt.Code[vmt.Logs[0].ValuePc]).String()
	for i := 0; i < len(vmt.Logs); i++ {
		if i > 0 {
			vmt.Logs[i].ValueStack = vmt.Logs[i-1].ValueStack

			if vmt.Logs[i-1].ValueOp == "CALL" {
				vmt.Logs[i].ValueStack = nil
			}
		}

		if i < len(vmt.Logs)-1 {
			opCode := ethereum.OpCode(vmt.Code[vmt.Logs[i+1].ValuePc])
			vmt.Logs[i+1].ValueOp = opCode.String()

			//if vmt.Logs[i+1].ValueOp == "EXTCODESIZE" {
			//	vmt.Logs[i].ValueStack = &[]string{}
			//	for j := 0; j < len(vmt.Logs[i].ValueEx.Push); j++ {
			//		vmt.Logs[i].ValueEx.Push[j] = "000000000000000000000000" + vmt.Logs[i].ValueEx.Push[j][2:]
			//		for len(vmt.Logs[i].ValueEx.Push[j]) < 64 {
			//			vmt.Logs[i].ValueEx.Push[j] = "0" + vmt.Logs[i].ValueEx.Push[j]
			//		}
			//	}
			//
			//	*vmt.Logs[i].ValueStack = append(*vmt.Logs[i].ValueStack, vmt.Logs[i].ValueEx.Push...)
			//}
		}

		traces = append(traces, vmt.Logs[i])
		if vmt.Logs[i].ValueSub != nil {
			subTraces := Walk(vmt.Logs[i].ValueSub, tx)
			subTraces[len(subTraces)-1].Terminating = true

			traces = append(traces, subTraces...)
		}
	}

	traces[len(traces)-1].Terminating = true

	return traces
}

type Action struct {
	CallType        string          `json:"callType"`
	Hash            *common.Hash    `json:"hash"`
	ParentHash      *common.Hash    `json:"hash"`
	TransactionHash *common.Hash    `json:"hash"`
	From            common.Address  `json:"from"`
	To              common.Address  `json:"to"`
	Input           hexutil.Bytes   `json:"input"`
	Gas             *hexutil.Uint64 `json:"gas,omitempty"`
	Value           *hexutil.Big    `json:"value,omitempty"`
}

type Result struct {
	GasUsed *hexutil.Uint64 `json:"gasUsed,omitempty"`
	Output  hexutil.Bytes   `json:"output"`
}

type Trace struct {
	ValueAction       Action `json:"action"`
	ValueResult       Result `json:"result"`
	ValueLogs         []Log  `json:"logs"`
	ValueSubtraces    int    `json:"subtraces"`
	ValueError        string `json:"error"`
	ValueTraceAddress []int  `json:"traceAddress"`
	ValueType         string `json:"type"`
}

func (t *Trace) Hash() *common.Hash {
	return t.ValueAction.Hash
}

func (t *Trace) ParentHash() *common.Hash {
	return t.ValueAction.ParentHash
}

func (t *Trace) TransactionHash() *common.Hash {
	return t.ValueAction.TransactionHash
}

func (t *Trace) Type() string {
	return t.ValueType
}

func (t *Trace) From() common.Address {
	return t.ValueAction.From
}

func (t *Trace) To() common.Address {
	return t.ValueAction.To
}

func (t *Trace) Input() hexutil.Bytes {
	return t.ValueAction.Input
}

func (t *Trace) Output() hexutil.Bytes {
	return t.ValueResult.Output
}

func (t *Trace) Gas() *hexutil.Uint64 {
	return t.ValueAction.Gas
}

func (t *Trace) GasUsed() *hexutil.Uint64 {
	return t.ValueResult.GasUsed
}

func (t *Trace) Value() *hexutil.Big {
	return t.ValueAction.Value
}

func (t *Trace) Error() string {
	return t.ValueError
}
