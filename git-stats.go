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

// JSON Periods

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

func IsAfter(t, other time.Time) bool {
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

// Contributions

func NewContribution(name string) *Contribution {
	return &Contribution{Additions: 0, Deletions: 0, Commits: 0, Name: name}
}

func NewContributionDate(name string, period PeriodTS) *Contribution {
	formattedName := fmt.Sprintf("%v (%v)", name, period.Alias)
	if period.Alias == "" {
		formattedName = name
	}
	return &Contribution{Additions: 0, Deletions: 0, Commits: 0, Name: formattedName, StartDate: period.Start, EndDate: period.End}
}

// Json Users

type User struct {
	Alias string `json:"alias"`
	Name  string `json:"name"`
}

type UserTS struct {
	Alias string
	Name  string
}

type UserArray struct {
	Users []User `json:"users"`
}

func NewUserArray() *UserArray {
	return &UserArray{Users: []User{}}
}

// Contributors

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

// Counters, score, report

func (c *Contribution) IncrementCounters(additions, deletions int) {
	c.Additions = additions + c.Additions
	c.Deletions = deletions + c.Deletions
}

func (c *Contribution) GetScore() float64 {
	threshold := 0.075
	score := 0.7*c.DifferenceScore + 0.15*c.AdditionScore + 0.15*c.CommitScore
	if (score < threshold) {
		score = 0.0
	}
	return score
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
	TotalScore     float64
}

func NewReport() *Report {
	return &Report{Contributors: make(map[string]*Contributor), TotalAdditions: 0, TotalDeletions: 0, TotalCommits: 0, TotalScore: 0.0}
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
		fmt.Println("This contributor does not exist: ", r.Contributors[name] )
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
		fmt.Println("This contributor does not exist: ", r.Contributors[name] )
		return errors.New("This contributor does not exist")
	}
	contrib := GetContribution(r.Contributors[name].Contributions, date)
	contrib.Commits++
	r.TotalCommits++
	return nil
}

func ExecGitHistory(repo string) (string, error) {
	command := exec.Command("git", "-C", repo, "log", "--numstat", "--pretty='%an|%ad'")
	fmt.Println("Gathering the stats in the repo (1/3)", repo)
	out, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func parseGitOutputHistory(gitOutput string, report *Report, subtree string, periodMap map[string][]PeriodTS, userMap map[string]string) {
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
			alias := strings.Replace(contribAndDate[0], "'", "", -1)
			_, exists := userMap[alias]
			if exists {
				currentContributor = userMap[alias]
				if (currentContributor == "") {
					fmt.Println(chalk.Yellow, "Skip user: ", alias)
					continue
				}
			} else {
				currentContributor = alias
			}
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
				additions = 0
			}
			deletions, err := strconv.Atoi(splittedLine[1])
			if err != nil {
				deletions = 0
			}
			
			date,_ := time.Parse("Mon Jan 2 15:04:05 2006 -0700", timeString)
				
			if !hasContributed {
				hasContributed = true
				report.AddContributor(currentContributor, periodMap)
				report.IncrementCommits(currentContributor, date)
			}
			report.IncrementCounters(currentContributor, additions, deletions, date)
		} else {
			fmt.Println(chalk.Yellow, "Error: unprocessed line (history): ", lineString)
		}
	}
}

func parseGitOutputBlame(gitOutput string, report *Report, userMap map[string]string) {
	reader := bufio.NewReader(strings.NewReader(gitOutput))
	currentContributor := ""
	var timeString string
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lineString := string(line)
		if len(string(line)) == 0 {
			continue
		}

		splittedLine := strings.Split(strings.Trim(lineString, " "), " ")

		if len(splittedLine) >= 3 {
			alias := strings.Join(splittedLine[2:], " ")
			_, exists := userMap[alias]
			if exists {
				currentContributor = userMap[alias]
				if (currentContributor == "") {
					fmt.Println(chalk.Yellow, "Skip user: ", alias)
					continue
				}
			} else {
				currentContributor = alias
			}

			additions, err := strconv.Atoi(splittedLine[0])
			if err != nil {
				fmt.Println(chalk.Yellow, "Skip blame contribution: ", lineString)
				additions = 0
			}

			//increment as additions
			date,_ := time.Parse("Mon Jan 2 15:04:05 2023 -0700", timeString)
			factor := 1
			report.IncrementCounters(currentContributor, additions * factor, 0, date)
		} else {
			fmt.Println(chalk.Yellow, "Error: unprocessed line (blame): ", len(splittedLine), lineString)
		}
	}
}

