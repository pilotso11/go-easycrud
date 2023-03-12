package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/pilotso11/go-easycrud/gormcrud"
	"github.com/xo/dburl"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Department struct {
	DepartmentID uint `gorm:"primaryKey"`
}

type Employee struct {
	gorm.Model
	EmployeeNo   int `crud:"key"`
	Name         string
	DepartmentID uint
}

type EmployeeDto struct {
	gorm.Model
	EmployeeNo   int `crud:"key"`
	Name         string
	DepartmentID string
}

// Exposes a CRUD API for "Employee" backed by GORM on http://127.0.0.1:8080/api/v1/employees
// Using the Employee type as both the data object and as the transport object.
// With GET employees/ to get all
// With GET employees/:EmployeeNo to get one
// With PUT employees/:EmployeeNo to change one
// With DELETE employees/:EmployeeNo to delete one
// With POST employees/ to create a new one
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
	gormcrud.RegisterApi(apiV1, db, "employees", gormcrud.DefaultOptions[Employee, Employee]())

	err = app.Listen("127.0.0.1:8080")
	if err != nil {
		log.Fatalf("Fiber errror: %v", err)
	}
}
