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

// This example exposes a CRUD API for "Employee" backed by GORM on http://127.0.0.1:8080/api/v1/employees
// Using the Employee type as both the data object and as the transport object.
// With GET employees/ to get all
// With POST employees/filter to search with exact matches ( "like" is not yet implemented )
// With GET employees/:id to get one
// With PUT employees/:id to change one
// With DELETE employees/:id to delete one
// With POST employees/ to create a new one

type Employee struct {
	gorm.Model
	Name       string
	EmployeeNo int
	Department string
}

func main() {
	app := fiber.New()
	dbUrl := "sqlite:test.db"
	dsn, _ := dburl.Parse(dbUrl)
	db, err := gorm.Open(sqlite.Open(dsn.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = db.AutoMigrate(&Employee{})
	if err != nil {
		log.Fatalf("Gorm migration error: %v", err)
	}

	api := app.Group("/api")
	apiV1 := api.Group("/v1")
	gormrest.RegisterApi(apiV1, db, "employees", gormrest.DefaultOptions[Employee, Employee]())

	// Add some test records
	db.Save(&Employee{
		Name:       "Sandra",
		EmployeeNo: 1,
		Department: "CEO",
	})
	db.Save(&Employee{
		Name:       "Simon",
		EmployeeNo: 2,
		Department: "Sales",
	})
	db.Save(&Employee{
		Name:       "Susan",
		EmployeeNo: 3,
		Department: "Engineering",
	})

	err = app.Listen("127.0.0.1:8080")
	if err != nil {
		log.Fatalf("Fiber errror: %v", err)
	}
}
