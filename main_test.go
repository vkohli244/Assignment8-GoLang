package main

import (
	"reflect"
	"strings"
	"testing"
	
)

var testEnv = Env{
	{name: "b", value: NumV{num_: 2}},
	{name: "a", value: NumV{num_: 1}},
	{name: "c", value: BoolV{true}},
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
	_, err := envLookup("d", testEnv)
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

func TestNumCInterp(t *testing.T) {
	input := NumC{3}
	expected := NumV{3}
	actual, err := interp(input, testEnv)
	
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if expected != actual {
		t.Errorf("interp(NumC{3}) failed, expected %v, got %v", expected, actual)
	}
}

func TestIdCInterp(t *testing.T) {
	input := idC{"a"}
	expected := NumV{1}
	actual, err := interp(input, testEnv)
	
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if expected != actual {
		t.Errorf("interp(IdC{\"a\"}) failed, expected %v, got %v", expected, actual)
	}
}

func TestLamCInterp(t *testing.T) {
	input := LamC{args: []string{"a", "b"}, body: NumC{1}}

	expected := CloV{[]string{"a", "b"}, NumC{1}, testEnv}
	actual, err := interp(input, testEnv)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("interp(CloV) failed, expected %v, got %v", expected, actual)
	}
}

func TestStringCInterp(t *testing.T) {
	input := StringC{"test"}
	expected := StringV{"test"}
	actual, err := interp(input, testEnv)
	
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if expected != actual {
		t.Errorf("interp(StringC) failed, expected %v, got %v", expected, actual)
	}
}

func TestIfCInterp(t *testing.T) {
	input := ifC{idC{"c"}, NumC{1}, NumC{2}}
	expected := NumV{1}
	actual, err := interp(input, testEnv)
	
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if expected != actual {
		t.Errorf("interp(IfC) failed, expected %v, got %v", expected, actual)
	}
}

func TestIfCNotPredicate(t *testing.T) {
	input := ifC{NumC{3}, NumC{1}, NumC{2}}
	_, err := interp(input, testEnv)
	
	if err == nil || !strings.Contains(err.Error(), "if test condition is not a predicate") {
		t.Errorf("expected error for bad condition in if, got %v", err)
	}
}

type FakeExprC struct {}
func (FakeExprC) isExpr() {}

func TestInterpBadInput(t *testing.T) {
	_, err := interp(FakeExprC{}, testEnv)

	if err == nil || !strings.Contains(err.Error(), "interp takes an ExprC") {
		t.Errorf("expected error for bad input to interp, got %v", err)
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