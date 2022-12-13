package configure

import (
	"fmt"
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/api"
	"github.com/soluble-ai/soluble-cli/pkg/config"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

type configureCommand struct {
	options.PrintClientOpts
	laceworkProfileName string
	reconfigure         bool
	clientHook          func(*api.Client) // for testing only
}

func (c *configureCommand) Register(cmd *cobra.Command) {
	c.PrintClientOpts.Register(cmd)
	flags := cmd.Flags()
	flags.StringVar(&c.laceworkProfileName, "lacework-profile", "",
		"Initialize using this lacework `profile`.  By default the IAC component will use the lacework profile with the same name as the current profile, or \"default\" if no profile is explicitly given.")
	flags.BoolVar(&c.reconfigure, "reconfigure", false, "Reconfigure the IAC component even if it already has been configured")
	flags.Lookup("iac-organization").Hidden = false
	flags.Lookup("api-server").Hidden = false
}

func (c *configureCommand) Run() (*jnode.Node, error) {
	cfg := config.Get()
	// we do not want to use the legacy token
	cfg.APIToken = ""
	os.Unsetenv("SOLUBLE_API_TOKEN")
	apiConfig := c.APIConfig.SetValues()
	if !config.IsRunningAsComponent() {
		// If we're not running as a component, then we need to link
		// this profile to a lacework profile.
		laceworkProfileName := c.laceworkProfileName
		if laceworkProfileName == "" {
			laceworkProfileName = cfg.LaceworkProfileName
		}
		if laceworkProfileName == "" || c.reconfigure {
			lwp := config.GetDefaultLaceworkProfile()
			if lwp != nil {
				laceworkProfileName = lwp.Name
			} else {
				lwps := config.GetLaceworkProfiles()
				switch {
				case len(lwps) == 1:
					laceworkProfileName = lwps[0].Name
				case len(lwps) == 0:
					log.Errorf("You must install and configure the lacework CLI first.")
					log.Infof("See {info:https://docs.lacework.com/cli} for more information.")
					return nil, fmt.Errorf("the lacework CLI must be configured")
				default:
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
		// now set values again, this will update based on the profile
		apiConfig.SetValues()
	}
	if err := apiConfig.Validate(false); err != nil {
		return nil, err
	}
	api := api.NewClient(apiConfig)
	if c.clientHook != nil {
		c.clientHook(api)
	}
	result, err := api.Get("/api/v1/users/profile")
	if err != nil {
		return nil, err
	}
	cfg.APIServer = api.APIServer
	cfg.Organization = api.Organization
	if cfg.LaceworkProfileName == "" {
		cfg.ConfiguredAccount = api.LaceworkAccount
	}
	if cfg.Organization == "" || c.reconfigure {
		// If no organization has been given and the user is a member
		// of a single org, then use that org.  Otherwise require the
		// user to be specific about which org to use.
		orgs := result.Path("data").Path("organizations")
		if orgs.IsArray() && orgs.Size() == 1 {
			cfg.Organization = orgs.Get(0).Path("orgId").AsText()
		} else {
			log.Errorf("You are a member of multiple IAC organizations.  A specific one must be specified with --iac-organization.")
			log.Infof("Your organizations are:")
			for _, org := range orgs.Elements() {
				orgID := org.Path("orgId").AsText()
				log.Infof("  {primary:%s} {info:(%s)}", orgID, org.Path("displayName").AsText())
			}
			return nil, fmt.Errorf("specify an IAC organization with --iac-organization")
		}
	}
	if err := config.Save(); err != nil {
		return nil, err
	}
	log.Infof("IAC has been configured for account {primary:%s} organization {primary:%s}",
		api.LaceworkAccount, cfg.Organization)
	return result, nil
}
