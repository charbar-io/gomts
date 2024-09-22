package gomts

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

const (
	integrationTestEnvVar = "GOMTS_INTEGRATION_TEST"

	testResourcePrefix = "gomtstest"
)

var shouldRunIntegrationTests bool

func init() {
	shouldRunIntegrationTests, _ = strconv.ParseBool(os.Getenv(integrationTestEnvVar))
}

func TestMain(m *testing.M) {
	if shouldRunIntegrationTests {
		client, _ := testClient()
		sweeper := &testSweeper{c: client}

		defer sweeper.sweep(context.Background())
	}

	m.Run()
}

func integrationTest(t *testing.T) {
	if !shouldRunIntegrationTests {
		t.Skipf("skipping integration test as %q is not truthy", integrationTestEnvVar)
	}
}

func testResourceName(name string) string {
	buff := make([]byte, int(math.Ceil(float64(4)/float64(1.33333333333))))
	rand.Read(buff)
	str := base64.RawURLEncoding.EncodeToString(buff)
	return testResourcePrefix + str[:4] + "-" + name
}

func testClient() (Client, *Config) {
	conf := new(Config)
	conf.Debug = true

	return NewClient(conf), conf
}

type testSweeper struct {
	c    Client
	opts testSweepOptions
}

type testSweepOptions struct {
	employeeIDs   []string
	departmentIDs []string
}

type errorList []error

func (e errorList) Error() string {
	sb := new(strings.Builder)
	sb.WriteString("errors:")

	for _, err := range e {
		sb.WriteString(fmt.Sprintf(" %v;", err))
	}

	return sb.String()
}

func (s *testSweeper) sweep(ctx context.Context) {
	if err := s.collectAll(ctx); err != nil {
		log.Fatalf("failed to collect: %v", err)
	}

	if err := s.deleteAll(ctx); err != nil {
		log.Fatalf("failed to delete resources: %v", err)
	}
}

func (s *testSweeper) collectAll(ctx context.Context) error {
	employees, err := s.c.Employees().List(ctx)
	if err != nil {
		return err
	}

	for _, employee := range employees {
		if strings.HasPrefix(employee.Name, testResourcePrefix) {
			s.opts.employeeIDs = append(s.opts.employeeIDs, employee.ID)
		}
	}

	departments, err := s.c.Departments().List(ctx)
	if err != nil {
		return err
	}

	for _, department := range departments {
		if strings.HasPrefix(department.Name, testResourcePrefix) {
			s.opts.departmentIDs = append(s.opts.departmentIDs, department.ID)
		}
	}

	return nil
}

func (s *testSweeper) deleteAll(ctx context.Context) error {
	var errList errorList

	for _, id := range s.opts.employeeIDs {
		if _, err := s.c.Employees().Delete(ctx, id); err != nil {
			errList = append(errList, err)
		}
	}

	for _, id := range s.opts.departmentIDs {
		if _, err := s.c.Departments().Delete(ctx, id); err != nil {
			errList = append(errList, err)
		}
	}

	if len(errList) == 0 {
		return nil
	}

	return errList
}
