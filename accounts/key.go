package accounts

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
)

type Keys struct {
	//PrivateKey string `json:"privateKey"`
	//PublicKey  string `json:"publicKey"`
	ecdPrivateKey ecdsa.PrivateKey
}

type PublicKey struct {
	Curve elliptic.Curve `json:"elliptic.Curve"`
	X     *big.Int       `json:"X"`
	Y     *big.Int       `json:"Y"`
}

//type PublicKey struct {
//	PublicKey ecdsa.PublicKey
//}

func (keys *Keys) GenerateKey() {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	keys.ecdPrivateKey = *privateKey
}

func (keys *Keys) GetKeys() (string, string) {
	x, _ := keys.ecdPrivateKey.PublicKey.X.GobEncode()
	y, _ := keys.ecdPrivateKey.PublicKey.Y.GobEncode()
	publicKeyBytesArray := append(x, y...)
	publicKey := hex.EncodeToString(publicKeyBytesArray)

	d, _ := keys.ecdPrivateKey.D.GobEncode()
	privateKeyBytesArray := append(publicKeyBytesArray, d...)
	privateKey := hex.EncodeToString(privateKeyBytesArray)

	return privateKey, publicKey
}

func (keys *Keys) EncodePublicKeyToJSON() (string, error) {
	publicKey := PublicKey{keys.ecdPrivateKey.PublicKey.Curve, keys.ecdPrivateKey.X, keys.ecdPrivateKey.Y}
	publicKeyJson, err := json.Marshal(publicKey)
	fmt.Println("bytes", publicKeyJson)
	if err != nil {
		log.Fatal("encode to publicKeyJson error:", err)
	}

	return string(publicKeyJson), err
}

func DecodePublicKeyFromJson(publicKeyJson string) (PublicKey, error) {
	var publicKey PublicKey
	fmt.Println("a", publicKeyJson)
	err := json.Unmarshal([]byte(publicKeyJson), &publicKey)
	if err != nil {
		log.Fatal("decode publicKeyJson to publicKey error:", err)
	}

	return publicKey, err
}

func (keys *Keys) Test() {
	msg := "hello, world"
	hash := sha256.Sum256([]byte(msg))

	r, s, err := ecdsa.Sign(rand.Reader, &keys.ecdPrivateKey, hash[:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("signature: (0x%x, 0x%x)\n", r, s)

	str, _ := keys.EncodePublicKeyToJSON()
	fmt.Println("str:", str)
	p, _ := DecodePublicKeyFromJson(str)

	fmt.Println("p", p)

	//p := ecdsa.PublicKey{p.Curve, p.X, p.Y}
	//
	//valid := ecdsa.Verify(p, hash[:], r, s)
	//fmt.Println("signature verified:", valid)
}
