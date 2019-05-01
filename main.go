package main

import (
	"./p1"
	"./p2"
	"./transaction"
	"fmt"
)

func main() {

	//router := p3.NewRouter()
	//if len(os.Args) > 1 {
	//	log.Fatal(http.ListenAndServe(":"+os.Args[1], router))
	//} else {
	//	log.Fatal(http.ListenAndServe(":6688", router))
	//}

	mpt := p1.MerklePatriciaTrie{}
	mpt.Initial()
	mpt.Insert("hello", "world")
	newBlock := p2.NewBlock(1, 11, "parentHash", "nonce", mpt)
	jsonString, _ := newBlock.EncodeToJSON()
	fmt.Println(jsonString)

	transactionFee := transaction.TransactionFee{"transSender", "transReceiver", 111, 0.0005,
		"ETH", "null"}
	fmt.Println("encode to transactionFeeJson:")
	transactionFeeJson, _ := transactionFee.EncodeTfToJSON()
	fmt.Println(transactionFeeJson)
	fmt.Println("---------------------")

	payment := transaction.Payment{"Sender", "Receiver", 222, 3,
		"ETH", "null", transactionFeeJson}
	fmt.Println("encode to paymentJson:")
	paymentJson, _ := payment.EncodePaymentToJSON()
	fmt.Println(paymentJson)
	fmt.Println("---------------------")

	song := transaction.Song{"Hello", "Adele", "25", "Adele",
		2, "Adele Adkins & Greg Kurstin", "pop", "2015",
		"lalalalalal", "great", transactionFeeJson}

	fmt.Println("original song:")
	fmt.Println(song)
	fmt.Println("---------------------")
	fmt.Println("encode to json")
	songJson, _ := song.EncodeSongToJSON()
	fmt.Println(songJson)
	fmt.Println("---------------------")
	fmt.Println("decode to song")
	newSong, _ := transaction.DecodeSongFromJson(songJson)
	fmt.Println(newSong)

}
