package main

import "testing"

func TestIncrementCounters(t *testing.T) {
	c := NewContributor("dummy")

	c.IncrementCounters(1, 1)

	if c.Additions != 1 && c.Deletions != 1 {
		t.Errorf("Additions and Deletions should be at 1 and the were %v, %v", c.Additions, c.Deletions)
	}

	c.IncrementCounters(0, 0)
	if c.Additions != 1 && c.Deletions != 1 {
		t.Errorf("Additions and Deletions should be at 1 and the were %v, %v", c.Additions, c.Deletions)
	}

	c.IncrementCounters(5, 4)
	if c.Additions != 6 && c.Deletions != 5 {
		t.Errorf("Additions and Deletions should be at 6 and 5 and the were %v, %v", c.Additions, c.Deletions)
	}

	c.IncrementCounters(-5, -4)
	if c.Additions != 1 && c.Deletions != 1 {
		t.Errorf("Additions and Deletions should be at 1 and the were %v, %v", c.Additions, c.Deletions)
	}

}
