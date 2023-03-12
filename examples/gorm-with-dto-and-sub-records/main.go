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
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/pilotso11/go-easyrest/gormrest"
	"github.com/xo/dburl"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// This example implements a more complex structure where DTOs are used and child objects are exposed
// For example:  http://localhost:8080/api/v1/department/Sales/employees to list the employees of the sales department.
// The Location is also exposed on the employee directly http://localhost:8080/api/v1/employees/1
// And Locations are also exposed http://localhost:8080/api/v§/locations/

type Department struct {
	ID        string     `gorm:"primaryKey;uniqueIndex"`
	Employees []Employee `rest:"child"`
}

type DepartmentDto struct {
	ID string
}

type Employee struct {
	gorm.Model
	Name         string
	DepartmentID string // Foreign key
	LocationID   string
	Location     Location
}

type EmployeeDto struct {
	ID           uint
	Name         string
	DepartmentID string
	Location     Location
}

type Location struct {
	Name    string `gorm:"primaryKey;uniqueIndex" rest:"key"`
	Address string
}

func main() {
	app := fiber.New()
	dbUrl := "sqlite:test.db"
	dsn, _ := dburl.Parse(dbUrl)
	db, err := gorm.Open(sqlite.Open(dsn.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = db.AutoMigrate(&Employee{}, &Department{}, &Location{})
	if err != nil {
		log.Fatalf("Gorm migration error: %v", err)
	}

	api := app.Group("/api")
	apiV1 := api.Group("/v1")
	gormrest.RegisterApi(apiV1, db, "employees", gormrest.DefaultOptions[Employee, EmployeeDto]())
	gormrest.RegisterApi(apiV1, db, "departments", gormrest.DefaultOptions[Department, DepartmentDto]())
	gormrest.RegisterApi(apiV1, db, "locations", gormrest.DefaultOptions[Location, Location]())

	// Create some test data
	elm := Location{Name: "Elm", Address: "1 Wall Street"}
	oak := Location{Name: "Oak", Address: "77 Oak Street"}
	db.Save(&elm)
	db.Save(&oak)
	management := Department{ID: "Management", Employees: []Employee{{Name: "Cherry", Location: elm}}}
	sales := Department{ID: "Sales", Employees: []Employee{{Name: "Sandy", Location: oak}, {Name: "Steven", Location: oak}}}
	engineering := Department{ID: "Engineering", Employees: []Employee{{Name: "Emily", Location: elm}, {Name: "Eugene", Location: oak}}}

	db.Save(&management)
	db.Save(&sales)
	db.Save(&engineering)

	err = app.Listen("127.0.0.1:8080")
	if err != nil {
		log.Fatalf("Fiber errror: %v", err)
	}
}
