package jsonrpc_client

import (
	"encoding/json"
	"testing"
)

func TestNewID(t *testing.T) {
	t.Run("string ID", func(t *testing.T) {
		id := NewID("test-id")
		if id.String == nil || *id.String != "test-id" {
			t.Errorf("expected ID: test-id, got: %v", id)
		}
		if id.Int != nil {
			t.Errorf("Int value is not nil: %v", *id.Int)
		}
	})

	t.Run("integer ID", func(t *testing.T) {
		id := NewID(42)
		if id.Int == nil || *id.Int != 42 {
			t.Errorf("expected ID: 42, got: %v", id)
		}
		if id.String != nil {
			t.Errorf("String value is not nil: %v", *id.String)
		}
	})

	t.Run("int32 ID", func(t *testing.T) {
		var val int32 = 42
		id := NewID(val)
		if id.Int == nil || *id.Int != 42 {
			t.Errorf("expected ID: 42, got: %v", id)
		}
		if id.String != nil {
			t.Errorf("String value is not nil: %v", *id.String)
		}
	})

	t.Run("uint32 ID", func(t *testing.T) {
		var val uint32 = 42
		id := NewID(val)
		if id.Int == nil || *id.Int != 42 {
			t.Errorf("expected ID: 42, got: %v", id)
		}
		if id.String != nil {
			t.Errorf("String value is not nil: %v", *id.String)
		}
	})

	t.Run("zero values", func(t *testing.T) {
		// Test with zero string
		id := NewID("")
		if id.String == nil || *id.String != "" {
			t.Errorf("expected empty string ID, got: %v", id)
		}

		// Test with zero int
		id = NewID(0)
		if id.Int == nil || *id.Int != 0 {
			t.Errorf("expected zero int ID, got: %v", id)
		}
	})
}

func TestJsonrpcIDNew(t *testing.T) {
	id := &IDValue{String: new(string)}
	*id.String = "test-id"

	newID := id.New()
	if newID == nil {
		t.Fatal("new ID is nil")
	}

	// New ID should be empty
	if newID.String != nil {
		t.Errorf("String value is not nil: %v", *newID.String)
	}
	if newID.Int != nil {
		t.Errorf("Int value is not nil: %v", *newID.Int)
	}
}

func TestJsonrpcIDIsZero(t *testing.T) {
	// For string ID
	strID := &IDValue{String: new(string)}
	*strID.String = "test-id"
	if strID.IsZero() {
		t.Error("string ID was evaluated as zero")
	}

	// For integer ID
	intID := &IDValue{Int: new(int)}
	*intID.Int = 42
	if intID.IsZero() {
		t.Error("integer ID was evaluated as zero")
	}

	// For empty ID
	emptyID := &IDValue{}
	if !emptyID.IsZero() {
		t.Error("empty ID was not evaluated as zero")
	}

	// For zero string value
	zeroStrID := &IDValue{String: new(string)}
	*zeroStrID.String = ""
	if zeroStrID.IsZero() {
		t.Error("empty string ID was evaluated as zero, but should not be")
	}

	// For zero int value
	zeroIntID := &IDValue{Int: new(int)}
	*zeroIntID.Int = 0
	if zeroIntID.IsZero() {
		t.Error("zero int ID was evaluated as zero, but should not be")
	}
}

