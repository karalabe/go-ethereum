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

package snap

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"os"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

type testPeer struct {
	id            string
	test          *testing.T
	remote        *Syncer
	log           log.Logger
	accountTrie   *trie.Trie
	accountValues entrySlice
	storageTries  map[common.Hash]*trie.Trie
	storageValues map[common.Hash]entrySlice

	reqAcc   func(t *testPeer, requestId uint64, root common.Hash, origin common.Hash, cap uint64) error
	cancelCh chan struct{}
}

func (t *testPeer) RequestTrieNodes(id uint64, root common.Hash, paths []TrieNodePathSet, bytes uint64) error {
	panic("implement me")
}

func (t *testPeer) ID() string {
	return t.id
}

func defaultRequestAccountRangeFn(t *testPeer, requestId uint64, root common.Hash, origin common.Hash, cap uint64) error {
	var proofs [][]byte
	var keys []common.Hash
	var vals [][]byte

	for _, entry := range t.accountValues {
		if bytes.Compare(origin[:], entry.k) <= 0 {
			keys = append(keys, common.BytesToHash(entry.k))
			vals = append(vals, entry.v)
		}
	}
	// Unless we send the entire trie, we need to supply proofs
	if true || len(vals) != len(t.accountValues) {
		proof := light.NewNodeSet()
		if err := t.accountTrie.Prove(origin[:], 0, proof); err != nil {
			t.log.Error("Could not prove inexistence of origin", "origin", origin,
				"error", err)
		}
		if len(keys) > 0 {
			lastK := (keys[len(keys)-1])[:]
			if err := t.accountTrie.Prove(lastK, 0, proof); err != nil {
				t.log.Error("Could not prove last item",
					"error", err)
			}
		}
		for _, blob := range proof.NodeList() {
			proofs = append(proofs, blob)
		}
	}
	if err := t.remote.OnAccounts(t, requestId, keys, vals, proofs); err != nil {
		t.log.Error("remote error on delivery", "error", err)
		t.test.Errorf("Remote side rejected our delivery: %v", err)
		t.cancelCh <- struct{}{}
		return err
	}
	return nil
}

func (t *testPeer) RequestAccountRange(requestId uint64, root, origin, limit common.Hash, cap uint64) error {
	t.Log().Info("<- AccRangeReq", "req_id", requestId, "root", root,
		"origin", origin, "limit", limit, "max", cap)

	// Pass the response
	go t.reqAcc(t, requestId, root, origin, cap)
	return nil
}

func (t *testPeer) RequestStorageRanges(requestId uint64, root common.Hash, accounts []common.Hash, origin, limit []byte, max uint64) error {
	t.test.Logf("%v <- StorRangeReq{id:%d, root: %x, account[0]: %x, origin: %x, limit %x max: %d}",
		t.id, requestId, root[:8], accounts[0], origin, limit, max)
	// Pass the response
	go func() {
		var hashes [][]common.Hash
		var slots [][][]byte
		var proofs [][]byte
		for _, account := range accounts {
			var keys []common.Hash
			var vals [][]byte
			for _, entry := range t.storageValues[account] {
				if bytes.Compare(origin[:], entry.k) <= 0 {
					keys = append(keys, common.BytesToHash(entry.k))
					vals = append(vals, entry.v)
				}
			}
			// If we're sending _nothing_, we need to prove the inexistence of the
			// requested element
			proof := light.NewNodeSet()
			if err := t.accountTrie.Prove(origin[:], 0, proof); err != nil {
				t.log.Error("Could not prove inexistence of origin", "origin", origin,
					"error", err)
			}
			if len(keys) > 0 {
				lastK := (keys[len(keys)-1])[:]
				if err := t.accountTrie.Prove(lastK, 0, proof); err != nil {
					t.log.Error("Could not prove last item",
						"error", err)
				}
			}
			for _, blob := range proof.NodeList() {
				proofs = append(proofs, blob)
			}
			hashes = append(hashes, keys)
			slots = append(slots, vals)
		}
		if err := t.remote.OnStorage(t, requestId, hashes, slots, proofs); err != nil {
			t.log.Error("remote error on delivery", "error", err)
			t.test.Errorf("Remote side rejected our delivery: %v", err)
			t.cancelCh <- struct{}{}
		}
	}()
	return nil
}

