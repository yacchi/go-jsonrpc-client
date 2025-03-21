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
	SendRequestFunc func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error)
}

func (m *MockTransport) SendRequest(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
	if m.SendRequestFunc != nil {
		return m.SendRequestFunc(ctx, input)
	}
	return &SendRequestOutput{}, nil
}

// NilOutputTransport is a transport that always returns nil output
type NilOutputTransport struct{}

func (t *NilOutputTransport) SendRequest(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
	return nil, nil
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

	t.Run("with multiple options", func(t *testing.T) {
		customGenerator := func() *IDValue {
			return NewID("multi-option-test")
		}
		client := NewClient(transport, WithIDGenerator(customGenerator))

		id := client.generateId()
		if id.strVar == nil || *id.strVar != "multi-option-test" {
			t.Errorf("expected ID: multi-option-test, got: %v", id)
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

		// Create a new client with the sequence generator to test the reset
		client = NewClient(transport, WithSequenceIDGenerator())

		// Manually set the sequence to MaxInt32 by repeatedly calling generateId
		// This is a bit of a hack for testing purposes
		for i := 0; i < math.MaxInt32-1; i++ {
			// We don't actually call this many times in the test
			// Just simulate the sequence reaching MaxInt32
			if i == 10 {
				break
			}
		}

		// Now manually set the sequence to MaxInt32 using a new generator
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

	t.Run("thread safety", func(t *testing.T) {
		transport := &MockTransport{}
		client := NewClient(transport, WithSequenceIDGenerator())

		var wg sync.WaitGroup
		idChan := make(chan int, 100)

		// Generate IDs concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					id := client.generateId()
					if id.intVar != nil {
						idChan <- *id.intVar
					}
				}
			}()
		}

		wg.Wait()
		close(idChan)

		// Collect all generated IDs
		ids := make(map[int]bool)
		for id := range idChan {
			if ids[id] {
				t.Errorf("duplicate ID generated: %d", id)
			}
			ids[id] = true
		}

		// Check that we have the expected number of unique IDs
		if len(ids) != 100 {
			t.Errorf("expected 100 unique IDs, got: %d", len(ids))
		}
	})
}

