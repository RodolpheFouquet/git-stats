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

func TestHasContributor(t *testing.T) {
	report := NewReport()

	name := "Pouet"
	if report.HasContributor(name) {
		t.Errorf("The report shouldn't have any contributor")
	}

	report.AddContributor(name)
	if !report.HasContributor(name) {
		t.Errorf("The report should have the contributor %v", name)
	}

	if report.HasContributor("Pouetpouet") {
		t.Error("Nah, this contributor wasn't added to the report!")
	}

}

func TestIncrementReportCounters(t *testing.T) {
	r := NewReport()
	name := "Pouet"
	err := r.IncrementCounters(name, 0, 0)

	if err == nil {
		t.Errorf("Incrementing a counter on a non existing contributor should return a valid error")
	}

	r.AddContributor(name)
	c := r.Contributors[name]
	err = r.IncrementCounters(name, 0, 0)
	if err != nil {
		t.Errorf("Incrementing a counter on a valid contributor should not return an error")
	}
	if c.Additions != 0 || c.Deletions != 0 {
		t.Errorf("Contributor Additions and Deletions should still be at 0")
	}

	if r.TotalAdditions != 0 || r.TotalDeletions != 0 {
		t.Errorf("Total Additions and Deletions should still be at 0")
	}

	addDiff := 10
	delDiff := 9
	r.IncrementCounters(name, addDiff, delDiff)
	if c.Additions != addDiff || c.Deletions != delDiff {
		t.Errorf("Contributor Additions and Deletions should be equal to %v and %v", addDiff, delDiff)
	}

	if r.TotalAdditions != addDiff || r.TotalDeletions != delDiff {
		t.Errorf("Total Additions and Deletions should be equal to %v and %v", addDiff, delDiff)
	}

	name2 := "Pouetpouet"
	r.AddContributor(name2)
	c2 := r.Contributors[name2]
	r.IncrementCounters(name2, addDiff, delDiff)
	if c2.Additions != addDiff || c2.Deletions != delDiff {
		t.Errorf("Contributor Additions and Deletions should be equal to %v and %v", addDiff, delDiff)
	}

	if c.Additions != addDiff || c.Deletions != delDiff {
		t.Errorf("Contributor Additions and Deletions should be equal to %v and %v", addDiff, delDiff)
	}
	if r.TotalAdditions != 2*addDiff || r.TotalDeletions != 2*delDiff {
		t.Errorf("Total Additions and Deletions should equal be to %v and %v", 2*addDiff, 2*delDiff)
	}
}

func TestIncrementCommit(t *testing.T) {
	r := NewReport()

	name := "Pouet"
	name2 := "Pouetpouet"
	err := r.IncrementCommits(name)
	if err == nil {
		t.Errorf("Incrementing the commits of a non existing user should fail")
	}

	r.AddContributor(name)
	c := r.Contributors[name]
	err = r.IncrementCommits(name)
	if err != nil {
		t.Errorf("Incrementing the commits of an existing user should not fail")
	}

	if c.Commits != 1 {
		t.Errorf("Failed to increment the contributor's commits")
	}

	if r.TotalCommits != 1 {
		t.Errorf("The total number of commits should have been incremented")
	}

	r.AddContributor(name2)
	c2 := r.Contributors[name2]
	r.IncrementCommits(name2)
	if c2.Commits != 1 {
		t.Errorf("Failed to increment the contributor's commits")
	}

	if r.TotalCommits != 2 {
		t.Errorf("The total number of commits should have been incremented")
	}

	if c.Commits != 1 {
		t.Errorf("The number of the first contributor should remain at 1")
	}
}
