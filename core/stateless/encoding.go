// Copyright 2024 The go-ethereum Authors
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

package stateless

import (
	"bytes"
	"io"
	"slices"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// EncodeRLP serializes a witness as RLP.
func (w *Witness) EncodeRLP(wr io.Writer) error {
	ext := &extwitness{
		Block:   w.Block,
		Headers: w.Headers,
	}
	ext.Codes = make([][]byte, 0, len(w.Codes))
	for code := range w.Codes {
		ext.Codes = append(ext.Codes, []byte(code))
	}
	slices.SortFunc(ext.Codes, bytes.Compare)

	ext.State = make([][]byte, 0, len(w.State))
	for node := range w.State {
		ext.State = append(ext.State, []byte(node))
	}
	slices.SortFunc(ext.State, bytes.Compare)

	return rlp.Encode(wr, ext)
}

// DecodeRLP decodes a witness from RLP.
func (w *Witness) DecodeRLP(s *rlp.Stream) error {
	var ew extwitness
	if err := s.Decode(&ew); err != nil {
		return err
	}
	w.Block, w.Headers = ew.Block, ew.Headers

	w.Codes = make(map[string]struct{}, len(ew.Codes))
	for _, code := range ew.Codes {
		w.Codes[string(code)] = struct{}{}
	}
	w.State = make(map[string]struct{}, len(ew.State))
	for _, node := range ew.State {
		w.State[string(node)] = struct{}{}
	}
	return nil
}

// extwitness is a witness RLP encoding for transferring across clients.
type extwitness struct {
	Block   *types.Block
	Headers []*types.Header
	Codes   [][]byte
	State   [][]byte
}
