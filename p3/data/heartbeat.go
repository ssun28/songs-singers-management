package data

import (
	"../../p1"
)

type HeartBeatData struct {
	IfNewBlock       bool   `json:"ifNewBlock"`
	Id               int32  `json:"id"`
	BlockJson        string `json:"blockJson"`
	PeerMapJson      string `json:"peerMapJson"`
	Addr             string `json:"addr"`
	Hops             int32  `json:"hops"`
	IfNewTransaction bool   `json:"ifNewTransaction"`
	TransactionJson  string `json:"transactionJson"`
	Timestamp        int64  `json:"timeStamp"`
}

func NewHeartBeatData(ifNewBlock bool, id int32, blockJson string, peerMapJson string,
	addr string, ifNewTransaction bool, transactionJson string, timestamp int64) HeartBeatData {
	newHeartBeatData := HeartBeatData{ifNewBlock, id, blockJson,
		peerMapJson, addr, 3, ifNewTransaction, transactionJson, timestamp}
	return newHeartBeatData
}

func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapJson string, addr string) HeartBeatData {
	newBlock := sbc.GenBlock(p1.MerklePatriciaTrie{})
	blockJson, _ := newBlock.EncodeToJSON()

	return NewHeartBeatData(false, selfId, blockJson, peerMapJson, addr, false, "", 0)
}
