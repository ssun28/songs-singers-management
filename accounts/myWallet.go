package accounts

import (
	"fmt"
	"sync"
)

type MyWallet struct {
	Address     string
	CurrencyMap map[string]float64
	SongMap     map[string]string
	privateKey  string
	publicKey   string
	mux         sync.Mutex
}

func NewMyWallet(address string) MyWallet {
	newCurrencyMap := make(map[string]float64)
	newSongMap := make(map[string]string)
	keys := Keys{}
	keys.GenerateKey()
	privateKey, publicKey := keys.GetKeys()
	newMyWallet := MyWallet{Address: address, CurrencyMap: newCurrencyMap, SongMap: newSongMap, privateKey: privateKey, publicKey: publicKey}
	newMyWallet.Initial()

	return newMyWallet
}

func (myWallet *MyWallet) Initial() {
	myWallet.CurrencyMap["ETH"] = 100
}

func (myWallet *MyWallet) Deposit(currency string, amount float64) string {
	myWallet.mux.Lock()
	defer myWallet.mux.Unlock()
	if val, ok := myWallet.CurrencyMap[currency]; ok {
		myWallet.CurrencyMap[currency] = val + amount
	} else {
		myWallet.CurrencyMap[currency] = amount
	}

	rs := fmt.Sprintf("Deposit %s %g successfully!\n", currency, amount)
	return rs
}

func (myWallet *MyWallet) Withdraw(currency string, amount float64) string {
	myWallet.mux.Lock()
	defer myWallet.mux.Unlock()
	fmt.Println("beforewithdraw", myWallet.CurrencyMap[currency])
	isValid, rs := myWallet.checkBalance(currency, amount)
	fmt.Println("afterwithdraw", myWallet.CurrencyMap[currency])
	if isValid {
		rs += fmt.Sprintf("Withdraw %s %g successfully!\n", currency, amount)

	} else {
		fail := fmt.Sprintf("Fail to withdraw %s %g!\n", currency, amount)
		rs = fail + rs
	}

	return rs
}

func (myWallet *MyWallet) checkBalance(currency string, amount float64) (bool, string) {
	rs := ""

	if val, ok := myWallet.CurrencyMap[currency]; ok {
		if val >= amount {
			fmt.Println("value:", val)
			myWallet.CurrencyMap[currency] = val - amount
			fmt.Println("inwithdraw", myWallet.CurrencyMap[currency])
			return true, rs
		} else {
			rs += fmt.Sprintf("No enough balance and please check in your wallet!\n")
		}
	} else {
		rs += fmt.Sprintf("No such currency %s! in your wallet!\n", currency)
	}

	return false, rs
}

func (myWallet *MyWallet) GetKeys() string {
	rs := "Here is your key pair :\n"
	rs += fmt.Sprintf("your privateKey:%s\nyour publicKey:%s\n", myWallet.privateKey, myWallet.publicKey)
	return rs
}

func (myWallet *MyWallet) ShowAllBalance() string {
	myWallet.mux.Lock()
	defer myWallet.mux.Unlock()

	rs := "The current balance in your wallet:\n"
	for currency, amount := range myWallet.CurrencyMap {
		rs += fmt.Sprintf("currency=%s, amount=%g", currency, amount)
	}

	return rs
}

func (myWallet *MyWallet) AddSoong(transactionId string, songUrl string) string {
	myWallet.mux.Lock()
	defer myWallet.mux.Unlock()
	myWallet.SongMap[transactionId] = songUrl

	rs := fmt.Sprintf("Add %s %s successfully!\n", transactionId, songUrl)
	return rs
}

func (myWallet *MyWallet) ShowSongs() string {
	myWallet.mux.Lock()
	defer myWallet.mux.Unlock()

	rs := "The current songs' URL in your wallet:\n"
	for transactionId, songUrl := range myWallet.SongMap {
		rs += fmt.Sprintf("transactionId=%s, songUrl=%s\n", transactionId, songUrl)
	}

	return rs
}