func (t *testPeer) RequestByteCodes(requestId uint64, hashes []common.Hash, max uint64) error {
	t.test.Logf("%v <- CodeReq{id:%d, #hashes: %d, max: %d}",
		t.id, requestId, len(hashes), max)
	panic("implement me")
}

func (t *testPeer) Log() log.Logger {
	return t.log
}

func newTestPeer(id string, t *testing.T, syncer *Syncer, cancelCh chan struct{}) *testPeer {
	peer := &testPeer{
		id:       id,
		test:     t,
		remote:   syncer,
		log:      log.New("id", id),
		reqAcc:   defaultRequestAccountRangeFn,
		cancelCh: cancelCh,
	}
	stdoutHandler := log.StreamHandler(os.Stdout, log.TerminalFormat(true))
	peer.log.SetHandler(stdoutHandler)
	return peer

}

type snapTester struct {
	//genesis *types.Block   // Genesis blocks used by the tester and peers
	stateDb ethdb.Database // Database used by the tester for syncing from peers
}

// TestSync tests a basic sync
func TestSync(t *testing.T) {
	stateDb := rawdb.NewMemoryDatabase()
	syncer := NewSyncer(stateDb, trie.NewSyncBloom(1, stateDb))
	sourceAccountTrie, elems := nonRandomAccountTrieNoStorage(100)
	cancel := make(chan struct{})
	source := newTestPeer("source", t, syncer, cancel)
	source.accountTrie = sourceAccountTrie
	source.accountValues = elems
	syncer.Register(source)
	if err := syncer.Sync(sourceAccountTrie.Hash(), cancel); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
}

// TestSyncWithStorage tests  basic sync using accounts + storage
func TestSyncWithStorage(t *testing.T) {
	stateDb := rawdb.NewMemoryDatabase()
	syncer := NewSyncer(stateDb, trie.NewSyncBloom(1, stateDb))
	sourceAccountTrie, elems, storageTries, storageElems := nonRandomAccountTrieWithStorage(100)
	cancel := make(chan struct{})
	source := newTestPeer("source", t, syncer, cancel)
	source.accountTrie = sourceAccountTrie
	source.accountValues = elems
	source.storageTries = storageTries
	source.storageValues = storageElems

	syncer.Register(source)
	if err := syncer.Sync(sourceAccountTrie.Hash(), cancel); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
}

// TestSyncBloatedProof tests a scenario where we provide only _one_ value, but
// also ship the entire trie inside the proof. If the attack is successfull,
// the remote side does not do any follow-up requests
func TestSyncBloatedProof(t *testing.T) {
	stateDb := rawdb.NewMemoryDatabase()
	syncer := NewSyncer(stateDb, trie.NewSyncBloom(1, stateDb))
	sourceAccountTrie, elems := nonRandomAccountTrieNoStorage(100)
	cancel := make(chan struct{})
	source := newTestPeer("source", t, syncer, cancel)
	source.accountTrie = sourceAccountTrie
	source.accountValues = elems
	syncer.Register(source)

	source.reqAcc = func(t *testPeer, requestId uint64, root common.Hash, origin common.Hash, cap uint64) error {
		var proofs [][]byte
		var keys []common.Hash
		var vals [][]byte
		proof := light.NewNodeSet()
		if err := t.accountTrie.Prove(origin[:], 0, proof); err != nil {
			t.log.Error("Could not prove origin", "origin", origin, "error", err)
		}

		for _, entry := range t.accountValues {
			if err := t.accountTrie.Prove(entry.k, 0, proof); err != nil {
				t.log.Error("Could not prove item", "error", err)
			}
		}
		{ // First item
			elem := t.accountValues[0]
			keys = append(keys, common.BytesToHash(elem.k))
			vals = append(vals, elem.v)
		}
		//{// Last item
		//	elem := t.accountValues[len(t.accountValues)-1]
		//	keys = append(keys, common.BytesToHash(elem.k))
		//	vals = append(vals, elem.v)
		//}

		for _, blob := range proof.NodeList() {
			proofs = append(proofs, blob)
		}
		if err := t.remote.OnAccounts(t, requestId, keys, vals, proofs); err != nil {
			t.log.Error("remote error on delivery", "error", err)
		}
		return nil
	}
	if err := syncer.Sync(sourceAccountTrie.Hash(), cancel); err != nil {
		t.Logf("sync failed: %v", err)
	} else {
		t.Fatal("Sync done, without any error")
	}

}

