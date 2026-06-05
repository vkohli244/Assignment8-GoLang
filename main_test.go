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

func TestPrimPlus(t *testing.T) {
	tests := []struct {
		name     string
		args     []Val
		expected int
	}{
		{"positive numbers", []Val{NumV{num_: 6}, NumV{num_: 4}}, 10},
		{"negative result component", []Val{NumV{num_: -1}, NumV{num_: 3}}, 2},
	}

	for _, test := range tests {
		value, err := primPlus(test.args)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		numValue, isNum := value.(NumV)
		if !isNum || numValue.num_ != test.expected {
			t.Fatalf("primPlus %s = %v, want NumV{%d}", test.name, value, test.expected)
		}
	}
}

func TestPrimMinus(t *testing.T) {
	tests := []struct {
		name     string
		args     []Val
		expected int
	}{
		{"positive result", []Val{NumV{num_: 6}, NumV{num_: 4}}, 2},
		{"negative result", []Val{NumV{num_: 4}, NumV{num_: 6}}, -2},
	}

	for _, test := range tests {
		value, err := primMinus(test.args)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		numValue, isNum := value.(NumV)
		if !isNum || numValue.num_ != test.expected {
			t.Fatalf("primMinus %s = %v, want NumV{%d}", test.name, value, test.expected)
		}
	}
}

func TestPrimDiv(t *testing.T) {
	tests := []struct {
		name     string
		args     []Val
		expected int
	}{
		{"even division", []Val{NumV{num_: 8}, NumV{num_: 2}}, 4},
		{"integer division", []Val{NumV{num_: 7}, NumV{num_: 2}}, 3},
	}

	for _, test := range tests {
		value, err := primDiv(test.args)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		numValue, isNum := value.(NumV)
		if !isNum || numValue.num_ != test.expected {
			t.Fatalf("primDiv %s = %v, want NumV{%d}", test.name, value, test.expected)
		}
	}
}

func TestPrimDivByZero(t *testing.T) {
	_, err := primDiv([]Val{NumV{num_: 8}, NumV{num_: 0}})
	if err == nil || !strings.Contains(err.Error(), "division by 0 undefined") {
		t.Fatalf("expected division by zero error, got %v", err)
	}
}

func TestPrimLessEqual(t *testing.T) {
	tests := []struct {
		name     string
		args     []Val
		expected bool
	}{
		{"less than", []Val{NumV{num_: 2}, NumV{num_: 3}}, true},
		{"equal", []Val{NumV{num_: 3}, NumV{num_: 3}}, true},
		{"greater than", []Val{NumV{num_: 4}, NumV{num_: 3}}, false},
	}

	for _, test := range tests {
		value, err := primLessEqual(test.args)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		boolValue, isBool := value.(BoolV)
		if !isBool || boolValue.bool_ != test.expected {
			t.Fatalf("primLessEqual %s = %v, want BoolV{%t}", test.name, value, test.expected)
		}
	}
}

func TestPrimitiveNumericOperatorsWrongArity(t *testing.T) {
	tests := []struct {
		name string
		op   string
		args []Val
		want string
	}{
		{"plus", "+", []Val{NumV{num_: 1}}, "wrong number of arguments to +"},
		{"minus", "-", []Val{NumV{num_: 1}}, "wrong number of arguments to -"},
		{"div", "/", []Val{NumV{num_: 1}}, "wrong number of arguments to /"},
		{"less equal", "<=", []Val{NumV{num_: 1}}, "wrong number of arguments to <="},
	}

	for _, test := range tests {
		_, err := applyPrimop(test.op, test.args)
		if err == nil || !strings.Contains(err.Error(), test.want) {
			t.Fatalf("%s: expected %q error, got %v", test.name, test.want, err)
		}
	}
}

func TestPrimOpsNonNums(t *testing.T) {
	tests := []struct {
		name string
		fn   func([]Val) (Val, error)
		args []Val
		want string
	}{
		{"plus", primPlus, []Val{StringV{string_: "left"}, NumV{num_: 1}}, "+ requires numbers"},
		{"minus", primMinus, []Val{StringV{string_: "left"}, NumV{num_: 1}}, "- requires numbers"},
		{"div", primDiv, []Val{StringV{string_: "left"}, NumV{num_: 1}}, "/ requires numbers"},
		{"less equal", primLessEqual, []Val{StringV{string_: "left"}, NumV{num_: 1}}, "<= requires numbers"},
	}

	for _, test := range tests {
		_, err := test.fn(test.args)
		if err == nil || !strings.Contains(err.Error(), test.want) {
			t.Fatalf("%s: expected %q error, got %v", test.name, test.want, err)
		}
	}
}