func TestJsonrpcIDEqual(t *testing.T) {
	// Compare same string IDs
	id1 := &IDValue{String: new(string)}
	*id1.String = "test-id"
	id2 := &IDValue{String: new(string)}
	*id2.String = "test-id"
	if !id1.Equal(id2) {
		t.Error("same string IDs are not equal")
	}

	// Compare different string IDs
	id3 := &IDValue{String: new(string)}
	*id3.String = "different-id"
	if id1.Equal(id3) {
		t.Error("different string IDs are considered equal")
	}

	// Compare same integer IDs
	id4 := &IDValue{Int: new(int)}
	*id4.Int = 42
	id5 := &IDValue{Int: new(int)}
	*id5.Int = 42
	if !id4.Equal(id5) {
		t.Error("same integer IDs are not equal")
	}

	// Compare different integer IDs
	id6 := &IDValue{Int: new(int)}
	*id6.Int = 100
	if id4.Equal(id6) {
		t.Error("different integer IDs are considered equal")
	}

	// Compare string ID with integer ID
	if id1.Equal(id4) {
		t.Error("string ID and integer ID are considered equal")
	}

	// Compare with nil
	if id1.Equal(nil) {
		t.Error("ID and nil are considered equal")
	}

	// Compare empty IDs - both are zero values
	emptyID1 := &IDValue{}
	emptyID2 := &IDValue{}
	// According to the Equal implementation, two empty IDs are not equal
	// because they don't have the same type of value (string or int)
	if emptyID1.Equal(emptyID2) {
		t.Error("empty IDs are considered equal, but should not be according to implementation")
	}
}

func TestJsonrpcIDMarshalJSON(t *testing.T) {
	// Serialize string ID
	id1 := &IDValue{String: new(string)}
	*id1.String = "test-id"
	bytes, err := id1.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	expected := `"test-id"`
	if string(bytes) != expected {
		t.Errorf("expected JSON: %s, got: %s", expected, string(bytes))
	}

	// Serialize integer ID
	id2 := &IDValue{Int: new(int)}
	*id2.Int = 42
	bytes, err = id2.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	expected = `42`
	if string(bytes) != expected {
		t.Errorf("expected JSON: %s, got: %s", expected, string(bytes))
	}

	// Serialize empty ID
	id3 := &IDValue{}
	bytes, err = id3.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	expected = `null`
	if string(bytes) != expected {
		t.Errorf("expected JSON: %s, got: %s", expected, string(bytes))
	}

	// Serialize zero string ID
	id4 := &IDValue{String: new(string)}
	*id4.String = ""
	bytes, err = id4.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	expected = `""`
	if string(bytes) != expected {
		t.Errorf("expected JSON: %s, got: %s", expected, string(bytes))
	}

	// Serialize zero int ID
	id5 := &IDValue{Int: new(int)}
	*id5.Int = 0
	bytes, err = id5.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	expected = `0`
	if string(bytes) != expected {
		t.Errorf("expected JSON: %s, got: %s", expected, string(bytes))
	}
}

func TestJsonrpcIDUnmarshalJSON(t *testing.T) {
	// Deserialize string ID
	id1 := &IDValue{}
	err := id1.UnmarshalJSON([]byte(`"test-id"`))
	if err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if id1.String == nil || *id1.String != "test-id" {
		t.Errorf("expected string ID: test-id, got: %v", id1)
	}
	if id1.Int != nil {
		t.Errorf("Int value is not nil: %v", *id1.Int)
	}

	// Deserialize integer ID
	id2 := &IDValue{}
	err = id2.UnmarshalJSON([]byte(`42`))
	if err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if id2.Int == nil || *id2.Int != 42 {
		t.Errorf("expected integer ID: 42, got: %v", id2)
	}
	if id2.String != nil {
		t.Errorf("String value is not nil: %v", *id2.String)
	}

	// Deserialize invalid JSON
	id3 := &IDValue{}
	err = id3.UnmarshalJSON([]byte(`{}`))
	if err == nil {
		t.Error("no error was returned for invalid JSON")
	}

	// Deserialize null - now the implementation handles null specially
	id4 := &IDValue{String: new(string)}
	*id4.String = "test-id"
	err = id4.UnmarshalJSON([]byte(`null`))
	if err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if !id4.IsZero() {
		t.Errorf("ID should be zero after unmarshaling null, got: %v", id4)
	}

	// Deserialize empty string
	id5 := &IDValue{}
	err = id5.UnmarshalJSON([]byte(`""`))
	if err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if id5.String == nil || *id5.String != "" {
		t.Errorf("expected empty string ID, got: %v", id5)
	}

	// Deserialize zero
	id6 := &IDValue{}
	err = id6.UnmarshalJSON([]byte(`0`))
	if err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if id6.Int == nil || *id6.Int != 0 {
		t.Errorf("expected zero int ID, got: %v", id6)
	}
}

