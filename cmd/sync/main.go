package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/xabi/dependabot-syncher/dependabot"
	"github.com/xabi/dependabot-syncher/github"
)

const CommitMsg = "Setup dependabot conf"

type Repos map[string]dependabot.Repository

func (repos Repos) Add(repoName string, lang dependabot.Lang, path string) {
	if r, ok := repos[repoName]; ok {

		r.Deps = append(r.Deps, dependabot.DependencyFile{
			Lang: lang,
			Path: path,
		})
		repos[repoName] = r
	} else {
		repos[repoName] = dependabot.Repository{
			Name: repoName,
			Deps: []dependabot.DependencyFile{
				{
					Lang: lang,
					Path: path,
				},
			},
		}
	}
}

func run(args []string, stdout io.Writer) error {
	if len(os.Args) < 2 {
		return errors.New("Missing repo owner")
	}

	ctx := context.Background()

	gh := github.New(ctx, os.Args[1], os.Getenv("GITHUB_TOKEN"))

	reposToUpdate := Repos{}

	for _, l := range dependabot.Langs {
		repos, err := gh.SearchForFile(ctx, l.DepsFile())
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range repos {
			reposToUpdate.Add(r.Repository.GetName(), l, r.GetPath())
		}

	}

	fmt.Fprintf(stdout, "%d repositories to update\n", len(reposToUpdate))

	for _, r := range reposToUpdate {
		b, err := dependabot.Generate(reposToUpdate[r.Name])
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintf(stdout, "updating %s repository\n", r.Name)

		if err := gh.UpdateFile(ctx, CommitMsg, r.Name, dependabot.ConfPathFile, b); err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

const (
	// exitFail is the exit code if the program
	// fails.
	exitFail = 1
)

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
}
