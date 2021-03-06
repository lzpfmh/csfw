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

package store

import (
	"sync"
	"sync/atomic"

	"github.com/corestoreio/csfw/config"
	"github.com/corestoreio/csfw/storage/dbr"
	"github.com/corestoreio/csfw/store/scope"
	"github.com/corestoreio/csfw/util/errors"
)

// IDbyCode returns for a website code or store code the id. Group scope is not
// supported because the group table does not contain a code string column. A
// not-supported error behaviour gets returned if an invalid scope has been
// provided. Default scope returns always 0.
type CodeToIDMapper interface {
	IDbyCode(scp scope.Scope, code string) (id int64, err error)
}

// AvailabilityChecker depends on the run mode from package scope. The Hash
// argument will be provided via scope.RunMode type or the
// scope.FromContextRunMode(ctx) function. runMode is called in M world:
// MAGE_RUN_CODE and MAGE_RUN_TYPE. The MAGE_RUN_TYPE can be either website or
// store scope and MAGE_RUN_CODE any defined website or store code from the
// database.
type AvailabilityChecker interface {
	// AllowedStoreIds returns all active store IDs for a run mode.
	AllowedStoreIds(runMode scope.Hash) ([]int64, error)
	// DefaultStoreID returns the default active store ID depending on the run mode.
	// Error behaviour is mostly of type NotValid.
	DefaultStoreID(runMode scope.Hash) (int64, error)
}

// Service represents type which handles the underlying storage and takes
// care of the default stores. A Service is bound a specific scope.Scope.
// Depending on the scope it is possible or not to switch stores. A Service
// contains also a config.Getter which gets passed to the scope of a
// Store(), Group() or Website() so that you always have the possibility to
// access a scoped based configuration value. This Service uses three
// internal maps to cache Websites, Groups and Stores.
type Service struct {

	// backend communicates with the database in reading mode and creates
	// new store, group and website pointers. If nil, panics.
	backend *factory
	// defaultStore someone must be always the default guy. Handled via atomic
	// package.
	defaultStoreID int64
	// mu protects the following fields
	mu sync.RWMutex
	// in general these caches can be optimized
	websites WebsiteSlice
	groups   GroupSlice
	stores   StoreSlice

	// int64 key identifies a website, group or store
	cacheWebsite map[int64]Website
	cacheGroup   map[int64]Group
	cacheStore   map[int64]Store
}

// NewService creates a new store Service which handles websites, groups and
// stores. You must either provide the functional options or call LoadFromDB()
// to setup the internal cache.
func NewService(cfg config.Getter, opts ...Option) (*Service, error) {
	srv := &Service{
		defaultStoreID: -1,
	}
	if err := srv.loadFromOptions(cfg, opts...); err != nil {
		return nil, errors.Wrap(err, "[store] NewService.ApplyStorage")
	}
	return srv, nil
}

// MustNewService same as NewService, but panics on error.
func MustNewService(cfg config.Getter, opts ...Option) *Service {
	m, err := NewService(cfg, opts...)
	if err != nil {
		panic(err)
	}
	return m
}

