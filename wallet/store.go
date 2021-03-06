// Copyright (c) 2017 The Alvalor Authors
//
// This file is part of Alvalor.
//
// Alvalor is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Alvalor is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.

package wallet

import (
	"bytes"

	"github.com/pkg/errors"
	argon2 "github.com/tvdburgt/go-argon2"
	"github.com/alvalor/alvalor-go/futhark"
)

// Salt variable.
var salt = []byte{
	0x53, 0x78, 0x3e, 0x4c,
	0x94, 0x78, 0x59, 0x18,
	0x8a, 0x9b, 0x31, 0xe7,
	0x4d, 0xed, 0x1d, 0x29,
}

// Store struct.
type Store struct {
	root []byte
	ctx  argon2.Context
}

// NewStore function.
func NewStore(seed []byte) (*Store, error) {
	s := &Store{
		ctx: argon2.Context{
			Iterations:  3,
			Memory:      1 << 16,
			Parallelism: 4,
			HashLen:     96,
			Mode:        argon2.ModeArgon2i,
			Version:     argon2.Version13,
		},
	}
	root, err := s.generate(seed, salt)
	if err != nil {
		return nil, errors.Wrap(err, "could not generate root key")
	}
	s.root = root
	return s, nil
}

// generate method.
func (s *Store) generate(input []byte, code []byte) ([]byte, error) {
	hash, err := argon2.Hash(&s.ctx, input, code)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute hash")
	}
	_, priv, err := futhark.GenerateKey(bytes.NewReader(hash[:64]))
	if err != nil {
		return nil, errors.Wrap(err, "could not generate private key")
	}
	key := append([]byte(priv), hash[64:]...)
	return key, nil
}

// derive method.
func (s *Store) derive(key []byte, index []byte) ([]byte, error) {
	data := append(key[:64], index...)
	key, err := s.generate(data, key[64:])
	if err != nil {
		return nil, errors.Wrap(err, "could not derive level one key")
	}
	return key, nil
}

// Key method.
func (s *Store) Key(levels [][]byte) ([]byte, error) {
	var err error
	key := s.root
	for i, level := range levels {
		key, err = s.derive(key, level)
		if err != nil {
			return nil, errors.Wrapf(err, "could not derive level %v key", i)
		}
	}
	return key, nil
}
