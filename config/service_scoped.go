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

package config

import (
	"time"

	"github.com/corestoreio/csfw/config/cfgpath"
	"github.com/corestoreio/csfw/store/scope"
	"github.com/corestoreio/csfw/util/errors"
)

// think about that segregation
//type ScopedStringer interface {
//  Parent() (scope.Scope, int64)
//	scope.Scoper
//	Bind(scope.Scope) ScopedGetter
//	String(r cfgpath.Route, s ...scope.Scope) (string, error)
//}
// and so on ...

// Scoped is equal to Getter but not an interface and the underlying
// implementation takes care of providing the correct scope: default, website or
// store and bubbling up the scope chain from store -> website -> default if a
// value won't get found in the desired scope. The cfgpath.Route for each
// primitive type represents always a path like "section/group/element" without
// the scope string and scope ID.
//
// To restrict bubbling up you can provide a second argument scope.Scope. You
// can restrict a configuration path to be only used with the default, website
// or store scope. See the examples. This second argument will mainly be used by
// the cfgmodel package to use a defined scope in a config.Structure. If you
// access the ScopedGetter from a store.Store, store.Website type the second
// argument must already be internally pre-filled.
//
// WebsiteID and StoreID must be in a relation like enforced in the database
// tables via foreign keys. Empty storeID triggers the website scope. Empty
// websiteID and empty storeID are triggering the default scope.
//
// You can use the function NewScoped() to create a new object but not
// mandatory. Returned error has mostly the behaviour of not found.
type Scoped struct {
	// Root holds the main functions for retrieving values by paths from the
	// storage.
	Root      Getter
	WebsiteID int64
	StoreID   int64
}

// NewScopedService instantiates a ScopedGetter implementation.  Getter specifies the
// root Getter which does not know about any scope.
func NewScoped(r Getter, websiteID, storeID int64) Scoped {
	return Scoped{
		Root:      r,
		WebsiteID: websiteID,
		StoreID:   storeID,
	}
}

// Parent tells you the parent underlying scope and its ID. Store falls back to
// website and website falls back to default.
func (ss Scoped) Parent() (scope.Scope, int64) {
	if ss.StoreID > 0 {
		return scope.Website, ss.WebsiteID
	}
	return scope.Default, 0
}

// Scope tells you the current underlying scope and its ID.
func (ss Scoped) Scope() (scope.Scope, int64) {
	if ss.StoreID > 0 {
		return scope.Store, ss.StoreID
	}
	if ss.WebsiteID > 0 {
		return scope.Website, ss.WebsiteID
	}
	return scope.Default, 0
}

func (ss Scoped) isAllowedStore(s ...scope.Scope) bool {
	scp, _ := ss.Scope()
	if len(s) > 0 && s[0] > scope.Absent {
		scp = s[0]
	}
	return ss.StoreID > 0 && scope.PermStoreReverse.Has(scp)
}

func (ss Scoped) isAllowedWebsite(s ...scope.Scope) bool {
	scp, _ := ss.Scope()
	if len(s) > 0 && s[0] > scope.Absent {
		scp = s[0]
	}
	return ss.WebsiteID > 0 && scope.PermWebsiteReverse.Has(scp)
}

// Byte traverses through the scopes store->website->default to find
// a matching byte slice value.
func (ss Scoped) Byte(r cfgpath.Route, s ...scope.Scope) ([]byte, scope.Hash, error) {
	// fallback to next parent scope if value does not exists
	p, err := cfgpath.New(r)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "[config] Byte. Route %q", r)
	}

	if ss.isAllowedStore(s...) {
		p = p.BindStore(ss.StoreID)
		v, err := ss.Root.Byte(p)
		if !errors.IsNotFound(err) || err == nil {
			// value found or err is not a NotFound error
			return v, p.ScopeHash, err
		}
	}
	if ss.isAllowedWebsite(s...) {
		p = p.BindWebsite(ss.WebsiteID)
		v, err := ss.Root.Byte(p)
		if !errors.IsNotFound(err) || err == nil {
			// value found or err is not a NotFound error
			return v, p.ScopeHash, err
		}
	}
	p.ScopeHash = scope.DefaultHash
	v, err := ss.Root.Byte(p)
	return v, scope.DefaultHash, err
}

// String traverses through the scopes store->website->default to find
// a matching string value.
func (ss Scoped) String(r cfgpath.Route, s ...scope.Scope) (string, scope.Hash, error) {
	// fallback to next parent scope if value does not exists
	p, err := cfgpath.New(r)
	if err != nil {
		return "", 0, errors.Wrapf(err, "[config] String. Route %q", r)
	}

	if ss.isAllowedStore(s...) {
		p = p.BindStore(ss.StoreID)
		v, err := ss.Root.String(p)
		if !errors.IsNotFound(err) || err == nil {
			// value found or err is not a NotFound error
			return v, p.ScopeHash, err
		}
	}

	if ss.isAllowedWebsite(s...) {
		p = p.BindWebsite(ss.WebsiteID)
		v, err := ss.Root.String(p)
		if !errors.IsNotFound(err) || err == nil {
			// value found or err is not a NotFound error
			return v, p.ScopeHash, err
		}
	}
	p.ScopeHash = scope.DefaultHash
	v, err := ss.Root.String(p)
	return v, scope.DefaultHash, err
}

