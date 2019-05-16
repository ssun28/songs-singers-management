package p3

import (
	"../accounts"
	"../p1"
	"../p2"
	"../songInfo"
	"../transaction"
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
var USER_TYPE = ""
var GAS_LIMIT = 2

//var SELF_ADDR = "http://localhost:6686"

var SBC data.SyncBlockChain
var Peers data.PeerList
var ifStarted bool
var ifTryNonce bool
var MyWallet accounts.MyWallet
var SL data.SynSongsLibrary
var STP data.SynTransactionPool

// This function will be executed before everything else.
// Do some initialization here.
func init() {
	if len(os.Args) > 1 {
		ID_STR = os.Args[1]
		USER_TYPE = os.Args[2]
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

	MyWallet = accounts.NewMyWallet(ID_STR)
	MyWallet.Initial()
	SL.Initial()
	STP.Initial()

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

	randNum := 5 + rand.Intn(6)
	time.Sleep(time.Duration(randNum) * time.Second)

	if USER_TYPE == "singer" {
		go TransactionPostSong()
	} else {
		go StartTryNonces()
	}

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

	if heartBeatData.IfNewTransaction {
		fmt.Println("$$$$$$have new transaction!!!")
	}
	//if heartBeatData.Addr != SELF_ADDR {

	Peers.Add(heartBeatData.Addr, heartBeatData.Id)

	Peers.InjectPeerMapJson(heartBeatData.PeerMapJson, SELF_ADDR)

	if heartBeatData.IfNewTransaction {
		STP.Add(heartBeatData.Id, heartBeatData.Timestamp, heartBeatData.TransactionJson)
		fmt.Println("$$$$$$", STP.Show())
		transaction, _ := transaction.DecodeTransactionFromJson(heartBeatData.TransactionJson)
		if transaction.Category == "song" {
			SL.Add(heartBeatData.Id, heartBeatData.Timestamp, transaction.Value)
		}
		heartBeatData.Hops--
		if heartBeatData.Hops > 0 {
			ForwardHeartBeat(heartBeatData)
		}
	}

	if heartBeatData.IfNewBlock {
		fmt.Println("heartbeatId:", heartBeatData.Id)
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
				if !SBC.CheckCurrentBlock(block.Header.Height, block.Header.Hash) {
					id, _ := strconv.Atoi(ID_STR)
					mpt := block.Value.Kv
					if heartBeatData.Id == int32(id) {
						fmt.Println("!!!!I got block reward!")
						MyWallet.Deposit("ETH", 10.0)
						mptSize := len(mpt)
						fmt.Println("size:@@@@", mptSize)
						MyWallet.Deposit("ETH", 0.5*float64(mptSize))
					}

					for key, transactionJson := range mpt {
						t, _ := transaction.DecodeTransactionFromJson(transactionJson)
						selfId := strings.Split(key, "id")[0]
						fmt.Println("selfid:", selfId)

						if t.Category == "song" {
							if selfId == ID_STR {
								MyWallet.Withdraw("ETH", 0.5)
							}
						} else if t.Category == "payment" {
							if selfId == ID_STR {
								MyWallet.Withdraw("ETH", 0.5)
								MyWallet.Withdraw("ETH", 1.0)
							}
							paymentJson := t.Value
							payment, _ := transaction.DecodePaymentFromJson(paymentJson)
							if payment.Receiver == ID_STR {
								MyWallet.Deposit("ETH", 1.0)
								others := payment.Others
								othersArray := strings.Split(others, ":")
								transactionId := othersArray[1]
								songName := othersArray[2]
								songUrl := "www.DApp-listen-" + songName + ".com"
								UploadSong(payment.Sender, transactionId, songUrl)
							}
						}
						fmt.Println(STP.Show())
						STP.Delete(key)
					}

					SBC.Insert(block)
				}
			}

			heartBeatData.Hops--
			if heartBeatData.Hops > 0 {
				ForwardHeartBeat(heartBeatData)
			}

			ifTryNonce = true
		}
	}

	fmt.Fprintf(w, "heartBeatReceive sucess")
	//}
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

func AskForSong(w http.ResponseWriter, r *http.Request) {
	u, _ := url.Parse(r.URL.Path)
	urlPath := strings.Split(u.Path, "/")

	transactionId := urlPath[2]
	addrId := strings.Split(transactionId, "id")[0]
	songName := urlPath[3]

	timestamp := int64(time.Now().Unix())
	transactionFee := transaction.TransactionFee{MyWallet.Address, "transReceiver", timestamp, 0.5,
		"ETH", "null"}

	transactionFeeJson, _ := transactionFee.EncodeTfToJSON()
	payment := transaction.Payment{MyWallet.Address, addrId, timestamp, 1.0, "ETH", "listen:" + transactionId + ":" + songName, transactionFeeJson}

	paymentJson, _ := payment.EncodePaymentToJSON()
	transaction := transaction.NewTransaction("payment", paymentJson, "pending")

	transactionJson, _ := transaction.EncodeTransactionToJSON()
	peerMapJsonStr, _ := Peers.PeerMapToJson()
	heartBeatData := data.NewHeartBeatData(false, Peers.GetSelfId(), "", peerMapJsonStr, SELF_ADDR, true, transactionJson, timestamp)
	ForwardHeartBeat(heartBeatData)

	fmt.Fprintf(w, "Waiting for the request:[transactionId:%s] be confirmed and the singer will send you an url to listen to the music!", transactionId)
}

func UploadSong(receiver string, transactionId string, songUrl string) {
	resp, err := http.Get(PRE_SELF_ADDR + receiver + "/getSongUrl" + "/" + transactionId + "/" + songUrl)
	if err != nil {
		log.Fatal("Ask for block resp error:", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(body)
}

func GetSongUrl(w http.ResponseWriter, r *http.Request) {
	u, _ := url.Parse(r.URL.Path)
	urlPath := strings.Split(u.Path, "/")

	transactionId := urlPath[2]
	songUrl := urlPath[3]
	rs := MyWallet.AddSoong(transactionId, songUrl)
	fmt.Fprintf(w, rs)
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
		//randNum := rand.Intn(10000)
		mpt := p1.MerklePatriciaTrie{}
		mpt.Initial()

		//key := strconv.FormatInt(int64(randNum), 10)
		//mpt.Insert(key, "")
		count := 0

		for len(STP.Copy()) == 0 {
			randNum := 5 + rand.Intn(6)
			time.Sleep(time.Duration(randNum) * time.Second)
		}

		transactionPoolCopy := STP.Copy()

		fmt.Println("length:", len(transactionPoolCopy))
		for id, transactionJson := range transactionPoolCopy {
			if count > GAS_LIMIT {
				break
			}
			transaction, _ := transaction.DecodeTransactionFromJson(transactionJson)
			transaction.Status = "confirm"
			newTransactionJson, _ := transaction.EncodeTransactionToJSON()
			mpt.Insert(id, newTransactionJson)
			STP.Delete(id)
			count++
		}

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
				heartBeatData := data.NewHeartBeatData(true, Peers.GetSelfId(), blockJson, peerMapJsonStr, SELF_ADDR, false, "", 0)

				//SBC.Insert(newBlock)

				ForwardHeartBeat(heartBeatData)
				ifTryNonce = false
			}
		}
		ifTryNonce = true
	}
}

