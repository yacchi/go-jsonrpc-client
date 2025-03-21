package jsonrpc_client

import (
	"errors"
	"fmt"
)

// Error is an interface for RPC errors
type Error interface {
	error
	IsRPCError() bool
}

// InvokeError represents an error that occurs during method invocation
type InvokeError struct {
	Method string
	Err    error
}

// Error returns a string representation of the invoke error
func (e *InvokeError) Error() string {
	return fmt.Sprintf("rpc: invoke error [%s]: %v", e.Method, e.Err)
}

// IsRPCError implements the Error interface
func (e *InvokeError) IsRPCError() bool {
	return true
}

// Unwrap returns the underlying error
func (e *InvokeError) Unwrap() error {
	return e.Err
}

// FunctionError represents an error that occurs inside a function
type FunctionError struct {
	Method  string
	Message string
}

// Error returns a string representation of the function error
func (e *FunctionError) Error() string {
	return fmt.Sprintf("rpc: function error [%s]: %s", e.Method, e.Message)
}

// IsRPCError implements the Error interface
func (e *FunctionError) IsRPCError() bool {
	return true
}

// StatusCodeError represents an error with a non-200 status code
type StatusCodeError struct {
	Method     string
	StatusCode int
}

// Error returns a string representation of the status code error
func (e *StatusCodeError) Error() string {
	return fmt.Sprintf("rpc: non-200 status code [%s]: %d", e.Method, e.StatusCode)
}

// IsRPCError implements the Error interface
func (e *StatusCodeError) IsRPCError() bool {
	return true
}

// EmptyPayloadError represents an error when the payload is empty
type EmptyPayloadError struct {
	Method string
}

// Error returns a string representation of the empty payload error
func (e *EmptyPayloadError) Error() string {
	return fmt.Sprintf("rpc: empty payload [%s]", e.Method)
}

// IsRPCError implements the Error interface
func (e *EmptyPayloadError) IsRPCError() bool {
	return true
}

// UnmarshalError represents an error during JSON deserialization
type UnmarshalError struct {
	Method string
	Err    error
}

// Error returns a string representation of the unmarshal error
func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("rpc: failed to unmarshal response [%s]: %v", e.Method, e.Err)
}

// IsRPCError implements the Error interface
func (e *UnmarshalError) IsRPCError() bool {
	return true
}

// Unwrap returns the underlying error
func (e *UnmarshalError) Unwrap() error {
	return e.Err
}

// EmptyResultError represents an error when the result is empty
type EmptyResultError struct {
	Method string
}

// Error returns a string representation of the empty result error
func (e *EmptyResultError) Error() string {
	return fmt.Sprintf("rpc: empty result [%s]", e.Method)
}

// IsRPCError implements the Error interface
func (e *EmptyResultError) IsRPCError() bool {
	return true
}

// MarshalError represents an error during JSON serialization
type MarshalError struct {
	Method string
	Err    error
}

// Error returns a string representation of the marshal error
func (e *MarshalError) Error() string {
	return fmt.Sprintf("rpc: failed to marshal request [%s]: %v", e.Method, e.Err)
}

// IsRPCError implements the Error interface
func (e *MarshalError) IsRPCError() bool {
	return true
}

// Unwrap returns the underlying error
func (e *MarshalError) Unwrap() error {
	return e.Err
}

// RPCError represents an error in a JSON-RPC error response
type RPCError struct {
	Method  string
	Code    int
	Message string
	Data    any
}

// Error returns a string representation of the RPC error
func (e *RPCError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("rpc: JSON-RPC error [%s] code=%d: %s, data=%v", e.Method, e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("rpc: JSON-RPC error [%s] code=%d: %s", e.Method, e.Code, e.Message)
}

// IsRPCError implements the Error interface
func (e *RPCError) IsRPCError() bool {
	return true
}

// InvalidRequestError represents an error when the request is invalid
type InvalidRequestError struct {
	Message string
}

// Error returns a string representation of the invalid request error
func (e *InvalidRequestError) Error() string {
	return fmt.Sprintf("rpc: invalid request: %s", e.Message)
}

// IsRPCError implements the Error interface
func (e *InvalidRequestError) IsRPCError() bool {
	return true
}

// EmptyResponseError represents an error when no response is received
type EmptyResponseError struct {
	Method string
}

// Error returns a string representation of the empty response error
func (e *EmptyResponseError) Error() string {
	return fmt.Sprintf("rpc: empty response [%s]", e.Method)
}

// IsRPCError implements the Error interface
func (e *EmptyResponseError) IsRPCError() bool {
	return true
}

// MissingResponseError represents an error when a response is missing for a request
type MissingResponseError struct {
	Method string
}

// Error returns a string representation of the missing response error
func (e *MissingResponseError) Error() string {
	return fmt.Sprintf("rpc: missing response for method [%s]", e.Method)
}

// IsRPCError implements the Error interface
func (e *MissingResponseError) IsRPCError() bool {
	return true
}

// IsRPCError determines if the given error is an RPC error
func IsRPCError(err error) bool {
	for err != nil {
		if _, ok := err.(Error); ok {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}
