package main

import (
	"bufio"
	"fmt"
	"github.com/RodolpheFouquet/termtables"
	"github.com/kardianos/osext"
	"github.com/ttacon/chalk"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Contributor struct {
	Name      string
	Additions int
	Deletions int
	Commits   int
}

func NewContributor(name string) *Contributor {
	return &Contributor{Name: name, Additions: 0, Deletions: 0, Commits: 0}
}

func (c *Contributor) IncrementCounters(additions, deletions int) {
	c.Additions = additions + c.Additions
	c.Deletions = deletions + c.Deletions
}

func PrintHelp(success bool) {
	execname, _ := osext.Executable()
	var color chalk.Color
	if success {
		color = chalk.Green
	} else {
		color = chalk.Red
	}
	fmt.Println(color, "Usage: ", execname, "repo_path", "subtree")
	fmt.Println(color, "\tExample: ", execname, "repo_path ", "/", " will give the stats for the whole repository")
	fmt.Println(color, "\tExample: ", execname, "repo_path ", "/module1/src", " will give the stats for the module1/src subpath")
}

func ExecGit(repo string) (string, error) {
	command := exec.Command("git", "-C", repo, "log", "--numstat", "--pretty='%an'")
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

func ParseStats(gitOutput, subtree string) (map[string]*Contributor, error) {
	fmt.Println("Parsing the stats from the repo")
	contributors := make(map[string]*Contributor)
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
			currentContributor = strings.Replace(lineString, "'", "", -1)
			_, exists := contributors[currentContributor]
			if !exists {
				contributors[currentContributor] = NewContributor(currentContributor)
			}
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
			if !hasContributed {
				hasContributed = true
				contributors[currentContributor].Commits++
			}
			contributors[currentContributor].IncrementCounters(additions, deletions)
		}
	}
	return contributors, nil
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--help" {
		PrintHelp(true)
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		PrintHelp(false)
		os.Exit(1)
	}

	gitOutput, err := ExecGit(os.Args[1])
	if err != nil {
		fmt.Println(chalk.Red, err)
		os.Exit(1)
	}

	contributors, err := ParseStats(gitOutput, os.Args[2])

	separator := strings.Repeat("#", 80)
	fmt.Println(chalk.Green, separator)
	fmt.Println(chalk.Green, "Summing up contributions for the repository ", os.Args[1], " subtree ", os.Args[2])
	fmt.Println(chalk.Green, separator)
	fmt.Println("")
	table := termtables.CreateTable()
	table.AddHeaders("Contributor", "Additions", "Deletions", "Commits")
	for _, v := range contributors {
		table.AddRow(v.Name, v.Additions, v.Deletions, v.Commits)
	}

	fmt.Println(table.Render())
}
