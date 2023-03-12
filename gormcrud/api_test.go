// MIT License
//
// Copyright (c) 2023 Seth Osher
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package gormcrud

import (
	"flag"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/pilotso11/go-easycrud"
	"github.com/pilotso11/go-easycrud/util"
	"github.com/stretchr/testify/assert"
	"github.com/xo/dburl"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestItem struct {
	gorm.Model
	Key      string      `gorm:"uniqueIndex" crud:"key"`
	Children []TestChild `crud:"child"`
	Field1   int
	Field2   int
	Field3   int `json:"-"`
}

type TestChild struct {
	ID         string `gorm:"primaryKey"`
	TestItemID uint
}

type TestItemDto struct {
	Key    string
	Field2 int
	Field3 int `json:"-"`
}

// Test object using gorm ID
type TestID struct {
	gorm.Model
	Value1 string
	Value2 string
}

type TestIntKey struct {
	ID   int `gorm:"primaryKey"`
	Name string
}

var allow bool
var db *gorm.DB

func setup(t *testing.T) (*fiber.App, *gorm.DB) {
	if db == nil {
		dbUrl := flag.String("db", "sqlite:test.db", "Database URL")
		// dbUrl := flag.String("db", "postgres://postgres:postgres@localhost:5438/test", "Database URL")
		dsn, err := dburl.Parse(*dbUrl)
		switch dsn.Driver {
		case "postgres":
			db, err = gorm.Open(postgres.Open(dsn.DSN), &gorm.Config{})
		case "sqlite3": // this cause cgo issues for some, especially on windows
			db, err = gorm.Open(sqlite.Open(dsn.DSN), &gorm.Config{})
		}
		if err != nil {
			t.Fatalf("%v", err)
		}
		err = db.AutoMigrate(&TestItem{}, &TestChild{})
		if err != nil {
			t.Fatalf("%v", err)
		}
	}

	app := fiber.New()

	RegisterApi(app, db, "test", Options[TestItem, TestItemDto]{
		Delete: true,
		Mutate: true,
		Create: true,
		Validator: func(c *fiber.Ctx, action easycrud.Action, item ...TestItem) bool {
			return allow
		},
	})

	RegisterApi(app, db, "test2", Options[TestItem, TestItem]{
		Delete: true,
		Mutate: true,
		Create: true,
		Validator: func(c *fiber.Ctx, action easycrud.Action, item ...TestItem) bool {
			return allow
		},
	})

	db.Exec("DELETE FROM test_children WHERE 1=1")
	db.Exec("DELETE FROM test_items WHERE 1=1")
	var all []TestItem
	db.Find(&all)
	for i := 0; i < len(all); i++ {
		db.Delete(&all[i])
	}

	item1 := TestItem{
		Key:      "id1",
		Children: []TestChild{{ID: "ch1.1"}, {ID: "ch1.2"}},
		Field1:   10,
		Field2:   20,
		Field3:   30,
	}
	item2 := TestItem{
		Key:      "id2",
		Children: []TestChild{{ID: "ch2.1"}, {ID: "ch2.2"}},
		Field1:   10,
		Field2:   20,
		Field3:   30,
	}
	db.Save(&item1)
	db.Save(&item2)

	_ = db.AutoMigrate(&TestID{}, &TestIntKey{})
	RegisterApi(app, db, "testid", Options[TestID, TestID]{
		Delete: true,
		Mutate: true,
		Create: true,
	})

	RegisterApi(app, db, "testint", Options[TestIntKey, TestIntKey]{
		Delete: true,
		Mutate: true,
		Create: true,
	})

	return app, db
}

func cleanup(a *fiber.App) {
	_ = a.Shutdown()
}

func TestFind(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)

	assert.NotPanics(t, func() {
		allow = true
		code, resp, err := util.GetJsonRequestResponse(app, "GET", "/test/id1", nil)
		assert.Equal(t, 200, code)
		assert.Nil(t, err)
		assert.Equal(t, "id1", resp["Key"])
		assert.EqualValues(t, 20, resp["Field2"])
		assert.Nil(t, resp["Field1"])
	})
}

func TestFindMissing(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)

	assert.NotPanics(t, func() {
		allow = true
		code, _, _ := util.GetJsonRequestResponse(app, "GET", "/test/idmissing", nil)
		assert.Equal(t, 404, code)
	})
}

func TestFindAll(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)

	assert.NotPanics(t, func() {
		allow = true
		code, ret, err := util.GetJsonSliceRequestResponse(app, "GET", "/test/", nil)
		assert.Equal(t, 200, code)
		assert.Nil(t, err)
		assert.Len(t, ret, 2)
		assert.Equal(t, "id1", ret[0]["Key"])
		assert.Equal(t, "id2", ret[1]["Key"])
	})

}

