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

type CIIntegration interface {
	Update(ctx context.Context, assessments Assessments)
}

type CIIntegrations []func(context.Context, *jnode.Node) CIIntegration

var ciIntegrations = CIIntegrations{}

func RegisterCIIntegration(integ func(context.Context, *jnode.Node) CIIntegration) {
	ciIntegrations = append(ciIntegrations, integ)
}

func getCIIntegrationToken(client *api.Client) *jnode.Node {
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

func (a Assessments) UpdateCI(client *api.Client) error {
	token := getCIIntegrationToken(client)
	if token == nil {
		return fmt.Errorf("CI integration has not been setup for this repository")
	}
	ctx := context.Background()
	for _, integ := range ciIntegrations {
		ci := integ(ctx, token)
		if ci != nil {
			ci.Update(ctx, a)
			return nil
		}
	}
	return fmt.Errorf("CI integration not supported")
}
