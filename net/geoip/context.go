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

package geoip

import (
	"context"
	"net/http"

	"github.com/corestoreio/csfw/util/errors"
)

// keyctxCountry type is unexported to prevent collisions with context keys
// defined in other packages.
type keyctxCountry struct{}

// ctxCountryWrapper to prevent too much calls to runtime.convT2*
type ctxCountryWrapper struct {
	c   *Country
	err error
}

// WithContextCountry creates a new context with geoip.Country attached.
func withContextCountry(ctx context.Context, c *Country) context.Context {
	return context.WithValue(ctx, keyctxCountry{}, ctxCountryWrapper{c: c})
}

// wrapContextError creates a new context with an error attached.
func wrapContextError(r *http.Request, c *Country, err error) *http.Request {
	return r.WithContext(context.WithValue(
		r.Context(),
		keyctxCountry{},
		ctxCountryWrapper{c: c, err: err},
	))
}

// FromContextCountry returns the geoip.Country in ctx if it exists or an error.
// The error has been previously set by WithContextError. An error can be for
// the first request, with a new IP address to fill the cache, of behaviour
// NotValid but all subsequent requests are of behaviour NotFound.
func FromContextCountry(ctx context.Context) (*Country, error) {
	wrp, ok := ctx.Value(keyctxCountry{}).(ctxCountryWrapper)
	if !ok {
		return nil, errors.NewNotFoundf(errContextCountryNotFound)
	}
	if wrp.err != nil {
		return nil, wrp.err
	}
	return wrp.c, nil
}
