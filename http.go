package gomts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"

	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
)

var (
	ErrMissingToken = errors.New("missing MyTimeStation API auth token")
)

// mtsTransport implements http.Transport for MyTimeStation API requests.
type mtsTransport struct {
	conf *Config

	// logr is used for logging dumped requests/responses if debug is enabled.
	logr *slog.Logger
}

// getWrappedTransport gets the underlying http.RoundTripper that will be used
// to perform the request (after MTS headers are added) and before the errors
// are coupled.
//
// If not set, http.DefaultTransport is used.
func (t *mtsTransport) getWrappedTransport() http.RoundTripper {
	if t.conf.Transport != nil {
		return t.conf.Transport
	}

	return http.DefaultTransport
}

// RoundTrip implements http.Transport.
func (t *mtsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.conf.GetAuthToken() == "" {
		return nil, ErrMissingToken
	}

	correlationID := uuid.New().String()

	// set user agent
	req.Header.Add("User-Agent", t.conf.GetUserAgent())

	// accept JSON only
	req.Header.Add("Accept", "application/json")

	// dump request if debug is enabled
	if t.conf.Debug {
		t.logRequest(req, correlationID)
	}

	// set basic auth
	req.SetBasicAuth(t.conf.GetAuthToken(), "")

	// perform request
	resp, err := t.getWrappedTransport().RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// dump response if debug is enabled
	if t.conf.Debug {
		t.logResponse(resp, correlationID)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// non 2XX status codes should be mapped to response errors
		return nil, mapResponseToError(resp)
	}

	return resp, nil
}

// mapResponseToError maps a non-2XX http.Response to an *Error.
func mapResponseToError(resp *http.Response) *Error {
	var errResp ErrorResponse

	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&errResp)

	err := errResp.Error

	if err.ErrorCode == 0 {
		err.ErrorCode = resp.StatusCode
	}

	if err.ErrorText == "" {
		err.ErrorText = http.StatusText(err.ErrorCode)
	}

	return &err
}

func (t *mtsTransport) logRequest(req *http.Request, correlationID string) {
	logr := t.logr.With(slog.String("correlationID", correlationID))

	reqBytes, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		// should never happen
		logr.ErrorContext(req.Context(), "failed to dump request", slog.Any("error", err))
	}

	t.logr.DebugContext(req.Context(), "outbound request", slog.String("request", string(reqBytes)))
}

func (t *mtsTransport) logResponse(resp *http.Response, correlationID string) {
	logr := t.logr.With(slog.String("correlationID", correlationID))

	respBytes, err := httputil.DumpResponse(resp, true)
	if err != nil {
		// should never happen
		logr.ErrorContext(resp.Request.Context(), "failed to dump response", slog.Any("error", err))
	}

	t.logr.DebugContext(resp.Request.Context(), "received response", slog.String("r", string(respBytes)))
}

// httpGet makes an HTTP GET request with the given client.
func httpGet[T any](ctx context.Context, c *client, path string) (*T, error) {
	return httpDo[T](ctx, c, http.MethodGet, path, nil)
}

// httpPut makes an HTTP PUT request with the given client.
func httpPut[T any](ctx context.Context, c *client, path string, body any) (*T, error) {
	return httpDo[T](ctx, c, http.MethodPut, path, body)
}

// httpPost makes an HTTP POST request with the given client.
func httpPost[T any](ctx context.Context, c *client, path string, body any) (*T, error) {
	return httpDo[T](ctx, c, http.MethodPost, path, body)
}

// httpDelete makes an HTTP DELETE request with the given client.
func httpDelete[T any](ctx context.Context, c *client, path string) (*T, error) {
	return httpDo[T](ctx, c, http.MethodDelete, path, nil)
}

func httpDo[T any](ctx context.Context, c *client, method, path string, body any) (*T, error) {
	url := c.conf.GetBaseURL() + path

	req, err := newHTTPRequest(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return mapResponseBody[T](c, resp)
}

func newHTTPRequest(ctx context.Context, method, reqURL string, body any) (*http.Request, error) {
	var (
		bodyReader  io.Reader
		contentType string
	)

	if body != nil {
		buf := new(bytes.Buffer)

		if _, ok := body.(formRequest); ok {
			contentType = "application/x-www-form-urlencoded"

			values, err := query.Values(body)
			if err != nil {
				return nil, fmt.Errorf("could not marshal url-form-encoded: %w", err)
			}

			buf.WriteString(values.Encode())
		} else {
			contentType = "application/json"

			if err := json.NewEncoder(buf).Encode(body); err != nil {
				return nil, fmt.Errorf("could not marshal json: %w", err)
			}
		}

		bodyReader = buf
	}

	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("could not build request: %w", err)
	}

	req.Header.Add("Content-Type", contentType)

	return req.WithContext(ctx), nil
}

// mapResponseBody maps resp.Body to type *T.
func mapResponseBody[T any](c *client, resp *http.Response) (*T, error) {
	var out T

	dec := json.NewDecoder(resp.Body)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logr.ErrorContext(resp.Request.Context(), "failed to close response body", slog.Any("error", err))
		}
	}()

	return &out, dec.Decode(&out)
}
