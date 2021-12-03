package sources

import (
	"encoding/json"
	"testing"
)

func TestCamelCaseMap(t *testing.T) {
	exampleMap := make(map[string]interface{})

	exampleMap["Name"] = "Dylan"
	exampleMap["Nested"] = map[string]string{
		"NestedKeyName":    "Value",
		"NestedAWSAcronym": "Wow",
	}

	i := interface{}(exampleMap)

	camel := CamelCaseMap(i)

	b, _ := json.Marshal(camel)

	expected := `{"name":"Dylan","nested":{"nestedAWSAcronym":"Wow","nestedKeyName":"Value"}}`

	if string(b) != expected {
		t.Fatalf("expected %v got %v", expected, string(b))
	}
}

func TestToAttributesCase(t *testing.T) {
	exampleMap := make(map[string]interface{})

	exampleMap["Name"] = "Dylan"
	exampleMap["Nested"] = map[string]string{
		"NestedKeyName":    "Value",
		"NestedAWSAcronym": "Wow",
	}
	exampleMap["Nil"] = nil

	i := interface{}(exampleMap)

	attrs, err := ToAttributesCase(i)

	if err != nil {
		t.Fatal(err)
	}

	if _, err := attrs.Get("nested"); err != nil {
		t.Error("could not find key nested")
	}

	if _, err := attrs.Get("nil"); err == nil {
		t.Error("expected nil attributes to be removed")
	}
}
