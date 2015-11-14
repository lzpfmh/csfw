// Copyright 2015, Cyrill @ Schumacher.fm and the CoreStore contributors
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

package store_test

import (
	"testing"

	"github.com/corestoreio/csfw/config/scope"
	"github.com/corestoreio/csfw/storage/csdb"
	"github.com/corestoreio/csfw/storage/dbr"
	"github.com/corestoreio/csfw/store"
	"github.com/corestoreio/csfw/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewWebsite(t *testing.T) {
	w, err := store.NewWebsite(
		&store.TableWebsite{WebsiteID: 1, Code: dbr.NewNullString("euro"), Name: dbr.NewNullString("Europe"), SortOrder: 0, DefaultGroupID: 1, IsDefault: dbr.NewNullBool(true)},
	)
	assert.NoError(t, err)
	assert.Equal(t, "euro", w.Data.Code.String)

	dg, err := w.DefaultGroup()
	assert.Nil(t, dg)
	assert.EqualError(t, store.ErrWebsiteDefaultGroupNotFound, err.Error())

	ds, err := w.DefaultStore()
	assert.Nil(t, ds)
	assert.EqualError(t, store.ErrWebsiteDefaultGroupNotFound, err.Error())
	assert.Nil(t, w.Stores)
	assert.Nil(t, w.Groups)
}

func TestMustNewWebsite(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.EqualError(t, r.(error), store.ErrArgumentCannotBeNil.Error())
		}
	}()
	_ = store.MustNewWebsite(nil, nil)
}

func TestNewWebsiteSetGroupsStores(t *testing.T) {
	w, err := store.NewWebsite(
		&store.TableWebsite{WebsiteID: 1, Code: dbr.NewNullString("euro"), Name: dbr.NewNullString("Europe"), SortOrder: 0, DefaultGroupID: 1, IsDefault: dbr.NewNullBool(true)},
		store.SetWebsiteGroupsStores(
			store.TableGroupSlice{
				&store.TableGroup{GroupID: 3, WebsiteID: 2, Name: "Australia", RootCategoryID: 2, DefaultStoreID: 5},
				&store.TableGroup{GroupID: 1, WebsiteID: 1, Name: "DACH Group", RootCategoryID: 2, DefaultStoreID: 2},
				&store.TableGroup{GroupID: 0, WebsiteID: 0, Name: "Default", RootCategoryID: 0, DefaultStoreID: 0},
				&store.TableGroup{GroupID: 2, WebsiteID: 1, Name: "UK Group", RootCategoryID: 2, DefaultStoreID: 4},
			},
			store.TableStoreSlice{
				&store.TableStore{StoreID: 0, Code: dbr.NewNullString("admin"), WebsiteID: 0, GroupID: 0, Name: "Admin", SortOrder: 0, IsActive: true},
				&store.TableStore{StoreID: 5, Code: dbr.NewNullString("au"), WebsiteID: 2, GroupID: 3, Name: "Australia", SortOrder: 10, IsActive: true},
				&store.TableStore{StoreID: 1, Code: dbr.NewNullString("de"), WebsiteID: 1, GroupID: 1, Name: "Germany", SortOrder: 10, IsActive: true},
				&store.TableStore{StoreID: 4, Code: dbr.NewNullString("uk"), WebsiteID: 1, GroupID: 2, Name: "UK", SortOrder: 10, IsActive: true},
				&store.TableStore{StoreID: 2, Code: dbr.NewNullString("at"), WebsiteID: 1, GroupID: 1, Name: "Österreich", SortOrder: 20, IsActive: true},
				&store.TableStore{StoreID: 6, Code: dbr.NewNullString("nz"), WebsiteID: 2, GroupID: 3, Name: "Kiwi", SortOrder: 30, IsActive: true},
				&store.TableStore{StoreID: 3, Code: dbr.NewNullString("ch"), WebsiteID: 1, GroupID: 1, Name: "Schweiz", SortOrder: 30, IsActive: true},
			},
		),
	)
	assert.NoError(t, err)

	dg, err := w.DefaultGroup()
	assert.NotNil(t, dg)
	assert.EqualValues(t, "DACH Group", dg.Data.Name, "get default group: %#v", dg)
	assert.NoError(t, err)

	ds, err := w.DefaultStore()
	assert.NotNil(t, ds)
	assert.EqualValues(t, "at", ds.Data.Code.String, "get default store: %#v", ds)
	assert.NoError(t, err)

	assert.NotNil(t, dg.Stores)
	assert.EqualValues(t, utils.StringSlice{"de", "at", "ch"}, dg.Stores.Codes())

	for _, st := range dg.Stores {
		assert.EqualValues(t, "DACH Group", st.Group.Data.Name)
		assert.EqualValues(t, "Europe", st.Website.Data.Name.String)
	}

	assert.NotNil(t, w.Stores)
	assert.EqualValues(t, utils.StringSlice{"de", "uk", "at", "ch"}, w.Stores.Codes())

	assert.NotNil(t, w.Groups)
	assert.EqualValues(t, utils.Int64Slice{1, 2}, w.Groups.IDs())

	assert.Exactly(t, int64(2), w.StoreID())
	assert.Exactly(t, int64(1), w.GroupID())
	assert.Equal(t, "euro", w.WebsiteCode())
}

