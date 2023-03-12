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

package easyrest

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/pilotso11/go-easyrest/util"
	"github.com/stretchr/testify/assert"
)

type ChildItem struct {
	Name string
}

type TestItem struct {
	Id       string
	Data     string
	Children []ChildItem
}

func ItemToDto(i TestItem) TestItemDto {
	return TestItemDto{
		Id:   i.Id,
		Data: i.Data,
	}
}

type TestItemDto struct {
	Id   string
	Data string
}

func (d TestItemDto) Match(v TestItem) bool {
	match := true
	if d.Id > "" {
		if v.Id != d.Id {
			match = false
		}
	}
	if d.Data > "" {
		if !strings.Contains(v.Data, d.Data) {
			match = false
		}
	}
	return match
}

type TestData struct {
	lock    sync.Mutex
	entries map[string]TestItem
	permit  bool
	fail    bool
}

func setup() (*fiber.App, *TestData) {
	app := fiber.New()
	data := TestData{
		entries: make(map[string]TestItem),
	}

	// This is a full controller
	fullApi := Api[TestItem, TestItemDto]{
		Path: "test",
		Find: func(key string) (TestItem, bool) {
			data.lock.Lock()
			defer data.lock.Unlock()
			item, ok := data.entries[key]
			return item, ok
		},
		FindAll: func() []TestItem {
			data.lock.Lock()
			defer data.lock.Unlock()
			var all []TestItem
			for _, v := range data.entries {
				all = append(all, v)
			}
			return all
		},
		Search: func(filter TestItemDto) []TestItem {
			data.lock.Lock()
			defer data.lock.Unlock()
			var all []TestItem
			for _, v := range data.entries {
				if filter.Match(v) {
					all = append(all, v)
				}
			}
			return all

		},
		Mutate: func(item TestItem, dto TestItemDto) (TestItem, error) {
			data.lock.Lock()
			defer data.lock.Unlock()
			if data.fail {
				return TestItem{}, errors.New("update error")
			}
			item.Data = dto.Data
			data.entries[item.Id] = item
			return item, nil

		},
		Create: func(dto TestItemDto) (TestItem, error) {
			data.lock.Lock()
			defer data.lock.Unlock()
			if data.fail {
				return TestItem{}, errors.New("create error")
			}
			newItem := TestItem{
				Id:       dto.Id,
				Data:     dto.Data,
				Children: []ChildItem{{"a"}, {"b"}},
			}
			data.entries[dto.Id] = newItem
			return newItem, nil
		},
		Delete: func(item TestItem) (TestItem, error) {
			data.lock.Lock()
			defer data.lock.Unlock()
			if data.fail {
				return TestItem{}, errors.New("delete error")
			}
			delete(data.entries, item.Id)
			return item, nil
		},
		SubEntities: []SubEntity[TestItem, TestItemDto]{
			{"children", func(item TestItem) []any {
				var ret []any
				for _, c := range item.Children {
					ret = append(ret, c)
				}
				return ret
			}},
		},
		Validator: func(ctx *fiber.Ctx, action Action, item ...TestItem) bool {
			return data.permit
		},
		Dto: ItemToDto,
	}

	_, _ = fullApi.Create(TestItemDto{"id1", "original data"})
	_, _ = fullApi.Create(TestItemDto{"id2", "original data2"})

	editOnlyApi := Api[TestItem, TestItemDto]{
		Path:        "test2",
		Find:        fullApi.Find,
		FindAll:     fullApi.FindAll,
		Mutate:      fullApi.Mutate,
		Create:      nil,
		Delete:      nil,
		SubEntities: nil,
		Validator:   fullApi.Validator,
		Dto:         ItemToDto,
	}

	readOnlyApi := Api[TestItem, TestItemDto]{
		Path:        "test3",
		Find:        fullApi.Find,
		FindAll:     fullApi.FindAll,
		Mutate:      nil,
		Create:      nil,
		Delete:      nil,
		SubEntities: nil,
		Validator:   nil,
		Dto:         ItemToDto,
	}

	RegisterAPI(app, fullApi)
	RegisterAPI(app, readOnlyApi)
	RegisterAPI(app, editOnlyApi)

	return app, &data

}

