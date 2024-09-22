package sweeper

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"go.charbar.io/gomts"
)

// Sweeper is responsible for cleaning up temporary or test resources.
type Sweeper struct {
	c gomts.Client

	logr *slog.Logger

	// mtx protects the following resources
	mtx           *sync.Mutex
	employeeIDs   []string
	departmentIDs []string
}

// NewSweeper creates a new Sweeper backed by the given client.
func NewSweeper(client gomts.Client, logger *slog.Logger) *Sweeper {
	return &Sweeper{
		c:    client,
		mtx:  new(sync.Mutex),
		logr: logger.WithGroup("sweeper"),
	}
}

// CollectWithPrefix collects all resources prefixed with names prefixed by
// the given string and slates them for deletion.
func (s *Sweeper) CollectWithPrefix(ctx context.Context, prefix string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// add employees for deletion
	employees, err := s.c.Employees().List(ctx)
	if err != nil {
		return err
	}

	for _, employee := range employees {
		if strings.HasPrefix(employee.Name, prefix) {
			s.employeeIDs = append(s.employeeIDs, employee.ID)
		}
	}

	// add departments for deletion
	departments, err := s.c.Departments().List(ctx)
	if err != nil {
		return err
	}

	for _, department := range departments {
		if strings.HasPrefix(department.Name, prefix) {
			s.departmentIDs = append(s.departmentIDs, department.ID)
		}
	}

	return nil
}

// Sweep cleans up all resources slated for deletion.
// Any individual errors are rolled up into an gomts.ErrorList and returned.
func (s *Sweeper) Sweep(ctx context.Context) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	var errList gomts.ErrorList

	// delete all employees
	for _, id := range s.employeeIDs {
		if _, err := s.c.Employees().Delete(ctx, id); err != nil {
			errList = append(errList, err)
		}

		s.logr.InfoContext(ctx, "deleted employee", slog.Any("employee_id", id))
	}

	// delete all departments
	for _, id := range s.departmentIDs {
		if _, err := s.c.Departments().Delete(ctx, id); err != nil {
			errList = append(errList, err)
		}

		s.logr.InfoContext(ctx, "deleted department", slog.Any("department_id", id))
	}

	if len(errList) == 0 {
		return nil
	}

	return errList
}

// AddEmployee adds an employee to be deleted.
func (s *Sweeper) AddEmployee(id string) {
	s.employeeIDs = append(s.employeeIDs, id)
}

// AddDepartment adds a department to be deleted.
func (s *Sweeper) AddDepartment(id string) {
	s.departmentIDs = append(s.departmentIDs, id)
}