func ExecGitBlameRaw(repo string) (string, error) {
	cmdGit := "git ls-tree -r -z --name-only HEAD -- | grep -z -Z -v extra_lib | sed 's/^/.\\//' | xargs -0 -n1 git blame --line-porcelain HEAD |grep -ae \"^author \"|sort|uniq -c|sort -nr"
	command := exec.Command("bash", "-c", cmdGit)
	command.Dir = repo
	out, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func ExecGitBlameSelected(repo string) (string, error) {
	cmdGit := "git ls-tree --name-only -z -r HEAD|egrep -z -Z -E 'configure|Makefile|\\.(h|cpp|c|js)$'|grep -z -Z -v extra_lib|xargs -0 -n1 git blame --line-porcelain|grep \"^author \"|sort|uniq -c|sort -nr"
	command := exec.Command("bash", "-c", cmdGit)
	command.Dir = repo
	fmt.Println("Gathering the stats in the repo (3/3)", repo)
	out, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func ParseStats(gitOutput1 string, gitOutput2 string, gitOutput3 string, subtree string, periods PeriodArray, users UserArray) (*Report, error) {
	periodMap := make(map[string][]PeriodTS)
	for _, period := range periods.Periods {
		periodMap[period.User] = append(periodMap[period.User], *NewPeriodTS(period))
	}
	userMap := make(map[string]string)
	for _, user := range users.Users {
		userMap[user.Alias] = user.Name
	}
	fmt.Println("Parsing the stats from the repo using ", subtree," as subtree" )
	report := NewReport()

	parseGitOutputHistory(gitOutput1, report, subtree, periodMap, userMap)
	parseGitOutputBlame(gitOutput2, report, userMap)
	parseGitOutputBlame(gitOutput3, report, userMap)

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

func DecodeJson(jsonBlob []byte) (PeriodArray, UserArray, error) {
	var periods PeriodArray
	var users UserArray
	err := json.Unmarshal(jsonBlob, &periods)
	if err != nil {
		fmt.Println("error:", err)
	}
	err = json.Unmarshal(jsonBlob, &users)
	if err != nil {
		fmt.Println("error:", err)
	}
	return periods, users, err
}

func main() {
	directory := flag.String("repo", "", "[mandatory] Path to the git repository")
	subtree := flag.String("subtree", "/", "[optional] Subtree you want to parse")
	config := flag.String("config", "", "[optional] Path to the configuration file")
	help := flag.Bool("help", false, "[optional] Displays this helps and quit")
	periods := *NewPeriodArray()
	users:= *NewUserArray()

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
		periods, users, err = DecodeJson(json)
		if err != nil {
			fmt.Println(chalk.Red, "Error while decoding the configuration file ", err)
			os.Exit(1)
		}
	}

	gitOutputHistory, err := ExecGitHistory(*directory)
	if err != nil {
		fmt.Println(chalk.Red, err)
		os.Exit(1)
	}

	gitOutputBlameRaw, err := ExecGitBlameRaw(*directory)
	if err != nil {
		fmt.Println(chalk.Red, err)
		os.Exit(1)
	}

	gitOutputBlameSelected, err := ExecGitBlameSelected(*directory)
	if err != nil {
		fmt.Println(chalk.Red, err)
		os.Exit(1)
	}

	report, err := ParseStats(gitOutputHistory, gitOutputBlameRaw, gitOutputBlameSelected, *subtree, periods, users)

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
				decreaseFactor := 3.0
				differenceScore := math.Max(float64(contribution.Additions-contribution.Deletions), float64(contribution.Deletions-contribution.Additions) / decreaseFactor) * 100.0 / float64(report.TotalAdditions-report.TotalDeletions)
				additionScore := float64(contribution.Additions) * 100.0 / float64(report.TotalAdditions)
				commitScore := float64(contribution.Commits) * 100.0 / float64(report.TotalCommits)
				contribution.SetScores(differenceScore, additionScore, commitScore)
				contributors = append(contributors, *(contribution))
				report.TotalScore += contribution.GetScore()
			}
		}
	}
	sort.Sort(OrderByScore(contributors))
	for index := range contributors {
		c := contributors[len(contributors)-index-1]
		if (c.GetScore() > 0) { // hide micro-contributors
			table.AddRow(c.Name, fmt.Sprintf("%.3f%%", c.DifferenceScore), fmt.Sprintf("%.3f%%", c.AdditionScore), fmt.Sprintf("%.3f%%", c.CommitScore), fmt.Sprintf("%.3f", c.GetScore() * 100.0 / report.TotalScore))
		}
	}

	table.AddSeparator()
	table.AddRow("Total", report.TotalAdditions, report.TotalDeletions, report.TotalCommits, "100.0")
	table.SetAlign(3, 2)
	table.SetAlign(3, 3)
	table.SetAlign(3, 4)
	table.SetAlign(3, 5)
	fmt.Println(table.Render())
}