func TestIDValueInJSON(t *testing.T) {
	// Using ID in JSONRPCRequest
	req := &JSONRPCRequest{
		Version: "2.0",
		ID:      NewID("test-id"),
		Method:  "test.method",
	}

	// JSON serialization
	bytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// JSON deserialization
	var newReq JSONRPCRequest
	err = json.Unmarshal(bytes, &newReq)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify ID
	if newReq.ID.String == nil || *newReq.ID.String != "test-id" {
		t.Errorf("expected ID: test-id, got: %v", newReq.ID)
	}
}

func TestJSONRPCError(t *testing.T) {
	err := &JSONRPCError{
		Code:    -32600,
		Message: "Invalid Request",
	}

	expected := "JSON-RPC Error -32600: Invalid Request"
	if err.Error() != expected {
		t.Errorf("expected error message: %s, got: %s", expected, err.Error())
	}

	// Error with data
	errWithData := &JSONRPCError{
		Code:    -32602,
		Message: "Invalid params",
		Data:    "Missing required parameter",
	}

	expectedWithData := "JSON-RPC Error -32602: Invalid params"
	if errWithData.Error() != expectedWithData {
		t.Errorf("expected error message: %s, got: %s", expectedWithData, errWithData.Error())
	}
}

func TestIDValue(t *testing.T) {
	t.Run("string ID", func(t *testing.T) {
		id := NewID("test-id")

		// MarshalJSON
		bytes, err := id.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON error: %v", err)
		}

		expected := `"test-id"`
		if string(bytes) != expected {
			t.Errorf("expected JSON: %s, got: %s", expected, string(bytes))
		}

		// UnmarshalJSON
		newID := &IDValue{}
		err = newID.UnmarshalJSON([]byte(`"another-id"`))
		if err != nil {
			t.Fatalf("UnmarshalJSON error: %v", err)
		}

		if newID.String == nil || *newID.String != "another-id" {
			t.Errorf("expected ID: another-id, got: %v", newID)
		}

		// Value method
		value := id.Value()
		if value != "test-id" {
			t.Errorf("expected Value() to return 'test-id', got: %v", value)
		}

		// Equal
		if !id.Equal(NewID("test-id")) {
			t.Error("same string IDs are not equal")
		}

		if id.Equal(NewID("different-id")) {
			t.Error("different string IDs are considered equal")
		}

		if id.Equal(NewID(123)) {
			t.Error("string ID and number ID are considered equal")
		}
	})

	t.Run("number ID", func(t *testing.T) {
		id := NewID(42)

		// MarshalJSON
		bytes, err := id.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON error: %v", err)
		}

		expected := `42`
		if string(bytes) != expected {
			t.Errorf("expected JSON: %s, got: %s", expected, string(bytes))
		}

		// UnmarshalJSON
		newID := &IDValue{}
		err = newID.UnmarshalJSON([]byte(`99`))
		if err != nil {
			t.Fatalf("UnmarshalJSON error: %v", err)
		}

		if newID.Int == nil || *newID.Int != 99 {
			t.Errorf("expected ID: 99, got: %v", newID)
		}

		// Value method
		value := id.Value()
		if value != 42 {
			t.Errorf("expected Value() to return 42, got: %v", value)
		}

		// Equal
		if !id.Equal(NewID(42)) {
			t.Error("same number IDs are not equal")
		}

		if id.Equal(NewID(100)) {
			t.Error("different number IDs are considered equal")
		}

		if id.Equal(NewID("42")) {
			t.Error("number ID and string ID are considered equal")
		}
	})

	t.Run("nil ID", func(t *testing.T) {
		id := &IDValue{}

		// Value method for nil ID
		value := id.Value()
		if value != nil {
			t.Errorf("expected Value() to return nil, got: %v", value)
		}
	})
}

