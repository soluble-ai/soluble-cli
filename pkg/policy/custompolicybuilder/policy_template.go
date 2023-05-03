package custompolicybuilder

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/policy"
	"gopkg.in/yaml.v3"
)

type PolicyTemplate struct {
	Name      string
	CheckType string
	Dir       string
	Tool      string
	Desc      string
	Title     string
	Severity  string
	Category  string
	RsrcType  string
	Provider  string
	InputPath string
}

var categories = []string{
	"IAM",
	"Storage",
	"Network",
	"Loadbalancers",
	"Compute",
	"Certs",
	"Secrets",
	"Encryption",
	"TLS",
	"Logging",
	"Dns",
	"Queues",
	"Containers",
	"Monitoring",
	"Tools",
	"Security",
	"General",
	"Backup & Recovery",
}
var gcpResourceTypes = []string{
	"multiple",
	"google_bigquery_dataset",
	"google_compute_instance",
	"google_dns_managed_zone",
	"google_kms_crypto_key",
	"google_storage_bucket",
	"google_logging_project_sink",
}

var azureResourceTypes = []string{
	"multiple",
	"azurerm_app_service",
	"azurerm_role_definition",
	"azurerm_key_vault",
	"azurerm_kubernetes_cluster",
	"azurerm_monitor_diagnostic_setting",
	"azurerm_mysql_server",
	"azurerm_application_gateway",
	"azurerm_network_security_group",
	"azurerm_network_watcher_flow_log",
	"azurerm_postgresql_server",
	"azurerm_security_center_contact",
	"azurerm_sql_server",
	"azurerm_sql_database",
	"azurerm_diagnostic_settings",
	"azurerm_storage_account",
}
var awsResourceTypes = []string{
	"multiple",
	"aws_api_gateway",
	"aws_cloudfront_distribution",
	"aws_cloudTrail",
	"aws_dynamodb_table",
	"aws_ebc_volume",
	"aws_elasticache_cluster",
	"aws_elb",
	"aws_lb",
	"aws_iam_policy",
	"aws_iam_group",
	"aws_iam_group_policy",
	"aws_iam_role",
	"aws_iam_role_policy",
	"aws_iam_role_policy_attachment",
	"aws_iam_user",
	"aws_iam_user_policy",
	"aws_iam_user_policy_attachment",
	"aws_iam_instance_profile",
	"aws_iam_account_password_policy",
	"aws_lambda_function",
	"aws_db_instance",
	"aws_rds_cluster",
	"aws_redshift_cluster",
	"aws_s3_bucket",
	"aws_security_group",
	"aws_sns_topic_subscription",
	"aws_vpc",
	"aws_flow_log",
	"aws_network_acl",
	"aws_wafv2_web_acl",
}

var severity = []string{
	"Info",
	"Low",
	"Medium",
	"High",
	"Critical",
}

var providers = []string{
	"AWS",
	"GCP",
	"Azure",
	"Kubernetes",
	"Github",
	"Oracle",
}

func getCheckTypes() []string {
	checkTypes := policy.InputTypeForTarget
	keys := make([]string, 0, len(checkTypes))
	for key := range checkTypes {
		keys = append(keys, string(key))
	}
	return keys
}

