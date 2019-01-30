package main

import "testing"

func TestNodeShouldBeDeleted(t *testing.T) {
	tests := []struct{
		node string
		removed []string
		expected bool
	}{
		{"cluster-abcdefghijkl-minion-1",[]string{"1", "2", "3"}, true},
		{"cluster-abcdefghijkl-minion-2",[]string{"1", "2", "3"}, true},
		{"cluster-abcdefghijkl-minion-5",[]string{"1", "2", "3"}, false},
		{"cluster-abcdefghijkl-minion-6",[]string{}, false},
		{"",[]string{"1", "2", "3"}, false},
	}

	for _, test := range tests {
		shouldDelete := NodeShouldBeDeleted(test.node, test.removed)
		if shouldDelete != test.expected {
			t.Errorf("NodeShouldBeDeleted(%s, %v) - got %t, want %t", test.node, test.removed, shouldDelete, test.expected)
		}
	}
}