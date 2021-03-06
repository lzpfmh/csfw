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

package cstesting

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/corestoreio/csfw/storage/dbr"
)

// MockDB creates a mocked database connection. Fatals on error.
func MockDB(t fataler) (*dbr.Connection, sqlmock.Sqlmock) {
	db, sm, err := sqlmock.New()
	fatalIfError(t, err)

	dbc, err := dbr.NewConnection(dbr.WithDB(db))
	fatalIfError(t, err)
	return dbc, sm
}