// loadFromOptions main function to set up the internal caches from the factory.
// Does nothing when the options have not been passed.
func (s *Service) loadFromOptions(cfg config.Getter, opts ...Option) error {
	if s == nil {
		s = new(Service)
		s.defaultStoreID = -1
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	be, err := newFactory(cfg, opts...)
	if err != nil {
		return errors.Wrap(err, "[store] NewService.NewFactory")
	}

	s.backend = be
	s.cacheWebsite = make(map[int64]Website)
	s.cacheGroup = make(map[int64]Group)
	s.cacheStore = make(map[int64]Store)

	ws, err := s.backend.Websites()
	if err != nil {
		return errors.Wrap(err, "[store] NewService.Websites")
	}
	s.websites = ws
	ws.Each(func(w Website) {
		s.cacheWebsite[w.Data.WebsiteID] = w
	})

	gs, err := s.backend.Groups()
	if err != nil {
		return errors.Wrap(err, "[store] NewService.Groups")
	}
	s.groups = gs
	gs.Each(func(g Group) {
		s.cacheGroup[g.Data.GroupID] = g
	})

	ss, err := s.backend.Stores()
	if err != nil {
		return errors.Wrap(err, "[store] NewService.Stores")
	}
	s.stores = ss
	ss.Each(func(str Store) {
		s.cacheStore[str.Data.StoreID] = str
	})
	return nil
}

// AllowedStoreIds returns all active store IDs for a run mode.
func (s *Service) AllowedStoreIds(runMode scope.Hash) ([]int64, error) {
	scp, id := runMode.Unpack()

	switch scp {
	case scope.Store:
		return s.stores.ActiveIDs(), nil

	case scope.Group:
		g, err := s.Group(id) // if ID == 0 then admin group
		if err != nil {
			return nil, errors.Wrapf(err, "[store] AllowedStoreIds.Group Scope %s ID %d", scp, id)
		}
		return g.Stores.ActiveIDs(), nil
	}

	var w Website
	if scp == scope.Website {
		var err error
		w, err = s.Website(id) // id ID == 0 then admin website
		if err != nil {
			return nil, errors.Wrapf(err, "[store] AllowedStoreIds.Website Scope %s ID %d", scp, id)
		}
	} else {
		var err error
		w, err = s.websites.Default()
		if err != nil {
			return nil, errors.Wrapf(err, "[store] AllowedStoreIds.Website.Default Scope %s ID %d", scp, id)
		}
	}
	g, err := w.DefaultGroup()
	if err != nil {
		return nil, errors.Wrapf(err, "[store] AllowedStoreIds.DefaultGroup Scope %s ID %d", scp, id)
	}
	return g.Stores.ActiveIDs(), nil
}

// DefaultStoreID returns the default active store ID depending on the run mode.
// Error behaviour is mostly of type NotValid.
func (s *Service) DefaultStoreID(runMode scope.Hash) (int64, error) {
	scp, id := runMode.Unpack()
	switch scp {
	case scope.Store:
		st, err := s.Store(id)
		if err != nil {
			return 0, errors.Wrapf(err, "[store] DefaultStoreID Scope %s ID %d", scp, id)
		}
		if !st.Data.IsActive {
			return 0, errors.NewNotValidf("[store] DefaultStoreID %s the store ID %d is not active", runMode, st.ID())
		}
		return st.ID(), nil

	case scope.Group:
		g, err := s.Group(id)
		if err != nil {
			return 0, errors.Wrapf(err, "[store] DefaultStoreID Scope %s ID %d", scp, id)
		}
		st, err := s.Store(g.Data.DefaultStoreID)
		if err != nil {
			return 0, errors.Wrapf(err, "[store] DefaultStoreID Scope %s ID %d", scp, id)
		}
		if !st.Data.IsActive {
			return 0, errors.NewNotValidf("[store] DefaultStoreID %s the store ID %d is not active", runMode, st.ID())
		}
		return st.ID(), nil
	}

	var w Website
	if scp == scope.Website {
		var err error
		w, err = s.Website(id)
		if err != nil {
			return 0, errors.Wrapf(err, "[store] DefaultStoreID.Website Scope %s ID %d", scp, id)
		}
	} else {
		var err error
		w, err = s.websites.Default()
		if err != nil {
			return 0, errors.Wrapf(err, "[store] DefaultStoreID.Website.Default Scope %s ID %d", scp, id)
		}
	}
	st, err := w.DefaultStore()
	if err != nil {
		return 0, errors.Wrapf(err, "[store] DefaultStoreID.Website.DefaultStore Scope %s ID %d", scp, id)
	}
	if !st.Data.IsActive {
		return 0, errors.NewNotValidf("[store] DefaultStoreID %s the store ID %d is not active", runMode, st.ID())
	}
	return st.Data.StoreID, nil
}

// IDbyCode returns for a website code or store code the id. It iterates over
// the internal cache maps. Group scope is not supported because the group table
// does not contain a code string column. A not-supported error behaviour gets
// returned if an invalid scope has been provided. Default scope returns always
// 0. Implements interface CodeToIDMapper.
func (s *Service) IDbyCode(scp scope.Scope, code string) (int64, error) {
	if code == "" {
		return 0, errors.NewEmptyf("[store] Service IDByCode: Code canot be empty.")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	// todo maybe add map cache
	switch scp {
	case scope.Store:
		if ts, ok := s.backend.stores.FindByCode(code); ok {
			return ts.StoreID, nil
		}
		return 0, errors.NewNotFoundf("[store] Code %q not found in %s", code, scp)
	case scope.Website:
		if tw, ok := s.backend.websites.FindByCode(code); ok {
			return tw.WebsiteID, nil
		}
		return 0, errors.NewNotFoundf("[store] Code %q not found in %s", code, scp)
	case scope.Default:
		return 0, nil
	}
	return 0, errors.NewNotSupportedf("[store] Scope %q not supported", scp)
}

// IsSingleStoreMode check if Single-Store mode is enabled in configuration and from Store count < 3.
// This flag only shows that admin does not want to show certain UI components at backend (like store switchers etc)
// if Magento has only one store view but it does not check the store view collection.
//func (sm *Service) IsSingleStoreMode() bool {
//	// refactor and remove dependency to backend.Backend
//	return sm.HasSingleStore() // && backend.Backend.GeneralSingleStoreModeEnabled.Get(sm.cr.NewScoped(0, 0)) // default scope
//}
//
//// HasSingleStore checks if we only have one store view besides the admin store view.
//// Mostly used in models to the set store id and in blocks to not display the store switch.
//func (sm *Service) HasSingleStore() bool {
//	ss, err := sm.Stores()
//	if err != nil {
//		return false
//	}
//	// that means: index 0 is admin store and always present plus one more store view.
//	return ss.Len() < 3
//}

// Website returns the cached Website from an ID including all of its groups and
// all related stores.
func (s *Service) Website(id int64) (Website, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if cs, ok := s.cacheWebsite[id]; ok {
		return cs, nil
	}
	return Website{}, errors.NewNotFoundf("[store] Cannot find Website ID %d", id)
}

// Websites returns a cached slice containing all Websites with its associated
// groups and stores. You shall not modify the returned slice.
func (s *Service) Websites() WebsiteSlice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.websites
}

// Group returns a cached Group which contains all related stores and its website.
func (s *Service) Group(id int64) (Group, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if cg, ok := s.cacheGroup[id]; ok {
		return cg, nil
	}
	return Group{}, errors.NewNotFoundf("[store] Cannot find Group ID %d", id)
}

// Groups returns a cached slice containing all  Groups with its associated
// stores and websites. You shall not modify the returned slice.
func (s *Service) Groups() GroupSlice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.groups
}

