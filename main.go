package main

import (
	"./p3"
	"log"
	"net/http"
	"os"
)

func main() {

	router := p3.NewRouter()
	if len(os.Args) > 1 {
		log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	} else {
		log.Fatal(http.ListenAndServe(":6688", router))
	}

	//mpt := p1.MerklePatriciaTrie{}
	//mpt.Initial()
	//mpt.Insert("hello", "world")
	//newBlock := p2.NewBlock(1, 11, "parentHash", "nonce", mpt)
	//jsonString, _ := newBlock.EncodeToJSON()
	//fmt.Println(jsonString)
	//
	//transactionFee := transaction.TransactionFee{"transSender", "transReceiver", 111, 0.0005,
	//	"ETH", "null"}
	//fmt.Println("encode to transactionFeeJson:")
	//transactionFeeJson, _ := transactionFee.EncodeTfToJSON()
	//fmt.Println(transactionFeeJson)
	//fmt.Println("---------------------")
	//
	//payment := transaction.Payment{"Sender", "Receiver", 222, 3,
	//	"ETH", "null", transactionFeeJson}
	//fmt.Println("encode to paymentJson:")
	//paymentJson, _ := payment.EncodePaymentToJSON()
	//fmt.Println(paymentJson)
	//fmt.Println("---------------------")
	//
	//song := songInfo.Song{"Hello", "Adele", "25", "Adele",
	//	2, "Adele Adkins & Greg Kurstin", "pop", "2015",
	//	"lalalalalal", "great", transactionFeeJson}
	//
	//fmt.Println("original songInfo:")
	//fmt.Println(song)
	//fmt.Println("---------------------")
	//fmt.Println("encode to json")
	//songJson, _ := song.EncodeSongToJSON()
	//fmt.Println(songJson)
	//fmt.Println("---------------------")
	//fmt.Println("decode to songInfo")
	//newSong, _ := songInfo.DecodeSongFromJson(songJson)
	//fmt.Println(newSong)
	//
	//
	//fmt.Println("---------------------")
	//myWallet := accounts.NewMyWallet("001")
	//fmt.Println(myWallet.ShowAllBalance())
	//
	//fmt.Println("-----------key----------")
	//key := accounts.Keys{}
	//key.GenerateKey()
	//private, public := key.GetKeys()
	//fmt.Println("private:", private)
	//fmt.Println("public:", public)
	////json , _ := key.EncodePublicKeyToJSON()
	////fmt.Println("encode:", json)
	//key.Test()

}
