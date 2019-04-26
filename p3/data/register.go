package data

import (
	"encoding/json"
	"log"
)

type RegisterData struct {
	AssignedId  int32  `json:"assignedId"`
	PeerMapJson string `json:"peerMapJson"`
}

func NewRegisterData(id int32, peerMapJson string) RegisterData {
	newRegisterData := RegisterData{id, peerMapJson}
	return newRegisterData
}

func (data *RegisterData) EncodeToJson() (string, error) {
	jsonRegisterData := RegisterData{data.AssignedId, data.PeerMapJson}
	jsonRegisterDataBytes, err := json.Marshal(jsonRegisterData)
	if err != nil {
		log.Fatal("encode to jsonRegisterData eror", err)
	}

	return string(jsonRegisterDataBytes), err
}