func TestNewWebsiteStoreIDError(t *testing.T) {
	w, err := store.NewWebsite(
		&store.TableWebsite{WebsiteID: 1, Code: dbr.NewNullString("euro"), Name: dbr.NewNullString("Europe"), SortOrder: 0, DefaultGroupID: 1, IsDefault: dbr.NewNullBool(true)},
	)
	assert.NoError(t, err)
	assert.Exactly(t, scope.UnavailableStoreID, w.StoreID())
}

func TestNewWebsiteSetGroupsStoresError1(t *testing.T) {
	w, err := store.NewWebsite(
		&store.TableWebsite{WebsiteID: 1, Code: dbr.NewNullString("euro"), Name: dbr.NewNullString("Europe"), SortOrder: 0, DefaultGroupID: 1, IsDefault: dbr.NewNullBool(true)},
		store.SetWebsiteGroupsStores(
			store.TableGroupSlice{
				&store.TableGroup{GroupID: 0, WebsiteID: 0, Name: "Default", RootCategoryID: 0, DefaultStoreID: 0},
			},
			store.TableStoreSlice{
				&store.TableStore{StoreID: 5, Code: dbr.NewNullString("au"), WebsiteID: 2, GroupID: 3, Name: "Australia", SortOrder: 10, IsActive: true},
				&store.TableStore{StoreID: 1, Code: dbr.NewNullString("de"), WebsiteID: 1, GroupID: 1, Name: "Germany", SortOrder: 10, IsActive: true},
				&store.TableStore{StoreID: 4, Code: dbr.NewNullString("uk"), WebsiteID: 1, GroupID: 2, Name: "UK", SortOrder: 10, IsActive: true},
				&store.TableStore{StoreID: 2, Code: dbr.NewNullString("at"), WebsiteID: 1, GroupID: 1, Name: "Österreich", SortOrder: 20, IsActive: true},
				&store.TableStore{StoreID: 6, Code: dbr.NewNullString("nz"), WebsiteID: 2, GroupID: 3, Name: "Kiwi", SortOrder: 30, IsActive: true},
				&store.TableStore{StoreID: 3, Code: dbr.NewNullString("ch"), WebsiteID: 1, GroupID: 1, Name: "Schweiz", SortOrder: 30, IsActive: true},
			},
		),
	)
	assert.Nil(t, w)
	assert.Contains(t, err.Error(), "Integrity error")
}

func TestTableWebsiteSlice(t *testing.T) {
	websites := store.TableWebsiteSlice{
		&store.TableWebsite{WebsiteID: 0, Code: dbr.NewNullString("admin"), Name: dbr.NewNullString("Admin"), SortOrder: 0, DefaultGroupID: 0, IsDefault: dbr.NewNullBool(false)},
		&store.TableWebsite{WebsiteID: 1, Code: dbr.NewNullString("euro"), Name: dbr.NewNullString("Europe"), SortOrder: 0, DefaultGroupID: 1, IsDefault: dbr.NewNullBool(true)},
		nil,
		&store.TableWebsite{WebsiteID: 2, Code: dbr.NewNullString("oz"), Name: dbr.NewNullString("OZ"), SortOrder: 20, DefaultGroupID: 3, IsDefault: dbr.NewNullBool(false)},
	}
	assert.True(t, websites.Len() == 4)

	w1, err := websites.FindByWebsiteID(999)
	assert.Nil(t, w1)
	assert.EqualError(t, store.ErrIDNotFoundTableWebsiteSlice, err.Error())

	w2, err := websites.FindByWebsiteID(2)
	assert.NotNil(t, w2)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), w2.WebsiteID)

	w3, err := websites.FindByCode("euro")
	assert.NotNil(t, w3)
	assert.NoError(t, err)
	assert.Equal(t, "euro", w3.Code.String)

	w4, err := websites.FindByCode("corestore")
	assert.Nil(t, w4)
	assert.EqualError(t, store.ErrIDNotFoundTableWebsiteSlice, err.Error())

	wf1 := websites.Filter(func(w *store.TableWebsite) bool {
		return w != nil && w.WebsiteID == 1
	})
	assert.EqualValues(t, "Europe", wf1[0].Name.String)
}

func TestTableWebsiteSliceLoad(t *testing.T) {
	dbc := csdb.MustOpenTest()
	defer func() { assert.NoError(t, dbc.Close()) }()
	dbrSess := dbc.NewSession()

	var websites store.TableWebsiteSlice
	_, err := websites.SQLSelect(dbrSess)
	assert.NoError(t, err)

	assert.True(t, websites.Len() >= 2)
	for _, s := range websites {
		assert.True(t, len(s.Code.String) > 1)
	}
}
