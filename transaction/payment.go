package transaction

import (
	"encoding/json"
	"log"
)

type Payment struct {
	Sender         string  `json:"sender"`
	Receiver       string  `json:"receiver"`
	Timestamp      int64   `json:"timestamp"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	Others         string  `json:"others"`
	TransactionFee string  `json:"transactionFee"`
}

func NewPayment(sender string, receiver string, timestamp int64, amount float64, currency string,
	others string, transactionFee string) Payment {
	newPayment := Payment{sender, receiver, timestamp, amount,
		currency, others, transactionFee}

	return newPayment
}

func (payment *Payment) EncodePaymentToJSON() (string, error) {
	paymentJson, err := json.Marshal(payment)
	if err != nil {
		log.Fatal("encode to paymentJson error:", err)
	}

	return string(paymentJson), err
}

func DecodePaymentFromJson(paymentJson string) (Payment, error) {
	var payment Payment
	err := json.Unmarshal([]byte(paymentJson), &payment)
	if err != nil {
		log.Fatal("decode paymentJson to payment error:", err)
	}

	return payment, err
}
