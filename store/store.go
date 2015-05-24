// Copyright 2015 CoreStore Authors
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

// Package store implements the handling of websites, groups and stores
package store

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/corestoreio/csfw/config"
	"github.com/corestoreio/csfw/directory"
	"github.com/corestoreio/csfw/storage/csdb"
	"github.com/corestoreio/csfw/storage/dbr"
	"github.com/corestoreio/csfw/utils"
	"github.com/dgrijalva/jwt-go"
)

const (
	// DefaultStoreID is always 0.
	DefaultStoreID int64 = 0
	// HTTPRequestParamStore name of the GET parameter to set a new store in a current website/group context
	HTTPRequestParamStore = `___store`
	// CookieName important when the user selects a different store within the current website/group context.
	// This cookie permanently saves the new selected store code for one year.
	// The cookie must be removed when the default store of the current website if equal to the current store.
	CookieName = `store`

	// PriceScopeGlobal prices are for all stores and websites the same.
	PriceScopeGlobal = `0` // must be string
	// PriceScopeWebsite prices are in each website different.
	PriceScopeWebsite = `1` // must be string
)

type (
	// Store contains two maps for faster retrieving of the store index and the store collection
	// Only used in generated code. Implements interface StoreGetter.
	Store struct {
		// Contains the current website for this store. No integrity checks
		w *Website
		g *Group
		// underlaying raw data
		s *TableStore
	}
	// StoreSlice a collection of pointers to the Store structs. StoreSlice has some nifty method receviers.
	StoreSlice []*Store
)

var (
	ErrStoreNotFound         = errors.New("Store not found")
	ErrStoreNotActive        = errors.New("Store not active")
	ErrStoreNewArgNil        = errors.New("An argument cannot be nil")
	ErrStoreIncorrectGroup   = errors.New("Incorrect group")
	ErrStoreIncorrectWebsite = errors.New("Incorrect website")
	ErrStoreCodeInvalid      = errors.New("The store code may contain only letters (a-z), numbers (0-9) or underscore(_). The first character must be a letter")
)

// NewStore returns a new pointer to a Store. Panics if one of the arguments is nil.
// The integrity checks are done by the database.
func NewStore(w *TableWebsite, g *TableGroup, s *TableStore) *Store {
	if w == nil || g == nil || s == nil {
		panic(ErrStoreNewArgNil)
	}

	if s.GroupID != g.GroupID {
		panic(ErrStoreIncorrectGroup)
	}

	if s.WebsiteID != w.WebsiteID {
		panic(ErrStoreIncorrectWebsite)
	}

	return &Store{
		w: NewWebsite(w),
		g: NewGroup(g, nil),
		s: s,
	}
}

/*
	@todo implement Magento\Store\Model\Store
*/

// ID satisfies the interface Retriever and mainly used in the StoreManager for selecting Website,Group ...
func (s *Store) ID() int64 {
	return s.s.StoreID
}

// Website returns the website associated to this store
func (s *Store) Website() *Website {
	return s.w
}

// Group returns the group associated to this store
func (s *Store) Group() *Group {
	return s.g
}

// Data returns the real store data from the database
func (s *Store) Data() *TableStore {
	return s.s
}

// Path returns the sub path from the URL where CoreStore is installed
func (s *Store) Path() string {
	url, err := url.ParseRequestURI(s.BaseURL(config.URLTypeWeb, false))
	if err != nil {
		return "/"
	}
	return url.Path
}

// BaseUrl returns the path from the URL or config where CoreStore is installed @todo
// @see https://github.com/magento/magento2/blob/0.74.0-beta7/app/code/Magento/Store/Model/Store.php#L539
func (s *Store) BaseURL(ut config.URLType, isSecure bool) string {
	var url string
	var p string
	switch ut {
	case config.URLTypeWeb:
		p = PathUnsecureBaseURL
		if isSecure {
			p = PathSecureBaseURL
		}
		break
	case config.URLTypeStatic:
		p = PathUnsecureBaseStaticURL
		if isSecure {
			p = PathSecureBaseStaticURL
		}
		break
	case config.URLTypeMedia:
		p = PathUnsecureBaseMediaURL
		if isSecure {
			p = PathSecureBaseMediaURL
		}
		break
	// @todo rethink that here and maybe add the other paths if needed.
	default:
		panic("Unsupported UrlType")
	}

	url = s.ConfigString(p)

	if strings.Contains(url, PlaceholderBaseURL) {
		// @todo replace placeholder with \Magento\Framework\App\Request\Http::getDistroBaseUrl()
		// getDistroBaseUrl will be generated from the $_SERVER variable,
		url = strings.Replace(url, PlaceholderBaseURL, mustReadConfig().GetString(config.Path(config.PathCSBaseURL)), 1)
	}
	url = strings.TrimRight(url, "/") + "/"

	return url
}