func TestMutate(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)

	assert.NotPanics(t, func() {
		allow = true
		code, ret, err := util.GetJsonRequestResponse(app, "PUT", "/test/id2", TestItemDto{
			Key:    "id2",
			Field2: 22,
			Field3: 33,
		})
		assert.Equal(t, 200, code)
		assert.Nil(t, err)
		assert.EqualValues(t, 22, ret["Field2"])

		dbItem := TestItem{Key: "id2"}
		db.Find(&dbItem, &dbItem)
		assert.Equal(t, 22, dbItem.Field2)
		assert.Equal(t, 10, dbItem.Field1)
		assert.Equal(t, 30, dbItem.Field3) // ensure not mutated json="-"
	})

}

func TestMutateMissing(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)

	assert.NotPanics(t, func() {
		allow = true
		code, _, _ := util.GetJsonRequestResponse(app, "PUT", "/test/idmising", TestItemDto{
			Key:    "id2",
			Field2: 22,
			Field3: 33,
		})
		assert.Equal(t, 404, code)
	})
}

func TestCreate(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)

	assert.NotPanics(t, func() {
		allow = true
		code, ret, err := util.GetJsonRequestResponse(app, "POST", "/test", TestItemDto{
			Key:    "idnew",
			Field2: 22,
			Field3: 33,
		})
		assert.Equal(t, 200, code)
		assert.Nil(t, err)
		assert.EqualValues(t, 22, ret["Field2"])
		assert.EqualValues(t, "idnew", ret["Key"])

		dbItem := TestItem{Key: "idnew"}
		db.Find(&dbItem, &dbItem)
		assert.Equal(t, 22, dbItem.Field2)
		assert.Equal(t, 0, dbItem.Field1)
		assert.Equal(t, 0, dbItem.Field3) // ensure not mutated json="-"
	})

}

func TestCreateMissingKey(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)

	assert.NotPanics(t, func() {
		allow = true
		code, _, _ := util.GetJsonRequestResponse(app, "POST", "/test", TestItemDto{
			Key:    "",
			Field2: 22,
			Field3: 33,
		})
		assert.Equal(t, 500, code)

	})

}

func TestCreateExistsAlready(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)

	assert.NotPanics(t, func() {
		allow = true
		code, _, _ := util.GetJsonRequestResponse(app, "POST", "/test", TestItemDto{
			Key:    "id1",
			Field2: 22,
			Field3: 33,
		})
		assert.Equal(t, 500, code)

		// Validate no mutation took place
		dbItem := TestItem{Key: "id1"}
		db.Find(&dbItem, &dbItem)
		assert.Equal(t, 20, dbItem.Field2)
		assert.Equal(t, 10, dbItem.Field1)
		assert.Equal(t, 30, dbItem.Field3)
	})
}

func TestDelete(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)

	assert.NotPanics(t, func() {
		allow = true
		code, _, _ := util.GetJsonRequestResponse(app, "DELETE", "/test/id2", nil)
		assert.Equal(t, 200, code)

		// Validate no mutation took place
		dbItem := TestItem{Key: "id1"}
		db.Find(&dbItem, &dbItem)
		assert.NotNil(t, dbItem.DeletedAt)
	})

}

func TestDeleteMissing(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)

	assert.NotPanics(t, func() {
		allow = true
		code, _, _ := util.GetJsonRequestResponse(app, "DELETE", "/test/idmissing", nil)
		assert.Equal(t, 404, code)
	})
}

func TestGetChildren(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)

	assert.NotPanics(t, func() {
		allow = true
		code, ret, err := util.GetJsonSliceRequestResponse(app, "GET", "/test/id1/children", nil)
		assert.Equal(t, 200, code)
		assert.Nil(t, err)
		assert.Len(t, ret, 2)
		assert.Equal(t, ret[0]["ID"], "ch1.1")
		assert.Equal(t, ret[1]["ID"], "ch1.2")
	})

}

func TestUseBaseAsDtoFind(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)

	assert.NotPanics(t, func() {
		allow = true
		code, resp, err := util.GetJsonRequestResponse(app, "GET", "/test2/id1", nil)
		assert.Equal(t, 200, code)
		assert.Nil(t, err)
		assert.Equal(t, "id1", resp["Key"])
		assert.EqualValues(t, 20, resp["Field2"])
		assert.EqualValues(t, 10, resp["Field1"])
	})

}

func TestUseBaseAsDtoMutate(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)

	assert.NotPanics(t, func() {
		allow = true
		code, ret, err := util.GetJsonRequestResponse(app, "PUT", "/test2/id2", TestItem{
			Key:    "id2",
			Field1: 11,
			Field2: 22,
			Field3: 33,
		})
		assert.Equal(t, 200, code)
		assert.Nil(t, err)
		assert.EqualValues(t, 22, ret["Field2"])

		dbItem := TestItem{Key: "id2"}
		db.Find(&dbItem, &dbItem)
		assert.Equal(t, 22, dbItem.Field2)
		assert.Equal(t, 11, dbItem.Field1)
		assert.Equal(t, 30, dbItem.Field3) // ensure not mutated json="-"
	})
}

type BadDto struct {
	Key          string
	Field1       int
	FieldMissing string
}

type DtoMissingKey struct {
	Field1 int
}

