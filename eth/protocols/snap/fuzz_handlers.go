package snap

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	fuzz "github.com/google/gofuzz"
)

func getChain() *core.BlockChain {
	db := rawdb.NewMemoryDatabase()
	//gspec := core.DefaultYoloV2GenesisBlock()
	ga := make(core.GenesisAlloc, 1000)
	var a = make([]byte, 20)
	for i := 0; i < 1000; i++ {
		binary.LittleEndian.PutUint64(a, uint64(i+0xff))
		ga[common.BytesToAddress(a)] = core.GenesisAccount{Balance: big.NewInt(int64(i))}
	}
	gspec := core.Genesis{
		Config: params.TestChainConfig,
		Alloc:  ga,
	}
	genesis := gspec.MustCommit(db)
	blocks, _ := core.GenerateChain(gspec.Config, genesis, ethash.NewFaker(), db, 2,
		func(i int, gen *core.BlockGen) {})
	cacheConf := &core.CacheConfig{
		TrieCleanLimit:      0,
		TrieDirtyLimit:      0,
		TrieTimeLimit:       5 * time.Minute,
		TrieCleanNoPrefetch: true,
		TrieCleanRejournal:  0,
		SnapshotLimit:       100,
		SnapshotWait:        true,
	}
	bc, _ := core.NewBlockChain(db, cacheConf, gspec.Config, ethash.NewFaker(), vm.Config{}, nil, nil)
	if _, err := bc.InsertChain(blocks); err != nil {
		panic(err)
	}
	return bc
}

type dummyBackend struct {
	chain *core.BlockChain
}

func (d *dummyBackend) Chain() *core.BlockChain {
	return d.chain
}

func (d *dummyBackend) RunPeer(peer *Peer, handler Handler) error {
	return nil
}

func (d *dummyBackend) PeerInfo(id enode.ID) interface{} {
	return "Oy vey"
}
func (d *dummyBackend) Handle(peer *Peer, packet Packet) error {
	return nil
}

type dummyRW struct {
	code       uint64
	data       []byte
	writeCount int
}

func (d *dummyRW) ReadMsg() (p2p.Msg, error) {
	return p2p.Msg{
		Code:       d.code,
		Payload:    bytes.NewReader(d.data),
		ReceivedAt: time.Now(),
		Size:       uint32(len(d.data)),
	}, nil
}

func (d *dummyRW) WriteMsg(msg p2p.Msg) error {
	d.writeCount++
	return nil
}

func doFuzz(input []byte, obj interface{}, code int) int {
	if len(input) > 1024*4 {
		return 1
	}
	bc := getChain()
	defer bc.Stop()
	backend := &dummyBackend{bc}
	fuzz.NewFromGoFuzz(input).Fuzz(obj)
	data, _ := rlp.EncodeToBytes(obj)
	cli := &dummyRW{
		code: uint64(code),
		data: data,
	}
	peer := &Peer{id: "lalal", version: 65, rw: cli}
	// It should only error on decoding failures, which should not happen
	if err := handleMessage(backend, peer); err != nil {
		panic(err)
	}
	// A response is always to be expected
	if cli.writeCount != 1 {
		panic(fmt.Sprintf("Expected 1 response, got %d", cli.writeCount))
	}
	return 0
}

func FuzzA(input []byte) int { return doFuzz(input, &GetAccountRangePacket{}, GetAccountRangeMsg) }
func FuzzB(input []byte) int { return doFuzz(input, &GetStorageRangesPacket{}, GetStorageRangesMsg) }
func FuzzC(input []byte) int { return doFuzz(input, &GetByteCodesPacket{}, GetByteCodesMsg) }
func FuzzD(input []byte) int { return doFuzz(input, &GetTrieNodesPacket{}, GetTrieNodesMsg) }