func TransactionPostSong() {
	timestamp := int64(time.Now().Unix())
	transactionFee := transaction.TransactionFee{MyWallet.Address, "transReceiver", timestamp, 0.5,
		"ETH", "null"}

	transactionFeeJson, _ := transactionFee.EncodeTfToJSON()
	song := songInfo.Song{}
	if ID_STR == "1111" {
		song = songInfo.Song{"Hello", "Adele", "25", "Adele",
			2, "Adele Adkins & Greg Kurstin", "pop", "2015",
			"lalalalalal", "great", transactionFeeJson}
	} else {
		song = songInfo.Song{"ME! (feat. Brendon Urie of Panic! At The Disco)", "Taylor Swift", "ME! (feat. Brendon Urie of Panic! At The Disco) - Single", "",
			0, "Taylor Swift, Brendon Urie & Joel Little", "Pop", "2019",
			"memememememe", "nice", transactionFeeJson}
	}

	songJson, _ := song.EncodeSongToJSON()
	transaction := transaction.NewTransaction("song", songJson, "pending")
	transactionJson, _ := transaction.EncodeTransactionToJSON()
	peerMapJsonStr, _ := Peers.PeerMapToJson()
	heartBeatData := data.NewHeartBeatData(false, Peers.GetSelfId(), "", peerMapJsonStr, SELF_ADDR, true, transactionJson, timestamp)
	ForwardHeartBeat(heartBeatData)
}

func verifyBlock(sha3StrResult string) bool {
	if strings.HasPrefix(sha3StrResult, "000000") {
		return true
	}
	return false
}

func Canonical(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n", SBC.Canonical())
}

func ShowSongs(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n", SL.Show())
}

func TransactionPool(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n", STP.Show())
}

func ShowMyWallet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n%s\n%s\n", MyWallet.GetKeys(), MyWallet.ShowAllBalance(), MyWallet.ShowSongs())
}

func hashToString(hashStr string) string {
	sum := sha3.Sum256([]byte(hashStr))
	return hex.EncodeToString(sum[:])
}