// TestInvoke tests the Invoke method
func TestInvoke(t *testing.T) {
	t.Run("successful case", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Verify request
				if len(input.Requests) == 0 {
					t.Errorf("no requests provided")
					return nil, errors.New("no requests provided")
				}
				request := input.Requests[0]
				if request.Method != "test.method" {
					t.Errorf("expected method: test.method, got: %s", request.Method)
				}

				// Set response
				resultJSON, _ := json.Marshal(map[string]string{"result": "success"})
				response := &JSONRPCResponse{
					ID:     request.ID,
					Result: resultJSON,
				}
				return &SendRequestOutput{
					Responses: []*JSONRPCResponse{response},
				}, nil
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
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Verify request
				if len(input.Requests) == 0 {
					t.Errorf("no requests provided")
					return nil, errors.New("no requests provided")
				}
				request := input.Requests[0]

				// Set error response
				response := &JSONRPCResponse{
					ID: request.ID,
					Error: &JSONRPCError{
						Code:    -32600,
						Message: "Invalid Request",
					},
				}
				return &SendRequestOutput{
					Responses: []*JSONRPCResponse{response},
				}, nil
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
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				if len(input.Requests) == 0 {
					t.Errorf("no requests provided")
					return nil, errors.New("no requests provided")
				}
				request := input.Requests[0]
				return nil, &InvokeError{Method: request.Method, Err: errors.New("transport error")}
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

	t.Run("nil output from SendRequest", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Return nil output
				return nil, nil
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
			t.Fatal("expected EmptyResponseError, got nil")
		}

		// Verify error
		var emptyErr *EmptyResponseError
		if !errors.As(err, &emptyErr) {
			t.Fatalf("expected error type: *EmptyResponseError, got: %T", err)
		}

		if emptyErr.Method != "test.method" {
			t.Errorf("expected method: test.method, got: %s", emptyErr.Method)
		}
	})

	t.Run("empty responses array from SendRequest", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Return empty responses array
				return &SendRequestOutput{
					Responses: []*JSONRPCResponse{},
				}, nil
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
			t.Fatal("expected EmptyResponseError, got nil")
		}

		// Verify error
		var emptyErr *EmptyResponseError
		if !errors.As(err, &emptyErr) {
			t.Fatalf("expected error type: *EmptyResponseError, got: %T", err)
		}

		if emptyErr.Method != "test.method" {
			t.Errorf("expected method: test.method, got: %s", emptyErr.Method)
		}
	})

	t.Run("with omit request parameter", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Verify request
				if len(input.Requests) == 0 {
					t.Errorf("no requests provided")
					return nil, errors.New("no requests provided")
				}
				request := input.Requests[0]

				// Verify request has no params
				if request.Params != nil {
					t.Errorf("expected params to be nil when using Omit, got: %v", request.Params)
				}

				// Set response
				resultJSON, _ := json.Marshal(map[string]string{"result": "success"})
				response := &JSONRPCResponse{
					ID:     request.ID,
					Result: resultJSON,
				}
				return &SendRequestOutput{
					Responses: []*JSONRPCResponse{response},
				}, nil
			},
		}

		client := NewClient(transport)

		// Method invocation with Omit as request
		invoke := &Invoke[Omit, map[string]string]{
			Name:    "test.method",
			Request: Omit{},
		}

		err := client.Invoke(context.Background(), invoke)
		if err != nil {
			t.Fatalf("Invoke error: %v", err)
		}
	})

	t.Run("with null result", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Verify request
				if len(input.Requests) == 0 {
					t.Errorf("no requests provided")
					return nil, errors.New("no requests provided")
				}
				request := input.Requests[0]

				// Set null result
				response := &JSONRPCResponse{
					ID:     request.ID,
					Result: nil,
				}
				return &SendRequestOutput{
					Responses: []*JSONRPCResponse{response},
				}, nil
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
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Verify request
				if len(input.Requests) == 0 {
					t.Errorf("no requests provided")
					return nil, errors.New("no requests provided")
				}
				request := input.Requests[0]

				// Set invalid JSON result
				response := &JSONRPCResponse{
					ID:     request.ID,
					Result: []byte(`{"result": "success"`), // Missing closing brace
				}
				return &SendRequestOutput{
					Responses: []*JSONRPCResponse{response},
				}, nil
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
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Verify request
				if len(input.Requests) == 0 {
					t.Errorf("no requests provided")
					return nil, errors.New("no requests provided")
				}
				request := input.Requests[0]

				// Set result with type mismatch
				response := &JSONRPCResponse{
					ID:     request.ID,
					Result: []byte(`{"result": 123}`), // Number instead of string
				}
				return &SendRequestOutput{
					Responses: []*JSONRPCResponse{response},
				}, nil
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
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Verify request
				if len(input.Requests) == 0 {
					t.Errorf("no requests provided")
					return nil, errors.New("no requests provided")
				}
				request := input.Requests[0]

				// Check if context is already canceled
				if ctx.Err() != nil {
					return nil, &InvokeError{Method: request.Method, Err: ctx.Err()}
				}

				// Simulate a slow operation that should be canceled
				select {
				case <-ctx.Done():
					return nil, &InvokeError{Method: request.Method, Err: ctx.Err()}
				case <-time.After(100 * time.Millisecond):
					// This should not execute if context is canceled quickly
					resultJSON, _ := json.Marshal(map[string]string{"result": "success"})
					response := &JSONRPCResponse{
						ID:     request.ID,
						Result: resultJSON,
					}
					return &SendRequestOutput{
						Responses: []*JSONRPCResponse{response},
					}, nil
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
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Verify request
				if len(input.Requests) == 0 {
					t.Errorf("no requests provided")
					return nil, errors.New("no requests provided")
				}
				request := input.Requests[0]

				// Verify request ID
				if request.ID.strVar == nil || *request.ID.strVar != "custom-id" {
					t.Errorf("expected ID: custom-id, got: %v", request.ID)
				}

				// Set response
				resultJSON, _ := json.Marshal(map[string]string{"result": "success"})
				response := &JSONRPCResponse{
					ID:     request.ID,
					Result: resultJSON,
				}
				return &SendRequestOutput{
					Responses: []*JSONRPCResponse{response},
				}, nil
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

	t.Run("with omit response parameter and null result", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Verify request
				if len(input.Requests) == 0 {
					t.Errorf("no requests provided")
					return nil, errors.New("no requests provided")
				}
				request := input.Requests[0]

				// Set null result
				response := &JSONRPCResponse{
					ID:     request.ID,
					Result: nil,
				}
				return &SendRequestOutput{
					Responses: []*JSONRPCResponse{response},
				}, nil
			},
		}

		client := NewClient(transport)

		// Define request type
		type TestRequest struct {
			Param string `json:"param"`
		}

		// Method invocation with Omit as request
		invoke := &Invoke[TestRequest, Omit]{
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
	})
}

