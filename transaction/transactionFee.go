package transaction

import (
	"encoding/json"
	"log"
)

type TransactionFee struct {
	Sender    string  `json:"sender"`
	Receiver  string  `json:"receiver"`
	Timestamp int64   `json:"timestamp"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Others    string  `json:"others"`
}

func NewTransactionFee(sender string, receiver string, timestamp int64, amount float64, currency string,
	others string) TransactionFee {
	newTransactionFee := TransactionFee{sender, receiver, timestamp,
		amount, currency, others}

	return newTransactionFee
}

func (transactionFee *TransactionFee) EncodeTfToJSON() (string, error) {
	transactionFeeJson, err := json.Marshal(transactionFee)
	if err != nil {
		log.Fatal("encode to transactionFeeJson error:", err)
	}

	return string(transactionFeeJson), err
}

func DecodeTfFromJson(transactionFeeJson string) (TransactionFee, error) {
	var transactionFee TransactionFee
	err := json.Unmarshal([]byte(transactionFeeJson), &transactionFee)
	if err != nil {
		log.Fatal("decode transactionFeeJson to transactionFee error:", err)
	}

	return transactionFee, err
}
