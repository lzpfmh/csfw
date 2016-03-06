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

package model

import (
	"github.com/corestoreio/csfw/config"
	"github.com/corestoreio/csfw/store/scope"
)

// Encryptor functions needed for encryption and decryption of
// string values. For example implements M1 and M2 encryption key functions.
type Encryptor interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}

// WithEncryptor sets the functions for reading and writing encrypted data
// to the configuration service
func WithEncryptor(e Encryptor) Option {
	return func(b *optionBox) Option {
		prev := b.Obscure.Encryptor
		b.Obscure.Encryptor = e
		return WithEncryptor(prev)
	}
}

// Obscure backend model for handling sensible values
type Obscure struct {
	Str
	Encryptor
}

// NewObscure creates a new Obscure with validation checks when writing values.
func NewObscure(path string, opts ...Option) Obscure {
	ret := Obscure{
		Str: NewStr(path),
	}
	(&ret).Option(opts...)
	return ret
}

// Option sets the options and returns the last set previous option
func (p *Obscure) Option(opts ...Option) (previous Option) {
	ob := &optionBox{
		baseValue: &p.baseValue,
		Obscure:   p,
	}
	for _, o := range opts {
		previous = o(ob)
	}
	p = ob.Obscure
	p.baseValue = *ob.baseValue
	return
}

// Get returns an encrypted value decrypted.
func (p Obscure) Get(sg config.ScopedGetter) (string, error) {
	s, err := p.Str.Get(sg)
	if err != nil {
		return "", err
	}
	return p.Decrypt(s)
}

// Write writes a raw value encrypted.
func (p Obscure) Write(w config.Writer, v string, s scope.Scope, scopeID int64) (err error) {
	v, err = p.Encrypt(v)
	if err != nil {
		return err
	}
	return p.Str.Write(w, v, s, scopeID)
}