func (pt *PolicyTemplate) PromptInput() error {
	var qs = []*survey.Question{
		{
			Name: "InputPath",
			Prompt: &survey.Input{
				Message: "Policies directory path:",
			},
			Validate: survey.ComposeValidators(survey.Required, pt.validatePolicyDirectory()),
		},
		{
			Name: "provider",
			Prompt: &survey.Select{
				Message: "Select provider:",
				Options: providers,
			},
		},
		{
			Name: "checkType",
			Prompt: &survey.Select{
				Message: "Select target:",
				Options: getCheckTypes(),
				Help:    "type of check (think: what is being examined?)",
			},
		},
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Policy name:",
				Help:    "Policy name may consist of lowercase letters, numbers and underscores. e.g. my_policy_1"},
			Validate: pt.validatePolicyName(),
		},
		{
			Name: "title",
			Prompt: &survey.Input{
				Message: "Title",
				Help:    "Title for your policy. Maximum length is 57 characters. e.g. 'Disk volumes must be encrypted'",
			},
			Validate: survey.ComposeValidators(survey.MinLength(1), survey.MaxLength(57)),
		},
		{
			Name: "desc",
			Prompt: &survey.Input{
				Message: "Description",
				Help:    "Longer description of your policy. You may want to include, for example, what the policy checks for, and the rationale for having the policy.",
			},
		},
		{
			Name: "category",
			Prompt: &survey.Input{
				Message: "Category",
				Help:    "functional grouping of the check",
				Suggest: func(input string) []string {
					return categories
				},
			},
			Validate: func(input interface{}) error {
				if isValid := regexp.MustCompile(`(^[A-Z][a-z_]*$)`).MatchString(input.(string)); !isValid {
					return fmt.Errorf("\ncategory must: \n-start with uppercase letter \n-only contain letters and underscores")
				}
				return nil
			},
		},
		{
			Name: "rsrcType",
			Prompt: &survey.Input{
				Message: "ResourceType\033[37m for suggestions type aws, google or azure then tab",
				Help:    "for multiple resource types use multiple",
				Suggest: func(input string) []string {
					var suggestions []string
					switch input {
					case "aws":
						suggestions = awsResourceTypes
					case "google", "gcp":
						suggestions = gcpResourceTypes
					case "azure":
						suggestions = azureResourceTypes
					case "m":
						suggestions = append(suggestions, "multiple")
					}
					return suggestions
				},
			},
			Validate: func(input interface{}) error {
				if isValid := regexp.MustCompile(`(^[a-z][a-z0-9_]*[a-z0-9]$)`).MatchString(input.(string)); !isValid {
					return fmt.Errorf("\nResource type must: \n- start with lowercase letter \n" +
						"- Only contain lowercase letters, numeric digits, and underscores\n" +
						"- End with a lowercase letter or numeric digit")
				}
				return nil
			},
		},
		{
			Name: "severity",
			Prompt: &survey.Select{
				Message: "Select severity:",
				Options: severity,
			},
		},
	}

	if err := survey.Ask(qs, pt); err != nil {
		return err
	}
	return nil
}

func (pt *PolicyTemplate) validatePolicyName() func(interface{}) error {
	return func(inputName interface{}) error {
		if isValid := regexp.MustCompile(`(^[a-z][a-z0-9_]*$)`).MatchString(inputName.(string)); !isValid {
			return fmt.Errorf("\nPolicy name must: \n-start with lowercase letter \n-only contain lowercase letters, numbers and underscores")
		}

		// avoid overwriting existing policy
		path := filepath.Join(pt.Dir, pt.Tool, inputName.(string), pt.CheckType)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			return fmt.Errorf("custom policy '%v' with check type '%v' already exists", inputName, pt.CheckType)
		}
		return nil
	}
}

func (pt *PolicyTemplate) createPoliciesDirectoryPrompt(dir, message string) error {
	create := false
	err := survey.AskOne(&survey.Confirm{
		Message: message,
	},
		&create)

	if err != nil {
		return err
	}
	if create {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			// if directory already exists, prompt user to input this path
			// to confirm this is the intended target directory
			return fmt.Errorf("\033[34m %s \033[0m already exists. Input this path to confirm this is the target directory", dir)
		}
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		} else {
			log.Infof("created: %s", dir)
		}
		pt.Dir = dir
	} else {
		return fmt.Errorf("please provide alternative path to a 'policies' directory")
	}
	return nil
}