// Bool traverses through the scopes store->website->default to find
// a matching bool value.
func (ss Scoped) Bool(r cfgpath.Route, s ...scope.Scope) (bool, scope.Hash, error) {
	// fallback to next parent scope if value does not exists
	p, err := cfgpath.New(r)
	if err != nil {
		return false, 0, errors.Wrapf(err, "[config] Bool. Route %q", r)
	}

	if ss.isAllowedStore(s...) {
		p = p.BindStore(ss.StoreID)
		v, err := ss.Root.Bool(p)
		if !errors.IsNotFound(err) || err == nil {
			// value found or err is not a NotFound error
			return v, p.ScopeHash, err
		}
	} // if not found in store scope go to website scope

	if ss.isAllowedWebsite(s...) {
		p = p.BindWebsite(ss.WebsiteID)
		v, err := ss.Root.Bool(p)
		if !errors.IsNotFound(err) || err == nil {
			// value found or err is not a NotFound error
			return v, p.ScopeHash, err
		}
	} // if not found in website scope go to default scope
	p.ScopeHash = scope.DefaultHash
	v, err := ss.Root.Bool(p)
	return v, scope.DefaultHash, err
}

// Float64 traverses through the scopes store->website->default to find
// a matching float64 value.
func (ss Scoped) Float64(r cfgpath.Route, s ...scope.Scope) (float64, scope.Hash, error) {
	// fallback to next parent scope if value does not exists
	p, err := cfgpath.New(r)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "[config] Float64. Route %q", r)
	}

	if ss.isAllowedStore(s...) {
		p = p.BindStore(ss.StoreID)
		v, err := ss.Root.Float64(p)
		if !errors.IsNotFound(err) || err == nil {
			// value found or err is not a NotFound error
			return v, p.ScopeHash, err
		}
	} // if not found in store scope go to website scope

	if ss.isAllowedWebsite(s...) {
		p = p.BindWebsite(ss.WebsiteID)
		v, err := ss.Root.Float64(p)
		if !errors.IsNotFound(err) || err == nil {
			// value found or err is not a NotFound error
			return v, p.ScopeHash, err
		}
	} // if not found in website scope go to default scope
	p.ScopeHash = scope.DefaultHash
	v, err := ss.Root.Float64(p)
	return v, scope.DefaultHash, err
}

// Int traverses through the scopes store->website->default to find
// a matching int value.
func (ss Scoped) Int(r cfgpath.Route, s ...scope.Scope) (int, scope.Hash, error) {
	// fallback to next parent scope if value does not exists
	p, err := cfgpath.New(r)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "[config] Int. Route %q", r)
	}

	if ss.isAllowedStore(s...) {
		p = p.BindStore(ss.StoreID)
		v, err := ss.Root.Int(p)
		if !errors.IsNotFound(err) || err == nil {
			// value found or err is not a NotFound error
			return v, p.ScopeHash, err
		}
	} // if not found in store scope go to website scope

	if ss.isAllowedWebsite(s...) {
		p = p.BindWebsite(ss.WebsiteID)
		v, err := ss.Root.Int(p)
		if !errors.IsNotFound(err) || err == nil {
			// value found or err is not a NotFound error
			return v, p.ScopeHash, err
		}
	} // if not found in website scope go to default scope
	p.ScopeHash = scope.DefaultHash
	v, err := ss.Root.Int(p)
	return v, scope.DefaultHash, err
}

// Time traverses through the scopes store->website->default to find
// a matching time.Time value.
func (ss Scoped) Time(r cfgpath.Route, s ...scope.Scope) (time.Time, scope.Hash, error) {
	// fallback to next parent scope if value does not exists
	p, err := cfgpath.New(r)
	if err != nil {
		return time.Time{}, 0, errors.Wrapf(err, "[config] Time. Route %q", r)
	}

	if ss.isAllowedStore(s...) {
		p = p.BindStore(ss.StoreID)
		v, err := ss.Root.Time(p)
		if !errors.IsNotFound(err) || err == nil {
			// value found or err is not a NotFound error
			return v, p.ScopeHash, err
		}
	} // if not found in store scope go to website scope

	if ss.isAllowedWebsite(s...) {
		p = p.BindWebsite(ss.WebsiteID)
		v, err := ss.Root.Time(p)
		if !errors.IsNotFound(err) || err == nil {
			// value found or err is not a NotFound error
			return v, p.ScopeHash, err
		}
	} // if not found in website scope go to default scope
	p.ScopeHash = scope.DefaultHash
	v, err := ss.Root.Time(p)
	return v, scope.DefaultHash, err
}
