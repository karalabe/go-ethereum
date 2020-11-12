// Copyright 2020 The go-ethereum Authors
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

package snapshot

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

// trieKV represents a trie key-value pair
type trieKV struct {
	key   common.Hash
	value []byte
}

type (
	// trieGeneratorFn is the interface of trie generation which can
	// be implemented by different trie algorithm.
	trieGeneratorFn func(db ethdb.KeyValueWriter, in chan (trieKV), out chan (common.Hash))

	// leafCallbackFn is the callback invoked at the leaves of the trie,
	// returns the subtrie root with the specified subtrie identifier.
	leafCallbackFn func(db ethdb.KeyValueWriter, accountHash, codeHash common.Hash, stat *generateStats) (common.Hash, error)
)

// GenerateAccountTrieRoot takes an account iterator and reproduces the root hash.
func GenerateAccountTrieRoot(it AccountIterator) (common.Hash, error) {
	return generateTrieRoot(nil, it, common.Hash{}, stackTrieGenerate, nil, &generateStats{start: time.Now()}, true)
}

// GenerateStorageTrieRoot takes a storage iterator and reproduces the root hash.
func GenerateStorageTrieRoot(account common.Hash, it StorageIterator) (common.Hash, error) {
	return generateTrieRoot(nil, it, account, stackTrieGenerate, nil, &generateStats{start: time.Now()}, true)
}

// VerifyState takes the whole snapshot tree as the input, traverses all the accounts
// as well as the corresponding storages and compares the re-computed hash with the
// original one(state root and the storage root).
func VerifyState(snaptree *Tree, root common.Hash) error {
	acctIt, err := snaptree.AccountIterator(root, common.Hash{})
	if err != nil {
		return err
	}
	defer acctIt.Release()

	got, err := generateTrieRoot(nil, acctIt, common.Hash{}, stackTrieGenerate, func(db ethdb.KeyValueWriter, accountHash, codeHash common.Hash, stat *generateStats) (common.Hash, error) {
		storageIt, err := snaptree.StorageIterator(root, accountHash, common.Hash{})
		if err != nil {
			return common.Hash{}, err
		}
		defer storageIt.Release()

		hash, err := generateTrieRoot(nil, storageIt, accountHash, stackTrieGenerate, nil, stat, false)
		if err != nil {
			return common.Hash{}, err
		}
		return hash, nil
	}, &generateStats{start: time.Now()}, true)

	if err != nil {
		return err
	}
	if got != root {
		return fmt.Errorf("state root hash mismatch: got %x, want %x", got, root)
	}
	return nil
}

// CommitAndVerifyState takes the whole snapshot tree as the input, traverses all the
// accounts as well as the corresponding storages and commits all re-constructed trie
// nodes to the given database(usually the given db is not the main one used in the
// system, acts as the temporary storage here).
//
// Besides, whenever we meet an account with additional contract code, the code will
// also be migrated to ensure the integrity of the newly created state.
func CommitAndVerifyState(snaptree *Tree, root common.Hash, db ethdb.Database, commitdb ethdb.KeyValueWriter) error {
	// Traverse all state by snapshot, re-construct the whole state trie
	// and commit to the given storage.
	acctIt, err := snaptree.AccountIterator(root, common.Hash{})
	if err != nil {
		return err // The required snapshot might not exist.
	}
	defer acctIt.Release()

	got, err := generateTrieRoot(commitdb, acctIt, common.Hash{}, stackTrieGenerate, func(commitdb ethdb.KeyValueWriter, accountHash, codeHash common.Hash, stat *generateStats) (common.Hash, error) {
		// Migrate the code first, commit the contract code into the tmp db.
		if codeHash != emptyCode {
			code := rawdb.ReadCode(db, codeHash)
			if len(code) == 0 {
				return common.Hash{}, errors.New("failed to migrate contract code")
			}
			rawdb.WriteCode(commitdb, codeHash, code)
		}
		// Then migrate all storage trie nodes into the tmp db.
		storageIt, err := snaptree.StorageIterator(root, accountHash, common.Hash{})
		if err != nil {
			return common.Hash{}, err
		}
		defer storageIt.Release()

		hash, err := generateTrieRoot(commitdb, storageIt, accountHash, stackTrieGenerate, nil, stat, false)
		if err != nil {
			return common.Hash{}, err
		}
		return hash, nil
	}, &generateStats{start: time.Now()}, true)

	if err != nil {
		return err
	}
	if got != root {
		return fmt.Errorf("State root hash mismatch, got %x, want %x", got, root)
	}
	return nil
}

