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

package csdb_test

import (
	"bytes"
	"testing"

	"errors"
	"fmt"

	"github.com/corestoreio/csfw/storage/csdb"
	"github.com/corestoreio/csfw/storage/dbr"
	"github.com/stretchr/testify/assert"
)

func TestGetColumns(t *testing.T) {
	// @todo fix test data retrieving from database ...
	dbc := csdb.MustOpenTest()
	defer dbc.Close()
	sess := dbc.NewSession()

	tests := []struct {
		table          string
		want           string
		wantErr        error
		wantJoinFields string
	}{
		{"core_config_data",
			"csdb.Column{Field:dbr.InitNullString(`config_id`, true), Type:dbr.InitNullString(`int(10) unsigned`, true), Null:dbr.InitNullString(`NO`, true), Key:dbr.InitNullString(`PRI`, true), Default:dbr.InitNullString(``, false), Extra:dbr.InitNullString(`auto_increment`, true)},\ncsdb.Column{Field:dbr.InitNullString(`scope`, true), Type:dbr.InitNullString(`varchar(8)`, true), Null:dbr.InitNullString(`NO`, true), Key:dbr.InitNullString(`MUL`, true), Default:dbr.InitNullString(`default`, true), Extra:dbr.InitNullString(``, true)},\ncsdb.Column{Field:dbr.InitNullString(`scope_id`, true), Type:dbr.InitNullString(`int(11)`, true), Null:dbr.InitNullString(`NO`, true), Key:dbr.InitNullString(``, true), Default:dbr.InitNullString(`0`, true), Extra:dbr.InitNullString(``, true)},\ncsdb.Column{Field:dbr.InitNullString(`path`, true), Type:dbr.InitNullString(`varchar(255)`, true), Null:dbr.InitNullString(`NO`, true), Key:dbr.InitNullString(``, true), Default:dbr.InitNullString(`general`, true), Extra:dbr.InitNullString(``, true)},\ncsdb.Column{Field:dbr.InitNullString(`value`, true), Type:dbr.InitNullString(`text`, true), Null:dbr.InitNullString(`YES`, true), Key:dbr.InitNullString(``, true), Default:dbr.InitNullString(``, false), Extra:dbr.InitNullString(``, true)}\n",
			nil,
			"config_id_scope_scope_id_path_value",
		},
		{"catalog_category_product",
			"csdb.Column{Field:dbr.InitNullString(`category_id`, true), Type:dbr.InitNullString(`int(10) unsigned`, true), Null:dbr.InitNullString(`NO`, true), Key:dbr.InitNullString(`PRI`, true), Default:dbr.InitNullString(`0`, true), Extra:dbr.InitNullString(``, true)},\ncsdb.Column{Field:dbr.InitNullString(`product_id`, true), Type:dbr.InitNullString(`int(10) unsigned`, true), Null:dbr.InitNullString(`NO`, true), Key:dbr.InitNullString(`PRI`, true), Default:dbr.InitNullString(`0`, true), Extra:dbr.InitNullString(``, true)},\ncsdb.Column{Field:dbr.InitNullString(`position`, true), Type:dbr.InitNullString(`int(11)`, true), Null:dbr.InitNullString(`NO`, true), Key:dbr.InitNullString(``, true), Default:dbr.InitNullString(`0`, true), Extra:dbr.InitNullString(``, true)}\n",
			nil,
			"category_id_product_id_position",
		},
		{"non_existent",
			"",
			errors.New("non_existent"),
			"",
		},
	}

	for _, test := range tests {
		cols1, err1 := csdb.GetColumns(sess, test.table)
		if test.wantErr != nil {
			assert.Error(t, err1)
			assert.Contains(t, err1.Error(), test.wantErr.Error())
			//t.Logf("%s\n%#v\n", err1.Error(), err1.(errgo.Locationer).Location())
		} else {
			assert.NoError(t, err1)
			assert.Equal(t, test.want, fmt.Sprintf("%#v\n", cols1))
			assert.Equal(t, test.wantJoinFields, cols1.JoinFields("_"))
		}
	}
}

