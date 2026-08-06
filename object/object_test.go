package object

import (
	"testing"
)

func TestFloatHashKey(t *testing.T) {
	one1 := &Float{Value: 1.5}
	one2 := &Float{Value: 1.5}
	diff1 := &Float{Value: 2.5}
	diff2 := &Float{Value: 2.5}

	if one1.HashKey() != one2.HashKey() {
		t.Errorf("floats with same value have different hash keys")
	}

	if diff1.HashKey() != diff2.HashKey() {
		t.Errorf("floats with different values have the same hash keys")
	}

	if one1.HashKey() == diff1.HashKey() {
		t.Errorf("floats with same hash keys have different values")
	}
}

func TestFloatInspect(t *testing.T) {
	tests := []struct {
		value    float64
		expected string
	}{
		{20.0, "20.0"},
		{3.14, "3.14"},
		{0.5, "0.5"},
		{100.0, "100.0"},
	}

	for _, tt := range tests {
		f := &Float{Value: tt.value}
		if f.Inspect() != tt.expected {
			t.Errorf("Inspect() wrong. expected=%q, got=%q", tt.expected, f.Inspect())
		}
	}
}
func TestStringHashKey(t *testing.T) {
	hello1 := &String{Value: "Hello World"}
	hello2 := &String{Value: "Hello World"}
	diff1 := &String{Value: "My name is johnny"}
	diff2 := &String{Value: "My name is johnny"}

	if hello1.HashKey() != hello2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}

	if diff1.HashKey() != diff2.HashKey() {
		t.Errorf("strings with different content have the same hash keys")
	}

	if hello1.HashKey() == diff1.HashKey() {
		t.Errorf("strings with same hash keys have different content")
	}
}