func cleanup(app *fiber.App) {
	_ = app.Shutdown()
}

func TestGetAll(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		// Not logged in should give 401
		data.permit = false
		code, resp, err := util.GetJsonSliceRequestResponse(app, "GET", "/test/", nil)
		assert.Equal(t, 401, code)

		// Logged in should give us 1
		data.permit = true
		code, resp, err = util.GetJsonSliceRequestResponse(app, "GET", "/test", nil)
		assert.Nil(t, err)
		assert.Equal(t, 200, code)
		assert.Len(t, resp, 2)
		assert.NotNil(t, resp[0]["Id"])
		assert.NotNil(t, resp[0]["Data"])

	})
}

func TestGetAllEditOnly(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		// Not logged in should give 401
		data.permit = false
		code, resp, err := util.GetJsonSliceRequestResponse(app, "GET", "/test2/", nil)
		assert.Equal(t, 401, code)

		// Logged in should give us 1
		data.permit = true
		code, resp, err = util.GetJsonSliceRequestResponse(app, "GET", "/test2/", nil)
		assert.Nil(t, err)
		assert.Equal(t, 200, code)
		assert.Len(t, resp, 2)
	})
}

func TestGetAllReadOnly(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		// No validator in should give 200
		data.permit = false
		code, resp, err := util.GetJsonSliceRequestResponse(app, "GET", "/test3/", nil)
		assert.Nil(t, err)
		assert.Equal(t, 200, code)
		assert.Len(t, resp, 2)
	})
}

func TestGetOne(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		// Not logged in should give 401
		data.permit = false
		code, resp, err := util.GetJsonRequestResponse(app, "GET", "/test/id1", nil)
		assert.Equal(t, 401, code)

		// Logged in should give us 404
		data.permit = true
		code, resp, err = util.GetJsonRequestResponse(app, "GET", "/test/id-not-found", nil)
		assert.Equal(t, 404, code)

		// Not found and not permitted should return access denied
		data.permit = false
		code, resp, err = util.GetJsonRequestResponse(app, "GET", "/test/id-not-found", nil)
		assert.Equal(t, 401, code)

		data.permit = true
		// Logged in should give us 1
		code, resp, err = util.GetJsonRequestResponse(app, "GET", "/test/id1", nil)
		assert.Nil(t, err)
		assert.Equal(t, 200, code)
		assert.Equal(t, resp["Id"], "id1")
	})
}

func TestGetOneReadOnly(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		// Not logged in should give 200 - no perms
		data.permit = false
		code, resp, err := util.GetJsonRequestResponse(app, "GET", "/test3/id2", nil)
		assert.Nil(t, err)
		assert.Equal(t, 200, code)
		assert.Equal(t, resp["Id"], "id2")

		// Logged in should give us 404
		data.permit = true
		code, resp, err = util.GetJsonRequestResponse(app, "GET", "/test2/id-missing", nil)
		assert.Equal(t, 404, code)

	})
}

func TestGetChildren(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		// Not logged in should give 401
		data.permit = false
		code, resp, err := util.GetJsonSliceRequestResponse(app, "GET", "/test/id1/children", nil)
		assert.Equal(t, 401, code)

		// Logged in should give us 404
		data.permit = true
		code, resp, err = util.GetJsonSliceRequestResponse(app, "GET", "/test/idnotfound/children", nil)
		assert.Equal(t, 404, code)

		data.permit = false
		code, resp, err = util.GetJsonSliceRequestResponse(app, "GET", "/test/idnotfound/children", nil)
		assert.Equal(t, 401, code)

		// Logged in should give us 1
		data.permit = true
		code, resp, err = util.GetJsonSliceRequestResponse(app, "GET", "/test/id1/children", nil)
		assert.Nil(t, err)
		assert.Equal(t, 200, code)
		assert.Len(t, resp, 2)
		assert.Equal(t, "a", resp[0]["Name"])
		assert.Equal(t, "b", resp[1]["Name"])
	})

}

