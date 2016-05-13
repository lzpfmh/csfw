// Copyright (c) 2014 Olivier Poitrey <rs@dailymotion.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is furnished
// to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package ctxcors

import (
	"net/http"
	"sync"

	"github.com/corestoreio/csfw/config"
	"github.com/corestoreio/csfw/net/httputil"
	"github.com/corestoreio/csfw/net/mw"
	"github.com/corestoreio/csfw/store"
	"github.com/corestoreio/csfw/store/scope"
	"github.com/corestoreio/csfw/util/errors"
)

// Service describes the CrossOriginResourceSharing which is used to create a
// Container Filter that implements CORS. Cross-origin resource sharing (CORS)
// is a mechanism that allows JavaScript on a web page to make XMLHttpRequests
// to another domain, not the domain the JavaScript originated from.
//
// http://en.wikipedia.org/wiki/Cross-origin_resource_sharing
// http://enable-cors.org/server.html
// http://www.html5rocks.com/en/tutorials/cors/#toc-handling-a-not-so-simple-request
type Service struct {

	// optionError use by functional option arguments to indicate that one
	// option has triggered an error and hence the other can options can
	// skip their process.
	optionError error

	// scpOptionFnc optional configuration closure, can be nil. It pulls
	// out the configuration settings during a request and caches the settings in the
	// internal map. ScopedOption requires a config.ScopedGetter
	scpOptionFnc ScopedOptionFunc

	defaultScopeCache scopedConfig

	mu sync.RWMutex
	// scopeCache internal cache of already created token configurations
	// scoped.Hash relates to the website ID.
	// this can become a bottle neck when multiple website IDs supplied by a
	// request try to access the map. we can use the same pattern like in freecache
	// to create a segment of 256 slice items to evenly distribute the lock.
	scopeCache map[scope.Hash]scopedConfig // see freecache to create high concurrent thru put

}

// New creates a new Cors handler with the provided options.
func New(opts ...Option) (*Service, error) {
	s := &Service{
		scopeCache: make(map[scope.Hash]scopedConfig),
	}
	if err := s.Options(WithDefaultConfig(scope.Default, 0)); err != nil {
		return nil, errors.Wrap(err, "[ctxcors] Options WithDefaultConfig")
	}
	if err := s.Options(opts...); err != nil {
		return nil, errors.Wrap(err, "[ctxcors] Options Any Config")
	}
	return s, nil
}

// MustNew same as New() but panics on error. Use only during app start up process.
func MustNew(opts ...Option) *Service {
	c, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return c
}

// Options applies option at creation time or refreshes them.
func (s *Service) Options(opts ...Option) error {
	for _, opt := range opts {
		opt(s)
	}
	if s.optionError != nil {
		return s.optionError
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for h := range s.scopeCache {
		if scp, _ := h.Unpack(); scp > scope.Website {
			return errors.NewNotSupportedf(errServiceUnsupportedScope, h)
		}
	}

	return nil
}

// AddError used by functional options to set an error. The error will only be
// then set if there is not yet an error otherwise it gets discarded. You can
// enable debug logging to find out more.
func (s *Service) AddError(err error) {
	if s.optionError != nil {
		if s.defaultScopeCache.log.IsDebug() {
			s.defaultScopeCache.log.Debug("jwtauth.Service.AddError", "err", err, "skipped", true, "currentError", s.optionError)
		}
		return
	}
	s.optionError = err
}

// WithCORS to be used as a middleware for ctxhttp.Handler.
// The applied configuration
// is used for the all store scopes or if the PkgBackend has been provided then
// on a website specific level.
// Middleware expects to find in a context a store.FromContextProvider().
func (s *Service) WithCORS() mw.Middleware {

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctx := r.Context()

			requestedStore, err := store.FromContextRequestedStore(ctx)
			if err != nil {
				if s.defaultScopeCache.log.IsDebug() {
					s.defaultScopeCache.log.Debug("Service.WithInitTokenAndStore.FromContextProvider", "err", err, "ctx", ctx, "req", r)
				}
				err = errors.Wrap(err, "[jwtauth] FromContextProvider")
				h.ServeHTTP(w, r.WithContext(withContextError(ctx, err)))
				return
			}

			// the scpCfg depends on how you have initialized the storeService during app boot.
			// requestedStore.Website.Config is the reason that all options only support
			// website scope and not group or store scope.
			scpCfg, err := s.configByScopedGetter(requestedStore.Website.Config)
			if err != nil {
				if s.defaultScopeCache.log.IsDebug() {
					s.defaultScopeCache.log.Debug("Service.WithInitTokenAndStore.ConfigByScopedGetter", "err", err, "requestedStore", requestedStore, "ctx", ctx, "req", r)
				}
				err = errors.Wrap(err, "[jwtauth] ConfigByScopedGetter")
				h.ServeHTTP(w, r.WithContext(withContextError(ctx, err)))
				return
			}

			if s.defaultScopeCache.log.IsInfo() {
				s.defaultScopeCache.log.Info("ctxcors.Service.WithCORS.handleActualRequest", "method", r.Method, "scopedConfig", scpCfg)
			}

			if r.Method == httputil.MethodOptions {
				if s.defaultScopeCache.log.IsDebug() {
					s.defaultScopeCache.log.Debug("ctxcors.Service.WithCORS.handlePreflight", "method", r.Method, "OptionsPassthrough", scpCfg.optionsPassthrough)
				}
				scpCfg.handlePreflight(w, r)
				// Preflight requests are standalone and should stop the chain as some other
				// middleware may not handle OPTIONS requests correctly. One typical example
				// is authentication middleware ; OPTIONS requests won't carry authentication
				// headers (see #1)
				if scpCfg.optionsPassthrough {
					h.ServeHTTP(w, r)
				}
				return
			}
			scpCfg.handleActualRequest(w, r)
			h.ServeHTTP(w, r)
		})
	}
}

