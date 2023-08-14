package tests_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	handler "github.com/abtinokhovat/file-handler-go"
)

var (
	data1 = testStructOne{
		Prop1: 123,
		Prop2: "Testing 123",
		Prop3: []string{"apple", "banana", "cherry"},
		Prop4: map[string]int{"one": 1, "two": 2, "three": 3},
		Prop5: 3.14159265359,
		Prop6: []int{5, 10, 15, 20},
	}

	data2 = testStructOne{
		Prop1: -987,
		Prop2: "Hello, Golang!",
		Prop3: []string{"dog", "cat", "fish"},
		Prop4: map[string]int{"red": 255, "green": 128, "blue": 0},
		Prop5: 2.71828182845,
		Prop6: []int{99, 88, 77},
	}

	data3 = testStructOne{
		Prop1: 0,
		Prop2: "Empty Data",
		Prop3: []string{},
		Prop4: map[string]int{},
		Prop5: 0.0,
		Prop6: []int{},
	}

	testCases = []struct {
		name string
		data testStructOne
	}{
		{name: "Data 1", data: data1},
		{name: "Data 2", data: data2},
		{name: "Empty Values", data: data3},
	}
)

type testStructOne struct {
	Prop1 int            `json:"prop_1"`
	Prop2 string         `json:"prop_2"`
	Prop3 []string       `json:"prop_3"`
	Prop4 map[string]int `json:"prop_4"`
	Prop5 float64        `json:"prop_5"`
	Prop6 []int          `json:"prop_6"`
}

type testStructTwo struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestSerializer_Serialize(t *testing.T) {
	s := handler.NewJsonSerializer[testStructOne]()

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)

	for _, tc := range testCases {
		_ = encoder.Encode(tc.data)
		jsonString, err := s.Serialize(tc.data)
		if err != nil {
			t.Errorf("Error while serializing:\n %s", err)
		}

		if buffer.String() != jsonString {
			t.Errorf("Expected %+v, got %+v", jsonString, buffer.String())
		}

		buffer.Reset()
	}

}

func TestJsonSerializer_Deserialize(t *testing.T) {
	jsonStr := `[{"name": "test1", "value": 42}, {"name": "test2", "value": 24}]`

	serializer := handler.NewJsonSerializer[testStructTwo]()

	result, err := serializer.Deserialize(jsonStr)
	if err != nil {
		t.Errorf("Error during deserialization: %v", err)
		return
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 items, but got %d", len(result))
		return
	}

	expectedData := []testStructTwo{
		{Name: "test1", Value: 42},
		{Name: "test2", Value: 24},
	}

	for i, expected := range expectedData {
		if result[i] != expected {
			t.Errorf("Data mismatch at index %d, expected %v, got %v", i, expected, result[i])
		}
	}
}

func TestJsonSerializer_Deserialize_Error(t *testing.T) {
	jsonStr := `invalid-json`

	serializer := handler.NewJsonSerializer[testStructTwo]()

	_, err := serializer.Deserialize(jsonStr)
	if err == nil {
		t.Error("Expected an error, but got nil")
		return
	}

	// Optionally check the specific error here
	var syntaxError *json.SyntaxError
	if !errors.As(err, &syntaxError) {
		t.Errorf("Expected a syntax error, but got %T", err)
	}
}
