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

package cfgmodel_test

import (
	"net/url"
	"testing"

	"github.com/corestoreio/csfw/config"
	"github.com/corestoreio/csfw/config/cfgmock"
	"github.com/corestoreio/csfw/config/cfgmodel"
	"github.com/corestoreio/csfw/config/cfgpath"
	"github.com/corestoreio/csfw/store/scope"
	"github.com/corestoreio/csfw/util/errors"
	"github.com/stretchr/testify/assert"
)

func TestURLGet(t *testing.T) {

	const pathWebURL = "web/unsecure/url"
	wantPath := cfgpath.MustNewByParts(pathWebURL).Bind(scope.Store, 1)
	b := cfgmodel.NewURL(pathWebURL, cfgmodel.WithFieldFromSectionSlice(configStructure))
	assert.Empty(t, b.Options())

	tests := []struct {
		scpcfg     config.Scoped
		wantErrBhf errors.BehaviourFunc
		wantHash   scope.Hash
		wantVal    interface{}
	}{
		{cfgmock.NewService().NewScoped(0, 1), nil, scope.DefaultHash, `http://john%20doe@corestore.io/?q=go+language#foo&bar`},
		{cfgmock.NewService(
			cfgmock.WithPV(cfgmock.PathValue{
				wantPath.String(): "http://cs.io",
			}),
		).NewScoped(0, 1), nil, scope.NewHash(scope.Store, 1), "http://cs.io"},
		{cfgmock.NewService(
			cfgmock.WithPV(cfgmock.PathValue{
				wantPath.String(): "http://192.168.0.%31/",
			}),
		).NewScoped(0, 1), errors.IsFatal, scope.NewHash(scope.Store, 1), nil},
		{cfgmock.NewService(
			cfgmock.WithPV(cfgmock.PathValue{
				wantPath.String(): "",
			}),
		).NewScoped(0, 1), nil, scope.NewHash(scope.Store, 1), nil},
	}
	for i, test := range tests {
		anURL, haveH, haveErr := b.Get(test.scpcfg)
		assert.Exactly(t, test.wantHash.String(), haveH.String(), "Index %d", i)
		if test.wantErrBhf != nil {
			assert.Nil(t, anURL, "Index %d", i)
			assert.True(t, test.wantErrBhf(haveErr), "Index %d Error %s", i, haveErr)
			continue
		}
		if test.wantVal != nil {
			assert.Exactly(t, test.wantVal, anURL.String(), "Index %d", i)
		} else {
			assert.Nil(t, anURL, "Index %d", i)
		}
		assert.NoError(t, haveErr, "Index %d", i)
	}

}

func TestURLWrite(t *testing.T) {
	const pathWebURL = "web/unsecure/url"
	wantPath := cfgpath.MustNewByParts(pathWebURL).Bind(scope.Store, 1)
	b := cfgmodel.NewURL(pathWebURL, cfgmodel.WithFieldFromSectionSlice(configStructure))

	data, err := url.Parse(`http://john%20doe@corestore.io/?q=go+language#foo&bar`)
	if err != nil {
		t.Fatal(err)
	}

	mw := &cfgmock.Write{}
	assert.NoError(t, b.Write(mw, data, scope.Store, 1))
	assert.Exactly(t, wantPath.String(), mw.ArgPath)
	assert.Exactly(t, `http://john%20doe@corestore.io/?q=go+language#foo&bar`, mw.ArgValue.(string))

	assert.NoError(t, b.Write(mw, nil, scope.Store, 1))
	assert.Exactly(t, wantPath.String(), mw.ArgPath)
	assert.Exactly(t, ``, mw.ArgValue.(string))
}

func TestBaseURLGet(t *testing.T) {
	const pathWebUnsecUrl = "web/unsecure/base_url"
	wantPath := cfgpath.MustNewByParts(pathWebUnsecUrl).Bind(scope.Store, 1)
	b := cfgmodel.NewBaseURL(pathWebUnsecUrl, cfgmodel.WithFieldFromSectionSlice(configStructure))

	assert.Empty(t, b.Options())

	sg, h, err := b.Get(cfgmock.NewService().NewScoped(0, 1))
	if err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, "{{base_url}}", sg)
	assert.Exactly(t, scope.DefaultHash.String(), h.String())

	sg, h, err = b.Get(cfgmock.NewService(
		cfgmock.WithPV(cfgmock.PathValue{
			wantPath.String(): "http://cs.io",
		}),
	).NewScoped(0, 1))
	if err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, "http://cs.io", sg)
	assert.Exactly(t, scope.NewHash(scope.Store, 1).String(), h.String())
}

func TestBaseURLWrite(t *testing.T) {

	const pathWebUnsecUrl = "web/unsecure/base_url"
	wantPath := cfgpath.MustNewByParts(pathWebUnsecUrl).Bind(scope.Store, 1)
	b := cfgmodel.NewBaseURL(pathWebUnsecUrl, cfgmodel.WithFieldFromSectionSlice(configStructure))

	mw := &cfgmock.Write{}
	assert.NoError(t, b.Write(mw, "dude", scope.Store, 1))
	assert.Exactly(t, wantPath.String(), mw.ArgPath)
	assert.Exactly(t, "dude", mw.ArgValue.(string))
}
