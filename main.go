package main

import (
	"fmt"
	"strings"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	"github.com/tj/docopt"

	c "github.com/hstove/gender/classifier"
)

var client *github.Client

const (
	Version = `1.0.0`
	Usage   = `Gender Stats.

Usage:
  stats <owner> <repo> <token>
  stats -h | --help
  stats --version

Options:
  -h --help     Show this screen.
  --token       Your Github token.
  --version     Show version.`
)

func main() {
	arguments, err := docopt.Parse(Usage, nil, true, Version, false)
	check(err)

	owner := arguments["<owner>"].(string)
	repo := arguments["<repo>"].(string)
	token := arguments["<token>"].(string)

	initializeClient(token)

	contributors := getContributors(owner, repo)
	names := getNames(contributors)

	percentFemale, percentMale := predictGenderStats(names)
	percentUnknown := (100 - percentFemale - percentMale)

	fmt.Println("\nContributors by Gender:")
	fmt.Printf("\n  - Female: %.2f%%\n", percentFemale)
	fmt.Printf("\n  - Male: %.2f%%\n", percentMale)
	fmt.Printf("\n  - Unknown: %.2f%%\n\n", percentUnknown)
}

func initializeClient(token string) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client = github.NewClient(tc)
}

func getContributors(owner, repo string) []github.Contributor {
	var contributors []github.Contributor

	options := &github.ListContributorsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		c, response, err := client.Repositories.ListContributors(owner, repo, options)
		check(err)
		contributors = append(contributors, c...)
		if response.NextPage == 0 {
			break
		}
		options.ListOptions.Page = response.NextPage
	}

	return contributors
}

func getNames(contributors []github.Contributor) []string {
	var names []string

	for _, c := range contributors {
		user, _, err := client.Users.Get(*c.Login)
		check(err)
		var name string
		if user.Name == nil {
			name = ""
		} else {
			name = strings.Split(*user.Name, " ")[0]
		}
		names = append(names, name)
	}

	return names
}

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func predictGenderStats(names []string) (f, m float64) {
	classifier := c.Classifier()

	var numFemale int
	var femaleNames []string

	var numMale int
	var maleNames []string

	for _, name := range names {
		gender, _ := c.Classify(classifier, name)

		if gender == string(c.Girl) {
			numFemale += 1
			femaleNames = append(femaleNames, name)
		}

		if gender == string(c.Boy) {
			numMale += 1
			maleNames = append(maleNames, name)
		}
	}

	printNames(maleNames, femaleNames)

	numTotal := len(names)

	f = percent(numFemale, numTotal)
	m = percent(numMale, numTotal)

	return
}

func printNames(maleNames, femaleNames []string) {
	fmt.Printf(
		"\nMALE (%d):\n%s\n",
		len(maleNames),
		strings.Join(maleNames, "\n"),
	)

	fmt.Printf(
		"\nFEMALE (%d):\n%s\n",
		len(femaleNames),
		strings.Join(femaleNames, "\n"),
	)
}

func percent(numForGender, numTotal int) float64 {
	return (float64(numForGender) / float64(numTotal) * 100)
}
