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
	"net/http"

	"github.com/corestoreio/csfw/store"
	"github.com/corestoreio/csfw/store/scope"
	"github.com/corestoreio/csfw/util"
	"github.com/corestoreio/csfw/util/errors"
)

// scopedConfig private internal scoped based configuration
type scopedConfig struct {
	// useDefault if true uses the default configuration and all other fields are
	// empty.
	useDefault bool
	// lastErr used during selecting the config from the scopeCache map.
	lastErr error
	// scopeHash defines the scope to which this configuration is bound to.
	scopeHash scope.Hash

	// AllowedCountries a slice which contains all allowed countries. An
	// incoming request for a scope checks if the country for an IP is contained
	// within this slice. Empty slice means that all countries are allowed.
	allowedCountries []string
	// IsAllowedFunc checks in middleware WithIsCountryAllowedByIP if the country is
	// allowed to process the request.
	IsAllowedFunc // func(s *store.Store, c *Country, allowedCountries []string) error

	// alternativeHandler if ip/country is denied we call this handler
	alternativeHandler http.Handler
}

func defaultScopedConfig(h scope.Hash) scopedConfig {
	return scopedConfig{
		scopeHash: h,
		IsAllowedFunc: func(_ *store.Store, c *Country, allowedCountries []string) error {
			var ac util.StringSlice = allowedCountries
			if ac.Contains(c.Country.IsoCode) {
				return nil
			}
			return errors.NewUnauthorizedf(errUnAuthorizedCountry, c.Country.IsoCode, allowedCountries)
		},
		alternativeHandler: DefaultAlternativeHandler,
	}
}

// IsValid a configuration for a scope is only then valid when the Key has been
// supplied, a non-nil signing method and a non-nil Verifier.
func (sc scopedConfig) isValid() error {
	if sc.lastErr != nil {
		return errors.Wrap(sc.lastErr, "[geoip] scopedConfig.isValid as an lastErr")
	}

	if sc.scopeHash == 0 || sc.IsAllowedFunc == nil ||
		sc.alternativeHandler == nil {
		return errors.NewNotValidf(errScopedConfigNotValid, sc.scopeHash, sc.IsAllowedFunc == nil, sc.alternativeHandler == nil)
	}
	return nil
}

func (sc scopedConfig) checkAllow(reqSt *store.Store, c *Country) error {
	if len(sc.allowedCountries) == 0 {
		return nil
	}
	return sc.IsAllowedFunc(reqSt, c, sc.allowedCountries)
}
