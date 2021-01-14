package assessments

import (
	"context"
	"errors"
	"fmt"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/xcp"
)

type PRIntegration interface {
	Update(ctx context.Context, assessments Assessments)
}

type PRIntegrations []func(context.Context, *jnode.Node) PRIntegration

var prIntegrations = PRIntegrations{}

func RegisterPRIntegration(integ func(context.Context, *jnode.Node) PRIntegration) {
	prIntegrations = append(prIntegrations, integ)
}

func getCIIntegrationToken(client *api.Client) *jnode.Node {
	res, err := client.Post("/api/v1/org/{org}/git/ci-token", nil, xcp.WithCIEnvBody(""))
	if err != nil {
		if !errors.Is(err, api.HTTPError) {
			log.Warnf("Could not get CI integration config: {danger:%s}", err)
		}
		return nil
	}
	return res
}

func (a *Assessment) UpdatePR(client *api.Client) error {
	as := Assessments{*a}
	return as.UpdatePR(client)
}

func (a Assessments) UpdatePR(client *api.Client) error {
	token := getCIIntegrationToken(client)
	if token == nil {
		return fmt.Errorf("pull request integration has not been setup for this repository")
	}
	log.Infof("Updating pull request for %d assessments", len(a))
	ctx := context.Background()
	for _, integ := range prIntegrations {
		ci := integ(ctx, token)
		if ci != nil {
			ci.Update(ctx, a)
			return nil
		}
	}
	return fmt.Errorf("pull request integration not supported")
}
