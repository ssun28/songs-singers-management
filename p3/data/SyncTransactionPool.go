package data

import (
	"fmt"
	"strconv"
	"sync"
)

type SynTransactionPool struct {
	stpMap map[string]string
	mux    sync.Mutex
}

func (stp *SynTransactionPool) Initial() {
	stp.stpMap = make(map[string]string)
}

func (stp *SynTransactionPool) Add(id int32, timestamp int64, transactionJson string) {
	stp.mux.Lock()
	key := strconv.FormatInt(int64(id), 10) + "id" + strconv.FormatInt(timestamp, 10)
	stp.stpMap[key] = transactionJson
	stp.mux.Unlock()
}

func (stp *SynTransactionPool) Delete(key string) {
	stp.mux.Lock()
	_, ok := stp.stpMap[key]
	if ok {
		delete(stp.stpMap, key)
	}
	stp.mux.Unlock()
}

func (stp *SynTransactionPool) Copy() map[string]string {
	stp.mux.Lock()
	defer stp.mux.Unlock()
	copyMap := make(map[string]string)
	for id, t := range stp.stpMap {
		copyMap[id] = t
	}
	return copyMap
}

func (stp *SynTransactionPool) Show() string {
	stp.mux.Lock()
	defer stp.mux.Unlock()
	rs := "Here are the transactions in the Transaction Pool:\n"
	fmt.Println("stpMapsize:", len(stp.stpMap))
	for key, transactionJson := range stp.stpMap {
		rs += fmt.Sprintf("id:%v; ", key)
		rs += fmt.Sprintf("Transaction:%s\n", transactionJson)
	}
	return rs
}