func TestColumns(t *testing.T) {

	tests := []struct {
		have  int
		want  int
		haveS string
		wantS string
	}{
		{
			mustStructure(table1).Columns.PrimaryKeys().Len(),
			0,
			mustStructure(table1).Columns.GoString(),
			"csdb.Column{Field:dbr.InitNullString(`category_id`, true), Type:dbr.InitNullString(`int(10) unsigned`, true), Null:dbr.InitNullString(`NO`, true), Key:dbr.InitNullString(`MUL`, true), Default:dbr.InitNullString(`0`, true), Extra:dbr.InitNullString(``, false)},\ncsdb.Column{Field:dbr.InitNullString(`path`, true), Type:dbr.InitNullString(`varchar(255)`, true), Null:dbr.InitNullString(`YES`, true), Key:dbr.InitNullString(`MUL`, true), Default:dbr.InitNullString(``, false), Extra:dbr.InitNullString(``, false)}",
		},
		{
			mustStructure(table2).Columns.PrimaryKeys().Len(),
			1,
			mustStructure(table2).Columns.GoString(),
			"csdb.Column{Field:dbr.InitNullString(`category_id`, true), Type:dbr.InitNullString(`int(10) unsigned`, true), Null:dbr.InitNullString(`NO`, true), Key:dbr.InitNullString(`PRI`, true), Default:dbr.InitNullString(`0`, true), Extra:dbr.InitNullString(``, false)},\ncsdb.Column{Field:dbr.InitNullString(`path`, true), Type:dbr.InitNullString(`varchar(255)`, true), Null:dbr.InitNullString(`YES`, true), Key:dbr.InitNullString(``, false), Default:dbr.InitNullString(``, false), Extra:dbr.InitNullString(``, false)}",
		},
		{
			mustStructure(table4).Columns.UniqueKeys().Len(), 1,
			mustStructure(table4).Columns.GoString(),
			"csdb.Column{Field:dbr.InitNullString(`user_id`, true), Type:dbr.InitNullString(`int(10) unsigned`, true), Null:dbr.InitNullString(`NO`, true), Key:dbr.InitNullString(`PRI`, true), Default:dbr.InitNullString(``, false), Extra:dbr.InitNullString(`auto_increment`, true)},\ncsdb.Column{Field:dbr.InitNullString(`email`, true), Type:dbr.InitNullString(`varchar(128)`, true), Null:dbr.InitNullString(`YES`, true), Key:dbr.InitNullString(``, false), Default:dbr.InitNullString(``, false), Extra:dbr.InitNullString(``, false)},\ncsdb.Column{Field:dbr.InitNullString(`username`, true), Type:dbr.InitNullString(`varchar(40)`, true), Null:dbr.InitNullString(`YES`, true), Key:dbr.InitNullString(`UNI`, true), Default:dbr.InitNullString(``, false), Extra:dbr.InitNullString(``, false)}",
		},
		{mustStructure(table4).Columns.PrimaryKeys().Len(), 1, "", ""},
	}

	for i, test := range tests {
		assert.Equal(t, test.want, test.have, "Incorrect length at index %d", i)
		assert.Equal(t, test.wantS, test.haveS)
	}

	tsN := mustStructure(table4).Columns.ByName("user_id_not_found")
	assert.NotNil(t, tsN)
	assert.False(t, tsN.Field.Valid)

	ts4 := mustStructure(table4).Columns.ByName("user_id")
	assert.True(t, ts4.Field.Valid)
	assert.True(t, ts4.IsAutoIncrement())

	ts4b := mustStructure(table4).Columns.ByName("email")
	assert.True(t, ts4b.Field.Valid)
	assert.True(t, ts4b.IsNull())

	assert.True(t, mustStructure(table4).Columns.First().IsPK())
	emptyTS := &csdb.Table{}
	assert.False(t, emptyTS.Columns.First().IsPK())

	hash, err := mustStructure(table3).Columns.Hash()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xd4, 0x12, 0x62, 0x5a, 0x9b, 0x3a, 0x68, 0xfe}, hash)

}