func TestInvalidDtoMapping(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)
	assert.Panics(t, func() {
		RegisterApi(app, db, "test", Options[TestItem, BadDto]{
			Delete: true,
			Mutate: true,
			Create: true,
			Validator: func(c *fiber.Ctx, action easycrud.Action, item ...TestItem) bool {
				return allow
			},
		})
	})
}

func TestMissingKey(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)
	assert.Panics(t, func() {
		RegisterApi(app, db, "testid", Options[TestItem, DtoMissingKey]{
			Delete: true,
			Mutate: true,
			Create: true,
			Validator: func(c *fiber.Ctx, action easycrud.Action, item ...TestItem) bool {
				return allow
			},
		})

	})
}

func TestGormId(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)
	assert.NotPanics(t, func() {
		db.Exec("DELETE FROM test_ids WHERE 1=1")
		id1 := TestID{Value1: "one", Value2: "two"}
		id2 := TestID{Value1: "one", Value2: "two"}
		id3 := TestID{Value1: "one", Value2: "two"}
		db.Save(&id1)
		db.Save(&id2)

		code, ret, err := util.GetJsonRequestResponse(app, "GET", fmt.Sprintf("/testid/%d", id1.ID), nil)
		assert.Equal(t, code, 200)
		assert.Nil(t, err)
		assert.Equal(t, "one", ret["Value1"])
		assert.EqualValues(t, id1.ID, ret["ID"])

		id1.Value1 = "new value"
		code, ret, err = util.GetJsonRequestResponse(app, "PUT", "/testid/1", id1)
		assert.Equal(t, code, 200)
		assert.Nil(t, err)
		db.Find(&id1, &id1)
		assert.Equal(t, "new value", id1.Value1)

		code, ret, err = util.GetJsonRequestResponse(app, "POST", "/testid/", id3)
		assert.Equal(t, code, 200)
		assert.Nil(t, err)
		db.Find(&id3, &id3)
		assert.NotEqual(t, 0, id3.ID)
		assert.Equal(t, "one", id3.Value1)
	})

}

func TestDefaultOptions(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)
	assert.NotPanics(t, func() {
		options := DefaultOptions[TestID, TestID]()
		assert.True(t, options.Mutate)
		assert.True(t, options.Create)
		assert.True(t, options.Delete)
		assert.NotNil(t, options.Validator)

		// Check validation is permitted
		RegisterApi(app, db, "testid2", options)
		code, _, _ := util.GetJsonRequestResponse(app, "GET", "/testid2/", nil)
		assert.Equal(t, 200, code)

	})
}

func TestDisabledOptions(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)
	assert.NotPanics(t, func() {
		options := Options[TestID, TestID]{}
		RegisterApi(app, db, "testid2", options)

		db.Exec("DELETE FROM test_ids WHERE 1=1")
		id1 := TestID{Value1: "one", Value2: "two"}
		db.Save(&id1)

		code, _, _ := util.GetJsonRequestResponse(app, "GET", "/testid2/", nil)
		assert.Equal(t, 200, code)

		code, _, _ = util.GetJsonRequestResponse(app, "GET", fmt.Sprintf("/testid2/%d", id1.ID), nil)
		assert.Equal(t, 200, code)

		code, _, _ = util.GetJsonRequestResponse(app, "PUT", fmt.Sprintf("/testid2/%d", id1.ID), id1)
		assert.Equal(t, 405, code)

		code, _, _ = util.GetJsonRequestResponse(app, "POST", "/testid2", id1)
		assert.Equal(t, 405, code)

		code, _, _ = util.GetJsonRequestResponse(app, "DELETE", fmt.Sprintf("/testid2/%d", id1.ID), nil)
		assert.Equal(t, 405, code)
	})
}

func TestWithIntKey(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)
	assert.NotPanics(t, func() {
		db.Exec("DELETE FROM test_int_keys WHERE 1=1")
		id1 := TestIntKey{ID: 1, Name: "one"}
		db.Save(&id1)
		id2 := TestIntKey{ID: 2, Name: "two"}

		code, res, _ := util.GetJsonRequestResponse(app, "GET", "/testint/1", nil)
		assert.Equal(t, 200, code)
		assert.EqualValues(t, 1, res["ID"])

		code, res, _ = util.GetJsonRequestResponse(app, "POST", "/testint/", id2)
		assert.Equal(t, 200, code)
		assert.EqualValues(t, 2, res["ID"])

		// Test parse errors

		code, res, _ = util.GetJsonRequestResponse(app, "GET", "/testint/one", nil)
		assert.Equal(t, 404, code)

		code, res, _ = util.GetJsonRequestResponse(app, "PUT", "/testint/one", id2)
		assert.Equal(t, 404, code)

	})
}

type NoId struct {
	Value string
}

type BaseId struct {
	ID    string
	Value string
}
type NoIdDto struct {
	Value string
}

func TestNoId(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)
	assert.Panics(t, func() {
		RegisterApi(app, db, "noid", DefaultOptions[NoId, NoId]())
	})
}

func TestNoIdOnDto(t *testing.T) {
	app, _ := setup(t)
	defer cleanup(app)
	assert.Panics(t, func() {
		RegisterApi(app, db, "noid", DefaultOptions[BaseId, NoIdDto]())
	})
}
