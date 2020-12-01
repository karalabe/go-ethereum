package main

import (
	"github.com/ethereum/go-ethereum/eth/protocols/snap"
)

func main() {
	snap.FuzzD([]byte("deodin\xff\xff"))

}