type kv struct {
	k, v []byte
	t    bool
}

// Some helpers for sorting
type entrySlice []*kv

func (p entrySlice) Len() int           { return len(p) }
func (p entrySlice) Less(i, j int) bool { return bytes.Compare(p[i].k, p[j].k) < 0 }
func (p entrySlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// nonRandomAccountTrieNoStorage spits out a trie, along with the leafs
func nonRandomAccountTrieNoStorage(n int) (*trie.Trie, entrySlice) {
	trie := new(trie.Trie)
	var entries entrySlice
	for i := uint64(0); i < uint64(n); i++ {
		ac := state.Account{
			Nonce:    i,
			Balance:  big.NewInt(int64(i)),
			Root:     emptyRoot,
			CodeHash: emptyCode[:],
		}
		value, _ := rlp.EncodeToBytes(ac)
		key := make([]byte, 32)
		binary.LittleEndian.PutUint64(key, i)
		elem := &kv{key, value, false}
		trie.Update(elem.k, elem.v)
		entries = append(entries, elem)
	}
	sort.Sort(entries)
	return trie, entries
}

// nonRandomAccountTrieWithStorage spits out a trie, along with the leafs
func nonRandomAccountTrieWithStorage(n int) (*trie.Trie, entrySlice,
	map[common.Hash]*trie.Trie, map[common.Hash]entrySlice) {
	accTrie := new(trie.Trie)
	var entries entrySlice

	stTrie, stEntries := nonRandomStorageTrie(3)

	var storageTries = make(map[common.Hash]*trie.Trie)
	var storageEntries = make(map[common.Hash]entrySlice)

	stRoot := stTrie.Hash()
	for i := uint64(0); i < uint64(n); i++ {
		ac := state.Account{
			Nonce:    i,
			Balance:  big.NewInt(int64(i)),
			Root:     stRoot,
			CodeHash: emptyCode[:],
		}
		value, _ := rlp.EncodeToBytes(ac)
		key := make([]byte, 32)
		binary.LittleEndian.PutUint64(key, i)
		elem := &kv{key, value, false}
		accTrie.Update(elem.k, elem.v)
		entries = append(entries, elem)

		// we reuse the same one for all accounts
		storageTries[common.BytesToHash(key)] = stTrie
		storageEntries[common.BytesToHash(key)] = stEntries
	}
	return accTrie, entries, storageTries, storageEntries
}

// nonRandomStorageTrie spits out a trie, along with the leafs
func nonRandomStorageTrie(n int) (*trie.Trie, entrySlice) {
	trie := new(trie.Trie)
	var entries entrySlice
	for i := uint64(0); i < uint64(n); i++ {
		// store 'i' at slot 'i'
		slotValue := make([]byte, 32)
		binary.LittleEndian.PutUint64(slotValue, i)
		rlpSlotValue, _ := rlp.EncodeToBytes(common.TrimLeftZeroes(slotValue[:]))

		slotKey := make([]byte, 32)
		binary.LittleEndian.PutUint64(slotKey, i)
		key := crypto.Keccak256Hash(slotKey[:])

		elem := &kv{key[:], rlpSlotValue, false}
		trie.Update(elem.k, elem.v)
		entries = append(entries, elem)
	}
	sort.Sort(entries)
	return trie, entries
}
