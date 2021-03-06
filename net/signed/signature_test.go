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

package signed_test

import (
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/corestoreio/csfw/net"
	"github.com/corestoreio/csfw/net/signed"
	"github.com/corestoreio/csfw/util/errors"
	"github.com/stretchr/testify/assert"
)

func TestSignature_Write(t *testing.T) {

	w := httptest.NewRecorder()
	sig := signed.Signature{
		KeyID:     "myKeyID",
		Algorithm: "hmac-sha1",
		Signature: []byte(`Hello Gophers`),
	}
	if err := sig.Write(w, hex.EncodeToString); err != nil {
		t.Fatalf("%+v", err)
	}

	const wantSig = `keyId="myKeyID",algorithm="hmac-sha1",signature="48656c6c6f20476f7068657273"`
	if have, want := w.Header().Get(net.ContentSignature), wantSig; have != want {
		t.Errorf("Have: %v Want: %v", have, want)
	}
}

func TestSignature_Parse(t *testing.T) {

	var newReqHeader = func(value string) *http.Request {
		req := httptest.NewRequest("GET", "http://corestore.io", nil)
		req.Header.Set(net.ContentSignature, value)
		return req
	}

	tests := []struct {
		req           *http.Request
		wantKeyID     string
		wantAlgorithm string
		wantSignature []byte
		wantErrBhf    errors.BehaviourFunc
	}{
		{
			newReqHeader(`keyId="myKeyID",algorithm="hmac-sha1",signature="48656c6c6f20476f7068657273"`),
			"myKeyID",
			"hmac-sha1",
			[]byte(`Hello Gophers`),
			nil,
		},
	}
	for _, test := range tests {
		sig := &signed.Signature{}
		haveErr := sig.Parse(test.req, hex.DecodeString)
		if test.wantErrBhf != nil {
			assert.True(t, test.wantErrBhf(haveErr), "Error: %+v", haveErr)
			continue
		}
		assert.Exactly(t, test.wantKeyID, sig.KeyID)
		assert.Exactly(t, test.wantAlgorithm, sig.Algorithm)
		assert.Exactly(t, test.wantSignature, sig.Signature)
		assert.NoError(t, haveErr)
	}
}
