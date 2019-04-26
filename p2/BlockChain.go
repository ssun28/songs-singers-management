package p2

import (
	mpt "../p1"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"log"
	"sort"
)

type Block struct {
	Header Header
	Value  mpt.MerklePatriciaTrie
}

type Header struct {
	Height     int32
	Timestamp  int64
	Hash       string
	ParentHash string
	Size       int32
	Nonce      string
}

type BlockJson struct {
	Height     int32             `json:"height"`
	Timestamp  int64             `json:"timeStamp"`
	Hash       string            `json:"hash"`
	ParentHash string            `json:"parentHash"`
	Size       int32             `json:"size"`
	Nonce      string            `json:"nonce"`
	MPT        map[string]string `json:"mpt"`
}

type BlockChain struct {
	Chain  map[int32][]Block
	Length int32
}

//Initial: This function takes arguments(such as height, parentHash, and value of MPT type) and forms a block.
func NewBlock(height int32, timestamp int64, parentHash string, nonce string, mpt mpt.MerklePatriciaTrie) Block {
	size := getByteArraySize(mpt)
	hashStr := string(height) + string(timestamp) + parentHash + mpt.Root + string(size)
	hashResult := hashToString(hashStr)
	header := Header{height, timestamp, hashResult, parentHash, size, nonce}
	newBlock := Block{header, mpt}

	return newBlock
}

func hashToString(hashStr string) string {
	sum := sha3.Sum256([]byte(hashStr))
	return hex.EncodeToString(sum[:])
}

func getByteArraySize(mpt mpt.MerklePatriciaTrie) int32 {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(mpt)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	size := int32(len(buf.Bytes()))

	return size
}

func (block *Block) EncodeToJSON() (string, error) {
	jsonBlock := BlockJson{block.Header.Height, block.Header.Timestamp, block.Header.Hash,
		block.Header.ParentHash, block.Header.Size, block.Header.Nonce, block.Value.Kv}
	jsonBlockBytes, err := json.Marshal(jsonBlock)
	if err != nil {
		log.Fatal("encode to jsonBlock error:", err)
	}

	return string(jsonBlockBytes), err
}

func DecodeFromJson(jsonBlock string) (Block, error) {
	var blockJson BlockJson
	err := json.Unmarshal([]byte(jsonBlock), &blockJson)
	if err != nil {
		log.Fatal("decode to block error:", err)
	}
	newMpt := mpt.MerklePatriciaTrie{}
	newMpt.Initial()
	for k, v := range blockJson.MPT {
		newMpt.Insert(k, v)
	}

	newHeader := Header{blockJson.Height, blockJson.Timestamp, blockJson.Hash,
		blockJson.ParentHash, blockJson.Size, blockJson.Nonce}
	newBlock := Block{newHeader, newMpt}

	return newBlock, err
}

func NewBlockChain() BlockChain {
	newBlockchain := BlockChain{map[int32][]Block{}, 0}
	return newBlockchain
}

func (blockchain *BlockChain) Get(height int32) []Block {
	if blockchain.Chain[height] == nil {
		return nil
	}
	blockArray := blockchain.Chain[height]

	return blockArray
}

func (blockchain *BlockChain) GetLatestBlocks() []Block {
	length := blockchain.Length
	if blockchain.Chain[length] == nil {
		return nil
	}
	latestBlockArray := blockchain.Chain[length]

	return latestBlockArray
}

func (blockchain *BlockChain) GetParentBlock(block Block) Block {
	blockHeight := block.Header.Height

	if blockHeight == 0 {
		return Block{}
	}

	if blockchain.Get(blockHeight-1) == nil {
		return Block{}
	}

	blockArray := blockchain.Get(blockHeight - 1)
	for _, element := range blockArray {
		if element.Header.Hash == block.Header.ParentHash {
			return element
		}
	}

	return Block{}
}

func (blockchain *BlockChain) Insert(block Block) {
	blockArray := blockchain.Chain[block.Header.Height]
	for _, element := range blockArray {
		if element.Header.Hash == block.Header.Hash {
			return
		}
	}

	blockArray = append(blockArray, block)
	blockchain.Chain[block.Header.Height] = blockArray

	if block.Header.Height > blockchain.Length {
		blockchain.Length = block.Header.Height
	}
}