// Store returns the cached Store view containing its group and its website.
func (s *Service) Store(id int64) (Store, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if cs, ok := s.cacheStore[id]; ok {
		return cs, nil
	}
	return Store{}, errors.NewNotFoundf("[store] Cannot find Store ID %d", id)
}

// Stores returns a cached Store slice containing all related websites and groups.
// You shall not modify the returned slice.
func (s *Service) Stores() StoreSlice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stores
}

// DefaultStoreView returns the overall default store view.
func (s *Service) DefaultStoreView() (Store, error) {
	if s.defaultStoreID >= 0 {
		s.mu.RLock()
		defer s.mu.RUnlock() // bug
		if cs, ok := s.cacheStore[atomic.LoadInt64(&s.defaultStoreID)]; ok {
			return cs, nil
		}
	}

	id, err := s.backend.DefaultStoreID()
	if err != nil {
		return Store{}, errors.Wrap(err, "[store] Service.storage.DefaultStoreView")
	}
	atomic.StoreInt64(&s.defaultStoreID, id)
	return s.Store(id)
}

// LoadFromDB reloads the website, store group and store view data from the database.
// After reloading internal cache will be cleared if there are no errors.
func (s *Service) LoadFromDB(dbrSess dbr.SessionRunner, cbs ...dbr.SelectCb) error {

	if err := s.backend.LoadFromDB(dbrSess, cbs...); err != nil {
		return errors.Wrap(err, "[store] LoadFromDB.Backend")
	}

	s.ClearCache()

	err := s.loadFromOptions(
		s.backend.baseConfig,
		WithTableWebsites(s.backend.websites...),
		WithTableGroups(s.backend.groups...),
		WithTableStores(s.backend.stores...),
	)
	return errors.Wrap(err, "[store] LoadFromDB.ApplyStorage")
}

// ClearCache resets the internal caches which stores the pointers to Websites,
// Groups or Stores. The ReInit() also uses this method to clear caches before
// the Storage gets reloaded.
func (s *Service) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.cacheWebsite) > 0 {
		for k := range s.cacheWebsite {
			delete(s.cacheWebsite, k)
		}
	}
	if len(s.cacheGroup) > 0 {
		for k := range s.cacheGroup {
			delete(s.cacheGroup, k)
		}
	}
	if len(s.cacheStore) > 0 {
		for k := range s.cacheStore {
			delete(s.cacheStore, k)
		}
	}
	s.defaultStoreID = -1
	s.websites = nil
	s.groups = nil
	s.stores = nil
}

// IsCacheEmpty returns true if the internal cache is empty.
func (s *Service) IsCacheEmpty() bool {
	return len(s.cacheWebsite) == 0 && len(s.cacheGroup) == 0 && len(s.cacheStore) == 0 &&
		s.defaultStoreID == -1
}
