package jsonrpc_client

import (
	"encoding/json"
	"fmt"
)

type IDValue struct {
	strVar *string
	intVar *int
}

// NewID creates a new IDValue from a string or integer value
func NewID[T ~string | ~int | ~int32 | ~uint32](id T) *IDValue {
	switch v := any(id).(type) {
	case string:
		return &IDValue{strVar: &v}
	case int:
		intValue := v
		return &IDValue{intVar: &intValue}
	case int32:
		intValue := int(v)
		return &IDValue{intVar: &intValue}
	case uint32:
		intValue := int(v)
		return &IDValue{intVar: &intValue}
	default:
		panic(fmt.Sprintf("unsupported ID type: %T", id))
	}
}

// New creates a new empty instance of jsonrpcID
func (i *IDValue) New() *IDValue {
	return &IDValue{}
}

// String returns the string value of the ID
func (i *IDValue) String() string {
	if i.strVar != nil {
		return *i.strVar
	}
	if i.intVar != nil {
		return fmt.Sprintf("%d", *i.intVar)
	}
	return "null"
}

// IsZero checks if the ID value is zero/empty
func (i *IDValue) IsZero() bool {
	return i.strVar == nil && i.intVar == nil
}

// Value returns the string or integer value of the ID
func (i *IDValue) Value() any {
	if i.strVar != nil {
		return *i.strVar
	}
	if i.intVar != nil {
		return *i.intVar
	}
	return nil
}

// Equal checks if this ID value equals another ID value
func (i *IDValue) Equal(other any) bool {
	if other == nil {
		return false
	}
	switch o := other.(*IDValue); {
	case i.strVar != nil && o.strVar != nil:
		return *i.strVar == *o.strVar
	case i.intVar != nil && o.intVar != nil:
		return *i.intVar == *o.intVar
	default:
		return false
	}
}

// UnmarshalJSON deserializes the ID value from JSON
func (i *IDValue) UnmarshalJSON(bytes []byte) error {
	// Handle null value
	if string(bytes) == "null" {
		i.strVar = nil
		i.intVar = nil
		return nil
	}

	var str string
	if err := json.Unmarshal(bytes, &str); err == nil {
		i.strVar = &str
		return nil
	}

	var intValue int
	if err := json.Unmarshal(bytes, &intValue); err == nil {
		i.intVar = &intValue
		return nil
	}

	return fmt.Errorf("invalid ID format")
}

// MarshalJSON serializes the ID value to JSON
func (i *IDValue) MarshalJSON() ([]byte, error) {
	if i.strVar != nil {
		return json.Marshal(*i.strVar)
	}
	if i.intVar != nil {
		return json.Marshal(*i.intVar)
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