// generateStats is a collection of statistics gathered by the trie generator
// for logging purposes.
type generateStats struct {
	accounts   uint64
	slots      uint64
	curAccount common.Hash
	curSlot    common.Hash
	start      time.Time
	lock       sync.RWMutex
}

// progress records the progress trie generator made recently.
func (stat *generateStats) progress(accounts, slots uint64, curAccount common.Hash, curSlot common.Hash) {
	stat.lock.Lock()
	defer stat.lock.Unlock()

	stat.accounts += accounts
	stat.slots += slots
	stat.curAccount = curAccount
	stat.curSlot = curSlot
}

// report prints the cumulative progress statistic smartly.
func (stat *generateStats) report() {
	stat.lock.RLock()
	defer stat.lock.RUnlock()

	var ctx []interface{}
	if stat.curSlot != (common.Hash{}) {
		ctx = append(ctx, []interface{}{
			"in", stat.curAccount,
			"at", stat.curSlot,
		}...)
	} else {
		ctx = append(ctx, []interface{}{"at", stat.curAccount}...)
	}
	// Add the usual measurements
	ctx = append(ctx, []interface{}{"accounts", stat.accounts}...)
	if stat.slots != 0 {
		ctx = append(ctx, []interface{}{"slots", stat.slots}...)
	}
	ctx = append(ctx, []interface{}{"elapsed", common.PrettyDuration(time.Since(stat.start))}...)
	log.Info("Iterating snapshot", ctx...)
}

// reportDone prints the last log when the whole generation is finished.
func (stat *generateStats) reportDone() {
	stat.lock.RLock()
	defer stat.lock.RUnlock()

	var ctx []interface{}
	ctx = append(ctx, []interface{}{"accounts", stat.accounts}...)
	if stat.slots != 0 {
		ctx = append(ctx, []interface{}{"slots", stat.slots}...)
	}
	ctx = append(ctx, []interface{}{"elapsed", common.PrettyDuration(time.Since(stat.start))}...)
	log.Info("Iterated snapshot", ctx...)
}

// runReport periodically prints the progress information.
func runReport(stats *generateStats, stop chan bool) {
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			stats.report()
			timer.Reset(time.Second * 8)
		case success := <-stop:
			if success {
				stats.reportDone()
			}
			return
		}
	}
}

// subTask wraps the necessary information of a task derived by generateTrieRoot.
type subTask struct {
	account common.Hash
	code    common.Hash
	root    common.Hash
	call    leafCallbackFn
	db      ethdb.KeyValueWriter
	stats   *generateStats
}

// runSubTasks is a helper function of generateTrieRoot. If generateTrieRoot has
// the callback for leaves, all sub-tasks can be accumulated and executed here.
// If any sub-task failed, a signal will be throw back very soon.
//
// Note we expect the out channel has at least 1 slot available so that sending
// won't be blocked.
func runSubTasks(in chan subTask, out chan error, stop chan struct{}) {
	var (
		running int
		failed  bool
		limit   = runtime.NumCPU()
		done    = make(chan error)
	)
	if cap(out) < 1 {
		panic("require buffered channel")
	}
	run := func(task subTask) {
		root, err := task.call(task.db, task.account, task.code, task.stats)
		if err != nil {
			done <- err
			return
		}
		if task.root != root {
			done <- fmt.Errorf("invalid subroot(%x), want %x, got %x", task.account, task.root, root)
			return
		}
		done <- nil
	}
	schedule := func(task subTask) {
		// Short circuit if failure already occurs.
		if failed {
			return
		}
		running += 1
		go run(task)

		// If there are too many runners, block here
		for running >= limit {
			failure := <-done
			running -= 1
			if failure != nil && !failed {
				failed = true
				out <- failure // won't be blocked
			}
		}
	}
	for {
		select {
		case task := <-in:
			schedule(task)
		case failure := <-done:
			running -= 1
			if failure != nil && !failed {
				failed = true
				out <- failure // won't be blocked
			}
		case <-stop:
			// Drain all cached tasks first
		drain:
			for {
				select {
				case task := <-in:
					schedule(task)
				default:
					break drain
				}
			}
			for running > 0 {
				failure := <-done
				running -= 1
				if failure != nil && !failed {
					failed = true
					out <- failure // won't be blocked
				}
			}
			if !failed {
				out <- nil // won't be blocked
			}
			return
		}
	}
}