func TestJSONRPCRequest(t *testing.T) {
	req := &JSONRPCRequest{
		Version: "2.0",
		ID:      NewID(1),
		Method:  "test.method",
		Params:  map[string]string{"key": "value"},
	}

	// JSON serialization
	bytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// JSON deserialization
	var newReq JSONRPCRequest
	err = json.Unmarshal(bytes, &newReq)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify values
	if newReq.Version != "2.0" {
		t.Errorf("expected version: 2.0, got: %s", newReq.Version)
	}

	if newReq.Method != "test.method" {
		t.Errorf("expected method: test.method, got: %s", newReq.Method)
	}

	// Verify ID
	if newReq.ID.Int == nil || *newReq.ID.Int != 1 {
		t.Errorf("expected ID: 1, got: %v", newReq.ID)
	}

	// Verify params
	paramsMap, ok := newReq.Params.(map[string]interface{})
	if !ok {
		t.Fatalf("Params type is incorrect: %T", newReq.Params)
	}

	if paramsMap["key"] != "value" {
		t.Errorf("expected parameter value: value, got: %v", paramsMap["key"])
	}

	// Test with nil params
	reqWithNilParams := &JSONRPCRequest{
		Version: "2.0",
		ID:      NewID(1),
		Method:  "test.method",
		Params:  nil,
	}

	bytes, err = json.Marshal(reqWithNilParams)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var newReqWithNilParams JSONRPCRequest
	err = json.Unmarshal(bytes, &newReqWithNilParams)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if newReqWithNilParams.Params != nil {
		t.Errorf("expected nil params, got: %v", newReqWithNilParams.Params)
	}
}

func TestJSONRPCResponse(t *testing.T) {
	// Success response
	resultJSON := json.RawMessage(`{"result":"success"}`)
	resp := &JSONRPCResponse{
		Version: "2.0",
		ID:      NewID(1),
		Result:  resultJSON,
	}

	// JSON serialization
	bytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// JSON deserialization
	var newResp JSONRPCResponse
	err = json.Unmarshal(bytes, &newResp)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify values
	if newResp.Version != "2.0" {
		t.Errorf("expected version: 2.0, got: %s", newResp.Version)
	}

	if newResp.Error != nil {
		t.Errorf("error is not nil: %v", newResp.Error)
	}

	// Error response
	errResp := &JSONRPCResponse{
		Version: "2.0",
		ID:      NewID(2),
		Error: &JSONRPCError{
			Code:    -32600,
			Message: "Invalid Request",
		},
	}

	// JSON serialization
	bytes, err = json.Marshal(errResp)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// JSON deserialization
	var newErrResp JSONRPCResponse
	err = json.Unmarshal(bytes, &newErrResp)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify error
	if newErrResp.Error == nil {
		t.Fatal("error is nil")
	}

	if newErrResp.Error.Code != -32600 {
		t.Errorf("expected error code: -32600, got: %d", newErrResp.Error.Code)
	}

	if newErrResp.Error.Message != "Invalid Request" {
		t.Errorf("expected error message: Invalid Request, got: %s", newErrResp.Error.Message)
	}

	// Test response with both result and error (should be valid JSON but semantically incorrect)
	bothResp := &JSONRPCResponse{
		Version: "2.0",
		ID:      NewID(3),
		Result:  json.RawMessage(`{"result":"success"}`),
		Error: &JSONRPCError{
			Code:    -32600,
			Message: "Invalid Request",
		},
	}

	// JSON serialization
	bytes, err = json.Marshal(bothResp)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// JSON deserialization
	var newBothResp JSONRPCResponse
	err = json.Unmarshal(bytes, &newBothResp)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Both fields should be present (though this is semantically incorrect for JSON-RPC)
	if newBothResp.Result == nil {
		t.Error("result is nil")
	}

	if newBothResp.Error == nil {
		t.Error("error is nil")
	}
}
