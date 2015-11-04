package main

import "testing"

func TestIncrementCounters(t *testing.T) {
	c := NewContributor("dummy")

	c.IncrementCounters(1, 1)

	if c.Additions != 1 || c.Deletions != 1 {
		t.Errorf("Additions and Deletions should be at 1 and the were %v, %v", c.Additions, c.Deletions)
	}

	c.IncrementCounters(0, 0)
	if c.Additions != 1 || c.Deletions != 1 {
		t.Errorf("Additions and Deletions should be at 1 and the were %v, %v", c.Additions, c.Deletions)
	}

	c.IncrementCounters(5, 4)
	if c.Additions != 6 || c.Deletions != 5 {
		t.Errorf("Additions and Deletions should be at 6 and 5 and the were %v, %v", c.Additions, c.Deletions)
	}

	c.IncrementCounters(-5, -4)
	if c.Additions != 1 || c.Deletions != 1 {
		t.Errorf("Additions and Deletions should be at 1 and the were %v, %v", c.Additions, c.Deletions)
	}

}

func TestNewContributor(t *testing.T) {
	c := NewContributor("Pouet")

	if c.Name != "Pouet" {
		t.Errorf("The expected name was Pouet, however we got %v", c.Name)
	}
}

func TestSetScores(t *testing.T) {
	c := NewContributor("")

	var diffScore, addScore, commitScore float32
	diffScore = 80
	addScore = 10
	commitScore = 10
	c.SetScores(diffScore, addScore, commitScore)

	if c.DifferenceScore != diffScore || c.AdditionScore != addScore || c.CommitScore != commitScore {
		t.Errorf("The expected scored should be %v %v %v and were %v %v %v", diffScore, addScore, commitScore, c.DifferenceScore, c.AdditionScore, c.CommitScore)
	}

}

func TestGetScores(t *testing.T) {
	c := NewContributor("")

	c.SetScores(80.0, 10.0, 50.0)
	expectedScore := c.DifferenceScore*0.8 + c.AdditionScore*0.1 + c.CommitScore*0.1
	if c.GetScore() != expectedScore {
		t.Errorf("The expected score was %v and we got %v", expectedScore, c.GetScore())
	}
}
