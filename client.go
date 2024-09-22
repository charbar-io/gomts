package gomts

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

const (
	defaultProtocol   = "https"
	defaultUserAgent  = "go.charbar.io/gomts"
	defaultHost       = "api.mytimestation.com"
	defaultAPIVersion = "v1.2"

	authTokenEnvVar = "MTS_AUTH_TOKEN"
)

// NewClient returns a new client with the given config.
func NewClient(conf *Config) Client {
	return newClient(conf)
}

// Client represents client to the MyTimeStation API.
type Client interface {
	// Employees returns the EmployeeClient.
	Employees() EmployeeClient

	Departments() DepartmentClient
}

// Config configures the underlying HTTP client that interfaces with
// MyTimeStation API methods.
type Config struct {
	// Protocol specifies the protocol to use.
	// NB: HTTPS is the only supported scheme for the production API.
	Protocol string

	// UserAgent specifies the value for the User-Agent header.
	UserAgent string

	// Host specifies the host of the MyTimeStation API to use.
	Host string

	// APIVersion specifies the version of the MyTimeStation API to use.
	APIVersion string

	// Debug enables request and response dumping.
	Debug bool

	// Token is the auth token to use for Basic Auth.
	// If not set $MTS_AUTH_TOKEN is used.
	AuthToken string

	// Transport can be specified to implement RoundTrip or the underlying
	// HTTP client. Will be called after the base MTS transport's RoundTrip is
	// called so all headers will be set on the request. Defaults to
	// http.DefaultTransport.
	Transport http.RoundTripper

	// LogHandler can be specified to cutomize the slog.Logger.
	LogHandler *slog.Handler
}

// GetAuthToken gets the configured auth token or the MTS_AUTH_TOKEN
// environment variable.
func (c *Config) GetAuthToken() string {
	if c.AuthToken == "" {
		return os.Getenv(authTokenEnvVar)
	}

	return c.AuthToken
}

// GetUserAgent gets the configured user agent or the default.
func (c *Config) GetUserAgent() string {
	if c.AuthToken == "" {
		return defaultUserAgent
	}

	return c.UserAgent
}

// GetProtocol gets the configured protocol or the default.
func (c *Config) GetProtocol() string {
	if c.Protocol == "" {
		return defaultProtocol
	}

	return c.Protocol
}

// GetAPIVersion gets the configured API version or the default.
func (c *Config) GetAPIVersion() string {
	if c.APIVersion == "" {
		return defaultAPIVersion
	}

	return c.APIVersion

}

// GetHost gets the configured API host or the default.
func (c *Config) GetHost() string {
	if c.Host == "" {
		return defaultHost
	}

	return c.Host
}

// GetBaseURL build the base MyTimeStation API URL from the configured
// protocol, host and API version.
//
// e.g. https://api.mytimestation.com/v1.0
func (c *Config) GetBaseURL() string {
	return fmt.Sprintf("%s://%s/%s",
		c.GetProtocol(),
		c.GetHost(),
		c.GetAPIVersion())
}

// GetLogger returns a *slog.Logger built from the configured slog.Handler or
// builds a default, text-based logger.
//
// Default log level will be `info`. If `debug` is `true`, it will be `debug`.
func (c *Config) GetLogger() *slog.Logger {
	if c.LogHandler != nil {
		// use user-specified log handler
		return slog.New(*c.LogHandler)
	}

	level := slog.LevelInfo.Level()

	if c.Debug {
		// up the log level to enable debug logging
		level = slog.LevelDebug.Level()
	}

	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
}

// GetTransport returns an http.Transport implementation for MyTimeStation
// authentication and request/response dumping.
func (c *Config) GetTransport() *mtsTransport {
	return &mtsTransport{
		conf: c,
		logr: slog.Default(),
	}
}

type client struct {
	conf       *Config
	httpClient *http.Client

	logr *slog.Logger

	departments *departmentClient
	employees   *employeeClient
}

func newClient(conf *Config) *client {
	logr := conf.GetLogger().WithGroup("gomts")

	transport := conf.GetTransport()
	transport.logr = logr.WithGroup("transport")

	httpClient := &http.Client{Transport: transport}

	c := &client{
		conf:       conf,
		logr:       logr,
		httpClient: httpClient,
	}

	c.employees = (*employeeClient)(c)
	c.departments = &departmentClient{c}

	return c
}

// Employees returns the EmployeeClient.
func (c *client) Employees() EmployeeClient {
	return c.employees
}

func (c *client) Departments() DepartmentClient {
	return c.departments
}

type formRequest interface {
	form()
}
