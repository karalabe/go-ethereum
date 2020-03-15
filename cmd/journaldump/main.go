package main

import (
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/crypto"
	"os"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "<datadir>")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
Exports the contents of the memory difflayers in the journal`)
	}
}

func main() {
	flag.Parse()
	var ddir string
	switch {
	case flag.NArg() == 1:
		ddir = flag.Arg(0)
	default:
		fmt.Fprintln(os.Stderr, "Error: one argument needed")
		flag.Usage()
		os.Exit(2)

	}
	diskdb, err := rawdb.NewLevelDBDatabase(ddir, 512, 100, "chaindata")
	if err != nil {
		fmt.Printf("Error opening db %v\n", err)
		os.Exit(1)
	}
	defer diskdb.Close()
	err = snapshot.ExportSnapshot(diskdb)
	if err != nil {
		fmt.Printf("Error exporting snapshot: %v\n", err)
		os.Exit(1)
	}
}
