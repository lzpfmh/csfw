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

package store

import (
	"errors"
	"net/http"
	"sync"

	"github.com/corestoreio/csfw/config"
	"github.com/corestoreio/csfw/storage/csdb"
	"github.com/corestoreio/csfw/storage/dbr"
	"github.com/juju/errgo"
)

type (
	// Manager implements the Storager interface and on request the StorageMutator interface.
	// Manager uses three internal maps to cache the pointers of Website, Group and Store.
	Manager struct {
		// storage get set of websites, groups and stores and also type assertion to StorageMutator for
		// ReInit and Persisting
		storage Storager
		mu      sync.RWMutex

		// the next six fields are for internal caching
		// map key is a hash value which is generated by either an int64 or a string.
		websiteMap map[uint64]*Website
		groupMap   map[uint64]*Group
		storeMap   map[uint64]*Store
		websites   WebsiteSlice
		groups     GroupSlice
		stores     StoreSlice

		// appStore (*cough*) contains the current selected store from init func. Cannot be cleared
		// when booting the app. This store is the main store under which the app runs.
		// In Magento slang it is called currentStore but current Store relates to a Store set
		// by InitByRequest()
		// @todo you can have per request a different store which overrides the currentStore
		appStore *Store

		// defaultStore some one must be always default.
		defaultStore *Store

		// HealthJob allows profiling and error handling. Default is a noop type
		// and can be overridden after creating a new Manager. @todo
		// HealthJob health.EventReceiver
	}
)

var (
	ErrUnsupportedScopeID         = errors.New("Unsupported scope id")
	ErrStoreChangeNotAllowed      = errors.New("Store change not allowed")
	ErrAppStoreNotSet             = errors.New("AppStore is not initialized")
	ErrAppStoreSet                = errors.New("AppStore already initialized")
	ErrManagerMutatorNotAvailable = errors.New("Storage Mutator is not implemented")
	ErrHashRetrieverNil           = errors.New("Hash argument is nil")
)

// NewManager creates a new store manager which handles websites, store groups and stores.
func NewManager(s Storager) *Manager {
	return &Manager{
		storage:    s,
		mu:         sync.RWMutex{},
		websiteMap: make(map[uint64]*Website),
		groupMap:   make(map[uint64]*Group),
		storeMap:   make(map[uint64]*Store),
		// HealthJob:  utils.HealthJobNoop, @todo
	}
}

// Init initializes the appStore from a scope code and a scope type.
// This function is mainly used when booting the app to set the environment configuration
// Also all other calls to any method receiver with nil arguments depends on the appStore.
// @see \Magento\Store\Model\StorageFactory::_reinitStores
func (sm *Manager) Init(scopeCode Retriever, scopeType config.ScopeID) error {
	if sm.appStore != nil {
		return ErrAppStoreSet
	}
	var err error
	switch scopeType {
	case config.ScopeStore:
		sm.appStore, err = sm.Store(scopeCode)
	case config.ScopeGroup:
		if g, errG := sm.Group(scopeCode); errG != nil { // this is the group_id
			return errgo.Mask(errG)
		} else { // else needed because of scoping of g
			sm.appStore, err = g.DefaultStore()
		}
	case config.ScopeWebsite:
		if w, errW := sm.Website(scopeCode); errW != nil {
			return errgo.Mask(errW)
		} else { // else needed because of scoping of w
			sm.appStore, err = w.DefaultStore()
		}
	default:
		return ErrUnsupportedScopeID
	}
	return errgo.Mask(err)
}

// InitByRequest returns a new Store read from a cookie or HTTP request param.
// The internal appStore must be set before hand.
// 1. check cookie store, always a string and the store code
// 2. check for ___store variable, always a string and the store code
// 3. May return nil,nil if nothing is set.
// This function must be used within an HTTP handler.
// The returned new Store must be used in the HTTP context and overrides the appStore.
// @see \Magento\Store\Model\StorageFactory::_reinitStores
func (sm *Manager) InitByRequest(res http.ResponseWriter, req *http.Request, scopeType config.ScopeID) (*Store, error) {
	if sm.appStore == nil {
		// that means you must call Init() before executing this function.
		return nil, ErrAppStoreNotSet
	}

	var reqStore *Store
	if keks := GetCookie(req); keks != nil {
		reqStore, _ = sm.GetRequestStore(keks, scopeType) // ignore errors
	}

	if reqStoreCode := req.URL.Query().Get(HTTPRequestParamStore); reqStoreCode != "" {
		var err error
		// @todo reqStoreCode if number ... cast to int64 because then group id if ScopeID is group.
		if reqStore, err = sm.GetRequestStore(Code(reqStoreCode), scopeType); err != nil {
			return nil, errgo.Mask(err)
		}
		// also delete and re-set a new cookie
		if reqStore != nil && reqStore.Data().Code.String == reqStoreCode {
			wds, err := reqStore.Website().DefaultStore()
			if err != nil {
				return nil, errgo.Mask(err)
			}
			if wds.Data().Code.String == reqStoreCode {
				reqStore.DeleteCookie(res) // cookie not needed anymore
			} else {
				reqStore.SetCookie(res) // make sure we force set the new store
			}
		}
	}
	return reqStore, nil // can be nil,nil
}