func TestGetGoPrimitive(t *testing.T) {

	tests := []struct {
		c           csdb.Column
		useNullType bool
		want        string
	}{
		{
			csdb.Column{
				Field:   dbr.NewNullString(`category_id131`, true),
				Type:    dbr.NewNullString(`int(10) unsigned`, true),
				Null:    dbr.NewNullString(`NO`, true),
				Key:     dbr.NewNullString(`PRI`, true),
				Default: dbr.NewNullString(`0`, true),
				Extra:   dbr.NewNullString(``, true),
			},
			false,
			"int64",
		},
		{
			csdb.Column{
				Field:   dbr.NewNullString(`category_id143`, true),
				Type:    dbr.NewNullString(`int(10) unsigned`, true),
				Null:    dbr.NewNullString(`YES`, true),
				Key:     dbr.NewNullString(`PRI`, true),
				Default: dbr.NewNullString(`0`, true),
				Extra:   dbr.NewNullString(``, true),
			},
			false,
			"int64",
		},
		{
			csdb.Column{
				Field:   dbr.NewNullString(`category_id155`, true),
				Type:    dbr.NewNullString(`int(10) unsigned`, true),
				Null:    dbr.NewNullString(`YES`, true),
				Key:     dbr.NewNullString(`PRI`, true),
				Default: dbr.NewNullString(`0`, true),
				Extra:   dbr.NewNullString(``, true),
			},
			true,
			"dbr.NullInt64",
		},

		{
			csdb.Column{
				Field:   dbr.NewNullString(`is_root_category155`, true),
				Type:    dbr.NewNullString(`smallint(2) unsigned`, true),
				Null:    dbr.NewNullString(`YES`, true),
				Key:     dbr.NewNullString(``, true),
				Default: dbr.NewNullString(`0`, true),
				Extra:   dbr.NewNullString(``, true),
			},
			false,
			"bool",
		},
		{
			csdb.Column{
				Field:   dbr.NewNullString(`is_root_category180`, true),
				Type:    dbr.NewNullString(`smallint(2) unsigned`, true),
				Null:    dbr.NewNullString(`YES`, true),
				Key:     dbr.NewNullString(``, true),
				Default: dbr.NewNullString(`0`, true),
				Extra:   dbr.NewNullString(``, true),
			},
			true,
			"dbr.NullBool",
		},

		{
			csdb.Column{
				Field:   dbr.NewNullString(`product_name193`, true),
				Type:    dbr.NewNullString(`varchar(255)`, true),
				Null:    dbr.NewNullString(`YES`, true),
				Key:     dbr.NewNullString(``, true),
				Default: dbr.NewNullString(`0`, true),
				Extra:   dbr.NewNullString(``, true),
			},
			true,
			"dbr.NullString",
		},
		{
			csdb.Column{
				Field: dbr.NewNullString(`product_name193`, true),
				Type:  dbr.NewNullString(`varchar(255)`, true),
				Null:  dbr.NewNullString(`YES`, true),
			},
			false,
			"string",
		},

		{
			csdb.Column{
				Field: dbr.NewNullString(`price`, true),
				Type:  dbr.NewNullString(`decimal(12,4)`, true),
				Null:  dbr.NewNullString(`YES`, true),
			},
			false,
			"money.Currency",
		},
		{
			csdb.Column{
				Field: dbr.NewNullString(`shipping_adjustment_230`, true),
				Type:  dbr.NewNullString(`decimal(12,4)`, true),
				Null:  dbr.NewNullString(`YES`, true),
			},
			true,
			"money.Currency",
		},
		{
			csdb.Column{
				Field: dbr.NewNullString(`grand_absolut_233`, true),
				Type:  dbr.NewNullString(`decimal(12,4)`, true),
				Null:  dbr.NewNullString(`YES`, true),
			},
			true,
			"money.Currency",
		},
		{
			csdb.Column{
				Field:   dbr.NewNullString(`some_currencies_242`, true),
				Type:    dbr.NewNullString(`decimal(12,4)`, true),
				Null:    dbr.NewNullString(`NO`, true),
				Default: dbr.NewNullString(`0.0000`, true),
			},
			true,
			"money.Currency",
		},

		{
			csdb.Column{
				Field:   dbr.NewNullString(`weight_252`, true),
				Type:    dbr.NewNullString(`decimal(10,4)`, true),
				Null:    dbr.NewNullString(`YES`, true),
				Default: dbr.NewNullString(`0.0000`, true),
			},
			true,
			"dbr.NullFloat64",
		},
		{
			csdb.Column{
				Field:   dbr.NewNullString(`weight_263`, true),
				Type:    dbr.NewNullString(`double(10,4)`, true),
				Null:    dbr.NewNullString(`YES`, true),
				Default: dbr.NewNullString(`0.0000`, true),
			},
			false,
			"float64",
		},

		{
			csdb.Column{
				Field:   dbr.NewNullString(`created_at_274`, true),
				Type:    dbr.NewNullString(`date`, true),
				Null:    dbr.NewNullString(`YES`, true),
				Default: dbr.NewNullString(`0000-00-00`, true),
			},
			false,
			"time.Time",
		},
		{
			csdb.Column{
				Field:   dbr.NewNullString(`created_at_274`, true),
				Type:    dbr.NewNullString(`date`, true),
				Null:    dbr.NewNullString(`YES`, true),
				Default: dbr.NewNullString(`0000-00-00`, true),
			},
			true,
			"dbr.NullTime",
		},
	}

	for _, test := range tests {
		have := test.c.GetGoPrimitive(test.useNullType)
		assert.Equal(t, test.want, have, "Test: %#v", test)
	}

}

