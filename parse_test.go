package main

import (
	"io/ioutil"
	"testing"
)

func TestIncrementCounters(t *testing.T) {
	c := NewContributor("dummy")

	c.Contributions[0].IncrementCounters(1, 1)

	if c.Contributions[0].Additions != 1 || c.Contributions[0].Deletions != 1 {
		t.Errorf("Additions and Deletions should be at 1 and the were %v, %v", c.Contributions[0].Additions, c.Contributions[0].Deletions)
	}

	c.Contributions[0].IncrementCounters(0, 0)
	if c.Contributions[0].Additions != 1 || c.Contributions[0].Deletions != 1 {
		t.Errorf("Additions and Deletions should be at 1 and the were %v, %v", c.Contributions[0].Additions, c.Contributions[0].Deletions)
	}

	c.Contributions[0].IncrementCounters(5, 4)
	if c.Contributions[0].Additions != 6 || c.Contributions[0].Deletions != 5 {
		t.Errorf("Additions and Deletions should be at 6 and 5 and the were %v, %v", c.Contributions[0].Additions, c.Contributions[0].Deletions)
	}

	c.Contributions[0].IncrementCounters(-5, -4)
	if c.Contributions[0].Additions != 1 || c.Contributions[0].Deletions != 1 {
		t.Errorf("Additions and Deletions should be at 1 and the were %v, %v", c.Contributions[0].Additions, c.Contributions[0].Deletions)
	}

}

func TestNewContributor(t *testing.T) {
	c := NewContributor("Pouet")

	if c.Contributions[0].Name != "Pouet" {
		t.Errorf("The expected name was Pouet, however we got %v", c.Contributions[0].Name)
	}
}

func TestSetScores(t *testing.T) {
	c := NewContributor("")

	var diffScore, addScore, commitScore float64
	diffScore = 80
	addScore = 10
	commitScore = 10
	c.Contributions[0].SetScores(diffScore, addScore, commitScore)

	if c.Contributions[0].DifferenceScore != diffScore || c.Contributions[0].AdditionScore != addScore || c.Contributions[0].CommitScore != commitScore {
		t.Errorf("The expected scored should be %v %v %v and were %v %v %v", diffScore, addScore, commitScore, c.Contributions[0].DifferenceScore, c.Contributions[0].AdditionScore, c.Contributions[0].CommitScore)
	}

}

