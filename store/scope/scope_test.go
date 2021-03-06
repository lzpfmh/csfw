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

package scope

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScopeBits(t *testing.T) {

	const (
		scope1 Scope = iota + 1
		scope2
		scope3
		scope4
		scope5
	)

	tests := []struct {
		have    []Scope
		want    Scope
		notWant Scope
		human   []string
		string
	}{
		{[]Scope{scope1, scope2}, scope2, scope3, []string{"Default", "Website"}, "Default,Website"},
		{[]Scope{scope3, scope4}, scope3, scope2, []string{"Group", "Store"}, "Group,Store"},
		{[]Scope{scope4, scope5}, scope4, scope2, []string{"Store"}, "Store"},
	}

	for _, test := range tests {
		var b = Perm(0).Set(test.have...)
		if b.Has(test.want) == false {
			t.Errorf("%d should contain %d", b, test.want)
		}
		if b.Has(test.notWant) {
			t.Errorf("%d should not contain %d", b, test.notWant)
		}
		assert.EqualValues(t, test.human, b.Human())
		assert.EqualValues(t, test.string, b.String())
	}
}

func TestFromString(t *testing.T) {

	tests := []struct {
		have string
		want Scope
	}{
		{"asdasd", Default},
		{strDefault, Default},
		{strWebsites, Website},
		{strStores, Store},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, FromString(test.have))
	}
}

func TestFromScope(t *testing.T) {

	tests := []struct {
		have Scope
		want StrScope
	}{
		{Default, StrDefault},
		{Absent, StrDefault},
		{Group, StrDefault},
		{Website, StrWebsites},
		{Store, StrStores},
	}
	for _, test := range tests {
		assert.Exactly(t, test.want, FromScope(test.have))
		assert.Exactly(t, test.want.String(), test.have.StrScope())
	}
}

func TestStrScope(t *testing.T) {

	assert.Equal(t, strDefault, StrDefault.String())
	assert.Equal(t, strWebsites, StrWebsites.String())
	assert.Equal(t, strStores, StrStores.String())

	assert.Exactly(t, Default, StrDefault.Scope())
	assert.Exactly(t, Website, StrWebsites.Scope())
	assert.Exactly(t, Store, StrStores.Scope())
}

func TestValid(t *testing.T) {

	tests := []struct {
		have string
		want bool
	}{
		{"Rust", false},
		{"default", true},
		{"website", false},
		{"websites", true},
		{"stores", true},
		{"Stores", false},
	}
	for i, test := range tests {
		assert.Exactly(t, test.want, Valid(test.have), "Index %d", i)
	}
}

func TestFromBytes(t *testing.T) {
	tests := []struct {
		have []byte
		want Scope
	}{
		{[]byte("asdasd"), Default},
		{[]byte(strDefault), Default},
		{[]byte(strWebsites), Website},
		{[]byte(strStores), Store},
	}
	for _, test := range tests {
		assert.Exactly(t, test.want, FromBytes(test.have))
	}
}

func TestValidBytes(t *testing.T) {
	tests := []struct {
		have []byte
		want bool
	}{
		{[]byte("Rust"), false},
		{[]byte("default"), true},
		{[]byte("website"), false},
		{[]byte("websites"), true},
		{[]byte("stores"), true},
		{[]byte("Stores"), false},
	}
	for i, test := range tests {
		assert.Exactly(t, test.want, ValidBytes(test.have), "Index %d", i)
	}
}

func TestStrScopeBytes(t *testing.T) {
	tests := []struct {
		id Scope
	}{
		{Default},
		{Website},
		{Store},
		{44},
	}
	for i, test := range tests {
		assert.Exactly(t, test.id.StrScope(), string(test.id.Bytes()), "Index %d", i)
	}
}

func TestValidParent(t *testing.T) {
	tests := []struct {
		c    Scope
		p    Scope
		want bool
	}{
		{Default, Default, true},
		{Website, Default, true},
		{Store, Website, true},
		{Default, Website, false},
		{Absent, Absent, false},
		{Absent, Default, false},
		{Default, Absent, false},
	}
	for i, test := range tests {
		if have, want := ValidParent(test.c, test.p), test.want; have != want {
			t.Errorf("(%d) Have: %v Want: %v", i, have, want)
		}
	}
}