var benchmarkGetColumns csdb.Columns
var benchmarkGetColumnsHashWant = []byte{0x3b, 0x2d, 0xdd, 0xf4, 0x4e, 0x2b, 0x3a, 0xd0}

// BenchmarkGetColumns-4	    1000	   3376128 ns/op	   24198 B/op	     196 allocs/op
// BenchmarkGetColumns-4	    1000	   1185381 ns/op	   21861 B/op	     179 allocs/op
func BenchmarkGetColumns(b *testing.B) {
	dbc := csdb.MustOpenTest()
	defer dbc.Close()
	sess := dbc.NewSession()
	var err error
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkGetColumns, err = csdb.GetColumns(sess, "eav_attribute")
		if err != nil {
			b.Error(err)
		}
	}
	hashHave, err := benchmarkGetColumns.Hash()
	if err != nil {
		b.Error(err)
	}
	if 0 != bytes.Compare(hashHave, benchmarkGetColumnsHashWant) {
		b.Errorf("\nHave %#v\nWant %#v\n", hashHave, benchmarkGetColumnsHashWant)
	}
	//	b.Log(benchmarkGetColumns.GoString())
}

var benchmarkColumnsJoinFields string
var benchmarkColumnsJoinFieldsWant = "category_id|product_id|position"
var benchmarkColumnsJoinFieldsData = csdb.Columns{
	csdb.Column{
		Field:   dbr.NewNullString("category_id"),
		Type:    dbr.NewNullString("int(10) unsigned"),
		Null:    dbr.NewNullString("NO"),
		Key:     dbr.NewNullString("", false),
		Default: dbr.NewNullString("0"),
		Extra:   dbr.NewNullString(""),
	},
	csdb.Column{
		Field:   dbr.NewNullString("product_id"),
		Type:    dbr.NewNullString("int(10) unsigned"),
		Null:    dbr.NewNullString("NO"),
		Key:     dbr.NewNullString(""),
		Default: dbr.NewNullString("0"),
		Extra:   dbr.NewNullString(""),
	},
	csdb.Column{
		Field:   dbr.NewNullString("position"),
		Type:    dbr.NewNullString("int(10) unsigned"),
		Null:    dbr.NewNullString("YES"),
		Key:     dbr.NewNullString(""),
		Default: dbr.NullString{},
		Extra:   dbr.NewNullString(""),
	},
}

// BenchmarkColumnsJoinFields-4	 2000000	       625 ns/op	     176 B/op	       5 allocs/op <- Go 1.5
func BenchmarkColumnsJoinFields(b *testing.B) {

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkColumnsJoinFields = benchmarkColumnsJoinFieldsData.JoinFields("|")
	}
	if benchmarkColumnsJoinFields != benchmarkColumnsJoinFieldsWant {
		b.Errorf("\nWant: %s\nHave: %s\n", benchmarkColumnsJoinFieldsWant, benchmarkColumnsJoinFields)
	}
}
