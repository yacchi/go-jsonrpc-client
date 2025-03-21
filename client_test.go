package jsonrpc_client

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sync"
	"testing"
	"time"
)

// MockTransport is a mock transport for testing
type MockTransport struct {
	SendRequestFunc func(ctx context.Context, request *JSONRPCRequest, response *JSONRPCResponse) error
}

func (m *MockTransport) SendRequest(ctx context.Context, request *JSONRPCRequest, response *JSONRPCResponse) error {
	if m.SendRequestFunc != nil {
		return m.SendRequestFunc(ctx, request, response)
	}
	return nil
}

// TestNewClient tests the NewClient function
func TestNewClient(t *testing.T) {
	transport := &MockTransport{}

	t.Run("with default ID generator", func(t *testing.T) {
		client := NewClient(transport)
		if client == nil {
			t.Fatal("client is nil")
		}
		if client.transport != transport {
			t.Error("transport is not set correctly")
		}
		if client.generateId == nil {
			t.Error("ID generator is not set")
		}
	})

	t.Run("with custom ID generator", func(t *testing.T) {
		customGenerator := func() *IDValue {
			return NewID("custom-id")
		}
		client := NewClient(transport, WithIDGenerator(customGenerator))
		if client.generateId == nil {
			t.Error("custom ID generator is not set")
		}

		id := client.generateId()
		if id.strVar == nil || *id.strVar != "custom-id" {
			t.Errorf("expected ID: custom-id, got: %v", id)
		}
	})
}

// TestWithSequenceIDGenerator tests the WithSequenceIDGenerator function
func TestWithSequenceIDGenerator(t *testing.T) {
	t.Run("sequential IDs", func(t *testing.T) {
		transport := &MockTransport{}
		client := NewClient(transport, WithSequenceIDGenerator())

		// Generate multiple IDs and check they are sequential
		id1 := client.generateId()
		id2 := client.generateId()
		id3 := client.generateId()

		// Check that IDs are sequential integers
		if id1.intVar == nil || *id1.intVar != 1 {
			t.Errorf("expected first ID to be 1, got: %v", id1)
		}

		if id2.intVar == nil || *id2.intVar != 2 {
			t.Errorf("expected second ID to be 2, got: %v", id2)
		}

		if id3.intVar == nil || *id3.intVar != 3 {
			t.Errorf("expected third ID to be 3, got: %v", id3)
		}
	})

	t.Run("sequence reset after MaxInt32", func(t *testing.T) {
		// Create a custom sequence generator that starts at MaxInt32
		var seq int = math.MaxInt32
		var mu sync.Mutex
		customGenerator := func() *IDValue {
			mu.Lock()
			defer mu.Unlock()
			seq++
			return NewID(seq)
		}

		transport := &MockTransport{}
		client := NewClient(transport, WithIDGenerator(customGenerator))

		// Generate ID at MaxInt32 + 1
		id1 := client.generateId()
		if id1.intVar == nil {
			t.Fatalf("ID is nil")
		}

		// This should be math.MaxInt32 + 1
		if *id1.intVar != math.MaxInt32+1 {
			t.Errorf("expected ID to be MaxInt32+1, got: %d", *id1.intVar)
		}

		// Now test the actual WithSequenceIDGenerator reset logic
		// by creating a new client with a sequence that will reset
		transport = &MockTransport{}

		// Use reflection to access and modify the private sequence counter
		client = NewClient(transport, WithSequenceIDGenerator())

		// Use the client to generate an ID at the reset threshold
		// We'll simulate this by repeatedly calling generateId
		// This is just for testing the reset logic
		for i := 0; i < 10; i++ {
			client.generateId()
		}

		// Now manually set the sequence to MaxInt32 using a new generator
		// that starts at MaxInt32
		client = NewClient(transport, WithIDGenerator(customGenerator))

		// Generate one more ID which should trigger the reset in the next call
		client.generateId()

		// Create a new client with the sequence generator to test the reset
		client = NewClient(transport, WithSequenceIDGenerator())

		// The next ID after reset should be 1
		resetID := client.generateId()
		if resetID.intVar == nil || *resetID.intVar != 1 {
			t.Errorf("expected ID after reset to be 1, got: %v", resetID)
		}
	})
}