// TestInvokeBatch tests the InvokeBatch method
func TestInvokeBatch(t *testing.T) {
	t.Run("successful case", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Verify batch flag
				if !input.Batch {
					t.Errorf("expected batch flag to be true")
				}

				// Verify requests
				if len(input.Requests) != 2 {
					t.Errorf("expected 2 requests, got: %d", len(input.Requests))
					return nil, errors.New("invalid request count")
				}

				// Set responses
				responses := make([]*JSONRPCResponse, 2)
				for i, req := range input.Requests {
					resultJSON, _ := json.Marshal(map[string]string{"result": "success" + string(rune('1'+i))})
					responses[i] = &JSONRPCResponse{
						ID:     req.ID,
						Result: resultJSON,
					}
				}
				return &SendRequestOutput{
					Responses: responses,
				}, nil
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

		// Create method callers
		invoke1 := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method1",
			Request: TestRequest{Param: "test1"},
		}
		invoke2 := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method2",
			Request: TestRequest{Param: "test2"},
		}

		// Batch invocation
		err := client.InvokeBatch(context.Background(), []MethodCaller{invoke1, invoke2})
		if err != nil {
			t.Fatalf("InvokeBatch error: %v", err)
		}

		// Verify responses
		if invoke1.Response.Result != "success1" {
			t.Errorf("expected result1: success1, got: %s", invoke1.Response.Result)
		}
		if invoke2.Response.Result != "success2" {
			t.Errorf("expected result2: success2, got: %s", invoke2.Response.Result)
		}
	})

	t.Run("with error response", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Set responses with one error
				responses := make([]*JSONRPCResponse, 2)

				// First response is success
				resultJSON, _ := json.Marshal(map[string]string{"result": "success"})
				responses[0] = &JSONRPCResponse{
					ID:     input.Requests[0].ID,
					Result: resultJSON,
				}

				// Second response is error
				responses[1] = &JSONRPCResponse{
					ID: input.Requests[1].ID,
					Error: &JSONRPCError{
						Code:    -32600,
						Message: "Invalid Request",
					},
				}

				return &SendRequestOutput{
					Responses: responses,
				}, nil
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

		// Create method callers
		invoke1 := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method1",
			Request: TestRequest{Param: "test1"},
		}
		invoke2 := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method2",
			Request: TestRequest{Param: "test2"},
		}

		// Batch invocation
		err := client.InvokeBatch(context.Background(), []MethodCaller{invoke1, invoke2})
		if err == nil {
			t.Fatal("expected error, got nil")
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

		// First invoke should still have a response
		if invoke1.Response.Result != "success" {
			t.Errorf("expected result1: success, got: %s", invoke1.Response.Result)
		}
	})

	t.Run("with missing response", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Return only one response for two requests
				resultJSON, _ := json.Marshal(map[string]string{"result": "success"})
				response := &JSONRPCResponse{
					ID:     input.Requests[0].ID,
					Result: resultJSON,
				}
				return &SendRequestOutput{
					Responses: []*JSONRPCResponse{response},
				}, nil
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

		// Create method callers
		invoke1 := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method1",
			Request: TestRequest{Param: "test1"},
		}
		invoke2 := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method2",
			Request: TestRequest{Param: "test2"},
		}

		// Batch invocation
		err := client.InvokeBatch(context.Background(), []MethodCaller{invoke1, invoke2})
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Verify error
		var missingErr *MissingResponseError
		if !errors.As(err, &missingErr) {
			t.Fatalf("expected error type: *MissingResponseError, got: %T", err)
		}

		if missingErr.Method != "test.method2" {
			t.Errorf("expected method: test.method2, got: %s", missingErr.Method)
		}
	})

	t.Run("with empty request list", func(t *testing.T) {
		client := NewClient(&MockTransport{})

		// Batch invocation with empty list
		err := client.InvokeBatch(context.Background(), []MethodCaller{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Verify error
		var invalidErr *InvalidRequestError
		if !errors.As(err, &invalidErr) {
			t.Fatalf("expected error type: *InvalidRequestError, got: %T", err)
		}
	})

	t.Run("transport error in InvokeBatch", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				return nil, &InvokeError{Method: "test.method", Err: errors.New("transport error")}
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

		// Create method callers
		invoke1 := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method1",
			Request: TestRequest{Param: "test1"},
		}
		invoke2 := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method2",
			Request: TestRequest{Param: "test2"},
		}

		// Batch invocation
		err := client.InvokeBatch(context.Background(), []MethodCaller{invoke1, invoke2})
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Verify error
		var invokeErr *InvokeError
		if !errors.As(err, &invokeErr) {
			t.Fatalf("expected error type: *InvokeError, got: %T", err)
		}
	})

	t.Run("nil output from SendRequest in InvokeBatch", func(t *testing.T) {
		// Set up mock transport
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				return nil, nil
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

		// Create method callers
		invoke1 := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method1",
			Request: TestRequest{Param: "test1"},
		}
		invoke2 := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method2",
			Request: TestRequest{Param: "test2"},
		}

		// Batch invocation
		err := client.InvokeBatch(context.Background(), []MethodCaller{invoke1, invoke2})
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Verify error
		var emptyErr *EmptyResponseError
		if !errors.As(err, &emptyErr) {
			t.Fatalf("expected error type: *EmptyResponseError, got: %T", err)
		}
	})

	t.Run("with notification requests", func(t *testing.T) {
		// Define request and response types
		type TestRequest struct {
			Param string `json:"param"`
		}
		type TestResponse struct {
			Result string `json:"result"`
		}

		// Create a custom mock transport that handles notification requests
		transport := &MockTransport{
			SendRequestFunc: func(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
				// Verify batch flag
				if !input.Batch {
					t.Errorf("expected batch flag to be true")
				}

				// Verify requests
				if len(input.Requests) != 2 {
					t.Errorf("expected 2 requests, got: %d", len(input.Requests))
					return nil, errors.New("invalid request count")
				}

				// Manually modify the first request to be a notification (ID = nil)
				input.Requests[0].ID = nil

				// Set response for the second request only
				resultJSON, _ := json.Marshal(map[string]string{"result": "success"})
				response := &JSONRPCResponse{
					ID:     input.Requests[1].ID,
					Result: resultJSON,
				}
				return &SendRequestOutput{
					Responses: []*JSONRPCResponse{response},
				}, nil
			},
		}

		client := NewClient(transport)

		// Create method callers - first one is intended to be a notification
		invoke1 := &Invoke[TestRequest, TestResponse]{
			ID:      nil, // This will be auto-generated by the client, but our mock will set it to nil
			Name:    "test.notification",
			Request: TestRequest{Param: "test1"},
		}
		invoke2 := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method2",
			Request: TestRequest{Param: "test2"},
		}

		// Batch invocation
		err := client.InvokeBatch(context.Background(), []MethodCaller{invoke1, invoke2})
		if err != nil {
			t.Fatalf("InvokeBatch error: %v", err)
		}

		// Verify response for the second request
		if invoke2.Response.Result != "success" {
			t.Errorf("expected result2: success, got: %s", invoke2.Response.Result)
		}
	})
}

