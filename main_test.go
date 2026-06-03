package main

import (
	"strings"
	"testing"
	
)

var testEnv = Env{
	{name: "b", value: NumV{num_: 2}},
	{name: "a", value: NumV{num_: 1}},
}

func TestEnvLookup(t *testing.T) {
	value, err := envLookup("a", testEnv)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	numValue, isNum := value.(NumV)
	if !isNum || numValue.num_ != 1 {
		t.Fatalf("envLookup('a') = %v, want NumV{num_: 1}", value)
	}
	value, err = envLookup("b", testEnv)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	numValue, isNum = value.(NumV)
	if !isNum || numValue.num_ != 2 {
		t.Fatalf("envLookup('b') = %v, want NumV{num_: 2}", value)
	}
}

func TestEnvLookupMissing(t *testing.T) {
	_, err := envLookup("c", testEnv)
	if err == nil {
		t.Fatal("expected error for unbound name 'c'")
	}
	if !strings.Contains(err.Error(), "value not found") {
		t.Fatalf("expected 'value not found' error, got %v", err)
	}
}

func TestSerialize(t *testing.T) {
	tests := []struct {
		name     string
		value    Val
		expected string
	}{
		{"number", NumV{num_: 3}, "3"},
		{"true", BoolV{bool_: true}, "true"},
		{"false", BoolV{bool_: false}, "false"},
		{"string", StringV{string_: "hello"}, "\"hello\""},
		{"closure", CloV{}, "#<procedure>"},
		{"primop", PrimopV{op: "+"}, "#<primop>"},
	}

	for _, test := range tests {
		result := serialize(test.value)

		if result != test.expected {
			t.Fatalf("serialize(%s) = %s, want %s",
				test.name, result, test.expected)
		}
	}
}

func TestPrimSubstring(t *testing.T) {
	result := primSubstring([]Val{
		StringV{string_: "hello"},
		NumV{num_: 0},
		NumV{num_: 2},
	})

	expected := StringV{string_: "he"}

	if result != expected {
		t.Fatalf("primSubstring result = %v, want %v", result, expected)
	}
}
func TestPrimSubstringEmpty(t *testing.T) {
	result := primSubstring([]Val{
		StringV{string_: "hello"},
		NumV{num_: 2},
		NumV{num_: 2},
	})

	expected := StringV{string_: ""}

	if result != expected {
		t.Fatalf("primSubstring empty = %v, want %v", result, expected)
	}
}
func TestPrimSubstringNonNaturals(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for non-natural indexes")
		}
	}()

	primSubstring([]Val{
		StringV{string_: "hello"},
		NumV{num_: 1.5},
		NumV{num_: 3.5},
	})
}
func TestPrimSubstringOutOfBounds(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for index out of bounds")
		}
	}()
	primSubstring([]Val{
		StringV{string_: "hello"},
		NumV{num_: 1},
		NumV{num_: 10},
	})
}
func TestPrimSubstringStopBeforeStart(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for stop before start")
		}
	}()
	primSubstring([]Val{
		StringV{string_: "hello"},
		NumV{num_: 4},
		NumV{num_: 1},
	})
}
func TestPrimSubstringBadArgumentTypes(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for bad argument types")
		}
	}()
	primSubstring([]Val{
		StringV{string_: "hello"},
		BoolV{bool_: true},
		NumV{num_: 3},
	})
}

func TestPrimErrorWrongNumberOfArgs(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for wrong number of arguments")
		}
	}()

	primError([]Val{})
}

func TestPrimErrorUserError(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for user error")
		}

		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected panic string, got %v", r)
		}

		if !strings.Contains(msg, "VEBG4 user-error") {
			t.Fatalf("expected user-error message, got %v", msg)
		}
	}()

	primError([]Val{
		NumV{num_: 5},
	})
}