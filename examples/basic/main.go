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

type NotificationParams struct {
	Message string `json:"message"`
}

func main() {
	// Create a new HTTP transport
	transport := jsonrpc.NewHTTPTransport("https://jsonrpc-example-server.com/api")

	// Create a new client with the transport and a sequence ID generator
	client := jsonrpc.NewClient(transport, jsonrpc.WithSequenceIDGenerator())

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

	// In a real application, this would make an actual API call
	// For this example, we'll simulate a successful response
	fmt.Println("Request:", invoke.Name, invoke.Request)
	fmt.Println("This would send a JSON-RPC request to the server")
	fmt.Println("Simulating successful response with result: 30")

	// Execute the method (in a real application)
	// err := client.Invoke(context.Background(), invoke)
	// if err != nil {
	//     fmt.Printf("Error: %v\n", err)
	//     return
	// }

	// Simulate successful response
	invoke.Response = 30

	// Use the response
	fmt.Printf("Result: %d\n", invoke.Response)
}

func errorHandling(client *jsonrpc.Client) {
	// Create a method invocation
	invoke := &jsonrpc.Invoke[AddParams, AddResult]{
		Name:    "divide",
		Request: AddParams{A: 10, B: 0}, // Division by zero will cause an error
	}

	fmt.Println("Request:", invoke.Name, invoke.Request)
	fmt.Println("This would send a JSON-RPC request to the server")
	fmt.Println("Simulating an error response: Division by zero")

	// In a real application, this would make an actual API call
	// For this example, we'll simulate an error response
	// err := client.Invoke(context.Background(), invoke)

	// Simulate RPC error
	err := &jsonrpc.RPCError{
		Method:  "divide",
		Code:    -32603,
		Message: "Internal error: division by zero",
	}

	// Error handling
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

	// This won't be reached in our example
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

	fmt.Println("Batch Request:")
	fmt.Printf("1. %s: %+v\n", invoke1.Name, invoke1.Request)
	fmt.Printf("2. %s: %+v\n", invoke2.Name, invoke2.Request)
	fmt.Println("This would send a batch JSON-RPC request to the server")
	fmt.Println("Simulating successful responses")

	// In a real application, this would make an actual API call
	// For this example, we'll simulate successful responses
	// err := client.InvokeBatch(context.Background(), []jsonrpc.MethodCaller{invoke1, invoke2})
	// if err != nil {
	//     fmt.Printf("Error: %v\n", err)
	//     return
	// }

	// Simulate successful responses
	invoke1.Response = 30
	invoke2.Response = 25

	// Use the responses
	fmt.Printf("Result 1: %d\n", invoke1.Response)
	fmt.Printf("Result 2: %d\n", invoke2.Response)
}

func notifications(client *jsonrpc.Client) {
	// Create a notification (no response expected)
	notification := &jsonrpc.Invoke[NotificationParams, jsonrpc.Omit]{
		Name:    "log",
		Request: NotificationParams{Message: "Hello, world!"},
	}

	fmt.Println("Notification:", notification.Name, notification.Request)
	fmt.Println("This would send a notification JSON-RPC request to the server")
	fmt.Println("No response is expected for notifications")

	// In a real application, this would make an actual API call
	// For this example, we'll just show the concept
	// err := client.Invoke(context.Background(), jsonrpc.AsNotification(notification))
	// if err != nil {
	//     fmt.Printf("Error: %v\n", err)
	//     return
	// }

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

	fmt.Println("Request with 2-second timeout:", invoke.Name, invoke.Request)
	fmt.Println("This would send a JSON-RPC request to the server with a context timeout")

	// In a real application, this would make an actual API call
	// For this example, we'll simulate a successful response
	// err := client.Invoke(ctx, invoke)
	// if err != nil {
	//     if errors.Is(err, context.DeadlineExceeded) {
	//         fmt.Println("Request timed out")
	//         return
	//     }
	//     fmt.Printf("Error: %v\n", err)
	//     return
	// }

	// Use the context in a simulated way to avoid unused variable warning
	select {
	case <-ctx.Done():
		fmt.Println("Context cancelled or timed out")
	default:
		fmt.Println("Context still valid")
	}

	// Simulate successful response
	invoke.Response = 30

	// Use the response
	fmt.Printf("Result: %d\n", invoke.Response)
}
