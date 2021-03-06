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

package cstesting_test

import (
	"fmt"
	"testing"

	"github.com/corestoreio/csfw/util/cstesting"
	"github.com/stretchr/testify/assert"
)

type mockErrorf struct {
	data string
}

func (m *mockErrorf) Errorf(format string, args ...interface{}) {
	m.data = fmt.Sprintf(format, args...)
}

func TestEqualPointers(t *testing.T) {
	var p1 = new(string)
	var p2 = new(string)

	me := &mockErrorf{}
	if have, want := cstesting.EqualPointers(me, p1, p2), false; have != want {
		t.Errorf("Have: %v Want: %v", have, want)
	}
	assert.Regexp(t, "Expecting equal pointers\nWant: 0xc[0-9]+\nHave: 0xc[0-9]+", me.data)
}

func TestContainsCount(t *testing.T) {

	me := &mockErrorf{}
	cstesting.ContainsCount(me, "Hello Gopher", "Rust", 1)
	assert.Exactly(t, "\"Hello Gopher\" should contain \"Rust\" times 1 Have: 0 Want: 1", me.data)
	me.data = ""
	cstesting.ContainsCount(me, "Hello Gopher", "Gopher", 1)
	assert.Empty(t, me.data)
}
