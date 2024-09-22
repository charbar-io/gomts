package gomts_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.charbar.io/gomts"
	"go.charbar.io/gomts/internal/sweeper"
)

const (
	integrationTestEnvVar = "GOMTS_INTEGRATION_TEST"

	testResourcePrefix = "gomtstest"
)

var shouldRunIntegrationTests bool

func init() {
	shouldRunIntegrationTests, _ = strconv.ParseBool(os.Getenv(integrationTestEnvVar))
}

// integrationTest conditionally sets up integration tests based on
// shouldRunIntegrationTests. If not enabled, the test will be skipped.
//
// Builds a client wrapped with testTransport which records created resources
// and slates them to be deleted by the sweeper on test clean up.
func integrationTest(t *testing.T) (gomts.Client, *gomts.Config) {
	if !shouldRunIntegrationTests {
		t.Skipf("skipping integration test as %q is not truthy", integrationTestEnvVar)
	}

	ctx := context.Background()

	client, conf := testClient()
	sweeper := sweeper.NewSweeper(client, conf.GetLogger())

	conf.Transport = &testTransport{
		logr:    conf.GetLogger().WithGroup("test_transport"),
		sweeper: sweeper,
	}

	t.Cleanup(func() {
		if err := sweeper.Sweep(ctx); err != nil {
			t.Fatalf("failed to clean up integration test resources: %v", err)
		}
	})

	return client, conf
}

// testTransport is used for intercepting request so we can track test
// resources and delete them on exit.
type testTransport struct {
	logr    *slog.Logger
	sweeper *sweeper.Sweeper
}

// RoundTrip implements http.RoundTripper. Any relevant POST requests are
// recorded so test resources can be cleaned up on teardown. See
// integrationTest.
func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if req.Method != http.MethodPost {
		// we only care about resources we've created
		return resp, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// this request failed, so there is nothing to clean up
		return resp, nil
	}

	// if we determine that the path matches a resource type we need to clean
	// up, we need to read the body, get the ID and add it to the sweeper
	// options to be deleted on teardown.

	buf := new(bytes.Buffer)

	if _, err := io.Copy(buf, resp.Body); err != nil {
		t.logr.ErrorContext(req.Context(), "could not copy resp body; resource may leak", slog.Any("error", err))
		return resp, nil
	}

	// replace response for downstream with nop closer
	resp.Body = io.NopCloser(buf)

	var parseErr error

	switch req.URL.Path {
	case "/v1.2/employees":
		var employeeResp gomts.EmployeeResponse
		if parseErr = json.Unmarshal(buf.Bytes(), &employeeResp); err == nil {
			t.sweeper.AddEmployee(employeeResp.Employee.ID)
			t.logr.Info("slated test employee for deletion", slog.Any("employee_id", employeeResp.Employee.ID))
		}

	case "/v1.2/departments":
		var departmentResp gomts.DepartmentResponse
		if parseErr = json.Unmarshal(buf.Bytes(), &departmentResp); err == nil {
			t.sweeper.AddDepartment(departmentResp.Department.ID)
			t.logr.Info("slated test department for deletion", slog.Any("department_id", departmentResp.Department.ID))
		}
	}

	if parseErr != nil {
		t.logr.ErrorContext(req.Context(), "could not unmarshal body; resource may leak", slog.Any("error", err))
	}

	return resp, nil
}

type testLogHandler struct {
	groups []string
}

func (h *testLogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	// log everything
	return true
}

func (h *testLogHandler) Handle(_ context.Context, record slog.Record) error {
	buf := new(bytes.Buffer)

	// format the time
	fmt.Fprint(buf, record.Time.Format(time.RFC3339))
	fmt.Fprint(buf, " ")

	// format the level
	fmt.Fprintf(buf, "[%s] ", record.Level.String())

	// format the groups
	fmt.Fprintf(buf, "[%s] ", strings.Join(h.groups, "::"))

	// Add message
	fmt.Fprint(buf, record.Message)
	fmt.Fprint(buf, " ")

	// Process attributes
	record.Attrs(func(attr slog.Attr) bool {
		strValue := attr.String()

		// unescape new line characters
		strValue = strings.ReplaceAll(strValue, "\\r\\n", "\n")

		fmt.Fprint(buf, strValue)
		fmt.Fprint(buf, " ")
		return true
	})

	// Output to stdout
	fmt.Fprint(os.Stdout, buf.String()+"\n")

	return nil
}

func (h *testLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *testLogHandler) WithGroup(name string) slog.Handler {
	h.groups = append(h.groups, name)
	return h
}

// testClient creates a basic client that depends on the auth token environment
// variable being set.
func testClient() (gomts.Client, *gomts.Config) {
	conf := new(gomts.Config)
	conf.LogHandler = new(testLogHandler)
	return gomts.NewClient(conf), conf
}

// testResourceName generates a unique-ish name for test resources so they can
// be cleaned up later if leaked by failed test teardown.
//
// format: ${PREFIX}${RANDOM_4_DIGITS}-${NAME}
func testResourceName(name string) string {
	buff := make([]byte, int(math.Ceil(float64(4)/float64(1.33333333333))))
	rand.Read(buff)
	str := base64.RawURLEncoding.EncodeToString(buff)
	return testResourcePrefix + str[:4] + "-" + name
}
