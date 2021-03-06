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

package memstore

import (
	"github.com/corestoreio/csfw/log"
	"github.com/corestoreio/csfw/net/ratelimit"
	"github.com/corestoreio/csfw/store/scope"
	"github.com/corestoreio/csfw/util/errors"
	"gopkg.in/throttled/throttled.v2/store/memstore"
)

// WithGCRA creates a memory based GCRA rate limiter.
// Duration: (s second,i minute,h hour,d day).
// GCRA => https://en.wikipedia.org/wiki/Generic_cell_rate_algorithm
// This function implements a debug log.
func WithGCRA(scp scope.Scope, id int64, maxKeys int, duration rune, requests, burst int) ratelimit.Option {
	return func(s *ratelimit.Service) error {
		rlStore, err := memstore.New(maxKeys)
		if err != nil {
			return errors.NewFatalf("[memstore] memstore.New MaxKeys(%d): %s", maxKeys, err)
		}
		if s.Log.IsDebug() {
			s.Log.Debug("ratelimit.memstore.WithGCRA",
				log.Stringer("scope", scp),
				log.Int64("scope_id", id),
				log.Int("max_keys", maxKeys),
				log.String("duration", string(duration)),
				log.Int("requests", requests),
				log.Int("burst", burst),
			)
		}
		return ratelimit.WithGCRAStore(scp, id, rlStore, duration, requests, burst)(s)
	}
}