func (blockchain *BlockChain) EncodeToJSON() (string, error) {
	var jsonBlockArray []BlockJson
	for _, v := range blockchain.Chain {
		for _, block := range v {
			jsonBlock := BlockJson{block.Header.Height, block.Header.Timestamp, block.Header.Hash,
				block.Header.ParentHash, block.Header.Size, block.Header.Nonce, block.Value.Kv}
			jsonBlockArray = append(jsonBlockArray, jsonBlock)
		}
	}

	jsonBlockchainBytes, err := json.Marshal(jsonBlockArray)
	if err != nil {
		log.Fatal("encode to jsonBlockchain:", err)
	}

	return string(jsonBlockchainBytes), err
}

func DecodeJsonToBlockChain(jsonBlockChain string) (BlockChain, error) {
	jsonBlockArray := make([]BlockJson, 0)
	err := json.Unmarshal([]byte(jsonBlockChain), &jsonBlockArray)

	if err != nil {
		log.Fatal("decode jsonBlockchain error:", err)
	}

	newBlockchain := NewBlockChain()
	for _, element := range jsonBlockArray {

		newMpt := mpt.MerklePatriciaTrie{}
		newMpt.Initial()
		for k, v := range element.MPT {
			newMpt.Insert(k, v)
		}

		newHeader := Header{element.Height, element.Timestamp, element.Hash,
			element.ParentHash, element.Size, element.Nonce}
		newBlock := Block{newHeader, newMpt}
		newBlockchain.Insert(newBlock)
	}

	return newBlockchain, err
}

func (bc *BlockChain) Show() string {
	rs := ""
	var idList []int
	for id := range bc.Chain {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)
	for _, id := range idList {
		var hashs []string
		for _, block := range bc.Chain[int32(id)] {
			hashs = append(hashs, block.Header.Hash+"<="+block.Header.ParentHash)
		}
		sort.Strings(hashs)
		rs += fmt.Sprintf("%v: ", id)
		for _, h := range hashs {
			rs += fmt.Sprintf("%s, ", h)
		}
		rs += "\n"
	}
	sum := sha3.Sum256([]byte(rs))
	rs = fmt.Sprintf("This is the BlockChain: %s\n", hex.EncodeToString(sum[:])) + rs
	return rs
}

func (bc *BlockChain) Canonical() string {
	rs := ""

	for h, ele := range bc.Chain {
		rs += fmt.Sprintf("height:%v, has %v block\n", h, len(ele))
	}
	rs += fmt.Sprintf("bc.length:%v\n", len(bc.Chain))
	rs += "\n"

	for _, block := range bc.Chain[bc.Length] {
		//currentHeight := bc.Length
		//rs += fmt.Sprintf("height=%v, timestamp=%d, hash=%s, parentHash=%s, size=%v\n",
		//	block.Header.Height, block.Header.Timestamp, block.Header.Hash, block.Header.ParentHash, block.Header.Size)

		rs += "Chain #1:\n"
		currentHeight := bc.Length
		rs += fmt.Sprintf("current height:%v\n", currentHeight)

		parentHash := block.Header.ParentHash
		rs += fmt.Sprintf("height=%v, timestamp=%d, hash=%s, parentHash=%s, size=%v\n",
			block.Header.Height, block.Header.Timestamp, block.Header.Hash, block.Header.ParentHash, block.Header.Size)
		currentHeight--
		for currentHeight > 0 {
			for _, b := range bc.Chain[currentHeight] {
				if b.Header.Hash == parentHash {
					rs += fmt.Sprintf("height=%v, timestamp=%d, hash=%s, parentHash=%s, size=%v\n",
						b.Header.Height, b.Header.Timestamp, b.Header.Hash, b.Header.ParentHash, b.Header.Size)
					parentHash = b.Header.ParentHash
					currentHeight--
					break
				}
			}
		}

		rs += "\n"
	}

	return rs
}