func TestGetChildrenNotProvided(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		// Not logged in should give 401
		data.permit = false
		code, _, _ := util.GetJsonSliceRequestResponse(app, "GET", "/test2/id1/children", nil)
		assert.Equal(t, 404, code)

		// Logged in should give us 404
		data.permit = true
		code, _, _ = util.GetJsonSliceRequestResponse(app, "GET", "/test2/id1/children", nil)
		assert.Equal(t, 404, code)
	})

}

func TestSaveOne(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		// Not logged in should give 401
		data.permit = false
		code, resp, err := util.GetJsonRequestResponse(app, "PUT", "/test/id1", TestItemDto{
			Id:   "id1",
			Data: "some new data",
		})
		assert.Equal(t, 401, code)

		// Logged in should save a new one
		data.permit = true
		// Logged in should save us 1
		code, resp, err = util.GetJsonRequestResponse(app, "PUT", "/test/id1", TestItemDto{
			Id:   "id1",
			Data: "some new data",
		})
		assert.Nil(t, err)
		assert.Equal(t, 200, code)

		newItem := data.entries["id1"]
		assert.NotNil(t, resp)
		assert.Equal(t, "some new data", resp["Data"])
		assert.Equal(t, "some new data", newItem.Data)

		data.permit = true
		data.fail = true
		code, resp, err = util.GetJsonRequestResponse(app, "PUT", "/test/id1", TestItemDto{
			Id:   "id1",
			Data: "some new data",
		})
		assert.Equal(t, 500, code)

	})
}

func TestAddOne(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		data.permit = false
		code, resp, _ := util.GetJsonRequestResponse(app, "POST", "/test/", TestItemDto{
			Id:   "idnew",
			Data: "some data",
		})
		assert.Equal(t, 401, code)

		data.permit = true
		code, resp, _ = util.GetJsonRequestResponse(app, "POST", "/test", TestItemDto{
			Id:   "idnew",
			Data: "some data",
		})
		assert.Equal(t, 200, code)
		newItem := data.entries["idnew"]
		assert.NotNil(t, resp)
		assert.Equal(t, "idnew", resp["Id"])
		assert.Equal(t, "some data", resp["Data"])
		assert.Equal(t, "some data", newItem.Data)

		data.permit = true
		data.fail = true
		code, resp, _ = util.GetJsonRequestResponse(app, "POST", "/test", TestItemDto{
			Id:   "idnew2",
			Data: "some data",
		})
		assert.Equal(t, 500, code)

	})
}

func TestMutateMissingNoPerms(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		data.permit = false
		code, _, _ := util.GetJsonRequestResponse(app, "PUT", "/test/idmissing", TestItemDto{
			Id:   "id1",
			Data: "New Data",
		})
		assert.Equal(t, 401, code)
	})
}

func TestMutateOneBadBody(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		data.permit = true
		code, _, _ := util.GetJsonRequestResponse(app, "PUT", "/test/id1", "just a string")
		assert.Equal(t, 400, code)
	})
}

func TestAddOneBadBody(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		data.permit = true
		code, _, _ := util.GetJsonRequestResponse(app, "POST", "/test", "just a string")
		assert.Equal(t, 400, code)
	})
}

func TestSaveOneSaveOneReadOnly(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		// Not logged in should give 401
		data.permit = false
		code, _, _ := util.GetJsonRequestResponse(app, "PUT", "/test/id1", TestItemDto{
			Id:   "id1",
			Data: "some new data",
		})
		assert.Equal(t, 401, code)

		// Logged in should save a new one
		data.permit = true
		// Logged in should save us 1
		code, _, _ = util.GetJsonRequestResponse(app, "PUT", "/test3/id1", TestItemDto{
			Id:   "id1",
			Data: "some new data",
		})
		assert.Equal(t, 405, code)

		code, _, _ = util.GetJsonRequestResponse(app, "PUT", "/test3/idnew", TestItemDto{
			Id:   "idnew",
			Data: "some data",
		})
		assert.Equal(t, 405, code)
	})
}

