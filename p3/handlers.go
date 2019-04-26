package p3

import (
	"../p1"
	"../p2"
	"./data"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var TA_SERVER = "http://localhost:6688"
var TA_SERVER_ID_STR = "6688"
var REGISTER_SERVER = TA_SERVER + "/peer"
var BC_DOWNLOAD_SERVER = TA_SERVER + "/upload"
var PRE_SELF_ADDR = "http://localhost:"
var SELF_ADDR = ""
var ID_STR = ""

//var SELF_ADDR = "http://localhost:6686"

var SBC data.SyncBlockChain
var Peers data.PeerList
var ifStarted bool
var ifTryNonce bool

// This function will be executed before everything else.
// Do some initialization here.
func init() {
	if len(os.Args) > 1 {
		ID_STR = os.Args[1]
		SELF_ADDR = PRE_SELF_ADDR + ID_STR
		SBC = data.NewBlockChain()
	} else {
		ID_STR = TA_SERVER_ID_STR
		SELF_ADDR = TA_SERVER
		SBC = data.FirstBlockChain()
	}

	id, _ := strconv.Atoi(ID_STR)
	Peers = data.NewPeerList(int32(id), 32)
	Peers.Register(int32(id))
}

// Register ID, download BlockChain, start HeartBeat
func Start(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "node start")
	ifStarted = true
	ifTryNonce = true
	if SELF_ADDR != TA_SERVER {
		Download()
	}

	go StartHeartBeat()

	go StartTryNonces()
}

// Display peerList and sbc
func Show(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

// Register to TA's server, get an ID
func Register() {}

// Download blockchain from TA server
func Download() {
	resp, err := http.Get(BC_DOWNLOAD_SERVER + "?id=" + ID_STR)
	if err != nil {
		log.Fatal("Download resp error:", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Read body error:", err)
	}

	blockChainJsonStr := string(body)
	SBC.UpdateEntireBlockChain(blockChainJsonStr)
}

// Upload blockchain to whoever called this method, return jsonStr
func Upload(w http.ResponseWriter, r *http.Request) {
	senderIdStr := r.URL.Query()["id"][0]
	senderId, _ := strconv.Atoi(senderIdStr)
	senderAddr := PRE_SELF_ADDR + senderIdStr

	Peers.Add(senderAddr, int32(senderId))

	blockChainJson, err := SBC.BlockChainToJson()
	if err != nil {
		data.PrintError(err, "Upload")
	}
	fmt.Fprint(w, blockChainJson)
}

// Upload a block to whoever called this method, return jsonStr
func UploadBlock(w http.ResponseWriter, r *http.Request) {
	u, _ := url.Parse(r.URL.Path)
	//fmt.Println("path in uploadBlock:", u.Path)
	urlPath := strings.Split(u.Path, "/")

	height, _ := strconv.Atoi(urlPath[2])
	hash := urlPath[3]
	//fmt.Println("path with height:", height)
	//fmt.Println("path with hash:", hash)

	block, haveBlockFlag := SBC.GetBlock(int32(height), hash)

	if haveBlockFlag {
		blockJson, err := block.EncodeToJSON()
		if err != nil {
			fmt.Fprintf(w, string(http.StatusInternalServerError))
		}
		fmt.Fprintf(w, blockJson)
	} else {
		fmt.Fprintf(w, string(http.StatusNoContent))
	}
}

// Received a heartbeat
func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {
	var heartBeatData data.HeartBeatData
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal("heartBeatReceive r.Body error:", err)
	}

	error := json.Unmarshal([]byte(body), &heartBeatData)

	if error != nil {
		log.Fatal("decode jsonPeerMap error:", err)
	}

	if heartBeatData.Addr != SELF_ADDR {

		Peers.Add(heartBeatData.Addr, heartBeatData.Id)

		Peers.InjectPeerMapJson(heartBeatData.PeerMapJson, SELF_ADDR)

		if heartBeatData.IfNewBlock {
			block, _ := p2.DecodeFromJson(heartBeatData.BlockJson)

			hashStr := block.Header.ParentHash + block.Header.Nonce + block.Value.Root
			sha3StrResult := hashToString(hashStr)

			if verifyBlock(sha3StrResult) {
				//fmt.Println("verify heartbeat sha3result success:", sha3StrResult)
				//fmt.Println("the block i need to insert 's height", block.Header.Height)
				//size, _ := SBC.Get(block.Header.Height)
				//fmt.Println("the blockarray size before insert", len(size))
				ifTryNonce = false
				if !SBC.CheckParentHash(block) {
					AskForBlock(block.Header.Height-1, block.Header.ParentHash)
				} else {
					SBC.Insert(block)
				}

				heartBeatData.Hops--
				if heartBeatData.Hops > 0 {
					ForwardHeartBeat(heartBeatData)
				}

				ifTryNonce = true
			}
		}

		fmt.Fprintf(w, "heartBeatReceive sucess")
	}
}

// Ask another server to return a block of certain height and hash
func AskForBlock(height int32, hash string) {
	if SBC.CheckCurrentBlock(height, hash) {
		return
	}

	peerMapCopy := Peers.Copy()
	for addr := range peerMapCopy {
		resp, err := http.Get(addr + "/block/" + string(height) + "/" + hash)
		if err != nil {
			log.Fatal("Ask for block resp error:", err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Ask for block resp.Body error:", err)
		}

		jsonBlock := string(body)
		if jsonBlock != string(http.StatusNoContent) && jsonBlock != string(http.StatusInternalServerError) {
			block, _ := p2.DecodeFromJson(jsonBlock)
			SBC.Insert(block)
			AskForBlock(block.Header.Height-1, block.Header.ParentHash)
			return
		}
	}
}

func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	heartBeatDataJson, err := json.Marshal(heartBeatData)
	if err != nil {
		log.Fatal("encode heartBeatData error:", err)
	}

	if len(Peers.Copy()) > 32 {
		Peers.Rebalance()
	}

	peerMapCopy := Peers.Copy()
	for addr := range peerMapCopy {
		resp, err := http.Post(addr+"/heartbeat/receive", "application/json; charset=UTF-8",
			strings.NewReader(string(heartBeatDataJson)))

		if err != nil {
			Peers.Delete(addr)
			log.Fatal("forward heartBeat post error:", err)
		}
		resp.Body.Close()
	}
}

func StartHeartBeat() {
	for {
		randNum := 5 + rand.Intn(6)
		time.Sleep(time.Duration(randNum) * time.Second)
		peerMapJsonStr, _ := Peers.PeerMapToJson()

		prepareHeartBeatData := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapJsonStr, SELF_ADDR)
		if prepareHeartBeatData.IfNewBlock {
			block, _ := p2.DecodeFromJson(prepareHeartBeatData.BlockJson)
			SBC.Insert(block)
		}
		ForwardHeartBeat(prepareHeartBeatData)
	}
}

