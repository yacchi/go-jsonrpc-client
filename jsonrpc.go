package jsonrpc_client

import (
	"encoding/json"
	"fmt"
)

type IDValue struct {
	String *string
	Int    *int
}

// NewID creates a new IDValue from a string or integer value
func NewID[T ~string | ~int | ~int32 | ~uint32](id T) *IDValue {
	switch v := any(id).(type) {
	case string:
		return &IDValue{String: &v}
	case int:
		intValue := v
		return &IDValue{Int: &intValue}
	case int32:
		intValue := int(v)
		return &IDValue{Int: &intValue}
	case uint32:
		intValue := int(v)
		return &IDValue{Int: &intValue}
	default:
		panic(fmt.Sprintf("unsupported ID type: %T", id))
	}
}

// New creates a new empty instance of jsonrpcID
func (i *IDValue) New() *IDValue {
	return &IDValue{}
}

// IsZero checks if the ID value is zero/empty
func (i *IDValue) IsZero() bool {
	return i.String == nil && i.Int == nil
}

// Value returns the string or integer value of the ID
func (i *IDValue) Value() any {
	if i.String != nil {
		return *i.String
	}
	if i.Int != nil {
		return *i.Int
	}
	return nil
}

// Equal checks if this ID value equals another ID value
func (i *IDValue) Equal(other any) bool {
	if other == nil {
		return false
	}
	switch o := other.(*IDValue); {
	case i.String != nil && o.String != nil:
		return *i.String == *o.String
	case i.Int != nil && o.Int != nil:
		return *i.Int == *o.Int
	default:
		return false
	}
}

// UnmarshalJSON deserializes the ID value from JSON
func (i *IDValue) UnmarshalJSON(bytes []byte) error {
	// Handle null value
	if string(bytes) == "null" {
		i.String = nil
		i.Int = nil
		return nil
	}

	var str string
	if err := json.Unmarshal(bytes, &str); err == nil {
		i.String = &str
		return nil
	}

	var intValue int
	if err := json.Unmarshal(bytes, &intValue); err == nil {
		i.Int = &intValue
		return nil
	}

	return fmt.Errorf("invalid ID format")
}

// MarshalJSON serializes the ID value to JSON
func (i *IDValue) MarshalJSON() ([]byte, error) {
	if i.String != nil {
		return json.Marshal(*i.String)
	}
	if i.Int != nil {
		return json.Marshal(*i.Int)
	}
	return json.Marshal(nil)
}

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	Version string   `json:"jsonrpc"`
	ID      *IDValue `json:"id,omitzero"`
	Method  string   `json:"method"`
	Params  any      `json:"params,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error returns a string representation of the JSON-RPC error
func (j *JSONRPCError) Error() string {
	return fmt.Sprintf("JSON-RPC Error %d: %s", j.Code, j.Message)
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	Version string          `json:"jsonrpc"`
	ID      *IDValue        `json:"id,omitzero"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}
