package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/RodolpheFouquet/termtables"
	"github.com/kardianos/osext"
	"github.com/ttacon/chalk"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Contribution struct {
	Additions       int
	Deletions       int
	Commits         int
	CommitScore     float64
	AdditionScore   float64
	DifferenceScore float64
	Name            string
	StartDate       time.Time
	EndDate         time.Time	
}

type Contributor struct {
	Name            string
	Contributions	[]*Contribution
}

type Period struct {
	User  string `json:"user"`
	Start string `json:"start"`
	End   string `json:"end"`
	Alias string `json:"alias"`
}

type PeriodTS struct {
	User string
	Start time.Time
	End   time.Time
	Alias string
}

type PeriodArray struct {
	Periods []Period `json:"periods"`
}

func  IsAfter(t, other time.Time) bool {
	return t.Unix() <= other.Unix()
}

func NewPeriodTS(period Period) *PeriodTS {
	start, err := time.Parse("2006-01-02", period.Start)
	if err != nil {
		fmt.Println(chalk.Red, err)
		return nil
	}
	stop, err := time.Parse("2006-01-02", period.End)
	if err != nil {
		fmt.Println(chalk.Red, err)
		return nil
	}
	return &PeriodTS{User: period.User, Start: start, End: stop, Alias: period.Alias}
}

func NewPeriodArray() *PeriodArray {
	return &PeriodArray{Periods: []Period{}}
}

func NewContribution(name string) *Contribution{
	return &Contribution{Additions: 0, Deletions: 0, Commits: 0, Name: name}
}

func NewContributionDate(name string,period PeriodTS) *Contribution{
	formattedName := fmt.Sprintf("%v (%v)", name, period.Alias)
	if period.Alias == "" {
		formattedName = name
	}
	return &Contribution{Additions: 0, Deletions: 0, Commits: 0, Name: formattedName, StartDate: period.Start, EndDate: period.End}
}

func NewContributor(name string, periods []PeriodTS) *Contributor {
	
	var contributions []*Contribution
	if len(periods) > 0 {
		contributions = append(contributions, NewContribution(fmt.Sprintf("%v %v", name, "(otherwise)"))) // otherwise
		for _, period := range periods {
			contributions = append(contributions, NewContributionDate(name, period)) 
		}
	} else {
		contributions = []*Contribution{NewContribution(name)}
	}
	
	return &Contributor{Name: name, Contributions: contributions}
}

func (c *Contribution) IncrementCounters(additions, deletions int) {
	c.Additions = additions + c.Additions
	c.Deletions = deletions + c.Deletions
}

func (c *Contribution) GetScore() float64 {
	return 0.8*c.DifferenceScore + 0.1*c.AdditionScore + 0.1*c.CommitScore
}

func (c *Contribution) SetScores(difference, addition, commits float64) {
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
func GetContribution(contributions []*Contribution, date time.Time) *Contribution {
	var ret  *Contribution
	if len(contributions) > 1 {
		for _, contrib := range contributions {
			if IsAfter(contrib.StartDate, date) && !IsAfter(contrib.EndDate, date) {
				return contrib
			}
		}
	} 
	ret = contributions[0]
	
	return ret
}

func (r *Report) AddContributor(name string, periodMap map[string][]PeriodTS) {
	if !r.HasContributor(name) {
		periods, _ := periodMap[name]
		r.Contributors[name] = NewContributor(name, periods)
	}
}

func (r *Report) IncrementCounters(name string, additions, deletions int, date time.Time) error {
	if !r.HasContributor(name) {
		return errors.New("This contributor does not exist")
	}
	contrib := GetContribution(r.Contributors[name].Contributions, date)
	contrib.IncrementCounters(additions, deletions)
	r.TotalAdditions += additions
	r.TotalDeletions += deletions
	return nil
}

func (r *Report) IncrementCommits(name string, date time.Time) error {
	if !r.HasContributor(name) {
		return errors.New("This contributor does not exist")
	}
	contrib := GetContribution(r.Contributors[name].Contributions, date)
	contrib.Commits++
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

func ParseStats(gitOutput, subtree string, periods PeriodArray) (*Report, error) {
	periodMap := make(map[string][]PeriodTS)
	for _, period := range periods.Periods {
		periodMap[period.User] = append(periodMap[period.User], *NewPeriodTS(period))
	}
	fmt.Println(periodMap)
	fmt.Println("Parsing the stats from the repo using ", subtree," as subtree" )
	report := NewReport()
	reader := bufio.NewReader(strings.NewReader(gitOutput))
	currentContributor := ""
	var timeString string
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
			timeString = strings.Replace(contribAndDate[1], "'", "", -1)
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
			
			date,_ := time.Parse("Mon Jan 2 15:04:05 2006 -0700", timeString)
				
			if !hasContributed {
				hasContributed = true
				report.AddContributor(currentContributor, periodMap)
				report.IncrementCommits(currentContributor, date)
			}
			report.IncrementCounters(currentContributor, additions, deletions, date)
		}
	}
	return report, nil
}

type OrderByScore []Contribution

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

func DecodeJson(jsonBlob []byte) (PeriodArray, error) {
	var periods PeriodArray
	err := json.Unmarshal(jsonBlob, &periods)
	if err != nil {
		fmt.Println("error:", err)
	}
	return periods, err
}

func main() {

	directory := flag.String("repo", "", "[mandatory] Path to the git repository")
	subtree := flag.String("subtree", "/", "[optionnal] Subtree you want to parse")
	config := flag.String("config", "", "[optionnal] Path to the configuration file")
	help := flag.Bool("help", false, "[optionnal] Displays this helps and quit")
	periods := *NewPeriodArray()

	flag.Parse()
	if *help {
		PrintHelp(true)
	}
	if *directory == "" {
		PrintHelp(false)
	}
	if *config != "" {
		fmt.Println("Using the config file ", *config)
		json, err := ioutil.ReadFile(*config)
		if err != nil {
			fmt.Println(chalk.Red, "Error while reading the configuration file ", err)
			os.Exit(1)
		}
		periods, err = DecodeJson(json)
		if err != nil {
			fmt.Println(chalk.Red, "Error while decoding the configuration file ", err)
			os.Exit(1)
		}
	}

	gitOutput, err := ExecGit(*directory)
	if err != nil {
		fmt.Println(chalk.Red, err)
		os.Exit(1)
	}

	report, err := ParseStats(gitOutput, *subtree, periods)

	separator := strings.Repeat("#", 80)
	fmt.Println(chalk.Green, separator)
	fmt.Println(chalk.Green, "Summing up contributions for the repository ", *directory, " subtree ", *subtree)
	fmt.Println(chalk.Green, separator)
	fmt.Println("")
	table := termtables.CreateTable()
	table.AddHeaders("Contributor", "Additions - Deletions", "Additions", "Commits", "Score")
	contributors := make([]Contribution, 0)
	for _, v := range report.Contributors {
		for _, contribution := range v.Contributions {
			if contribution.Commits > 0 {
				differenceScore := math.Abs(float64(contribution.Additions-contribution.Deletions)) * 100.0 / float64(report.TotalAdditions-report.TotalDeletions)
				additionScore := float64(contribution.Additions) * 100.0 / float64(report.TotalAdditions)
				commitScore := float64(contribution.Commits) * 100.0 / float64(report.TotalCommits)
				contribution.SetScores(differenceScore, additionScore, commitScore)
				contributors = append(contributors, *(contribution))
			}
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