func StartTryNonces() {
	for {
		randNum := rand.Intn(10000)
		mpt := p1.MerklePatriciaTrie{}
		mpt.Initial()
		key := strconv.FormatInt(int64(randNum), 10)
		mpt.Insert(key, "")
		newBlock := SBC.GenBlock(mpt)

		for ifTryNonce {
			randNum := int64(time.Now().UnixNano())
			nonce := strconv.FormatInt(randNum, 16)
			hashStr := newBlock.Header.ParentHash + nonce + newBlock.Value.Root
			sha3StrResult := hashToString(hashStr)

			if verifyBlock(sha3StrResult) {
				//fmt.Println("I got the answer********:", sha3StrResult)
				//fmt.Println("height for the next block is :", newBlock.Header.Height)
				//size, _ := SBC.Get(newBlock.Header.Height)
				//fmt.Println("the blockarray size before insert", len(size))

				newBlock.Header.Nonce = nonce
				blockJson, _ := newBlock.EncodeToJSON()
				peerMapJsonStr, _ := Peers.PeerMapToJson()
				heartBeatData := data.NewHeartBeatData(true, Peers.GetSelfId(), blockJson, peerMapJsonStr, SELF_ADDR)
				SBC.Insert(newBlock)

				ForwardHeartBeat(heartBeatData)
				ifTryNonce = false
			}
		}
		ifTryNonce = true
	}
}

func Canonical(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n", SBC.Canonical())
}

func hashToString(hashStr string) string {
	sum := sha3.Sum256([]byte(hashStr))
	return hex.EncodeToString(sum[:])
}

func verifyBlock(sha3StrResult string) bool {
	if strings.HasPrefix(sha3StrResult, "000000") {
		return true
	}
	return false
}