func TestSaveOneSaveOneEditOnly(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		// Not logged in should give 401
		data.permit = false
		code, resp, err := util.GetJsonRequestResponse(app, "PUT", "/test2/id1", TestItemDto{
			Id:   "id1",
			Data: "some new data",
		})
		assert.Equal(t, 401, code)

		// Logged in should save a new one
		data.permit = true
		// Logged in should save us 1
		code, resp, err = util.GetJsonRequestResponse(app, "PUT", "/test2/id1", TestItemDto{
			Id:   "id1",
			Data: "some new data",
		})
		assert.Nil(t, err)
		assert.Equal(t, 200, code)

		newItem := data.entries["id1"]
		assert.NotNil(t, resp)
		assert.Equal(t, "some new data", resp["Data"])
		assert.Equal(t, "some new data", newItem.Data)

		code, resp, err = util.GetJsonRequestResponse(app, "PUT", "/test2/idnew", TestItemDto{
			Id:   "idnew",
			Data: "some data",
		})
		assert.Equal(t, 404, code)
	})
}

func TestRemoveOne(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		// Not logged in should give 401
		data.permit = false
		code, resp, err := util.GetStringRequestResponse(app, "DELETE", "/test/id1", "")
		assert.Nil(t, err)
		assert.Equal(t, 401, code)

		// Logged in should save a new one
		data.permit = true
		code, resp, err = util.GetStringRequestResponse(app, "DELETE", "/test/id1", "")
		assert.Nil(t, err)
		assert.Equal(t, 200, code)
		assert.Contains(t, resp, "deleted")
		_, ok := data.entries["id"]
		assert.False(t, ok)

		code, resp, err = util.GetStringRequestResponse(app, "DELETE", "/test/id1", "")
		assert.Equal(t, 404, code)

		// Missing should return 401 unauthorised if not found and not permitted
		data.permit = false
		code, resp, err = util.GetStringRequestResponse(app, "DELETE", "/test/id1", "")
		assert.Equal(t, 401, code)

		data.permit = true
		data.fail = true
		code, resp, err = util.GetStringRequestResponse(app, "DELETE", "/test/id2", "")
		assert.Equal(t, 500, code)

	})

}

func TestRemoveEditOnly(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		// Not logged in should give 401
		data.permit = false
		code, _, err := util.GetStringRequestResponse(app, "DELETE", "/test2/id1", "")
		assert.Nil(t, err)
		assert.Equal(t, 405, code)

		// Logged in should save a new one
		data.permit = true
		code, _, _ = util.GetStringRequestResponse(app, "DELETE", "/test2/id1", "")
		assert.Nil(t, err)
		assert.Equal(t, 405, code)

	})

}

func TestRemoveReadOnly(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)

		// Not logged in should give 401
		data.permit = false
		code, _, err := util.GetStringRequestResponse(app, "DELETE", "/test3/id1", "")
		assert.Nil(t, err)
		assert.Equal(t, 405, code)

		// Logged in should save a new one
		data.permit = true
		code, _, _ = util.GetStringRequestResponse(app, "DELETE", "/test3/id1", "")
		assert.Nil(t, err)
		assert.Equal(t, 405, code)

	})

}

func TestFilter(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)
		data.permit = false
		code, resp, err := util.GetJsonSliceRequestResponse(app, "POST", "/test/filter", TestItemDto{Data: "data"})
		assert.Equal(t, 401, code)
		data.permit = true

		code, resp, err = util.GetJsonSliceRequestResponse(app, "POST", "/test/filter", TestItemDto{Data: "data"})
		assert.Equal(t, 200, code)
		assert.Nil(t, err)
		assert.Len(t, resp, 2)

		code, resp, err = util.GetJsonSliceRequestResponse(app, "POST", "/test/filter", TestItemDto{Data: "data2"})
		assert.Equal(t, 200, code)
		assert.Nil(t, err)
		assert.Len(t, resp, 1)

	})
}

func TestFilterBadBody(t *testing.T) {
	assert.NotPanics(t, func() {
		app, data := setup()
		defer cleanup(app)
		data.permit = true
		code, _, _ := util.GetStringSliceRequestResponse(app, "POST", "/test/filter", "")
		assert.Equal(t, 400, code)
		data.permit = true

	})
}