// ConfigString tries to get a value from the scopeStore if empty
// falls back to default global scope.
// If using etcd or consul maybe this can lead to round trip times because of network access.
func (s *Store) ConfigString(path ...string) string {
	val := mustReadConfig().GetString(config.ScopeStore(s), config.Path(path...))
	if val == "" {
		val = mustReadConfig().GetString(config.Path(path...))
	}
	return val
}

// NewCookie creates a new pre-configured cookie.
// @todo create cookie manager to stick to the limits of http://www.ietf.org/rfc/rfc2109.txt page 15
// @see http://browsercookielimits.squawky.net/
func (s *Store) NewCookie() *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     s.Path(),
		Domain:   "",
		Secure:   false,
		HttpOnly: true,
	}
}

// SetCookie adds a cookie which contains the store code and is valid for one year.
func (s *Store) SetCookie(res http.ResponseWriter) {
	if res != nil {
		keks := s.NewCookie()
		keks.Value = s.Data().Code.String
		keks.Expires = time.Now().AddDate(1, 0, 0) // one year valid
		http.SetCookie(res, keks)
	}
}

// DeleteCookie deletes the store cookie
func (s *Store) DeleteCookie(res http.ResponseWriter) {
	if res != nil {
		keks := s.NewCookie()
		keks.Expires = time.Now().AddDate(-10, 0, 0)
		http.SetCookie(res, keks)
	}
}

// AddClaim adds the store code to a JSON web token
func (s *Store) AddClaim(t *jwt.Token) {
	t.Claims[CookieName] = s.Data().Code.String
}

// RootCategoryId returns the root category ID assigned to this store view.
func (s *Store) RootCategoryId() int64 {
	return s.Group().Data().RootCategoryID
}

/*
	Store Currency
*/

// AllowedCurrencies returns all installed currencies from global scope.
func (s *Store) AllowedCurrencies() []string {
	return strings.Split(mustReadConfig().GetString(config.Path(directory.PathSystemCurrencyInstalled)), ",")
}

// CurrentCurrency @todo
// @see app/code/Magento/Store/Model/Store.php::getCurrentCurrency
func (s *Store) CurrentCurrency() *directory.Currency {
	return nil
}

/*
	Global functions
*/
// GetClaim returns a valid store code from a JSON web token or nil
func GetCodeFromClaim(t *jwt.Token) Retriever {
	if t == nil {
		return nil
	}
	c, ok := t.Claims[CookieName]
	if cs, okcs := c.(string); okcs && ok && nil == ValidateStoreCode(cs) {
		return Code(cs)
	}
	return nil
}

// GetCookie returns from a Request the value of the store cookie or nil.
func GetCodeFromCookie(req *http.Request) Retriever {
	if req == nil {
		return nil
	}
	if keks, err := req.Cookie(CookieName); nil == err && nil == ValidateStoreCode(keks.Value) {
		return Code(keks.Value)
	}
	return nil
}

// ValidateStoreCode checks if a store code is valid. Returns an error if the  first letter is not a-zA-Z
// and followed by a-zA-Z0-9_ or store code length is greater than 32 characters.
func ValidateStoreCode(c string) error {
	if c == "" || len(c) > 32 {
		return ErrStoreCodeInvalid
	}
	c1 := c[0]
	if false == ((c1 >= 'a' && c1 <= 'z') || (c1 >= 'A' && c1 <= 'Z')) {
		return ErrStoreCodeInvalid
	}
	if false == utils.IsAlphaNumeric(c) {
		return ErrStoreCodeInvalid
	}
	return nil
}

/*
	StoreSlice method receivers
*/

// Len returns the length
func (s StoreSlice) Len() int { return len(s) }