func TestGetScores(t *testing.T) {
	c := NewContributor("")

	c.Contributions[0].SetScores(80.0, 10.0, 50.0)
	expectedScore := c.Contributions[0].DifferenceScore*0.8 + c.Contributions[0].AdditionScore*0.1 + c.Contributions[0].CommitScore*0.1
	if c.Contributions[0].GetScore() != expectedScore {
		t.Errorf("The expected score was %v and we got %v", expectedScore, c.Contributions[0].GetScore())
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
	if c.Contributions[0].Additions != 0 || c.Contributions[0].Deletions != 0 {
		t.Errorf("Contributor Additions and Deletions should still be at 0")
	}

	if r.TotalAdditions != 0 || r.TotalDeletions != 0 {
		t.Errorf("Total Additions and Deletions should still be at 0")
	}

	addDiff := 10
	delDiff := 9
	r.IncrementCounters(name, addDiff, delDiff)
	if c.Contributions[0].Additions != addDiff || c.Contributions[0].Deletions != delDiff {
		t.Errorf("Contributor Additions and Deletions should be equal to %v and %v", addDiff, delDiff)
	}

	if r.TotalAdditions != addDiff || r.TotalDeletions != delDiff {
		t.Errorf("Total Additions and Deletions should be equal to %v and %v", addDiff, delDiff)
	}

	name2 := "Pouetpouet"
	r.AddContributor(name2)
	c2 := r.Contributors[name2]
	r.IncrementCounters(name2, addDiff, delDiff)
	if c2.Contributions[0].Additions != addDiff || c2.Contributions[0].Deletions != delDiff {
		t.Errorf("Contributor Additions and Deletions should be equal to %v and %v", addDiff, delDiff)
	}

	if c.Contributions[0].Additions != addDiff || c.Contributions[0].Deletions != delDiff {
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

	if c.Contributions[0].Commits != 1 {
		t.Errorf("Failed to increment the contributor's commits")
	}

	if r.TotalCommits != 1 {
		t.Errorf("The total number of commits should have been incremented")
	}

	r.AddContributor(name2)
	c2 := r.Contributors[name2]
	r.IncrementCommits(name2)
	if c2.Contributions[0].Commits != 1 {
		t.Errorf("Failed to increment the contributor's commits")
	}

	if r.TotalCommits != 2 {
		t.Errorf("The total number of commits should have been incremented")
	}

	if c.Contributions[0].Commits != 1 {
		t.Errorf("The number of the first contributor should remain at 1")
	}
}

func CheckContributors(report *Report, contributors []string) bool {
	ret := true

	for index := range contributors {
		ret = ret && report.HasContributor(contributors[index])
	}
	return ret
}

func testParse(t *testing.T, file string) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		t.Errorf("Could not read the test file %v", err)
	}

	report, err := ParseStats(string(content), "/", *NewPeriodArray())
	if err != nil {
		t.Errorf("Reading a valid git log should not return an error")
	}
	contributors := []string{"Contributor1", "Contributor2", "Contributor3"}
	if !CheckContributors(report, contributors) {
		t.Errorf("There's at least a missing contributor in the output")
	}

	numOfTotalCommits := 9 // there is a full binary commit

	if report.TotalCommits != numOfTotalCommits {
		t.Errorf("The total number of commits should be %v and it was %v", numOfTotalCommits, report.TotalCommits)
	}
	totalAdd := 189
	totalDel := 8
	if report.TotalAdditions != totalAdd {
		t.Errorf("The total number of additions should be %v and it was %v", totalAdd, report.TotalAdditions)
	}

	if report.TotalDeletions != totalDel {
		t.Errorf("The total number of deletions should be %v and it was %v", totalDel, report.TotalDeletions)
	}

	if report.Contributors["Contributor3"].Contributions[0].Commits != 1 {
		t.Errorf("Contributor3 should only have one commit")
	}

	if report.Contributors["Contributor1"].Contributions[0].Commits != 6 {
		t.Errorf("Contributor1 should have 6 commits")
	}

	if report.Contributors["Contributor2"].Contributions[0].Commits != 2 {
		t.Errorf("Contributor2 should have 2 commits")
	}

	if report.Contributors["Contributor3"].Contributions[0].Additions != 1 {
		t.Errorf("Contributor3 should only have one addition")
	}

	if report.Contributors["Contributor3"].Contributions[0].Deletions != 1 {
		t.Errorf("Contributor3 should only have one deletion")
	}

	if report.Contributors["Contributor1"].Contributions[0].Additions != 159 {
		t.Errorf("Contributor1 should have 159 additions")
	}

	if report.Contributors["Contributor1"].Contributions[0].Deletions != 3 {
		t.Errorf("Contributor1 should have 3 deletions")
	}

	if report.Contributors["Contributor2"].Contributions[0].Additions != 29 {
		t.Errorf("Contributor2 should have 29 additions")
	}

	if report.Contributors["Contributor2"].Contributions[0].Deletions != 4 {
		t.Errorf("Contributor2 should have 4 deletions")
	}

	report, err = ParseStats(string(content), "/test", *NewPeriodArray())
	contributors = []string{"Contributor1", "Contributor2"}
	if !CheckContributors(report, contributors) {
		t.Errorf("There's at least a missing contributor in the output")
	}

	numOfTotalCommits = 7 // there is a full binary commit

	if report.TotalCommits != numOfTotalCommits {
		t.Errorf("The total number of commits should be %v and it was %v", numOfTotalCommits, report.TotalCommits)
	}

	totalAdd = 176
	totalDel = 4
	if report.TotalAdditions != totalAdd {
		t.Errorf("The total number of additions should be %v and it was %v", totalAdd, report.TotalAdditions)
	}

	if report.TotalDeletions != totalDel {
		t.Errorf("The total number of deletions should be %v and it was %v", totalDel, report.TotalDeletions)
	}

	if report.Contributors["Contributor2"].Contributions[0].Commits != 2 {
		t.Errorf("Contributor3 should have 2 commits")
	}

	if report.Contributors["Contributor1"].Contributions[0].Commits != 5 {
		t.Errorf("Contributor1 should have 5 commits")
	}

	if report.Contributors["Contributor1"].Contributions[0].Additions != 147 {
		t.Errorf("Contributor1 should have 147 additions")
	}

	if report.Contributors["Contributor1"].Contributions[0].Deletions != 0 {
		t.Errorf("Contributor1 should have 0 deletions")
	}

	if report.Contributors["Contributor2"].Contributions[0].Additions != 29 {
		t.Errorf("Contributor2 should have 29 additions")
	}

	if report.Contributors["Contributor2"].Contributions[0].Deletions != 4 {
		t.Errorf("Contributor2 should have 4 deletions")
	}

	report, err = ParseStats(string(content), "/tests", *NewPeriodArray())
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(report.Contributors) != 0 {
		t.Errorf("This path doesn't exist, there shouldn't be any contributor %v", len(report.Contributors))
	}

	if report.TotalCommits != 0 || report.TotalAdditions != 0 || report.TotalDeletions != 0 {
		t.Errorf("This path does not exist, there should be no additions/deletions/commits")
	}
}

func TestParseUnix(t *testing.T) {
	testParse(t, "test_assets/test_gitlog.txt")
}

func TestParseWindows(t *testing.T) {
	testParse(t, "test_assets/test_gitlogwin.txt")
}
