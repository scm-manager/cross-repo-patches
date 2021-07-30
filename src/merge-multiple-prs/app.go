package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v37/github"
	"golang.org/x/oauth2"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage merge-multiple-prs 'title of prs to merge'")
		os.Exit(1)
	}

	title := os.Args[1]

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

	var mergeRequests []mergeRequest
	for _, repo := range repos {
		fmt.Println("check", repo.GetName())
		mergeRequests = append(mergeRequests, checkRepository(ctx, client, repo.GetName(), title)...)
	}

	if len(mergeRequests) == 0 {
		fmt.Println("no matching pr found")
		os.Exit(0)
	}

	var response string
	fmt.Printf("type doit to merge %d pull requests: \n", len(mergeRequests))
	_, err = fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	if response != "doit" {
		fmt.Println("abort")
		os.Exit(1)
	}

	for _, mr := range mergeRequests {
		result, _, err := client.PullRequests.Merge(ctx, "scm-manager", mr.repo, mr.pr, "", &github.PullRequestOptions{
			MergeMethod: "squash",
		})
		if err != nil {
			log.Fatal(err)
		}
		if !*result.Merged {
			fmt.Printf("failed to merge %d of %s\n", mr.pr, mr.repo)
		} else {
			fmt.Printf("successfully merged pr %d of %s", mr.pr, mr.repo)
		}

		fmt.Printf("Delete branch %s of repository %s", mr.ref, mr.repo)
		_, err = client.Git.DeleteRef(ctx, "scm-manager", mr.repo, mr.ref)
		if err != nil {
			fmt.Printf("failed to delete branch %v\n", err)
		}
	}
}

type mergeRequest struct {
	repo string
	pr   int
	ref  string
}

func checkRepository(ctx context.Context, client *github.Client, repo string, title string) []mergeRequest {
	prs, _, err := client.PullRequests.List(ctx, "scm-manager", repo, nil)
	if err != nil {
		log.Fatal(err)
	}

	var mergeRequests []mergeRequest

	for _, pr := range prs {
		prTitle := pr.GetTitle()
		if prTitle == title {
			if !pr.GetMergeable() {
				mergeRequests = append(mergeRequests, mergeRequest{
					pr:   pr.GetNumber(),
					repo: repo,
					ref:  pr.GetHead().GetRef(),
				})
			} else {
				fmt.Printf("matching pr %d of repository %s is not mergable\n", pr.GetNumber(), repo)
			}
		}
	}
	return mergeRequests
}
