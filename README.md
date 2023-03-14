# GO Fiber Easy Rest API With GORM
GORM (https://gorm.io) is an amazing DB ORM for go.   It lets you define a CRUD api from just the type structure.
go-easyrest attempts the same, exposing the same GORM object as a REST/CRUD API with just one line of code.

# Installation
`go get github.com/pilotso11/go-easyrest`

# Usage
```go

type Employee struct {
	gorm.Model
	Name       string
	EmployeeNo int
	Department string
}

// connect to your DB
// Create your gorm data model
err = db.AutoMigrate(&Employee{})

// Setup Fiber
app := fiber.New()

// Create the REST API
easyrest.RegisterApi(apiV1, db, "employees", easyrest.DefaultOptions[Employee, Employee]())

fiber.Serve("localhost:8080")

```
With this single line a the following are exposed:
- GET http://localhost:8080/employees - to get all employees
- POST http://localhost:8080/employees/filter - to search for employees (exact match only at the moment)
- POST http://localhost:8080/employees/ - to create a single employee
- GET http://localhost:8080/employees/ID - to get a single employee
- PUT http://localhost:8080/employees/ID - to update a single employee
- DELETE http://localhost:8080/employees/ID - to delete a single employee

More advanced uses allow for custom DTO types, authentication, limit of functions and exposing child object lists. 
Look in [/examples](/examples) .