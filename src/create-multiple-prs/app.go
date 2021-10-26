package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v37/github"
	"golang.org/x/oauth2"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage create-prs 'source-branch' 'title of prs to create'")
		os.Exit(1)
	}

	sourceBranch := os.Args[1]
	title := os.Args[2]

	accessToken := os.Getenv("GITHUB_OAUTH_TOKEN")
	if accessToken == "" {
		fmt.Println("could not find env GITHUB_OAUTH_TOKEN")
		os.Exit(1)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	repos, _, err := client.Repositories.List(ctx, "scm-manager", &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	var reposWithMissingPrs []*github.Repository
	for _, repo := range repos {
		fmt.Println("check", repo.GetName())
		if hasOpenPr(ctx, client, repo.GetName(), title) {
			fmt.Println("repository has already an open pr with the title")
		} else if hasSourceBranch(ctx, client, repo.GetName(), sourceBranch) {
			fmt.Println("add repository to list")
			reposWithMissingPrs = append(reposWithMissingPrs, repo)
		} else {
			fmt.Println("repository has no such source branch")
		}
	}

	if len(reposWithMissingPrs) == 0 {
		fmt.Println("no matching repositories found")
		os.Exit(0)
	}

	var response string
	fmt.Printf("type doit to create %d pull requests: \n", len(reposWithMissingPrs))
	_, err = fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	if response != "doit" {
		fmt.Println("abort")
		os.Exit(1)
	}

	for _, repo := range reposWithMissingPrs {
		fmt.Println("create pr for repo", repo.GetName())

		target := repo.GetDefaultBranch()

		_, resp, err := client.PullRequests.Create(ctx, "scm-manager", repo.GetName(), &github.NewPullRequest{
			Title: &title,
			Head:  &sourceBranch,
			Body:  &title,
			Base:  &target,
		})
		if err != nil {
			if resp != nil && resp.StatusCode == 403 {
				fmt.Println("hit rate limit, sleeping 1 Minute")
				time.Sleep(1 * time.Minute)
				_, _, err = client.PullRequests.Create(ctx, "scm-manager", repo.GetName(), &github.NewPullRequest{
					Title: &title,
					Head:  &sourceBranch,
					Body:  &title,
					Base:  &target,
				})
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
		}
	}
}

func hasOpenPr(ctx context.Context, client *github.Client, repo string, title string) bool {
	prs, _, err := client.PullRequests.List(ctx, "scm-manager", repo, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, pr := range prs {
		prTitle := pr.GetTitle()
		if prTitle == title {
			return true
		}
	}
	return false
}

func hasSourceBranch(ctx context.Context, client *github.Client, repo string, sourceBranch string) bool {
	_, response, err := client.Repositories.GetBranch(ctx, "scm-manager", repo, sourceBranch, false)
	if response != nil {
		if response.StatusCode == 200 {
			return true
		} else if response.StatusCode == 404 {
			return false
		}
		log.Fatal(err)
	}
	log.Fatal(err)
	return false
}
