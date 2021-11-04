// Copyright 2018 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package downloader

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// Test chain parameters.
var (
	testKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	testAddress = crypto.PubkeyToAddress(testKey.PublicKey)
	testDB      = rawdb.NewMemoryDatabase()
	testGenesis = core.GenesisBlockForTesting(testDB, testAddress, big.NewInt(1000000000000000))
)

// The common prefix of all test chains:
var testChainBase = newTestChain(blockCacheMaxItems+200, testGenesis)

// Different forks on top of the base chain:
var testChainForkLightA, testChainForkLightB, testChainForkHeavy *testChain

func init() {
	var forkLen = int(fullMaxForkAncestry + 50)
	var wg sync.WaitGroup
	wg.Add(3)
	go func() { testChainForkLightA = testChainBase.makeFork(forkLen, false, 1); wg.Done() }()
	go func() { testChainForkLightB = testChainBase.makeFork(forkLen, false, 2); wg.Done() }()
	go func() { testChainForkHeavy = testChainBase.makeFork(forkLen, true, 3); wg.Done() }()
	wg.Wait()
}

type testChain struct {
	blocks []*types.Block
}

// newTestChain creates a blockchain of the given length.
func newTestChain(length int, genesis *types.Block) *testChain {
	tc := &testChain{
		blocks: []*types.Block{genesis},
	}
	tc.generate(length-1, 0, genesis, false)
	return tc
}

// makeFork creates a fork on top of the test chain.
func (tc *testChain) makeFork(length int, heavy bool, seed byte) *testChain {
	fork := tc.copy(len(tc.blocks) + length)
	fork.generate(length, seed, tc.blocks[len(tc.blocks)-1], heavy)
	return fork
}

// shorten creates a copy of the chain with the given length. It panics if the
// length is longer than the number of available blocks.
func (tc *testChain) shorten(length int) *testChain {
	if length > len(tc.blocks) {
		panic(fmt.Errorf("can't shorten test chain to %d blocks, it's only %d blocks long", length, len(tc.blocks)))
	}
	return tc.copy(length)
}

func (tc *testChain) copy(newlen int) *testChain {
	if newlen > len(tc.blocks) {
		newlen = len(tc.blocks)
	}
	cpy := &testChain{
		blocks: append([]*types.Block{}, tc.blocks[:newlen]...),
	}
	return cpy
}

// generate creates a chain of n blocks starting at and including parent.
// the returned hash chain is ordered head->parent. In addition, every 22th block
// contains a transaction and every 5th an uncle to allow testing correct block
// reassembly.
func (tc *testChain) generate(n int, seed byte, parent *types.Block, heavy bool) {
	blocks, _ := core.GenerateChain(params.TestChainConfig, parent, ethash.NewFaker(), testDB, n, func(i int, block *core.BlockGen) {
		block.SetCoinbase(common.Address{seed})
		// If a heavy chain is requested, delay blocks to raise difficulty
		if heavy {
			block.OffsetTime(-9)
		}
		// Include transactions to the miner to make blocks more interesting.
		if parent == tc.blocks[0] && i%22 == 0 {
			signer := types.MakeSigner(params.TestChainConfig, block.Number())
			tx, err := types.SignTx(types.NewTransaction(block.TxNonce(testAddress), common.Address{seed}, big.NewInt(1000), params.TxGas, block.BaseFee(), nil), signer, testKey)
			if err != nil {
				panic(err)
			}
			block.AddTx(tx)
		}
		// if the block number is a multiple of 5, add a bonus uncle to the block
		if i > 0 && i%5 == 0 {
			block.AddUncle(&types.Header{
				ParentHash: block.PrevBlock(i - 2).Hash(),
				Number:     big.NewInt(block.Number().Int64() - 1),
			})
		}
	})
	tc.blocks = append(tc.blocks, blocks...)
}