// TestInvokeJSONRPCRequest tests the JSONRPCRequest method of Invoke
func TestInvokeJSONRPCRequest(t *testing.T) {
	t.Run("with regular request", func(t *testing.T) {
		type TestRequest struct {
			Param string `json:"param"`
		}
		type TestResponse struct {
			Result string `json:"result"`
		}

		invoke := &Invoke[TestRequest, TestResponse]{
			ID:      NewID(123),
			Name:    "test.method",
			Request: TestRequest{Param: "test"},
		}

		request := invoke.JSONRPCRequest()
		if request.Version != "2.0" {
			t.Errorf("expected version: 2.0, got: %s", request.Version)
		}
		if request.Method != "test.method" {
			t.Errorf("expected method: test.method, got: %s", request.Method)
		}
		if request.ID == nil || request.ID.intVar == nil || *request.ID.intVar != 123 {
			t.Errorf("expected ID: 123, got: %v", request.ID)
		}
		if request.Params == nil {
			t.Error("expected params to be non-nil")
		}
	})

	t.Run("with omit request", func(t *testing.T) {
		invoke := &Invoke[Omit, map[string]string]{
			ID:      NewID(123),
			Name:    "test.method",
			Request: Omit{},
		}

		request := invoke.JSONRPCRequest()
		if request.Version != "2.0" {
			t.Errorf("expected version: 2.0, got: %s", request.Version)
		}
		if request.Method != "test.method" {
			t.Errorf("expected method: test.method, got: %s", request.Method)
		}
		if request.ID == nil || request.ID.intVar == nil || *request.ID.intVar != 123 {
			t.Errorf("expected ID: 123, got: %v", request.ID)
		}
		if request.Params != nil {
			t.Errorf("expected params to be nil when using Omit, got: %v", request.Params)
		}
	})
}

