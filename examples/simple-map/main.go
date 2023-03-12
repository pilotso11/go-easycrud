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

package main

import (
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/pilotso11/go-easyrest"
)

// This example shows a simple rest API backed by an in memory map
// The map is exposed at http://localhost:8080/api/v1/data/

type Employee struct {
	Name       string
	Department string
}

type data struct {
	lock    sync.Mutex
	entries map[string]Employee
}

func main() {
	store := data{
		entries: make(map[string]Employee),
	}

	app := fiber.New()
	api := app.Group("/api")
	apiV1 := api.Group("/v1")

	restApi := easyrest.Api[Employee, Employee]{
		Path: "data",
		Find: func(key string) (Employee, bool) {
			store.lock.Lock()
			defer store.lock.Unlock()
			e, ok := store.entries[key]
			return e, ok
		},
		FindAll: func() []Employee {
			store.lock.Lock()
			defer store.lock.Unlock()
			var all []Employee
			for _, e := range store.entries {
				all = append(all, e)
			}
			return all
		},
		Search: func(filter Employee) []Employee {
			store.lock.Lock()
			defer store.lock.Unlock()
			var all []Employee
			for _, e := range store.entries {
				match := true
				if filter.Name > "" && !strings.Contains(e.Name, filter.Name) {
					match = false
				}
				if filter.Department > "" && !strings.Contains(e.Department, filter.Department) {
					match = false
				}
				if match {
					all = append(all, e)
				}
			}
			return all

		},
		Mutate: func(_ Employee, edit Employee) (Employee, error) {
			store.lock.Lock()
			defer store.lock.Unlock()
			store.entries[edit.Name] = edit
			return edit, nil
		},
		Create: func(toAdd Employee) (Employee, error) {
			store.lock.Lock()
			defer store.lock.Unlock()
			store.entries[toAdd.Name] = toAdd
			return toAdd, nil
		},
		Delete: func(toRemove Employee) (Employee, error) {
			store.lock.Lock()
			defer store.lock.Unlock()
			delete(store.entries, toRemove.Name)
			return Employee{}, nil
		},
		Dto: func(employee Employee) Employee {
			return employee
		},
	}

	easyrest.RegisterAPI(apiV1, restApi)

	_ = app.Listen("127.0.0.1:8080")
}
