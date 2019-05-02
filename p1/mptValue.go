package p1

import (
	"encoding/json"
	"log"
)

type MptValue struct {
	Category string `json:"type"`
	Value    string `json:"value"`
	Status   string `json:"status"`
}

func NewMptValue(category string, value string, status string) MptValue {
	newMptValue := MptValue{category, value, status}

	return newMptValue
}

func (mptValue *MptValue) EncodeMptValueToJSON() (string, error) {
	mptValueJson, err := json.Marshal(mptValue)
	if err != nil {
		log.Fatal("encode to mptValueJson error:", err)
	}

	return string(mptValueJson), err
}

func DecodeMptValueFromJson(mptValueJson string) (MptValue, error) {
	var mptValue MptValue
	err := json.Unmarshal([]byte(mptValueJson), &mptValue)
	if err != nil {
		log.Fatal("decode mptValueJson to mptValue error:", err)
	}

	return mptValue, err
}
