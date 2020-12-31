package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/assessments"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"golang.org/x/oauth2"
)

type Integration struct {
	owner  string
	repo   string
	commit string
	gh     *github.Client
}

func NewIntegration(ctx context.Context, config *jnode.Node) assessments.CIIntegration {
	gitRepo := config.Path("gitRepo").AsText()
	if gitRepo == "" {
		return nil
	}
	p := strings.Split(gitRepo, "/")
	if len(p) == 3 && p[0] == "github.com" {
		return &Integration{
			owner:  p[1],
			repo:   p[2],
			commit: config.Path("gitCommit").AsText(),
			gh: github.NewClient(oauth2.NewClient(ctx,
				oauth2.StaticTokenSource(&oauth2.Token{
					AccessToken: config.Path("token").AsText(),
				}))),
		}
	}
	return nil
}

func (it *Integration) Update(ctx context.Context, assessments assessments.Assessments) {
	if len(assessments) == 0 {
		return
	}
	for _, a := range assessments {
		if a.URL == "" {
			continue
		}
		output := &github.CheckRunOutput{
			Title:   &a.Title,
			Summary: &a.Markdown,
			Text:    stringp(fmt.Sprintf("See <%s>", a.URL)),
		}
		checkRunOptions := github.CreateCheckRunOptions{
			HeadSHA:     it.commit,
			Name:        a.Title,
			DetailsURL:  &a.URL,
			ExternalID:  &a.ID,
			Status:      stringp("completed"),
			CompletedAt: &github.Timestamp{Time: time.Now()},
			Output:      output,
		}
		if a.Failed {
			checkRunOptions.Conclusion = stringp("failure")
		} else {
			checkRunOptions.Conclusion = stringp("success")
		}
		checkRun, _, err := it.gh.Checks.CreateCheckRun(ctx, it.owner, it.repo, checkRunOptions)
		if err != nil {
			log.Warnf("Could not create github check: {danger:%s}", err)
		} else {
			log.Infof("Updated github check run {primary:%s} id %d HeadSHA %s", a.Title, checkRun.GetID(), it.commit)
		}
	}
}

func stringp(s string) *string {
	if s != "" {
		return &s
	}
	return nil
}

func init() {
	assessments.RegisterCIIntegration(NewIntegration)
}