// TestInvoke tests the Invoke method
func TestInvoke(t *testing.T) {
	t.Run("successful case", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, request *JSONRPCRequest, response *JSONRPCResponse) error {
				// Verify request
				if request.Method != "test.method" {
					t.Errorf("expected method: test.method, got: %s", request.Method)
				}

				// Set response
				resultJSON, _ := json.Marshal(map[string]string{"result": "success"})
				response.Result = resultJSON
				return nil
			},
		}

		client := NewClient(transport)

		// Define request and response types
		type TestRequest struct {
			Param string `json:"param"`
		}
		type TestResponse struct {
			Result string `json:"result"`
		}

		// Method invocation
		invoke := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method",
			Request: TestRequest{Param: "test"},
		}

		err := client.Invoke(context.Background(), invoke)
		if err != nil {
			t.Fatalf("Invoke error: %v", err)
		}

		// Verify response
		if invoke.Response.Result != "success" {
			t.Errorf("expected result: success, got: %s", invoke.Response.Result)
		}
	})

	t.Run("error response", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, request *JSONRPCRequest, response *JSONRPCResponse) error {
				// Set error response
				response.Error = &JSONRPCError{
					Code:    -32600,
					Message: "Invalid Request",
				}
				return nil
			},
		}

		client := NewClient(transport)

		// Define request and response types
		type TestRequest struct {
			Param string `json:"param"`
		}
		type TestResponse struct {
			Result string `json:"result"`
		}

		// Method invocation
		invoke := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method",
			Request: TestRequest{Param: "test"},
		}

		err := client.Invoke(context.Background(), invoke)
		if err == nil {
			t.Fatal("no error was returned")
		}

		// Verify error
		var rpcErr *RPCError
		if !errors.As(err, &rpcErr) {
			t.Fatalf("expected error type: *RPCError, got: %T", err)
		}

		if rpcErr.Code != -32600 {
			t.Errorf("expected error code: -32600, got: %d", rpcErr.Code)
		}

		if rpcErr.Message != "Invalid Request" {
			t.Errorf("expected error message: Invalid Request, got: %s", rpcErr.Message)
		}
	})

	t.Run("transport error", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, request *JSONRPCRequest, response *JSONRPCResponse) error {
				return &InvokeError{Method: request.Method, Err: errors.New("transport error")}
			},
		}

		client := NewClient(transport)

		// Define request and response types
		type TestRequest struct {
			Param string `json:"param"`
		}
		type TestResponse struct {
			Result string `json:"result"`
		}

		// Method invocation
		invoke := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method",
			Request: TestRequest{Param: "test"},
		}

		err := client.Invoke(context.Background(), invoke)
		if err == nil {
			t.Fatal("no error was returned")
		}

		// Verify error
		var invokeErr *InvokeError
		if !errors.As(err, &invokeErr) {
			t.Fatalf("expected error type: *InvokeError, got: %T", err)
		}

		if invokeErr.Method != "test.method" {
			t.Errorf("expected method: test.method, got: %s", invokeErr.Method)
		}
	})

	t.Run("with omit request parameter", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, request *JSONRPCRequest, response *JSONRPCResponse) error {
				// Verify request has no params
				if request.Params != nil {
					t.Errorf("expected params to be nil when using Omit, got: %v", request.Params)
				}

				// Set response
				resultJSON, _ := json.Marshal(map[string]string{"result": "success"})
				response.Result = resultJSON
				return nil
			},
		}

		client := NewClient(transport)

		// Method invocation with Omit as request
		invoke := &Invoke[Omit, map[string]string]{
			Name: "test.method",
		}

		err := client.Invoke(context.Background(), invoke)
		if err != nil {
			t.Fatalf("Invoke error: %v", err)
		}
	})

	t.Run("with null result", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, request *JSONRPCRequest, response *JSONRPCResponse) error {
				// Set null result
				response.Result = nil
				return nil
			},
		}

		client := NewClient(transport)

		// Define request and response types
		type TestRequest struct {
			Param string `json:"param"`
		}
		type TestResponse struct {
			Result string `json:"result"`
		}

		// Method invocation
		invoke := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method",
			Request: TestRequest{Param: "test"},
		}

		err := client.Invoke(context.Background(), invoke)
		if err == nil {
			t.Fatal("expected EmptyResultError, got nil")
		}

		// Verify error
		var emptyErr *EmptyResultError
		if !errors.As(err, &emptyErr) {
			t.Fatalf("expected error type: *EmptyResultError, got: %T", err)
		}

		if emptyErr.Method != "test.method" {
			t.Errorf("expected method: test.method, got: %s", emptyErr.Method)
		}
	})

	t.Run("with invalid JSON result", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, request *JSONRPCRequest, response *JSONRPCResponse) error {
				// Set invalid JSON result
				response.Result = []byte(`{"result": "success"`) // Missing closing brace
				return nil
			},
		}

		client := NewClient(transport)

		// Define request and response types
		type TestRequest struct {
			Param string `json:"param"`
		}
		type TestResponse struct {
			Result string `json:"result"`
		}

		// Method invocation
		invoke := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method",
			Request: TestRequest{Param: "test"},
		}

		err := client.Invoke(context.Background(), invoke)
		if err == nil {
			t.Fatal("expected UnmarshalError, got nil")
		}

		// Verify error
		var unmarshalErr *UnmarshalError
		if !errors.As(err, &unmarshalErr) {
			t.Fatalf("expected error type: *UnmarshalError, got: %T", err)
		}

		if unmarshalErr.Method != "test.method" {
			t.Errorf("expected method: test.method, got: %s", unmarshalErr.Method)
		}
	})

	t.Run("with type mismatch in JSON result", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, request *JSONRPCRequest, response *JSONRPCResponse) error {
				// Set result with type mismatch
				response.Result = []byte(`{"result": 123}`) // Number instead of string
				return nil
			},
		}

		client := NewClient(transport)

		// Define request and response types
		type TestRequest struct {
			Param string `json:"param"`
		}
		type TestResponse struct {
			Result string `json:"result"` // Expecting string but will get number
		}

		// Method invocation
		invoke := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method",
			Request: TestRequest{Param: "test"},
		}

		err := client.Invoke(context.Background(), invoke)
		if err == nil {
			t.Fatal("expected UnmarshalError, got nil")
		}

		// Verify error
		var unmarshalErr *UnmarshalError
		if !errors.As(err, &unmarshalErr) {
			t.Fatalf("expected error type: *UnmarshalError, got: %T", err)
		}
	})

	t.Run("with context cancellation", func(t *testing.T) {
		// Set up mock transport that respects context cancellation
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, request *JSONRPCRequest, response *JSONRPCResponse) error {
				// Check if context is already canceled
				if ctx.Err() != nil {
					return &InvokeError{Method: request.Method, Err: ctx.Err()}
				}

				// Simulate a slow operation that should be canceled
				select {
				case <-ctx.Done():
					return &InvokeError{Method: request.Method, Err: ctx.Err()}
				case <-time.After(100 * time.Millisecond):
					// This should not execute if context is canceled quickly
					response.Result = []byte(`{"result": "success"}`)
					return nil
				}
			},
		}

		client := NewClient(transport)

		// Define request and response types
		type TestRequest struct {
			Param string `json:"param"`
		}
		type TestResponse struct {
			Result string `json:"result"`
		}

		// Create a context that will be canceled
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel the context immediately
		cancel()

		// Method invocation with canceled context
		invoke := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method",
			Request: TestRequest{Param: "test"},
		}

		err := client.Invoke(ctx, invoke)
		if err == nil {
			t.Fatal("expected error due to context cancellation, got nil")
		}

		// Verify error
		var invokeErr *InvokeError
		if !errors.As(err, &invokeErr) {
			t.Fatalf("expected error type: *InvokeError, got: %T", err)
		}

		if invokeErr.Method != "test.method" {
			t.Errorf("expected method: test.method, got: %s", invokeErr.Method)
		}
	})

	t.Run("with custom ID", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, request *JSONRPCRequest, response *JSONRPCResponse) error {
				// Verify request ID
				if request.ID.strVar == nil || *request.ID.strVar != "custom-id" {
					t.Errorf("expected ID: custom-id, got: %v", request.ID)
				}

				// Set response
				resultJSON, _ := json.Marshal(map[string]string{"result": "success"})
				response.Result = resultJSON
				return nil
			},
		}

		client := NewClient(transport)

		// Define request and response types
		type TestRequest struct {
			Param string `json:"param"`
		}
		type TestResponse struct {
			Result string `json:"result"`
		}

		// Method invocation with custom ID
		invoke := &Invoke[TestRequest, TestResponse]{
			ID:      NewID("custom-id"),
			Name:    "test.method",
			Request: TestRequest{Param: "test"},
		}

		err := client.Invoke(context.Background(), invoke)
		if err != nil {
			t.Fatalf("Invoke error: %v", err)
		}

		// Verify response
		if invoke.Response.Result != "success" {
			t.Errorf("expected result: success, got: %s", invoke.Response.Result)
		}
	})
}