func TestPrimEqNums(t *testing.T) {
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

func TestPrimEqStrs(t *testing.T) {
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

func TestPrimEqBools(t *testing.T) {
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

func TestPrimEqMixedTypes(t *testing.T) {
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

func TestPrimEqClosFalse(t *testing.T) {
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

func TestPrimStrlen(t *testing.T) {
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

func TestPrimStrlenNotStr(t *testing.T) {
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

func TestAppCOneArg(t *testing.T) {
	input := AppC{
		f: LamC{
			args: []string{"x"},
			body: idC{"x"},
		},
		args: []ExprC{NumC{7}},
	}
	expected := NumV{7}
	// x is bound to 7
	actual, err := interp(input, testEnv)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if expected != actual {
		t.Errorf("interp(AppC) failed, expected %v, got %v", expected, actual)
	}
}

func TestAppCTwoArgs(t *testing.T) {
	input := AppC{
		f: LamC{
			args: []string{"x", "y"},
			body: idC{"y"},
		},
		args: []ExprC{NumC{7}, StringC{"done"}},
	}
	expected := StringV{"done"}
	// second param bound to "done"
	actual, err := interp(input, testEnv)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if expected != actual {
		t.Errorf("interp(AppC multiple args) failed, expected %v, got %v", expected, actual)
	}
}

func TestAppCClosureEnv(t *testing.T) {
	input := AppC{
		f: LamC{
			args: []string{"x"},
			body: idC{"a"},
		},
		args: []ExprC{NumC{99}},
	}
	expected := NumV{1}
	// a is bound to 1 in test_env
	actual, err := interp(input, testEnv)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if expected != actual {
		t.Errorf("interp(AppC closure env) failed, expected %v, got %v", expected, actual)
	}
}

func TestAppCArity(t *testing.T) {
	input := AppC{
		f: LamC{
			args: []string{"x", "y"},
			body: idC{"x"},
		},
		args: []ExprC{NumC{7}},
	}
	// expects two args, recieves one
	_, err := interp(input, testEnv)

	if err == nil || !strings.Contains(err.Error(), "wrong number of arguments") {
		t.Errorf("expected wrong arity error, got %v", err)
	}
}

func TestAppCNonFun(t *testing.T) {
	input := AppC{
		f:    NumC{3},
		args: []ExprC{NumC{7}},
	}
	// The function position evaluates to a number, not a closure, so applying it should error.
	_, err := interp(input, testEnv)

	if err == nil || !strings.Contains(err.Error(), "application expected a function") {
		t.Errorf("expected error for applying non-function, got %v", err)
	}
}

type FakeExprC struct{}

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
	result, err := primSubstring([]Val{
		StringV{string_: "hello"},
		NumV{num_: 0},
		NumV{num_: 2},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := StringV{string_: "he"}

	if result != expected {
		t.Fatalf("primSubstring result = %v, want %v", result, expected)
	}
}
func TestPrimSubstringEmpty(t *testing.T) {
	result, err := primSubstring([]Val{
		StringV{string_: "hello"},
		NumV{num_: 2},
		NumV{num_: 2},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := StringV{string_: ""}

	if result != expected {
		t.Fatalf("primSubstring empty = %v, want %v", result, expected)
	}
}
func TestPrimSubstringNonNaturals(t *testing.T) {
	_, err := primSubstring([]Val{
		StringV{string_: "hello"},
		NumV{num_: -1},
		NumV{num_: 3},
	})
	if err == nil || !strings.Contains(err.Error(), "non-naturals") {
		t.Fatalf("expected non-natural indexes error, got %v", err)
	}
}
func TestPrimSubstringOutOfBounds(t *testing.T) {
	_, err := primSubstring([]Val{
		StringV{string_: "hello"},
		NumV{num_: 1},
		NumV{num_: 10},
	})
	if err == nil || !strings.Contains(err.Error(), "index out of bounds") {
		t.Fatalf("expected index out of bounds error, got %v", err)
	}
}
func TestPrimSubstringStopBeforeStart(t *testing.T) {
	_, err := primSubstring([]Val{
		StringV{string_: "hello"},
		NumV{num_: 4},
		NumV{num_: 1},
	})
	if err == nil || !strings.Contains(err.Error(), "stop before start") {
		t.Fatalf("expected stop before start error, got %v", err)
	}
}
func TestPrimSubstringBadArgumentTypes(t *testing.T) {
	_, err := primSubstring([]Val{
		StringV{string_: "hello"},
		BoolV{bool_: true},
		NumV{num_: 3},
	})
	if err == nil || !strings.Contains(err.Error(), "bad argument types") {
		t.Fatalf("expected bad argument types error, got %v", err)
	}
}

func TestPrimErrorWrongNumberOfArgs(t *testing.T) {
	_, err := applyPrimop("error", []Val{})
	if err == nil || !strings.Contains(err.Error(), "wrong number of arguments to error") {
		t.Fatalf("expected wrong number of arguments error, got %v", err)
	}
}

func TestPrimErrorUserError(t *testing.T) {
	_, err := primError([]Val{
		NumV{num_: 5},
	})
	if err == nil || !strings.Contains(err.Error(), "VEBG4 user-error") {
		t.Fatalf("expected user-error message, got %v", err)
	}
}

// Zip tests
// tests that zip correctly pairs names with values
func TestZip(t *testing.T) {
	result := zip(
		[]string{"a", "b"},
		[]Val{
			NumV{num_: 1},
			NumV{num_: 2},
		},
	)
	expected := Env{
		{name: "a", value: NumV{num_: 1}},
		{name: "b", value: NumV{num_: 2}},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("zip result = %v, want %v", result, expected)
	}
}

// Tesing empty
// Tests that zip returns an empty environment witho names and no values
func TestZipEmpty(t *testing.T) {
	result := zip([]string{}, []Val{})
	expected := Env{}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("zip empty = %v, want %v", result, expected)
	}
}
