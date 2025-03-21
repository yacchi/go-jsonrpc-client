# go-jsonrpc-client

[![Go Reference](https://pkg.go.dev/badge/github.com/yacchi/go-jsonrpc-client.svg)](https://pkg.go.dev/github.com/yacchi/go-jsonrpc-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/yacchi/go-jsonrpc-client)](https://goreportcard.com/report/github.com/yacchi/go-jsonrpc-client)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A type-safe JSON-RPC 2.0 client implementation for Go.

## Features

- Full [JSON-RPC 2.0](https://www.jsonrpc.org/specification) specification support
- Type-safe request and response handling with generics
- No external dependencies - built entirely with Go standard library
- Support for both single requests and batch requests
- Support for notifications (requests without responses)
- Customizable ID generation
- Comprehensive error handling
- HTTP transport with customizable options
- Context support for cancellation and timeouts

## Installation

```bash
go get github.com/yacchi/go-jsonrpc-client
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	jsonrpc "github.com/yacchi/go-jsonrpc-client"
)

func main() {
	// Create a new HTTP transport
	transport := jsonrpc.NewHTTPTransport(
		"https://api.example.com/jsonrpc",
		jsonrpc.WithHTTPHeaders(map[string]string{
			"Authorization": "Bearer token",
		}),
	)

	// Create a new client with the transport
	client := jsonrpc.NewClient(transport)

	// Define request and response types
	type AddParams struct {
		A int `json:"a"`
		B int `json:"b"`
	}
	type AddResult int

	// Create a method invocation
	invoke := &jsonrpc.Invoke[AddParams, AddResult]{
		Name:    "add",
		Request: AddParams{A: 10, B: 20},
	}

	// Execute the method
	err := client.Invoke(context.Background(), invoke)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Use the response
	fmt.Printf("Result: %d\n", invoke.Response)
}
```

## Batch Requests

```go
// Create multiple method invocations
invoke1 := &jsonrpc.Invoke[AddParams, AddResult]{
	Name:    "add",
	Request: AddParams{A: 10, B: 20},
}

invoke2 := &jsonrpc.Invoke[AddParams, AddResult]{
	Name:    "add",
	Request: AddParams{A: 30, B: 40},
}

// Execute batch request
err := client.InvokeBatch(context.Background(), []jsonrpc.MethodCaller{invoke1, invoke2})
if err != nil {
	log.Fatalf("Error: %v", err)
}

// Use the responses
fmt.Printf("Result 1: %d\n", invoke1.Response)
fmt.Printf("Result 2: %d\n", invoke2.Response)
```

## Notifications

```go
// Create a notification (no response expected)
notification := &jsonrpc.Invoke[LogParams, jsonrpc.Omit]{
	Name:    "log",
	Request: LogParams{Message: "Hello, world!"},
}

// Send notification
err := client.Invoke(context.Background(), jsonrpc.AsNotification(notification))
if err != nil {
	log.Fatalf("Error: %v", err)
}
```

## Custom ID Generator

```go
// Create a client with a custom ID generator
client := jsonrpc.NewClient(
	transport,
	jsonrpc.WithIDGenerator(func() *jsonrpc.IDValue {
		return jsonrpc.NewID("custom-id-" + uuid.New().String())
	}),
)
```

## Error Handling

```go
err := client.Invoke(context.Background(), invoke)
if err != nil {
	switch {
	case jsonrpc.IsRPCError(err):
		// Handle JSON-RPC protocol errors
		var rpcErr *jsonrpc.RPCError
		if errors.As(err, &rpcErr) {
			fmt.Printf("RPC Error: Code=%d, Message=%s\n", rpcErr.Code, rpcErr.Message)
		}
	default:
		// Handle other errors
		fmt.Printf("Error: %v\n", err)
	}
	return
}
```

## Documentation

For more detailed documentation, please refer to the [Go Reference](https://pkg.go.dev/github.com/yacchi/go-jsonrpc-client).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