// GetRequestStore is in Magento named setCurrentStore and only used by InitByRequest().
// First argument is the store ID or store code, 2nd arg the scope from the init process.
// Also prevents running a store from another website or store group,
// if website or store group was specified explicitly.
// It returns either an error or the new Store. The returning errors can get ignored because if
// a Store Code is invalid the parent calling function must fall back to the appStore.
// This function must be used within an RPC handler.
func (sm *Manager) GetRequestStore(r Retriever, scopeType config.ScopeID) (*Store, error) {
	if sm.appStore == nil {
		// that means you must call Init() before executing this function.
		return nil, ErrAppStoreNotSet
	}

	activeStore, err := sm.activeStore(r) // this is the active store from Cookie or Request.
	if activeStore == nil || err != nil {
		// store is not active so ignore
		return nil, errgo.Mask(err)
	}

	allowStoreChange := false
	switch scopeType {
	case config.ScopeStore:
		allowStoreChange = true
		break
	case config.ScopeGroup:
		allowStoreChange = activeStore.Data().GroupID == sm.appStore.Data().GroupID
		break
	case config.ScopeWebsite:
		allowStoreChange = activeStore.Data().WebsiteID == sm.appStore.Data().WebsiteID
		break
	}

	if allowStoreChange {
		return activeStore, nil
	}
	return nil, ErrStoreChangeNotAllowed
}

// IsSingleStoreMode check if Single-Store mode is enabled in configuration and from Store count < 3.
// This flag only shows that admin does not want to show certain UI components at backend (like store switchers etc)
// if Magento has only one store view but it does not check the store view collection.
func (sm *Manager) IsSingleStoreMode(cfg config.ScopeReader) bool {
	return sm.HasSingleStore() && cfg.IsSetFlag(PathSingleStoreModeEnabled, config.ScopeStore, sm.appStore)
}

// HasSingleStore checks if we only have one store view besides the admin store view.
// Mostly used in models to the set store id and in blocks to not display the store switch.
func (sm *Manager) HasSingleStore() bool {
	ss, err := sm.Stores()
	if err != nil {
		return false
	}
	// that means: index 0 is admin store and always present plus one more store view.
	return ss.Len() < 3
}

