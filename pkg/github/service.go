package github

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
	"time"
)

func GetFile(githubOwner, githubRepo, filePath, githubUsername, githubToken string) ([]byte, error) {
	os.RemoveAll("/tmp/srebot")
	r, err := git.PlainClone("/tmp/srebot", false, &git.CloneOptions{
		URL:      fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", githubUsername, githubToken, githubOwner, githubRepo),
		Progress: os.Stdout,
	})
	if err != nil {
		return nil, err
	}

	_, err = r.Head()
	if err != nil {
		return nil, err
	}

	return os.ReadFile(fmt.Sprintf("/tmp/srebot/%s", filePath))
}
func ApplyChanges(
	githubOwner, githubRepo, githubUsername, githubToken, baseBranch, commitMsg,
	filePath, fileContent, prTitle, prDescription string) error {
	os.RemoveAll("/tmp/srebot")
	newBranch := fmt.Sprintf("sre-bot-%s", time.Now().Format("2006-01-02-15-04-05"))
	// Clone the repository
	r, err := git.PlainClone("/tmp/srebot", false, &git.CloneOptions{
		URL:      fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", githubUsername, githubToken, githubOwner, githubRepo),
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}

	// Create a new branch
	headRef, err := r.Head()
	if err != nil {
		return err
	}

	ref := plumbing.NewHashReference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", newBranch)), headRef.Hash())
	err = r.Storer.SetReference(ref)
	if err != nil {
		return err
	}

	// Checkout the new branch
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: ref.Name(),
	})
	if err != nil {
		return err
	}

	// Modify a file
	err = os.WriteFile(fmt.Sprintf("/tmp/srebot/%s", filePath), []byte(fileContent), 0644)
	if err != nil {
		return err
	}

	// Add the modified file to the staging area
	_, err = w.Add(filePath)
	if err != nil {
		return err
	}

	// Commit the changes
	_, err = w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "srebot",
			Email: "srebot@kaytu.io",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	// Push the changes to the remote repository
	err = r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: githubUsername, // this can be anything except an empty string
			Password: githubToken,
		},
	})
	if err != nil {
		return err
	}

	// Create a pull request
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	newPR := &github.NewPullRequest{
		Title:               github.String(prTitle),
		Head:                github.String(newBranch),
		Base:                github.String(baseBranch),
		Body:                github.String(prDescription),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(ctx, githubOwner, githubRepo, newPR)
	if err != nil {
		return err
	}

	fmt.Println(pr.GetHTMLURL())
	return nil
}
