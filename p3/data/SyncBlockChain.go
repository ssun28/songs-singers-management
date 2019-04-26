package data

import (
	"../../p1"
	"../../p2"
	"math/rand"
	"sync"
	"time"
)

type SyncBlockChain struct {
	bc  p2.BlockChain
	mux sync.Mutex
}

func NewBlockChain() SyncBlockChain {
	return SyncBlockChain{bc: p2.NewBlockChain()}
}
func FirstBlockChain() SyncBlockChain {
	mpt := p1.MerklePatriciaTrie{}
	mpt.Initial()
	mpt.Insert("first", "blockChain")
	b1 := p2.NewBlock(1, 1234567890, "genesis", "1f7b169c846f218a", mpt)
	firstBc := p2.NewBlockChain()
	firstBc.Insert(b1)

	return SyncBlockChain{bc: firstBc}
}

func (sbc *SyncBlockChain) Get(height int32) ([]p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	if sbc.bc.Get(height) == nil {
		return nil, false
	}

	return sbc.bc.Get(height), true
}

func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	block := p2.Block{}
	if sbc.bc.Get(height) == nil {
		return block, false
	}
	blockArray := sbc.bc.Get(height)
	for _, element := range blockArray {
		if element.Header.Hash == hash {
			block = element
			return block, true
		}
	}
	return block, false
}

func (sbc *SyncBlockChain) GetLatestBlocks() []p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	return sbc.bc.GetLatestBlocks()
}

func (sbc *SyncBlockChain) GetParentBlock(block p2.Block) p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()

	return sbc.bc.GetParentBlock(block)
}

func (sbc *SyncBlockChain) Insert(block p2.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(block)
	sbc.mux.Unlock()
}

func (sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	blockHeight := insertBlock.Header.Height

	if blockHeight == 0 {
		return true
	}

	if sbc.bc.Get(blockHeight-1) == nil {
		return false
	}

	blockArray := sbc.bc.Get(blockHeight - 1)
	for _, element := range blockArray {
		if element.Header.Hash == insertBlock.Header.ParentHash {
			return true
		}
	}

	return false
}

func (sbc *SyncBlockChain) CheckCurrentBlock(height int32, hash string) bool {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	if height == 0 {
		return true
	}

	blockArray := sbc.bc.Get(height)
	for _, element := range blockArray {
		if element.Header.Hash == hash {
			return true
		}
	}

	return false
}

func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJson string) {
	sbc.mux.Lock()
	sbc.bc, _ = p2.DecodeJsonToBlockChain(blockChainJson)
	sbc.mux.Unlock()
}

func (sbc *SyncBlockChain) BlockChainToJson() (string, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.EncodeToJSON()
}

func (sbc *SyncBlockChain) GenBlock(mpt p1.MerklePatriciaTrie) p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	currentHeight := sbc.bc.Length
	blockArray := sbc.bc.Get(currentHeight)

	if blockArray == nil {
		return p2.Block{}
	}
	blockArrayLength := len(blockArray)
	randIndex := rand.Intn(blockArrayLength)

	parentHash := blockArray[randIndex].Header.Hash
	newTimeStamp := int64(time.Now().Unix())
	newBlock := p2.NewBlock(currentHeight+1, newTimeStamp, parentHash, "", mpt)

	return newBlock
}

func (sbc *SyncBlockChain) Show() string {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Show()
}

func (sbc *SyncBlockChain) Canonical() string {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Canonical()
}
