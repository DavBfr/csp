package main

import "testing"

func TestRemoveDuplicates(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b"}
	result := removeDuplicates(input)
	expected := []string{"a", "b", "c"}

	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}

	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("At position %d, expected %s, got %s", i, expected[i], result[i])
		}
	}
}
