// Copyright 2019 The go-ethereum Authors
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

package rawdb

import (
	"errors"
	"math"
	"math/big"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/prque"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

type (
	prepareCallback func(*types.Block)              // Callback for custom concurrent pre-computations on block data
	actionCallback  func(ethdb.Batch, *types.Block) // Callback for custom sequential operations on block data.
)

// iterateCanonicalChain iterates the specified range of canonical blocks and applies
// the given action callback. Note, both for forward and backward iteration, the range
// is [from, to).
func iterateCanonicalChain(db ethdb.Database, from uint64, to uint64, prepare prepareCallback, action actionCallback, reverse bool, progMsg, doneMsg string) error {
	// Short circuit if the action is nil or if the range is invalid
	if action == nil {
		return nil
	}
	if from >= to {
		return nil
	}
	// Spawn multi-routines, iterate over the specified blocks and invoke prepare
	// callback concurrently.
	var (
		number  int64
		results = make(chan *types.Block, 4*runtime.NumCPU())
	)
	if !reverse {
		number = int64(from - 1)
	} else {
		number = int64(to)
	}
	abort := make(chan struct{})
	defer close(abort)

	threads := to - from
	if cpus := runtime.NumCPU(); threads > uint64(cpus) {
		threads = uint64(cpus)
	}
	for i := 0; i < int(threads); i++ {
		go func() {
			for {
				// Fetch the next task number, terminating if everything's done
				var n int64
				if !reverse {
					n = atomic.AddInt64(&number, 1)
					if n >= int64(to) {
						return
					}
				} else {
					n = atomic.AddInt64(&number, -1)
					if n < int64(from) {
						return
					}
				}
				block := ReadBlock(db, ReadCanonicalHash(db, uint64(n)), uint64(n))
				if prepare != nil && block != nil {
					prepare(block)
				}
				// Feed the block to the aggregator, or abort on interrupt
				select {
				case results <- block:
				case <-abort:
					return
				}
			}
		}()
	}
	// Reassemble the blocks into a contiguous stream and apply the action callback
	var (
		next, first, last int64
		queue             = prque.New(nil)

		batch  = db.NewBatch()
		start  = time.Now()
		logged = start.Add(-7 * time.Second) // Unindex during import is fast, don't double log
	)
	if !reverse {
		next, first, last = int64(from), int64(from), int64(to)
	} else {
		next, first, last = int64(to-1), int64(to-1), int64(from-1)
	}
	for i := from; i < to; i++ {
		// Retrieve the next result and bail if it's nil
		block := <-results
		if block == nil {
			return errors.New("broken database")
		}
		// Push the block into the import queue and process contiguous ranges
		priority := -int64(block.NumberU64())
		if reverse {
			priority = int64(block.NumberU64())
		}
		queue.Push(block, priority)
		for !queue.Empty() {
			// If the next available item is gapped, return
			if _, priority := queue.Peek(); !reverse && -priority != next || reverse && priority != next {
				break
			}
			// Next block available, pop it off and index it
			block = queue.PopItem().(*types.Block)

			if !reverse {
				next++
			} else {
				next--
			}
			// Invoke action to inject specified data into key-value database.
			action(batch, block)

			// If enough data was accumulated in memory or we're at the last block, dump to disk
			if batch.ValueSize() > ethdb.IdealBatchSize || next == last {
				if err := batch.Write(); err != nil {
					return err
				}
				batch.Reset()
			}
			// If we've spent too much time already, notify the user of what we're doing
			if time.Since(logged) > 8*time.Second {
				log.Info(progMsg, "blocks", int64(math.Abs(float64(next-first))), "total", to-from, "tail", block.Number(), "hash", block.Hash(), "elapsed", common.PrettyDuration(time.Since(start)))
				logged = time.Now()
			}
		}
	}
	tail := to
	if reverse {
		tail = from
	}
	log.Info(doneMsg, "blocks", to-from, "tail", tail, "elapsed", common.PrettyDuration(time.Since(start)))
	return nil
}

// InitDatabaseFromFreezer reinitializes an empty database from a previous batch
// of frozen ancient blocks. The method iterates over all the frozen blocks and
// injects into the database the block hash->number mappings.
func InitDatabaseFromFreezer(db ethdb.Database) {
	// If we can't access the freezer or it's empty, abort
	frozen, err := db.Ancients()
	if err != nil || frozen == 0 {
		return
	}
	var (
		batch  = db.NewBatch()
		start  = time.Now()
		logged = start.Add(-7 * time.Second) // Unindex during import is fast, don't double log
		hash   common.Hash
	)
	for i := uint64(0); i < frozen; i++ {
		if h, err := db.Ancient(freezerHashTable, i); err != nil {
			log.Crit("Failed to init database from freezer", "err", err)
		} else {
			hash = common.BytesToHash(h)
		}

		WriteHeaderNumber(batch, hash, i)
		// If enough data was accumulated in memory or we're at the last block, dump to disk
		if batch.ValueSize() > ethdb.IdealBatchSize {
			if err := batch.Write(); err != nil {
				log.Crit("Failed to write data to db", "err", err)
			}
			batch.Reset()
		}
		// If we've spent too much time already, notify the user of what we're doing
		if time.Since(logged) > 8*time.Second {
			log.Info("Initializing database from freezer", "blocks", i, "total", frozen, "tail", i, "hash", hash, "elapsed", common.PrettyDuration(time.Since(start)))
			logged = time.Now()
		}
	}
	if err := batch.Write(); err != nil {
		log.Crit("Failed to write data to db", "err", err)
	}
	batch.Reset()

	WriteHeadHeaderHash(db, hash)
	WriteHeadFastBlockHash(db, hash)
	log.Info("Initialized database from freezer", "blocks", frozen, "tail", frozen, "elapsed", common.PrettyDuration(time.Since(start)))
}

// IndexTransactions creates txlookup indices of the specified block range.
//
// This function iterates canonical chain in reverse order, it has one main advantage:
// We can write tx index tail flag periodically even without the whole indexing
// procedure is finished. So that we can resume indexing procedure next time quickly.
func IndexTransactions(db ethdb.Database, from uint64, to uint64) {
	// One thread sequentially reads data from db
	type numberRlp struct {
		number uint64
		rlp    rlp.RawValue
	}
	// A set of threads takes the rlp, and hashes the transactions
	type delivery struct {
		number uint64
		hashes []common.Hash
	}

	var (
		batch   = db.NewBatch()
		start   = time.Now()
		logged  = start.Add(-7 * time.Second) // Unindex during import is fast, don't double log
		threads = to - from

		rlpCh    = make(chan *numberRlp, threads*2) // we send raw rlp over this channel
		hashesCh = make(chan *delivery, threads*2)  // send hashes over hashesCh
		abort    = make(chan struct{})
	)
	if cpus := runtime.NumCPU(); threads > uint64(cpus) {
		threads = uint64(cpus)
	}

	// lookup runs in one instance
	lookup := func() {
		for n := from; n < to; n++ {
			data := ReadCanonicalBodyRLP(db, uint64(n))
			// Feed the block to the aggregator, or abort on interrupt
			select {
			case rlpCh <- &numberRlp{n, data}:
			case <-abort:
				return
			}
		}
	}
	// process runs in parallell
	process := func() {
		var hasher = sha3.NewLegacyKeccak256()
		for data := range rlpCh {
			it, err := rlp.NewListIterator(data.rlp)
			if err != nil {
				log.Crit("tx indexing error", "error", err)
				return
			}
			it.Next()
			txs := it.Value()
			txIt, err := rlp.NewListIterator(txs)
			if err != nil {
				log.Crit("tx indexing error", "error", err)
				return
			}
			var hashes []common.Hash
			for txIt.Next() {
				if err := txIt.Err(); err != nil {
					log.Crit("tx indexing error", "error", err)
					return
				}
				var txHash common.Hash
				hasher.Reset()
				hasher.Write(txIt.Value())
				hasher.Sum(txHash[:0])
				hashes = append(hashes, txHash)
			}
			result := &delivery{
				hashes: hashes,
				number: data.number,
			}
			// Feed the block to the aggregator, or abort on interrupt
			select {
			case hashesCh <- result:
			case <-abort:
				return
			}
		}
	}
	defer close(abort)
	go lookup() // start the sequential db accessor
	for i := 0; i < int(threads); i++ {
		go process()
	}
	for delivery := range hashesCh {
		number := new(big.Int).SetUint64(delivery.number).Bytes()
		WriteTxLookupEntriesByHash(batch, number, delivery.hashes)

		// If enough data was accumulated in memory or we're at the last block, dump to disk
		if batch.ValueSize() > ethdb.IdealBatchSize {
			if err := batch.Write(); err != nil {
				log.Crit("Failed writing batch to db", "error", err)
				return
			}
			batch.Reset()
		}
		// If we've spent too much time already, notify the user of what we're doing
		if time.Since(logged) > 8*time.Second {
			log.Info("Indexing transactions", "blocks", int64(math.Abs(float64(delivery.number-from))), "total", to-from, "elapsed", common.PrettyDuration(time.Since(start)))
			logged = time.Now()
		}
	}
	if err := batch.Write(); err != nil {
		log.Crit("Failed writing batch to db", "error", err)
		return
	}
	log.Info("Indexed transactions", "blocks", int64(math.Abs(float64(to-from))), "total", to-from, "elapsed", common.PrettyDuration(time.Since(start)))

	return
}

// UnindexTransactions removes txlookup indices of the specified block range.
func UnindexTransactions(db ethdb.Database, from uint64, to uint64) {
	// Write flag first and then unindex the transaction indices. Some indices
	// will be left in the database if crash happens but it's fine.
	WriteTxIndexTail(db, to)

	// If only one block is unindexed, do it directly
	if from+1 == to {
		DeleteTxLookupEntries(db, ReadBlock(db, ReadCanonicalHash(db, from), from))
		log.Info("Unindexed transactions", "blocks", 1, "tail", to)
		return
	}
	// Otherwise spin up the concurrent iterator and unindexer
	deleteIndices := func(batch ethdb.Batch, block *types.Block) { DeleteTxLookupEntries(batch, block) }
	if err := iterateCanonicalChain(db, from, to, nil, deleteIndices, false, "Unindexing transactions", "Unindexed transactions"); err != nil {
		log.Crit("Failed to unindex transactions", "err", err)
	}
}
