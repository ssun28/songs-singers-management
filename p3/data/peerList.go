package data

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"sync"
)

type PeerList struct {
	selfId    int32
	peerMap   map[string]int32
	maxLength int32
	mux       sync.Mutex
}

type Peer struct {
	Addr string `json:"addr"`
	Id   int32  `json:"id"`
}

func NewPeerList(id int32, maxLength int32) PeerList {
	newPeerMap := make(map[string]int32)
	newPeerList := PeerList{selfId: id, peerMap: newPeerMap, maxLength: maxLength}
	return newPeerList
}

func (peers *PeerList) Add(addr string, id int32) {
	peers.mux.Lock()
	peers.peerMap[addr] = id
	peers.mux.Unlock()
}

func (peers *PeerList) Delete(addr string) {
	peers.mux.Lock()
	delete(peers.peerMap, addr)
	peers.mux.Unlock()
}

func (peers *PeerList) Rebalance() {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	maxLength := peers.maxLength

	if int(maxLength) >= len(peers.peerMap) {
		return
	}

	tempArray := make([]int32, 0)
	tempArray = append(tempArray, peers.selfId)
	for _, id := range peers.peerMap {
		tempArray = append(tempArray, id)
	}

	sort.Slice(tempArray, func(i, j int) bool {
		return tempArray[i] < tempArray[j]
	})

	index := binarySearch(tempArray, peers.selfId)
	if index == -1 {
		return
	}

	newPeersArray := make([]int32, 0)
	halfMaxLength := int(maxLength / 2)

	if index-halfMaxLength < 0 && index+halfMaxLength < len(tempArray) {
		newPeersArray = append(newPeersArray, tempArray[0:index]...)
		addLeftIndex := len(tempArray) - (halfMaxLength - index)
		newPeersArray = append(newPeersArray, tempArray[addLeftIndex:]...)
		newPeersArray = append(newPeersArray, tempArray[index+1:index+halfMaxLength+1]...)
	} else if index-halfMaxLength >= 0 && index+halfMaxLength >= len(tempArray) {
		newPeersArray = append(newPeersArray, tempArray[index-halfMaxLength:index]...)
		newPeersArray = append(newPeersArray, tempArray[index+1:]...)
		addRightIndex := halfMaxLength - (len(tempArray) - 1 - index)
		newPeersArray = append(newPeersArray, tempArray[0:addRightIndex]...)

	} else {
		newPeersArray = append(newPeersArray, tempArray[index-halfMaxLength:index]...)
		newPeersArray = append(newPeersArray, tempArray[index+1:index+halfMaxLength+1]...)
	}

	sort.Slice(newPeersArray, func(i, j int) bool {
		return newPeersArray[i] < newPeersArray[j]
	})

	for addr, id := range peers.peerMap {
		if binarySearch(newPeersArray, id) == -1 {
			delete(peers.peerMap, addr)
		}
	}
}

func (peers *PeerList) Show() string {
	peers.mux.Lock()
	defer peers.mux.Unlock()

	rs := ""
	for addr, id := range peers.peerMap {
		rs += fmt.Sprintf("addr=%s, id=%v;", addr, id)
	}

	rs = fmt.Sprintf("This is PeerMap:\n") + rs

	return rs
}

func (peers *PeerList) Register(id int32) {
	peers.selfId = id
	fmt.Printf("SelfId=%v\n", id)
}

func (peers *PeerList) Copy() map[string]int32 {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	copyMap := make(map[string]int32)
	for addr, id := range peers.peerMap {
		copyMap[addr] = id
	}
	return copyMap
}

func (peers *PeerList) GetSelfId() int32 {
	return peers.selfId
}

func (peers *PeerList) PeerMapToJson() (string, error) {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	var jsonPeerArray []Peer

	for addr, id := range peers.peerMap {
		jsonPeer := Peer{addr, id}
		jsonPeerArray = append(jsonPeerArray, jsonPeer)
	}

	jsonPeerMapBytes, err := json.Marshal(jsonPeerArray)
	if err != nil {
		log.Fatal("encode to jsonPeerMap error:", err)
	}

	return string(jsonPeerMapBytes), err
}

func (peers *PeerList) InjectPeerMapJson(peerMapJsonStr string, selfAddr string) {
	peers.mux.Lock()
	jsonPeerArray := make([]Peer, 0)
	err := json.Unmarshal([]byte(peerMapJsonStr), &jsonPeerArray)

	if err != nil {
		log.Fatal("decode jsonPeerMap error:", err)
	}

	for _, peer := range jsonPeerArray {
		if peer.Addr == selfAddr {
			continue
		}
		peers.peerMap[peer.Addr] = peer.Id
	}
	peers.mux.Unlock()
}

func binarySearch(nums []int32, target int32) int {
	if nums == nil || len(nums) == 0 {
		return -1
	}

	start := 0
	end := len(nums) - 1
	for start+1 < end {
		mid := start + (end-start)/2
		if nums[mid] == target {
			return mid
		} else if nums[mid] < target {
			start = mid
		} else {
			end = mid
		}
	}

	if nums[start] == target {
		return start
	}
	if nums[end] == target {
		return end
	}

	return -1
}

func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	expected.Add("-1-1", -1)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("7777", 7)
	peers.Add("9999", 9)
	peers.Add("11111111", 11)
	peers.Add("2020", 20)
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("7777", 7)
	expected.Add("9999", 9)
	expected.Add("2020", 20)
	fmt.Println(reflect.DeepEqual(peers, expected))
}
