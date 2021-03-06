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

package backendauth

import (
	"github.com/corestoreio/csfw/config/cfgpath"
	"github.com/corestoreio/csfw/config/element"
	"github.com/corestoreio/csfw/storage/text"
	"github.com/corestoreio/csfw/store/scope"
)

// NewConfigStructure global configuration structure for this package.
// Used in frontend (to display the user all the settings) and in
// backend (scope checks and default values). See the source code
// of this function for the overall available sections, groups and fields.
func NewConfigStructure() (element.SectionSlice, error) {
	return element.NewConfiguration(
		element.Section{
			ID: cfgpath.NewRoute(`net`),
			Groups: element.NewGroupSlice(
				element.Group{
					ID:      cfgpath.NewRoute(`auth`),
					Label:   text.Chars(`Authentication (Basic, IP)`),
					Comment: text.Chars(` `),

					SortOrder: 160,
					Scopes:    scope.PermWebsite,
					Fields: element.NewFieldSlice(
						element.Field{
							// Path: `net/auth/enable`,
							ID:        cfgpath.NewRoute(`eanble`),
							Label:     text.Chars(`Is Active`),
							Comment:   text.Chars(` `),
							Type:      element.TypeSelect,
							SortOrder: 10,
							Visible:   element.VisibleYes,
							Scopes:    scope.PermStore,
						},
					),
				},
			),
		},
	)
}
