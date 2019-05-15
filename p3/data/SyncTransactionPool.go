package data

import (
	"fmt"
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
	key := string(id) + "id" + string(timestamp)
	stp.stpMap[key] = transactionJson
	stp.mux.Unlock()
}

func (stp *SynTransactionPool) Delete(id int32, timestamp int64, transactionJson string) {
	stp.mux.Lock()
	key := string(id) + "id" + string(timestamp)
	_, ok := stp.stpMap[key]
	if ok {
		delete(stp.stpMap, key)
	}
	stp.mux.Unlock()
}

func (stp *SynTransactionPool) Show() string {
	stp.mux.Lock()
	defer stp.mux.Unlock()
	rs := "Here are the transaction in the Transaction Pool:\n"
	for key, transactionJson := range stp.stpMap {
		rs += fmt.Sprintf("id %v: ", key)
		rs += fmt.Sprintf("Transaction:%s\n", transactionJson)
	}
	return rs
}