// configByScopedGetter returns the internal configuration depending on the ScopedGetter.
// Mainly used within the middleware. Exported here to build your own middleware.
// A nil argument falls back to the default scope configuration.
// If you have applied the option WithBackend() the configuration will be pulled out
// one time from the backend service.
func (s *Service) configByScopedGetter(sg config.ScopedGetter) (scopedConfig, error) {

	h := scope.DefaultHash
	if sg != nil {
		h = scope.NewHash(sg.Scope())
	}

	if (s.scpOptionFnc == nil || sg == nil) && h == scope.DefaultHash && s.defaultScopeCache.IsValid() {
		return s.defaultScopeCache, nil
	}

	sc, err := s.getConfigByScopeID(false, h)
	if err == nil {
		// cached entry found and ignore the error because we fall back to
		// default scope at the end of this function.
		return sc, nil
	}

	if s.scpOptionFnc != nil {
		if err := s.Options(s.scpOptionFnc(sg)...); err != nil {
			return scopedConfig{}, errors.Wrap(err, "[jwtauth] Options by scpOptionFnc")
		}
	}

	// after applying the new config try to fetch the new scoped token configuration
	return s.getConfigByScopeID(true, h)
}

func (s *Service) getConfigByScopeID(fallback bool, hash scope.Hash) (scopedConfig, error) {
	var empty scopedConfig
	// requested scope plus ID
	scpCfg, ok := s.getScopedConfig(hash)
	if ok {
		if scpCfg.IsValid() {
			return scpCfg, nil
		}
		return empty, errors.NewNotValidf(errScopedConfigNotValid, hash)
	}

	if fallback {
		// fallback to default scope
		var err error
		if !s.defaultScopeCache.IsValid() {
			err = errConfigNotFound
			if s.defaultScopeCache.log.IsDebug() {
				s.defaultScopeCache.log.Debug("ctxcors.Service.getConfigByScopeID.default", "err", err, "scope", scope.DefaultHash.String(), "fallback", fallback)
			}
		}
		return s.defaultScopeCache, err
	}

	// give up, nothing found
	return empty, errConfigNotFound
}

// getScopedConfig part of lookupScopedConfig and doesn't use a lock because the lock
// has been acquired in lookupScopedConfig()
func (s *Service) getScopedConfig(h scope.Hash) (sc scopedConfig, ok bool) {
	s.mu.RLock()
	sc, ok = s.scopeCache[h]
	s.mu.RUnlock()

	if ok {
		var hasChanges bool
		// do some init stuff ...
		if sc.log == nil {
			sc.log = s.defaultScopeCache.log // copy logger
			hasChanges = true
		}

		if hasChanges {
			s.mu.Lock()
			s.scopeCache[h] = sc
			s.mu.Unlock()
		}
	}
	return sc, ok
}
