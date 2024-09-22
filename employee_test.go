package gomts_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.charbar.io/gomts"
)

var numRunes = []rune("1234567890")

func randomPin() string {
	b := make([]rune, 4)
	for i := range b {
		b[i] = numRunes[rand.Intn(len(numRunes))]
	}
	return string(b)
}

func TestEmployeesCreate(t *testing.T) {
	client, _ := integrationTest(t)

	ctx := context.Background()

	dept, err := client.Departments().Create(ctx, &gomts.DepartmentCreateRequest{
		Name: testResourceName("something"),
	})
	assert.NoError(t, err)

	createRequest := &gomts.EmployeeCreateRequest{
		Name:  testResourceName("bob ross"),
		PIN:   randomPin(),
		Title: "Senior Artist",

		DepartmentID: dept.ID,
	}

	newEmployee, err := client.Employees().Create(ctx, createRequest)
	assert.NoError(t, err)

	employee, err := client.Employees().Get(ctx, newEmployee.ID)
	assert.NoError(t, err)

	assert.Equal(t, createRequest.Name, employee.Name)
	assert.Equal(t, createRequest.PIN, employee.PIN)
	assert.Equal(t, createRequest.Title, employee.Title)
	assert.NotEmpty(t, employee.CardNumber)
	assert.NotEmpty(t, employee.CardQRCode)
	assert.NotEmpty(t, employee.PrimaryDepartment)
}