// Filter returns a new slice filtered by predicate f
func (s StoreSlice) Filter(f func(*Store) bool) StoreSlice {
	var stores StoreSlice
	for _, v := range s {
		if v != nil && f(v) {
			stores = append(stores, v)
		}
	}
	return stores
}

// Codes returns a StringSlice with all store codes
func (s StoreSlice) Codes() utils.StringSlice {
	if len(s) == 0 {
		return nil
	}
	var c utils.StringSlice
	for _, st := range s {
		if st != nil {
			c.Append(st.Data().Code.String)
		}
	}
	return c
}

// IDs returns an Int64Slice with all store ids
func (s StoreSlice) IDs() utils.Int64Slice {
	if len(s) == 0 {
		return nil
	}
	var ids utils.Int64Slice
	for _, st := range s {
		if st != nil {
			ids.Append(st.Data().StoreID)
		}
	}
	return ids
}

// LastItem returns the last item of this slice or nil
func (s StoreSlice) LastItem() *Store {
	if s.Len() > 0 {
		return s[s.Len()-1]
	}
	return nil
}

/*
	TableStore and TableStoreSlice method receivers
*/

// IsDefault returns true if the current store is the default store.
func (s TableStore) IsDefault() bool {
	return s.StoreID == DefaultStoreID
}

// Load uses a dbr session to load all data from the core_store table into the current slice.
// The variadic 2nd argument can be a call back function to manipulate the select.
// Additional columns or joins cannot be added. This method receiver should only be used in development.
// @see https://github.com/magento/magento2/blob/0.74.0-beta7/app%2Fcode%2FMagento%2FStore%2FModel%2FResource%2FStore%2FCollection.php#L147
// regarding the sort order.
func (s *TableStoreSlice) Load(dbrSess dbr.SessionRunner, cbs ...csdb.DbrSelectCb) (int, error) {
	return csdb.LoadSlice(dbrSess, TableCollection, TableIndexStore, &(*s), append(cbs, func(sb *dbr.SelectBuilder) *dbr.SelectBuilder {
		sb.OrderBy("CASE WHEN main_table.store_id = 0 THEN 0 ELSE 1 END ASC")
		sb.OrderBy("main_table.sort_order ASC")
		return sb.OrderBy("main_table.name ASC")
	})...)
}

// Len returns the length
func (s TableStoreSlice) Len() int { return len(s) }

// FindByID returns a TableStore if found by id or an error
func (s TableStoreSlice) FindByID(id int64) (*TableStore, error) {
	for _, st := range s {
		if st != nil && st.StoreID == id {
			return st, nil
		}
	}
	return nil, ErrStoreNotFound
}

// FindByCode returns a TableStore if found by id or an error
func (s TableStoreSlice) FindByCode(code string) (*TableStore, error) {
	for _, st := range s {
		if st != nil && st.Code.Valid && st.Code.String == code {
			return st, nil
		}
	}
	return nil, ErrStoreNotFound
}

// FilterByGroupID returns a new slice with all TableStores belonging to a group id
func (s TableStoreSlice) FilterByGroupID(id int64) TableStoreSlice {
	return s.Filter(func(ts *TableStore) bool {
		return ts.GroupID == id
	})
}

// FilterByWebsiteID returns a new slice with all TableStores belonging to a website id
func (s TableStoreSlice) FilterByWebsiteID(id int64) TableStoreSlice {
	return s.Filter(func(ts *TableStore) bool {
		return ts.WebsiteID == id
	})
}

// Filter returns a new slice containing TableStores filtered by predicate f
func (s TableStoreSlice) Filter(f func(*TableStore) bool) TableStoreSlice {
	if len(s) == 0 {
		return nil
	}
	var tss TableStoreSlice
	for _, v := range s {
		if v != nil && f(v) {
			tss = append(tss, v)
		}
	}
	return tss
}

// Codes returns a StringSlice with all store codes
func (s TableStoreSlice) Codes() utils.StringSlice {
	if len(s) == 0 {
		return nil
	}
	var c utils.StringSlice
	for _, store := range s {
		if store != nil {
			c.Append(store.Code.String)
		}
	}
	return c
}

// IDs returns an Int64Slice with all store ids
func (s TableStoreSlice) IDs() utils.Int64Slice {
	if len(s) == 0 {
		return nil
	}
	var ids utils.Int64Slice
	for _, store := range s {
		if store != nil {
			ids.Append(store.StoreID)
		}
	}
	return ids
}
