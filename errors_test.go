package jsonrpc_client

import (
	"errors"
	"fmt"
	"testing"
)

func TestInvokeError(t *testing.T) {
	err := &InvokeError{
		Method: "test.method",
		Err:    errors.New("connection error"),
	}

	// Test Error() method
	expected := "rpc: invoke error [test.method]: connection error"
	if err.Error() != expected {
		t.Errorf("expected error message: %s, got: %s", expected, err.Error())
	}

	// Test IsRPCError() method
	if !err.IsRPCError() {
		t.Error("IsRPCError() returned false")
	}

	// Test Unwrap() method
	unwrapped := err.Unwrap()
	if unwrapped == nil || unwrapped.Error() != "connection error" {
		t.Errorf("expected unwrapped error: connection error, got: %v", unwrapped)
	}
}

func TestFunctionError(t *testing.T) {
	err := &FunctionError{
		Method:  "test.method",
		Message: "invalid parameter",
	}

	// Test Error() method
	expected := "rpc: function error [test.method]: invalid parameter"
	if err.Error() != expected {
		t.Errorf("expected error message: %s, got: %s", expected, err.Error())
	}

	// Test IsRPCError() method
	if !err.IsRPCError() {
		t.Error("IsRPCError() returned false")
	}
}

func TestStatusCodeError(t *testing.T) {
	err := &StatusCodeError{
		Method:     "test.method",
		StatusCode: 404,
	}

	// Test Error() method
	expected := "rpc: non-200 status code [test.method]: 404"
	if err.Error() != expected {
		t.Errorf("expected error message: %s, got: %s", expected, err.Error())
	}

	// Test IsRPCError() method
	if !err.IsRPCError() {
		t.Error("IsRPCError() returned false")
	}
}

func TestEmptyPayloadError(t *testing.T) {
	err := &EmptyPayloadError{
		Method: "test.method",
	}

	// Test Error() method
	expected := "rpc: empty payload [test.method]"
	if err.Error() != expected {
		t.Errorf("expected error message: %s, got: %s", expected, err.Error())
	}

	// Test IsRPCError() method
	if !err.IsRPCError() {
		t.Error("IsRPCError() returned false")
	}
}

func TestUnmarshalError(t *testing.T) {
	err := &UnmarshalError{
		Method: "test.method",
		Err:    errors.New("invalid JSON"),
	}

	// Test Error() method
	expected := "rpc: failed to unmarshal response [test.method]: invalid JSON"
	if err.Error() != expected {
		t.Errorf("expected error message: %s, got: %s", expected, err.Error())
	}

	// Test IsRPCError() method
	if !err.IsRPCError() {
		t.Error("IsRPCError() returned false")
	}

	// Test Unwrap() method
	unwrapped := err.Unwrap()
	if unwrapped == nil || unwrapped.Error() != "invalid JSON" {
		t.Errorf("expected unwrapped error: invalid JSON, got: %v", unwrapped)
	}
}

func TestEmptyResultError(t *testing.T) {
	err := &EmptyResultError{
		Method: "test.method",
	}

	// Test Error() method
	expected := "rpc: empty result [test.method]"
	if err.Error() != expected {
		t.Errorf("expected error message: %s, got: %s", expected, err.Error())
	}

	// Test IsRPCError() method
	if !err.IsRPCError() {
		t.Error("IsRPCError() returned false")
	}
}

func TestMarshalError(t *testing.T) {
	err := &MarshalError{
		Method: "test.method",
		Err:    errors.New("invalid type"),
	}

	// Test Error() method
	expected := "rpc: failed to marshal request [test.method]: invalid type"
	if err.Error() != expected {
		t.Errorf("expected error message: %s, got: %s", expected, err.Error())
	}

	// Test IsRPCError() method
	if !err.IsRPCError() {
		t.Error("IsRPCError() returned false")
	}

	// Test Unwrap() method
	unwrapped := err.Unwrap()
	if unwrapped == nil || unwrapped.Error() != "invalid type" {
		t.Errorf("expected unwrapped error: invalid type, got: %v", unwrapped)
	}
}

func TestRPCError(t *testing.T) {
	// Error without data
	err := &RPCError{
		Method:  "test.method",
		Code:    -32600,
		Message: "Invalid Request",
	}

	// Test Error() method
	expected := "rpc: JSON-RPC error [test.method] code=-32600: Invalid Request"
	if err.Error() != expected {
		t.Errorf("expected error message: %s, got: %s", expected, err.Error())
	}

	// Error with data
	errWithData := &RPCError{
		Method:  "test.method",
		Code:    -32602,
		Message: "Invalid params",
		Data:    "Missing required parameter",
	}

	// Test Error() method
	expectedWithData := "rpc: JSON-RPC error [test.method] code=-32602: Invalid params, data=Missing required parameter"
	if errWithData.Error() != expectedWithData {
		t.Errorf("expected error message: %s, got: %s", expectedWithData, errWithData.Error())
	}

	// Test IsRPCError() method
	if !err.IsRPCError() {
		t.Error("IsRPCError() returned false")
	}
}

func TestIsRPCError(t *testing.T) {
	// For RPC error
	rpcErr := &RPCError{
		Method:  "test.method",
		Code:    -32600,
		Message: "Invalid Request",
	}
	if !IsRPCError(rpcErr) {
		t.Error("RPC error was evaluated as false")
	}

	// For wrapped RPC error
	wrappedErr := fmt.Errorf("wrapped error: %w", rpcErr)
	if !IsRPCError(wrappedErr) {
		t.Error("wrapped RPC error was evaluated as false")
	}

	// For normal error
	normalErr := errors.New("normal error")
	if IsRPCError(normalErr) {
		t.Error("normal error was evaluated as RPC error")
	}

	// For nil
	if IsRPCError(nil) {
		t.Error("nil was evaluated as RPC error")
	}
}
