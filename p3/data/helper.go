package data

import (
	"fmt"
	"log"
)

func PrintError(err error, str string) {
	log.Fatal("encode to jsonPeerMap error:", err)
	fmt.Println("%s method error: %s\n", str, err)
}
