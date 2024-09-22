[![Go Report Card](https://goreportcard.com/badge/go.charbar.io/gomts)](https://goreportcard.com/report/go.charbar.io/gomts) [![GoDoc](https://godoc.org/go.charbar.io/gomts?status.svg)](https://godoc.org/go.charbar.io/gomts)

# gomts

A [MyTimeStation] client for Go.

## Installation

```shell
go get go.charbar.io/gomts
```

## Documentation

See [godoc] for a complete reference.

Documentation for the [MyTimeStation] API can be found [here](https://www.mytimestation.com/API.asp#).

### Usage

```golang
ctx := context.Background()

client := gomts.NewClient(&gomts.Configuration{
    UserAgent:  "my-custom-user-agent",    // defaults to go.charbar.io/gomts
    Debug:      true,                      // defaults to false
    AuthToken:  os.Getenv("MY_MTS_TOKEN"), // defaults $MY_MTS_TOKEN
    Transport:  myTransport{},             // defaults to http.DefaultTransport
    LogHandler: &myLogHandler{},           // defaults to slog.NewTextHandler
})

employees, err := client.Employees.List(ctx)
if err != nil {
    return nil, err
}

fmt.Println(employees) // []gomts.Employee{gomts.Employee{ID: "emp_12345", Name: "Bob Ross", ...}}
```

[MyTimeStation]: https://mytimestation.com
[godoc]: https://go.charbar.io/gomts

## Development

### Testing

Running the integration tests requires you to have a MyTimeStation API token
and to explicitly set the `GOMTS_INTEGRATION_TESTS` environment variable to a
truthy value.

```shell
MTS_AUTH_TOKEN="XXXXXXXXXXXXXX" \
    GOMTS_INTEGRATION_TESTS=1 \
    go test -v ./...
```

## License

[MIT License]

[MIT License]: ./README.md
[MyTimeStation]: https://mytimestation.com
[godoc]: https://go.charbar.io/gomts