func (pt *PolicyTemplate) validatePolicyDirectory() func(interface{}) error {
	return func(inputDir interface{}) error {
		dir := inputDir.(string)
		pt.Dir = dir

		switch {
		case isPoliciesPath(dir):
			// path points to a policies dir
			// check dir exists
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				// if dir doesn't exist offer to create it
				err := pt.createPoliciesDirectoryPrompt(dir,
					dir+" is not an existing policies directory. Create this directory now?")
				if err != nil {
					return err
				}
			}
		case dir == "." || dir == "./":
			// check current dir is named policies
			workingDir, _ := os.Getwd()
			if !isPoliciesPath(workingDir) {
				// offer to create `policies` dir in current dir
				err := pt.createPoliciesDirectoryPrompt("./policies",
					"current directory is not named policies. Create policies directory in current directory?")
				if err != nil {
					return err
				}
			}
		default:
			// path does not point to a policies dir
			// offer to create `policies dir in provided path
			err := pt.createPoliciesDirectoryPrompt(filepath.Join(dir, "policies"),
				dir+" path does not point to a policies directory. Create policies directory here?")
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func isPoliciesPath(path string) bool {
	if path == "policies" {
		return true
	}
	split := strings.LastIndex(path, "/")
	if split == -1 {
		return false
	}
	last := path[split:]
	return last == "/policies"
}

func (pt *PolicyTemplate) CreateCustomPolicyTemplate() error {
	if err := pt.CreateDirectoryStructure(); err != nil {
		return err
	}
	if err := pt.GenerateMetadataYaml(); err != nil {
		return err
	}
	if err := pt.GeneratePolicyTemplate(); err != nil {
		return err
	}
	return nil
}

func (pt *PolicyTemplate) CreateDirectoryStructure() error {
	// full directory path
	pt.Dir += "/" + pt.Tool + "/" + pt.Name + "/" + pt.CheckType + "/tests"
	if err := os.MkdirAll(pt.Dir, os.ModePerm); err != nil {
		return err
	} else {
		fmt.Println("created: ", pt.Dir)
	}
	return nil
}

func (pt *PolicyTemplate) GenerateMetadataYaml() error {
	// metadata.yaml (in PolicyName dir)
	metadataPath := strings.Split(pt.Dir, pt.CheckType)[0] + "/metadata.yaml"
	// shouldn't overwrite an existing metadata.yaml file
	if _, err := os.Stat(metadataPath); !os.IsNotExist(err) {
		return nil
	}

	type Metadata struct {
		Category    string    `yaml:"category"`
		CheckTool   string    `yaml:"checkTool"`
		CheckType   string    `yaml:"checkType"`
		Description yaml.Node `yaml:"description"`
		Provider    string    `yaml:"provider"`
		Severity    string    `yaml:"severity"`
		Title       yaml.Node `yaml:"title"`
	}

	// Ensure CheckType is capitalised for metadata
	caser := cases.Title(language.English)
	metadata := Metadata{
		Category:    pt.Category,
		CheckTool:   pt.Tool,
		CheckType:   caser.String(pt.CheckType),
		Description: doubleQuote(pt.Desc),
		Provider:    pt.Provider,
		Severity:    pt.Severity,
		Title:       doubleQuote(pt.Title),
	}

	data, err := yaml.Marshal(&metadata)

	if err != nil {
		return err
	}

	err2 := os.WriteFile(metadataPath, data, os.ModePerm)

	if err2 != nil {
		return err2
	}
	return nil
}

func doubleQuote(val string) yaml.Node {
	node := yaml.Node{
		Value: val,
		Kind:  yaml.ScalarNode,
		Style: yaml.DoubleQuotedStyle,
	}
	return node
}

func (pt *PolicyTemplate) GeneratePolicyTemplate() error {
	regoPath := strings.Split(pt.Dir, "tests")[0] + "/policy.rego"
	regoTemplate :=
		"package policies." + pt.Name +
			"\n\ninput_type := \"" + policy.InputTypeForTarget[policy.Target(pt.CheckType)] + "\""

	if pt.RsrcType != "" {
		if pt.RsrcType == "multiple" {
			pt.RsrcType = strings.ToUpper(pt.RsrcType)
		}
		regoTemplate += "\n\nresource_type := \"" + pt.RsrcType + "\""
	}

	regoTemplate +=
		"\n\ndefault allow = false" +
			"\n\n# Add Rego Policy # \n"

	err := os.WriteFile(regoPath, []byte(regoTemplate), os.ModePerm)

	if err != nil {
		return err
	}
	return nil
}
