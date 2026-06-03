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