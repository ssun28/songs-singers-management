package transaction

import (
	"encoding/hex"
	"encoding/json"
	"golang.org/x/crypto/sha3"
	"log"
)

type Transaction struct {
	Category string `json:"type"`
	Value    string `json:"value"`
	Status   string `json:"status"`
	Hash     string `json:"hash"`
}

func NewTransaction(category string, value string, status string) Transaction {
	hash := hashToString(value)
	newTransaction := Transaction{category, value, status, hash}

	return newTransaction
}

func (transaction *Transaction) EncodeTransactionToJSON() (string, error) {
	transactionJson, err := json.Marshal(transaction)
	if err != nil {
		log.Fatal("encode to transactionJson error:", err)
	}

	return string(transactionJson), err
}

func DecodeTransactionFromJson(transactionJson string) (Transaction, error) {
	var transaction Transaction
	err := json.Unmarshal([]byte(transactionJson), &transaction)
	if err != nil {
		log.Fatal("decode transactionJson to transaction error:", err)
	}

	return transaction, err
}

func hashToString(hashStr string) string {
	sum := sha3.Sum256([]byte(hashStr))
	return hex.EncodeToString(sum[:])
}