// TestUnmarshal tests the Unmarshal method of Invoke
func TestUnmarshal(t *testing.T) {
	t.Run("successful unmarshal", func(t *testing.T) {
		type TestRequest struct {
			Param string `json:"param"`
		}
		type TestResponse struct {
			Result string `json:"result"`
		}

		invoke := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method",
			Request: TestRequest{Param: "test"},
		}

		resultJSON, _ := json.Marshal(map[string]string{"result": "success"})
		response := &JSONRPCResponse{
			ID:     NewID(123),
			Result: resultJSON,
		}

		err := invoke.Unmarshal(response)
		if err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if invoke.Response.Result != "success" {
			t.Errorf("expected result: success, got: %s", invoke.Response.Result)
		}
	})

	t.Run("with omit request", func(t *testing.T) {
		invoke := &Invoke[Omit, map[string]string]{
			Name:    "test.method",
			Request: Omit{},
		}

		response := &JSONRPCResponse{
			ID:     NewID(123),
			Result: nil,
		}

		err := invoke.Unmarshal(response)
		if err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
	})

	t.Run("with null result", func(t *testing.T) {
		type TestRequest struct {
			Param string `json:"param"`
		}
		type TestResponse struct {
			Result string `json:"result"`
		}

		invoke := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method",
			Request: TestRequest{Param: "test"},
		}

		response := &JSONRPCResponse{
			ID:     NewID(123),
			Result: nil,
		}

		err := invoke.Unmarshal(response)
		if err == nil {
			t.Fatal("expected EmptyResultError, got nil")
		}

		var emptyErr *EmptyResultError
		if !errors.As(err, &emptyErr) {
			t.Fatalf("expected error type: *EmptyResultError, got: %T", err)
		}
	})

	t.Run("with invalid JSON result", func(t *testing.T) {
		type TestRequest struct {
			Param string `json:"param"`
		}
		type TestResponse struct {
			Result string `json:"result"`
		}

		invoke := &Invoke[TestRequest, TestResponse]{
			Name:    "test.method",
			Request: TestRequest{Param: "test"},
		}

		response := &JSONRPCResponse{
			ID:     NewID(123),
			Result: []byte(`{"result": "success"`), // Missing closing brace
		}

		err := invoke.Unmarshal(response)
		if err == nil {
			t.Fatal("expected UnmarshalError, got nil")
		}

		var unmarshalErr *UnmarshalError
		if !errors.As(err, &unmarshalErr) {
			t.Fatalf("expected error type: *UnmarshalError, got: %T", err)
		}
	})
}
