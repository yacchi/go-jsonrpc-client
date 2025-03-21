package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	jsonrpc "github.com/yacchi/go-jsonrpc-client"
)

// Define request and response types
type AddParams struct {
	A int `json:"a"`
	B int `json:"b"`
}

type AddResult int

type SubtractParams struct {
	Minuend    int `json:"minuend"`
	Subtrahend int `json:"subtrahend"`
}

type SubtractResult int

type NotifyParams struct {
	Message string `json:"message"`
}

func main() {
	// Create a new HTTP transport pointing to our local server
	transport := jsonrpc.NewHTTPTransport(
		"http://localhost:8080",
		jsonrpc.WithHTTPHeaders(map[string]string{
			"Content-Type": "application/json",
		}),
	)

	// Create a new client with the transport
	client := jsonrpc.NewClient(transport, jsonrpc.WithSequenceIDGenerator())

	fmt.Println("JSON-RPC Client Example")
	fmt.Println("======================")
	fmt.Println("Make sure the server is running with: go run examples/http_server/main.go")
	fmt.Println()

	// Example 1: Simple method invocation
	fmt.Println("Example 1: Simple method invocation")
	simpleInvocation(client)

	// Example 2: Error handling
	fmt.Println("\nExample 2: Error handling")
	errorHandling(client)

	// Example 3: Batch requests
	fmt.Println("\nExample 3: Batch requests")
	batchRequests(client)

	// Example 4: Notifications
	fmt.Println("\nExample 4: Notifications")
	notifications(client)

	// Example 5: Context with timeout
	fmt.Println("\nExample 5: Context with timeout")
	contextWithTimeout(client)
}

func simpleInvocation(client *jsonrpc.Client) {
	// Create a method invocation
	invoke := &jsonrpc.Invoke[AddParams, AddResult]{
		Name:    "add",
		Request: AddParams{A: 10, B: 20},
	}

	fmt.Printf("Calling %s with params: %+v\n", invoke.Name, invoke.Request)

	// Execute the method
	err := client.Invoke(context.Background(), invoke)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Use the response
	fmt.Printf("Result: %d\n", invoke.Response)
}

func errorHandling(client *jsonrpc.Client) {
	// Create a method invocation that will cause an error (division by zero)
	invoke := &jsonrpc.Invoke[AddParams, AddResult]{
		Name:    "divide",
		Request: AddParams{A: 10, B: 0},
	}

	fmt.Printf("Calling %s with params: %+v\n", invoke.Name, invoke.Request)

	// Execute the method
	err := client.Invoke(context.Background(), invoke)
	if err != nil {
		switch {
		case jsonrpc.IsRPCError(err):
			// Handle JSON-RPC protocol errors
			var rpcErr *jsonrpc.RPCError
			if errors.As(err, &rpcErr) {
				fmt.Printf("RPC Error: Code=%d, Message=%s\n", rpcErr.Code, rpcErr.Message)
				if rpcErr.Data != nil {
					fmt.Printf("Error data: %v\n", rpcErr.Data)
				}
			}
		default:
			// Handle other errors
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	// This won't be reached if there's an error
	fmt.Printf("Result: %d\n", invoke.Response)
}

func batchRequests(client *jsonrpc.Client) {
	// Create multiple method invocations
	invoke1 := &jsonrpc.Invoke[AddParams, AddResult]{
		Name:    "add",
		Request: AddParams{A: 10, B: 20},
	}

	invoke2 := &jsonrpc.Invoke[SubtractParams, SubtractResult]{
		Name:    "subtract",
		Request: SubtractParams{Minuend: 30, Subtrahend: 5},
	}

	fmt.Println("Sending batch request:")
	fmt.Printf("1. %s: %+v\n", invoke1.Name, invoke1.Request)
	fmt.Printf("2. %s: %+v\n", invoke2.Name, invoke2.Request)

	// Execute batch request
	err := client.InvokeBatch(context.Background(), []jsonrpc.MethodCaller{invoke1, invoke2})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Use the responses
	fmt.Printf("Result 1: %d\n", invoke1.Response)
	fmt.Printf("Result 2: %d\n", invoke2.Response)
}

func notifications(client *jsonrpc.Client) {
	// Create a notification (no response expected)
	notification := &jsonrpc.Invoke[NotifyParams, jsonrpc.Omit]{
		Name:    "notify",
		Request: NotifyParams{Message: "Hello, world!"},
	}

	fmt.Printf("Sending notification: %s with params: %+v\n", notification.Name, notification.Request)

	// Send notification
	err := client.Invoke(context.Background(), jsonrpc.AsNotification(notification))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Notification sent successfully")
}

func contextWithTimeout(client *jsonrpc.Client) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create a method invocation
	invoke := &jsonrpc.Invoke[AddParams, AddResult]{
		Name:    "add",
		Request: AddParams{A: 10, B: 20},
	}

	fmt.Printf("Calling %s with 2-second timeout and params: %+v\n", invoke.Name, invoke.Request)

	// Execute the method with timeout context
	err := client.Invoke(ctx, invoke)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("Request timed out")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	// Use the response
	fmt.Printf("Result: %d\n", invoke.Response)
}
