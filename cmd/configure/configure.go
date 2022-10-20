package configure

import (
	"fmt"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

type configureCommand struct {
	options.PrintClientOpts
	laceworkProfileName string
}

func (c *configureCommand) Register(cmd *cobra.Command) {
	c.PrintClientOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVar(&c.laceworkProfileName, "lacework-profile", "",
		"Initialize using this lacework `profile`.  By default the IAC component will use the lacework profile with the same name as the current profile, or \"default\" if no profile is explicitly given.")
	flags.Lookup("organization").Hidden = false
	flags.Lookup("api-server").Hidden = false
}

func (c *configureCommand) Run() (*jnode.Node, error) {
	cfg := config.Config
	// we do not want to use the legacy token
	cfg.APIToken = ""
	if cfg.GetLaceworkAccount() == "" {
		// If we don't have an account then we'll get it (along with
		// the api key and secret) from a lacework profile
		laceworkProfileName := c.laceworkProfileName
		if laceworkProfileName == "" {
			laceworkProfileName = cfg.LaceworkProfileName
		}
		if laceworkProfileName == "" {
			lwp := config.GetDefaultLaceworkProfile()
			if lwp != nil {
				laceworkProfileName = lwp.Name
			} else {
				lwps := config.GetLaceworkProfiles()
				if len(lwps) == 1 {
					laceworkProfileName = lwps[0].Name
				} else {
					log.Errorf("You have multiple lacework profiles configured.  Choose a specific one with --lacework-profile.")
					log.Infof("Your lacework profiles are:")
					for _, lwp := range lwps {
						log.Infof("   {primary:%s} {info:(%s)}", lwp.Name, lwp.Account)
					}
					return nil, fmt.Errorf("choose a lacework profile with --lacework-profile")
				}
			}
		}
		cfg.SetLaceworkProfile(laceworkProfileName)
	}
	api, err := c.GetAPIClient()
	if err != nil {
		return nil, err
	}
	if api.LaceworkAPIToken == "" {
		return nil, fmt.Errorf("lacework CLI configuration is missing, run 'lacework configure' first")
	}
	result, err := api.Get("/api/v1/users/profile")
	if err != nil {
		return nil, err
	}
	cfg.APIServer = api.APIServer
	cfg.Organization = api.Organization
	if cfg.Organization == "" {
		// If no organization has been given and the user is a member
		// of a single org, then use that org.  Otherwise require the
		// user to be specific about which org to use.
		orgs := result.Path("data").Path("organizations")
		if orgs.IsArray() && orgs.Size() == 1 {
			cfg.Organization = orgs.Get(0).Path("orgId").AsText()
		} else {
			log.Errorf("You are a member of multiple IAC organizations.  A specific one must be specified with --organization.")
			log.Infof("Your organizations are:")
			for _, org := range orgs.Elements() {
				orgID := org.Path("orgId").AsText()
				log.Infof("  {primary:%s} {info:(%s)}", orgID, org.Path("displayName").AsText())
			}
			return nil, fmt.Errorf("specify an organization with --organization")
		}
	}
	if err := config.Save(); err != nil {
		return nil, err
	}
	log.Infof("IAC has been configured for account {primary:%s} organization {primary:%s}",
		cfg.GetLaceworkAccount(), cfg.Organization)
	return result, nil
}
