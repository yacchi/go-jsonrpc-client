package jsonrpc_client

import (
	"context"
	"encoding/json"
	"math"
	"sync"
)

// Client represents a JSON-RPC client
type Client struct {
	transport  Transport
	generateId func() *IDValue
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithIDGenerator sets a custom ID generator function for the client
func WithIDGenerator(generateId func() *IDValue) ClientOption {
	return func(c *Client) {
		c.generateId = generateId
	}
}

// WithSequenceIDGenerator sets a sequence-based ID generator for the client
func WithSequenceIDGenerator() ClientOption {
	var seq int
	var mu sync.Mutex
	return WithIDGenerator(func() *IDValue {
		mu.Lock()
		defer mu.Unlock()
		seq++
		if seq > math.MaxInt32 {
			seq = 1
		}
		return NewID(seq)
	})
}

// NewClient creates a new JSON-RPC client
func NewClient(transport Transport, opts ...ClientOption) *Client {
	c := &Client{
		transport: transport,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.generateId == nil {
		WithSequenceIDGenerator()(c)
	}
	return c
}

// MethodCaller is an interface for method invocation
type MethodCaller interface {
	JSONRPCRequest() *JSONRPCRequest
	Unmarshal(resp *JSONRPCResponse) error
}

// Omit is used to indicate that a parameter should be omitted
type Omit struct{}

// Invoke represents method invocation information
type Invoke[Tin any, Tout any] struct {
	ID       *IDValue
	Name     string
	Request  Tin
	Response Tout
}

// JSONRPCRequest generates a JSON-RPC request
func (i *Invoke[Tin, Tout]) JSONRPCRequest() *JSONRPCRequest {
	var params any
	if _, isOmit := any(i.Request).(Omit); !isOmit {
		params = i.Request
	}
	return &JSONRPCRequest{
		Version: "2.0",
		ID:      i.ID,
		Method:  i.Name,
		Params:  params,
	}
}

// Unmarshal decodes a JSON-RPC response
func (i *Invoke[Tin, Tout]) Unmarshal(resp *JSONRPCResponse) error {
	if _, isOmit := any(i.Request).(Omit); isOmit {
		return nil
	}
	if resp.Result == nil {
		return &EmptyResultError{Method: i.Name}
	}
	if err := json.Unmarshal(resp.Result, &i.Response); err != nil {
		return &UnmarshalError{Method: i.Name, Err: err}
	}
	return nil
}

// Invoke calls a method
func (c *Client) Invoke(ctx context.Context, req MethodCaller) error {
	// Get request information
	request := req.JSONRPCRequest()
	if request.ID == nil {
		// Generate a new ID if ID is nil
		request.ID = c.generateId()
	}

	// Send request
	input := &SendRequestInput{
		Requests: []*JSONRPCRequest{request},
		Batch:    false,
	}

	output, err := c.transport.SendRequest(ctx, input)
	if err != nil {
		return err // already wrapped in an appropriate error type
	}

	if output == nil || len(output.Responses) == 0 {
		return &EmptyResponseError{Method: request.Method}
	}

	response := output.Responses[0]

	// Check JSON-RPC error
	if response.Error != nil {
		return &RPCError{
			Method:  request.Method,
			Code:    response.Error.Code,
			Message: response.Error.Message,
			Data:    response.Error.Data,
		}
	}

	// Decode response
	return req.Unmarshal(response)
}

// InvokeBatch calls multiple methods in a batch
func (c *Client) InvokeBatch(ctx context.Context, reqs []MethodCaller) error {
	if len(reqs) == 0 {
		return &InvalidRequestError{Message: "no requests provided"}
	}

	// Prepare requests
	requests := make([]*JSONRPCRequest, len(reqs))
	for i, req := range reqs {
		request := req.JSONRPCRequest()
		// Generate ID if this is not a notification request (ID = nil)
		if request.ID == nil {
			// Generate ID for regular request
			request.ID = c.generateId()
		}
		requests[i] = request
	}

	// Send request
	input := &SendRequestInput{
		Requests: requests,
		Batch:    true,
	}

	output, err := c.transport.SendRequest(ctx, input)
	if err != nil {
		return err
	}

	// Process responses
	if output == nil {
		return &EmptyResponseError{Method: requests[0].Method}
	}
	// Map responses based on ID
	responseMap := make(map[string]*JSONRPCResponse)
	for _, resp := range output.Responses {
		if resp.ID != nil {
			responseMap[resp.ID.String()] = resp
		}
	}

	// Process response for each request
	for i, req := range reqs {
		request := requests[i]
		if request.ID == nil {
			// No response expected for notifications
			continue
		}

		resp, ok := responseMap[request.ID.String()]
		if !ok {
			return &MissingResponseError{Method: request.Method}
		}

		// Check for JSON-RPC error
		if resp.Error != nil {
			return &RPCError{
				Method:  request.Method,
				Code:    resp.Error.Code,
				Message: resp.Error.Message,
				Data:    resp.Error.Data,
			}
		}

		// Decode response
		if err := req.Unmarshal(resp); err != nil {
			return err
		}
	}

	return nil
}
