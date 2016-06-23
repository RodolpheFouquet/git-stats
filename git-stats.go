package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/RodolpheFouquet/termtables"
	"github.com/kardianos/osext"
	"github.com/ttacon/chalk"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Contributor struct {
	Name            string
	Additions       int
	Deletions       int
	Commits         int
	CommitScore     float64
	AdditionScore   float64
	DifferenceScore float64
	Alias           string
	StartDate       time.Time
	EndDate         time.Time
}

type Period struct {
	User  string `json:"user"`
	Start string `json:"start"`
	End   string `json:"end"`
	Alias string `json:"alias"`
}

func NewContributor(name string) *Contributor {
	return &Contributor{Name: name, Additions: 0, Deletions: 0, Commits: 0}
}

func (c *Contributor) IncrementCounters(additions, deletions int) {
	c.Additions = additions + c.Additions
	c.Deletions = deletions + c.Deletions
}

func (c *Contributor) GetScore() float64 {
	return 0.8*c.DifferenceScore + 0.1*c.AdditionScore + 0.1*c.CommitScore
}

func (c *Contributor) SetScores(difference, addition, commits float64) {
	c.DifferenceScore = difference
	c.AdditionScore = addition
	c.CommitScore = commits
}

type Report struct {
	Contributors   map[string]*Contributor
	TotalAdditions int
	TotalDeletions int
	TotalCommits   int
}

func NewReport() *Report {
	return &Report{Contributors: make(map[string]*Contributor), TotalAdditions: 0, TotalDeletions: 0, TotalCommits: 0}
}

func (r *Report) HasContributor(name string) bool {
	_, exists := r.Contributors[name]
	return exists
}

func (r *Report) AddContributor(name string) {
	if !r.HasContributor(name) {
		r.Contributors[name] = NewContributor(name)
	}
}

func (r *Report) IncrementCounters(name string, additions, deletions int) error {
	if !r.HasContributor(name) {
		return errors.New("This contributor does not exist")
	}
	r.Contributors[name].IncrementCounters(additions, deletions)
	r.TotalAdditions += additions
	r.TotalDeletions += deletions
	return nil
}

func (r *Report) IncrementCommits(name string) error {
	if !r.HasContributor(name) {
		return errors.New("This contributor does not exist")
	}
	r.Contributors[name].Commits++
	r.TotalCommits++
	return nil
}

func ExecGit(repo string) (string, error) {
	command := exec.Command("git", "-C", repo, "log", "--numstat", "--pretty='%an|%ad'")
	fmt.Println("Gathering the stats in the repo", repo)
	out, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func ParseStats(gitOutput, subtree string) (*Report, error) {
	fmt.Println("Parsing the stats from the repo using %v as subtree", subtree)
	report := NewReport()
	reader := bufio.NewReader(strings.NewReader(gitOutput))
	currentContributor := ""
	hasContributed := false
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lineString := string(line)
		if len(string(line)) == 0 {
			continue
		}

		splittedLine := strings.Split(lineString, "\t")

		if len(splittedLine) == 1 {
			contribAndDate := strings.Split(lineString, "|")
			currentContributor = strings.Replace(contribAndDate[0], "'", "", -1)
			timeString := strings.Replace(contribAndDate[1], "'", "", -1)

			date, _ := time.Parse("Mon Jan 2 15:04:05 2006 -0700", timeString)
			fmt.Println(date)
			hasContributed = false
		} else if len(splittedLine) == 3 {
			pathModified := fmt.Sprintf("/%s", splittedLine[2])
			rel, err := filepath.Rel(subtree, pathModified)
			if err != nil {
				fmt.Println(chalk.Yellow, "Relative Warning: ", err)
			}
			if strings.Contains(rel, "..") {
				continue
			}

			additions, err := strconv.Atoi(splittedLine[0])
			if err != nil {
				continue
			}
			deletions, err := strconv.Atoi(splittedLine[1])
			if err != nil {
				continue
			}

			if !hasContributed {
				hasContributed = true
				report.AddContributor(currentContributor)
				report.IncrementCommits(currentContributor)
			}
			report.IncrementCounters(currentContributor, additions, deletions)
		}
	}
	return report, nil
}

type OrderByScore []Contributor

func (a OrderByScore) Len() int           { return len(a) }
func (a OrderByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a OrderByScore) Less(i, j int) bool { return a[i].GetScore() < a[j].GetScore() }

func PrintHelp(success bool) {
	execname, _ := osext.Executable()
	var color chalk.Color
	if success {
		color = chalk.Green
	} else {
		color = chalk.Red
	}
	fmt.Println(color, "Usage: ", execname, "--repo=repo_path", "[options]")
	flag.PrintDefaults()
	if success {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func main() {

	directory := flag.String("repo", "", "[mandatory] Path to the git repository")
	subtree := flag.String("subtree", "/", "[optionnal] Subtree you want to parse")
	config := flag.String("config", "", "[optionnal] Path to the configuration file")
	help := flag.Bool("help", false, "[optionnal] Displays this helps and quit")

	flag.Parse()
	if *help {
		PrintHelp(true)
	}
	if *directory == "" {
		PrintHelp(false)
	}
	fmt.Println(*config)

	gitOutput, err := ExecGit(*directory)
	if err != nil {
		fmt.Println(chalk.Red, err)
		os.Exit(1)
	}

	report, err := ParseStats(gitOutput, *directory)

	separator := strings.Repeat("#", 80)
	fmt.Println(chalk.Green, separator)
	fmt.Println(chalk.Green, "Summing up contributions for the repository ", *directory, " subtree ", *subtree)
	fmt.Println(chalk.Green, separator)
	fmt.Println("")
	table := termtables.CreateTable()
	table.AddHeaders("Contributor", "Additions - Deletions", "Additions", "Commits", "Score")
	contributors := make([]Contributor, 0)
	for _, v := range report.Contributors {
		if v.Commits > 0 {
			differenceScore := math.Abs(float64(v.Additions-v.Deletions)) * 100.0 / float64(report.TotalAdditions-report.TotalDeletions)
			additionScore := float64(v.Additions) * 100.0 / float64(report.TotalAdditions)
			commitScore := float64(v.Commits) * 100.0 / float64(report.TotalCommits)
			v.SetScores(differenceScore, additionScore, commitScore)
			contributors = append(contributors, *v)
		}
	}
	sort.Sort(OrderByScore(contributors))
	for index := range contributors {
		c := contributors[len(contributors)-index-1]
		table.AddRow(c.Name, fmt.Sprintf("%.3f%%", c.DifferenceScore), fmt.Sprintf("%.3f%%", c.AdditionScore), fmt.Sprintf("%.3f%%", c.CommitScore), fmt.Sprintf("%.3f", c.GetScore()))
	}

	table.AddSeparator()
	table.AddRow("Total", report.TotalAdditions, report.TotalDeletions, report.TotalCommits, "-----")
	table.SetAlign(3, 2)
	table.SetAlign(3, 3)
	table.SetAlign(3, 4)
	table.SetAlign(3, 5)
	fmt.Println(table.Render())
}
