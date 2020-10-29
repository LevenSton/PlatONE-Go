package backend

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/consensus"
	"github.com/PlatONEnetwork/PlatONE-Go/core"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"
	"github.com/PlatONEnetwork/PlatONE-Go/life/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/p2p/discover"
)

var (
	ErrContractNotFound = errors.New("contract not found")
)

type ChainContext struct {
	// Engine retrieves the chain's consensus engine.
	chain *consensus.ChainReader

	engine consensus.Engine
}

func (cc *ChainContext) GetHeader(hash common.Hash, number uint64) *types.Header {
	return cc.GetHeader(hash, number)
}

func (cc *ChainContext) Engine() consensus.Engine {
	return cc.engine
}

type CBFTProduceBlockCfg struct {
	ProduceDuration int32 `json:"ProduceDuration"`
	BlockInterval   int32 `json:"BlockInterval"`
}

type commonResult struct {
	RetCode int32      `json:"code"`
	RetMsg  string     `json:"msg"`
	Data    []nodeInfo `json:"data"`
}

type nodeInfo struct {
	Name       string `json:"name,omitempty"`
	Owner      string `json:"owner,omitempty"`
	Desc       string `json:"desc,omitempty"`
	Types      int32  `json:"type,omitempty"`
	Status     int32  `json:"status,omitempty"`
	ExternalIP string `json:"externalIP,omitempty"`
	InternalIP string `json:"internalIP,omitempty"`
	PublicKey  string `json:"publicKey,omitempty"`
	RpcPort    int32  `json:"rpcPort,omitempty"`
	P2pPort    int32  `json:"p2pPort,omitempty"`
}

// getInitialNodesList catch initial nodes List from paramManager contract when
// new a dpos and miner a new block
func getConsensusNodesList(chain consensus.ChainReader, sb *backend, number uint64) ([]discover.NodeID, error) {
	var tmp []common.NodeInfo
	isOldBlock := number < chain.CurrentHeader().Number.Uint64()

	if !isOldBlock {
		tmp = common.SysCfg.GetConsensusNodesFilterDelay(number, []common.NodeInfo{}, isOldBlock)
	} else {
		res := CallSystemContractAtBlockNumber(chain, sb, number, "__sys_NodeManager", CallSystemContractRes)
		nodes := ParseResultToNodeInfos(res)
		tmp = common.SysCfg.GetConsensusNodesFilterDelay(number, nodes, isOldBlock)
	}

	nodeIDs := make([]discover.NodeID, 0, len(tmp))
	for _, dataObj := range tmp {
		if pubKey := dataObj.PublicKey; len(pubKey) > 0 {
			log.Debug("Consensus node", "PublicKey", pubKey)
			if nodeID, err := discover.HexID(pubKey); err == nil {
				nodeIDs = append(nodeIDs, nodeID)
			}
		}
	}
	return nodeIDs, nil
}

func ParseResultToNodeInfos(res []byte) []common.NodeInfo {
	strRes := common.CallResAsString(res)
	var tmp common.CommonResult
	if err := json.Unmarshal(utils.String2bytes(strRes), &tmp); err != nil {
		log.Warn("ParseResultToNodeInfos: unmarshal consensus node list failed", "result", strRes, "err", err.Error())
		return nil
	} else if tmp.RetCode != 0 {
		log.Debug("ParseResultToNodeInfos: contract inner error", "code", tmp.RetCode, "msg", tmp.RetMsg)
		return nil
	} else {
		return tmp.Data
	}
}

func CallSystemContractRes(conAddr common.Address, callContract func(conAddr common.Address, data []byte) []byte) []byte {
	return callContract(conAddr, common.GenCallData("getAllNodes", []interface{}{}))
}

func CallSystemContractAtBlockNumber(
	chain consensus.ChainReader,
	sb *backend,
	number uint64,
	sysContractName string,
	fn func(conAddr common.Address, callContract func(conAddr common.Address, data []byte) []byte) []byte,
) []byte {
	_state, _ := state.New(chain.GetHeaderByNumber(number).Root, state.NewDatabase(sb.db))
	if _state == nil {
		log.Warn("load state fail at block number", "number", number)
		return nil
	}
	msg := types.NewMessage(common.Address{}, nil, 1, big.NewInt(1), 0x1, big.NewInt(1), nil, false, types.NormalTxType)
	cc := ChainContext{&chain, sb}
	context := core.NewEVMContext(msg, chain.CurrentHeader(), &cc, nil)
	evm := vm.NewEVM(context, _state, chain.Config(), vm.Config{})
	callContract := func(conAddr common.Address, data []byte) []byte {
		res, _, err := evm.Call(vm.AccountRef(common.Address{}), conAddr, data, uint64(0xffffffffff), big.NewInt(0))
		if err != nil {
			return nil
		}
		return res
	}

	callParams := []interface{}{sysContractName, "latest"}
	btsRes := callContract(syscontracts.CnsManagementAddress, common.GenCallData("getContractAddress", callParams))
	strRes := common.CallResAsString(btsRes)
	if len(strRes) == 0 || common.IsHexZeroAddress(strRes) {
		log.Warn("call system contract address fail")
		return nil
	}
	contractAddr := common.HexToAddress(strRes)
	return fn(contractAddr, callContract)
}
