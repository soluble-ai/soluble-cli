package buildreport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/exit"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.ToolOpts
	SkipCI bool
}

type CIIntegration interface {
	Update(ctx context.Context, assessments []tools.Assessment)
}

type CIIntegrations []func(context.Context, *jnode.Node) CIIntegration

var _ tools.Interface = &Tool{}

var integrations = CIIntegrations{}

func (t *Tool) Name() string {
	return "build-report"
}

func (t *Tool) Register(c *cobra.Command) {
	t.Internal = true
	t.NotUploadable = true
	t.ToolOpts.Register(c)
	flags := c.Flags()
	flags.BoolVar(&t.SkipCI, "skip-ci-integration", false, "Don't integrate with CI (e.g. don't create github checks)")
}

func (t *Tool) Run() (*tools.Result, error) {
	n, err := t.GetAPIClient().Get("/api/v1/org/{org}/assessments",
		func(r *resty.Request) {
			r.SetQueryParam("detail", "true")
			r.SetQueryParam("searchType", "ci")
		},
		xcp.WithCIEnv)
	if err != nil {
		return nil, err
	}
	assessments := []tools.Assessment{}
	findings := jnode.NewArrayNode()
	failures := []string{}
	for _, n := range n.Path("data").Elements() {
		assessment := tools.Assessment{}
		if err := json.Unmarshal([]byte(n.String()), &assessment); err != nil {
			return nil, err
		}
		var count int
		var level string
		assessment.Failed, count, level = assessment.HasFailures(t.ParsedFailThresholds)
		if assessment.Failed {
			failures = append(failures, fmt.Sprintf("{warning:%s} has {danger:%d %s findings}", assessment.Title, count, level))
		}
		assessments = append(assessments, assessment)
		for _, finding := range assessment.Findings {
			findings.AppendObject().Put("sid", finding.SID).
				Put("module", assessment.Module).
				Put("pass", finding.Pass).
				Put("severity", finding.Severity).
				Put("file", finding.FilePath).
				Put("title", finding.GetTitle())
		}
	}
	if !t.SkipCI {
		t.updateCI(assessments)
	}
	result := &tools.Result{
		Data:         jnode.NewObjectNode().Put("findings", findings),
		PrintPath:    []string{"findings"},
		PrintColumns: []string{"sid", "module", "pass", "severity", "file", "title"},
	}
	if len(failures) > 0 {
		exit.Code = 2
		exit.Func = func() {
			for _, m := range failures {
				log.Errorf(m)
			}
		}
	}
	return result, nil
}

func (t *Tool) updateCI(assessments []tools.Assessment) {
	token := getCIToken(t.GetAPIClient())
	if token == nil {
		return
	}
	if !t.SkipCI {
		ctx := context.Background()
		for _, integ := range integrations {
			ci := integ(ctx, token)
			if ci != nil {
				ci.Update(ctx, assessments)
				break
			}
		}
	}
}

func getCIToken(client *api.Client) *jnode.Node {
	body := jnode.NewObjectNode()
	for k, v := range xcp.GetCIEnv() {
		body.Put(k, v)
	}
	res, err := client.Post("/api/v1/org/{org}/git/ci-token", body)
	if err != nil {
		if !errors.Is(err, api.HTTPError) {
			log.Warnf("Could not get CI integration config: {danger:%s}", err)
		}
		return nil
	}
	return res
}

func RegisterIntegration(integ func(context.Context, *jnode.Node) CIIntegration) {
	integrations = append(integrations, integ)
}