// generateTrieRoot generates the trie hash based on the snapshot iterator.
// It can be used for generating account trie, storage trie or even the
// whole state which connects the accounts and the corresponding storages.
func generateTrieRoot(db ethdb.KeyValueWriter, it Iterator, account common.Hash, generatorFn trieGeneratorFn, leafCallback leafCallbackFn, stats *generateStats, report bool) (common.Hash, error) {
	var (
		in      = make(chan trieKV)         // chan to pass leaves
		out     = make(chan common.Hash, 1) // chan to collect result
		stoplog = make(chan bool, 1)        // 1-size buffer, works when logging is not enabled
		wg      sync.WaitGroup
	)
	// Spin up a go-routine for trie hash re-generation
	wg.Add(1)
	go func() {
		defer wg.Done()
		generatorFn(db, in, out)
	}()
	// Spin up a go-routine for progress logging
	if report && stats != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runReport(stats, stoplog)
		}()
	}
	// If there is a callback specified, spin up
	// a go-routine for processing the sub-tasks
	// in the background.
	var (
		// The channel size here is quite arbitrary. We don't
		// want to block the main thread for iterating the state
		// trie, but also need to prevent OOM.
		subIn   = make(chan subTask, 1024)
		subOut  = make(chan error, 1)
		subStop = make(chan struct{})
	)
	if leafCallback != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runSubTasks(subIn, subOut, subStop)
		}()
	}
	// stop is a helper function to shutdown the background threads
	// and return the re-generated trie hash. There are three scenarios
	// for calling this function:
	// (a) the failure already occurs when processing the sub-task(e.g.
	//     the storage root is not matched).
	// (b) the failure already occurs when iterating the state trie.
	// (c) there is no failure yet.
	// In case (a) we won't fetch the sub-failure again.
	stop := func(success bool, subFailed bool) (common.Hash, bool) {
		close(in)
		if leafCallback != nil {
			close(subStop)
		}
		result := <-out

		if leafCallback != nil && !subFailed {
			subErr := <-subOut
			success = success && subErr == nil
		}
		stoplog <- success
		wg.Wait()
		return result, success
	}
	var (
		logged    = time.Now()
		processed = uint64(0)
		leaf      trieKV
		last      common.Hash
	)
	// Start to feed leaves
	for it.Next() {
		if account == (common.Hash{}) {
			var (
				err      error
				fullData []byte
			)
			if leafCallback == nil {
				fullData, err = FullAccountRLP(it.(AccountIterator).Account())
				if err != nil {
					stop(false, false)
					return common.Hash{}, err
				}
			} else {
				// Check if there is any sub failure occurs
				select {
				case err := <-subOut:
					if err != nil {
						stop(false, true)
						return common.Hash{}, err
					}
				default:
				}
				account, err := FullAccount(it.(AccountIterator).Account())
				if err != nil {
					stop(false, false)
					return common.Hash{}, err
				}
				subIn <- subTask{
					account: it.Hash(),
					code:    common.BytesToHash(account.CodeHash),
					root:    common.BytesToHash(account.Root),
					call:    leafCallback,
					db:      db,
					stats:   stats,
				}
				fullData, err = rlp.EncodeToBytes(account)
				if err != nil {
					stop(false, false)
					return common.Hash{}, err
				}
			}
			leaf = trieKV{it.Hash(), fullData}
		} else {
			leaf = trieKV{it.Hash(), common.CopyBytes(it.(StorageIterator).Slot())}
		}
		in <- leaf

		// Accumulate the generaation statistic if it's required.
		processed++
		if time.Since(logged) > 3*time.Second && stats != nil {
			if account == (common.Hash{}) {
				stats.progress(processed, 0, it.Hash(), common.Hash{})
			} else {
				stats.progress(0, processed, account, it.Hash())
			}
			logged, processed = time.Now(), 0
		}
		last = it.Hash()
	}
	// Commit the last part statistic.
	if processed > 0 && stats != nil {
		if account == (common.Hash{}) {
			stats.progress(processed, 0, last, common.Hash{})
		} else {
			stats.progress(0, processed, account, last)
		}
	}
	result, success := stop(true, false)
	if !success {
		return common.Hash{}, errors.New("Failed to generate root")
	}
	return result, nil
}

func stackTrieGenerate(db ethdb.KeyValueWriter, in chan trieKV, out chan common.Hash) {
	t := trie.NewStackTrie(db)
	for leaf := range in {
		t.TryUpdate(leaf.key[:], leaf.value)
	}
	var root common.Hash
	if db == nil {
		root = t.Hash()
	} else {
		root, _ = t.Commit()
	}
	out <- root
}
