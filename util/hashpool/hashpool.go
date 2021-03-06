// Copyright 2015-2016, Cyrill @ Schumacher.fm and the CoreStore contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hashpool

import (
	"encoding/hex"
	"hash"
	"sync"

	"github.com/corestoreio/csfw/util/bufferpool"
)

// Tank implements a sync.Pool for hash.Hash
type Tank struct {
	p *sync.Pool
	// BufferSize used in SumBase64() to append the hashed data to. Default 1024.
	BufferSize int
}

// Get returns type safe a hash.
func (t Tank) Get() hash.Hash {
	return t.p.Get().(hash.Hash)
}

// Sum calculates the hash of data and appends the current hash to appendTo and
// returns the resulting slice. It does not change the underlying hash state. It
// fetches a hash from the pool and returns it after writing the sum. No need to
// call Get() and Put().
func (t Tank) Sum(data, appendTo []byte) []byte {
	h := t.Get()
	defer t.Put(h)
	_, _ = h.Write(data)
	return h.Sum(appendTo)
}

// SumHex writes the hashed data into the hex encoder.
func (t Tank) SumHex(data []byte) string {
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)
	bs := 1024
	if t.BufferSize > 0 {
		bs = t.BufferSize
	}
	buf.Grow(bs)
	tmpBuf := t.Sum(data, buf.Bytes())
	buf.Reset()
	_, _ = buf.Write(tmpBuf)
	return hex.EncodeToString(buf.Bytes())
}

// Put empties the hash and returns it back to the pool.
//
//		hp := New(func() hash.Hash { return fnv.New64() })
//		hsh := hp.Get()
//		defer hp.Put(hsh)
//		// your code
//		return hsh.Sum([]byte{})
//
func (t Tank) Put(h hash.Hash) {
	h.Reset()
	t.p.Put(h)
}

// New instantiates a new hash pool with a custom pre-allocated hash.Hash.
func New(h func() hash.Hash) Tank {
	return Tank{
		p: &sync.Pool{
			New: func() interface{} {
				nh := h()
				nh.Reset()
				return nh
			},
		},
	}
}

// New32 instantiates a new hash pool with a custom pre-allocated hash.Hash32.
func New32(h func() hash.Hash32) Tank {
	return New(func() hash.Hash { return h() })
}

// New64 instantiates a new hash pool with a custom pre-allocated hash.Hash64.
func New64(h func() hash.Hash64) Tank {
	return New(func() hash.Hash { return h() })
}
