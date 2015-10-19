package main

import (
	"bufio"
	"fmt"
	"github.com/apcera/termtables"
	"github.com/kardianos/osext"
	"github.com/ttacon/chalk"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Contributor struct {
	Name      string
	Additions int
	Deletions int
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
}

func ExecGit(repo string) (string, error) {
	command := exec.Command("git", "-C", "/home/tamareu/Code/gpac", "log", "--numstat", "--pretty='%an'")
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

func ParseStats(gitOutput string) (map[string]*Contributor, error) {
	fmt.Println("Parsing the stats from the repo")
	contributors := make(map[string]*Contributor)
	reader := bufio.NewReader(strings.NewReader(gitOutput))
	currentContributor := ""
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		lineString := string(line)
		if len(string(line)) == 0 {
			continue
		}

		switch lineString[0] {
		case '-':
			continue
		case '\'':
			currentContributor = strings.Replace(lineString, "'", "", -1)
			_, exists := contributors[currentContributor]
			if !exists {
				contributors[currentContributor] = &Contributor{Name: currentContributor, Additions: 0, Deletions: 0}
			}
		default:
			splittedLine := strings.Split(lineString, "\t")
			additions, err := strconv.Atoi(splittedLine[0])
			if err != nil {
				fmt.Println(chalk.Yellow, "Warning: ", err)
				continue
			}
			deletions, err := strconv.Atoi(splittedLine[1])
			if err != nil {
				fmt.Println(chalk.Yellow, "Warning: ", err)
				continue
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

	contributors, err := ParseStats(gitOutput)

	separator := strings.Repeat("#", 80)
	fmt.Println(chalk.Green, separator)
	fmt.Println(chalk.Green, "Summing up contributions for the repository ", os.Args[1], " subtree ", os.Args[2])
	fmt.Println(chalk.Green, separator)
	fmt.Println("")
	table := termtables.CreateTable()
	table.AddHeaders("Contributor", "Additions", "Deletions")
	for _, v := range contributors {
		table.AddRow(v.Name, v.Additions, v.Deletions)
	}

	fmt.Println(table.Render())
}
