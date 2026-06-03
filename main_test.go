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

func TestPrimEqualNumbers(t *testing.T) {
	args := []Val{NumV{num_: 6}, NumV{num_: 6}}
	value, err := primEqual(args)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	boolValue, isBool := value.(BoolV)
	if !isBool || !boolValue.bool_ {
		t.Fatalf("primEqual nums = %v, want true", value)
	}
}

func TestPrimEqualStrings(t *testing.T) {
	args := []Val{StringV{string_: "hello"}, StringV{string_: "hello"}}
	value, err := primEqual(args)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	boolValue, isBool := value.(BoolV)
	if !isBool || !boolValue.bool_ {
		t.Fatalf("primEqual strings = %v, want true", value)
	}
}

func TestPrimEqualBools(t *testing.T) {
	args := []Val{BoolV{bool_: true}, BoolV{bool_: true}}
	value, err := primEqual(args)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	boolValue, isBool := value.(BoolV)
	if !isBool || !boolValue.bool_ {
		t.Fatalf("primEqual bools = %v, want true", value)
	}
}

func TestPrimEqualMixedKindsFalse(t *testing.T) {
	args := []Val{NumV{num_: 6}, StringV{string_: "hello"}}
	value, err := primEqual(args)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	boolValue, isBool := value.(BoolV)
	if !isBool || boolValue.bool_ {
		t.Fatalf("primEqual mixed = %v, want false", value)
	}
}

func TestPrimEqualClosuresFalse(t *testing.T) {
	leftClosure := CloV{params_: []string{}, body_: NumC{n: 8}, env_: Env{}}
	rightClosure := CloV{params_: []string{}, body_: NumC{n: 8}, env_: Env{}}
	value, err := primEqual([]Val{leftClosure, rightClosure})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	boolValue, isBool := value.(BoolV)
	if !isBool || boolValue.bool_ {
		t.Fatalf("primEqual closures = %v, want false", value)
	}
}

func TestPrimEqualWrongArity(t *testing.T) {
	_, err := primEqual([]Val{NumV{num_: 6}})
	if err == nil || !strings.Contains(err.Error(), "equal? requires two values") {
		t.Fatalf("expected arity error, got %v", err)
	}
}

func TestPrimStrlenHello(t *testing.T) {
	args := []Val{StringV{string_: "hello"}}
	value, err := primStrlen(args)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	numValue, isNum := value.(NumV)
	if !isNum || numValue.num_ != 5 {
		t.Fatalf("primStrlen hello = %v, want NumV{5}", value)
	}
}

func TestPrimStrlenNotAString(t *testing.T) {
	args := []Val{NumV{num_: 6}}
	_, err := primStrlen(args)
	if err == nil || !strings.Contains(err.Error(), "not a string") {
		t.Fatalf("expected not a string error, got %v", err)
	}
	args = []Val{BoolV{bool_: true}}
	_, err = primStrlen(args)
	if err == nil || !strings.Contains(err.Error(), "not a string") {
		t.Fatalf("expected not a string error, got %v", err)
	}
}

func TestPrimStrlenWrongArity(t *testing.T) {
	_, err := primStrlen([]Val{})
	if err == nil || !strings.Contains(err.Error(), "strlen requires one value") {
		t.Fatalf("expected arity error, got %v", err)
	}
	_, err = primStrlen([]Val{NumV{num_: 6}, NumV{num_: 6}})
	if err == nil || !strings.Contains(err.Error(), "strlen requires one value") {
		t.Fatalf("expected arity error, got %v", err)
	}
}