// Website returns the cached Website pointer from an ID or code including all of its
// groups and all related stores. It panics when the integrity is incorrect.
// If ID and code are available then the non-empty code has precedence.
// If no argument has been supplied then the Website of the internal appStore
// will be returned. If more than one argument has been provided it returns an error.
func (sm *Manager) Website(r ...Retriever) (*Website, error) {
	notR := notRetriever(r...)
	switch {
	case notR && sm.appStore == nil:
		return nil, ErrAppStoreNotSet
	case notR && sm.appStore != nil:
		return sm.appStore.Website(), nil
	}

	key, err := hash(r[0])
	if err != nil {
		return nil, err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	if w, ok := sm.websiteMap[key]; ok && w != nil {
		return w, nil
	}

	w, err := sm.storage.Website(r[0])
	sm.websiteMap[key] = w
	return sm.websiteMap[key], errgo.Mask(err)
}

// Websites returns a cached slice containing all pointers to Websites with its associated
// groups and stores. It panics when the integrity is incorrect.
func (sm *Manager) Websites() (WebsiteSlice, error) {
	if sm.websites != nil {
		return sm.websites, nil
	}
	var err error
	sm.websites, err = sm.storage.Websites()
	return sm.websites, err
}

// Group returns a cached Group which contains all related stores and its website.
// Only the argument ID is supported.
// If no argument has been supplied then the Group of the internal appStore
// will be returned. If more than one argument has been provided it returns an error.
func (sm *Manager) Group(r ...Retriever) (*Group, error) {
	notR := notRetriever(r...)
	switch {
	case notR && sm.appStore == nil:
		return nil, ErrAppStoreNotSet
	case notR && sm.appStore != nil:
		return sm.appStore.Group(), nil
	}

	key, err := hash(r[0])
	if err != nil {
		return nil, err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	if g, ok := sm.groupMap[key]; ok && g != nil {
		return g, nil
	}

	g, err := sm.storage.Group(r[0])
	sm.groupMap[key] = g
	return sm.groupMap[key], errgo.Mask(err)
}

// Groups returns a cached slice containing all pointers to Groups with its associated
// stores and websites. It panics when the integrity is incorrect.
func (sm *Manager) Groups() (GroupSlice, error) {
	if sm.groups != nil {
		return sm.groups, nil
	}
	var err error
	sm.groups, err = sm.storage.Groups()
	return sm.groups, err
}

// Store returns the cached Store view containing its group and its website.
// If ID and code are available then the non-empty code has precedence.
// If no argument has been supplied then the appStore
// will be returned. If more than one argument has been provided it returns an error.
func (sm *Manager) Store(r ...Retriever) (*Store, error) {
	notR := notRetriever(r...)
	switch {
	case notR && sm.appStore == nil:
		return nil, ErrAppStoreNotSet
	case notR && sm.appStore != nil:
		return sm.appStore, nil
	}

	key, err := hash(r[0])
	if err != nil {
		return nil, err
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	if s, ok := sm.storeMap[key]; ok && s != nil {
		return s, nil
	}

	s, err := sm.storage.Store(r[0])
	sm.storeMap[key] = s
	return sm.storeMap[key], errgo.Mask(err)
}

// Stores returns a cached Store slice. Can return an error when the website or
// the group cannot be found.
func (sm *Manager) Stores() (StoreSlice, error) {
	if sm.stores != nil {
		return sm.stores, nil
	}
	var err error
	sm.stores, err = sm.storage.Stores()
	return sm.stores, err
}

// DefaultStoreView returns the default store view.
func (sm *Manager) DefaultStoreView() (*Store, error) {
	if sm.defaultStore != nil {
		return sm.defaultStore, nil
	}
	var err error
	sm.defaultStore, err = sm.storage.DefaultStoreView()
	return sm.defaultStore, err
}

// activeStore returns a new non-cached Store with all its Websites and Groups but only if the Store
// is marked as active. Argument can be an ID or a Code. Returns nil if Store not found or inactive.
// No need here to return an error.
func (sm *Manager) activeStore(r Retriever) (*Store, error) {
	s, err := sm.storage.Store(r)
	if err != nil {
		return nil, err
	}
	if s.Data().IsActive {
		return s, nil
	}
	return nil, ErrStoreNotActive
}

// ReInit reloads the website, store group and store view data from the database @todo
func (sm *Manager) ReInit(dbrSess dbr.SessionRunner) error {
	if mut, ok := sm.storage.(StorageMutator); ok {
		defer sm.ClearCache() // hmmm .... defer ...
		return mut.ReInit(dbrSess)
	}
	return ErrManagerMutatorNotAvailable
}

// ClearCache resets the internal caches which stores the pointers to a Website, Group or Store and
// all related slices. Please use with caution. ReInit() also uses this method.
// Providing argument true clears also the internal appStore cache.
func (sm *Manager) ClearCache(clearAll ...bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if len(sm.websiteMap) > 0 {
		for k := range sm.websiteMap {
			delete(sm.websiteMap, k)
		}
	}
	if len(sm.groupMap) > 0 {
		for k := range sm.groupMap {
			delete(sm.groupMap, k)
		}
	}
	if len(sm.storeMap) > 0 {
		for k := range sm.storeMap {
			delete(sm.storeMap, k)
		}
	}
	sm.websites = nil
	sm.groups = nil
	sm.stores = nil
	sm.defaultStore = nil
	// do not clear currentStore as this one depends on the init funcs
	if 1 == len(clearAll) && clearAll[0] {
		sm.appStore = nil
	}
}

// IsCacheEmpty returns true if the internal cache is empty.
func (sm *Manager) IsCacheEmpty() bool {
	return len(sm.websiteMap) == 0 && len(sm.groupMap) == 0 && len(sm.storeMap) == 0 &&
		sm.websites == nil && sm.groups == nil && sm.stores == nil && sm.defaultStore == nil
}

// notRetriever checks if variadic Retriever is nil or has more than two entries
// or the first index is nil.
func notRetriever(r ...Retriever) bool {
	lr := len(r)
	return r == nil || (lr == 1 && r[0] == nil) || lr > 1
}

// hash generates the key for the map from either an id int64 or a code string.
// If both interfaces are nil it returns 0 which is default for website, group or store.
// fnv64a used to calculate the uint64 value of a string, especially website code and store code.
func hash(r Retriever) (uint64, error) {
	uz := uint64(0)
	if r == nil {
		return uz, ErrHashRetrieverNil
	}

	if c, ok := r.(CodeRetriever); ok && c.Code() != "" {
		data := []byte(c.Code())
		var hash uint64 = 14695981039346656037
		for _, c := range data {
			hash ^= uint64(c)
			hash *= 1099511628211
		}
		return hash, nil
	}
	return uint64(r.ID()), nil
}

// loadSlice internal global helper func to execute a SQL select. @todo refactor and remove dependency of GetTableS...
func loadSlice(dbrSess dbr.SessionRunner, table csdb.Index, dest interface{}, cbs ...csdb.DbrSelectCb) (int, error) {
	ts, err := GetTableStructure(table)
	if err != nil {
		return 0, errgo.Mask(err)
	}

	sb, err := ts.Select(dbrSess)
	if err != nil {
		return 0, errgo.Mask(err)
	}

	for _, cb := range cbs {
		sb = cb(sb)
	}
	return sb.LoadStructs(dest)
}
