package policyimporter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"gopkg.in/yaml.v3"
)

type Converter struct {
	OpalRegoPath string
	DestPath     string
}

type Policy struct {
	Name      string
	CheckType string
	Tool      string
	Dir       string
	Desc      string
	Title     string
	Severity  string
	RsrcType  string
	Provider  string
}

type Metadata struct {
	Category    string                 `yaml:"category"`
	CheckTool   string                 `yaml:"checkTool"`
	CheckType   []string               `yaml:"checkType"`
	Description yaml.Node              `yaml:"description"`
	Provider    string                 `yaml:"provider"`
	Severity    string                 `yaml:"severity"`
	Title       yaml.Node              `yaml:"title"`
	ID          yaml.Node              `yaml:"id"`
	LwIds       map[string]interface{} `yaml:"lwids"`
}

type Custom struct {
	Controls map[string]interface{}
	Severity string
}
type Metadoc struct {
	Custom      Custom
	Description string
	Title       string
	ID          string
}

var ManualCheck []string

func validatePath(expectedPath string) func(interface{}) error {
	return func(inputDir interface{}) error {
		dir := inputDir.(string)
		sub := len(dir) - len(expectedPath)
		if len(dir) < len(expectedPath) || dir[sub:] != expectedPath {
			return fmt.Errorf("invalid directory path: %v", dir)
		}
		return nil
	}
}

func (c *Converter) PromptInput() error {
	var qs = []*survey.Question{
		{
			Name: "opalRegoPath",
			Prompt: &survey.Input{
				Message: "Opal policies directory path",
				Help:    "provide path to opal built-in policies. EG: 'rego/policies'",
			},
			Validate: validatePath("policies"),
		},
		{
			Name: "destPath",
			Prompt: &survey.Input{
				Message: "Converted policies destination path",
				Help:    "must point to a 'policies/opal' directory",

			},
			Validate: validatePath("policies/opal"),
		},
	}
	if err := survey.Ask(qs, c); err == nil {
		return err
	}
	return nil
}

func Find(root, ext string) []string {
	var regoFilePaths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(d.Name()) == ext {
			regoFilePaths = append(regoFilePaths, path)
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return regoFilePaths
}
func getName(fileName, rsrcType string) string {
	if rsrcType != "" {
		return rsrcType + "_" + fileName[:len(fileName)-5] // 5 = len(".rego")
	} else {
		return fileName[:len(fileName)-5]
	}
}

func setupBaseDirStructure(path string) error {
	// path = <...>/policies/opal/
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (p *Policy) createPolicyDirStructure(destPath string) error {
	p.Dir = destPath + "/" + p.Name + "/" + p.CheckType
	// Ensure policy dir doesn't already exist
	if _, err := os.Stat(p.Dir); !os.IsNotExist(err) {
		return fmt.Errorf("policy '%v' with check type '%v' already exists: %v", p.Name, p.CheckType, p.Dir)
	}
	if err := os.MkdirAll(p.Dir, os.ModePerm); err != nil {
		return err
	} else {
		fmt.Println("created: ", p.Dir)
	}
	return nil
}
func (p *Policy) copyRegoFile(regoPath string) error {
	destination := p.Dir + "/policy.rego"
	input, err := os.ReadFile(regoPath)
	if err != nil {
		return err
	}

	err = os.WriteFile(destination, input, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (p *Policy) getID() string {
	idName := strings.ReplaceAll(p.Name, "_", "-")
	return "lacework-opl-" + idName
}
func doubleQuote(val string) yaml.Node {
	node := yaml.Node{
		Value: val,
		Kind:  yaml.ScalarNode,
		Style: yaml.DoubleQuotedStyle,
	}
	return node
}

func (md *Metadoc) getMetadoc(regoPath string) error {
	readFile, err := os.Open(regoPath)
	if err != nil {
		return err
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	metadoc := "__rego__metadoc__ := {"
	open := 0
	closed := 0

	for fileScanner.Scan() {
		line := fileScanner.Text()
		if open > 0 {
			metadoc += line
		}
		if strings.Contains(line, metadoc) || strings.ContainsAny(line, "{}") {
			open += strings.Count(line, "{")
			closed += strings.Count(line, "}")
			if closed == open {
				break
			}
		}
	}
	if err = readFile.Close(); err != nil {
		return err
	}

	metaJSON := strings.Split(metadoc, "__rego__metadoc__ :=")[1]
	if err = json.Unmarshal([]byte(metaJSON), &md); err != nil {
		return err
	}
	return nil
}

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func updateMetadata(name string, existing, new Metadata) Metadata {
	// if metadata has conflicting values - add to manual check list
	// else update check type and return NEW metadata
	// merge LWIDs as many policies with same ID dont have all the lwids

	new.CheckType = append(new.CheckType, existing.CheckType...)
	if existing.Category != new.Category ||
		existing.Description.Value != new.Description.Value ||
		existing.Severity != new.Severity ||
		existing.Title.Value != new.Title.Value {
		ManualCheck = append(ManualCheck, name)
	} else if !reflect.DeepEqual(existing.LwIds, new.LwIds) {
		new.LwIds = mergeMaps(existing.LwIds, new.LwIds)
	}
	return new
}

func (p *Policy) generateMetadataYaml(regoPath string) error {
	// don't blindly overwrite an existing metadata.yaml file
	existingMd := Metadata{}
	metadataPath := strings.Split(p.Dir, p.CheckType)[0] + "metadata.yaml"
	if _, err := os.Stat(metadataPath); !os.IsNotExist(err) {
		md, err := os.ReadFile(metadataPath)
		if err != nil {
			return err
		}

		if err = yaml.Unmarshal(md, &existingMd); err != nil {
			return err
		}
	}

	metadoc := Metadoc{}
	if err := metadoc.getMetadoc(regoPath); err != nil {
		return err
	}

	metadata := Metadata{
		Category:    p.RsrcType,
		CheckTool:   p.Tool,
		CheckType:   []string{p.CheckType},
		Description: doubleQuote(metadoc.Description),
		Provider:    p.Provider,
		Severity:    metadoc.Custom.Severity,
		Title:       doubleQuote(metadoc.Title),
		LwIds:       metadoc.Custom.Controls,
		ID:          doubleQuote(p.getID()),
	}

	// CheckTool isn't set for an empty Metadata struct
	if existingMd.CheckTool == "opal" {
		metadata = updateMetadata(p.Name, existingMd, metadata)
	}

	data, err := yaml.Marshal(&metadata)
	if err != nil {
		return err
	}

	if err = os.WriteFile(metadataPath, data, 0644); err != nil {
		return err
	}
	return nil
}

func (p *Policy) convertPolicy(regoPath, destPath string) error {
	if err := p.createPolicyDirStructure(destPath); err != nil {
		return err
	}

	err := p.generateMetadataYaml(regoPath)
	if err != nil {
		return err
	}

	if err = p.copyRegoFile(regoPath); err != nil {
		return err
	}
	return nil
}

func (p *Policy) Convert(regoFile, destPath string) error {
	relPath := strings.Split(regoFile, "/policies/")[1]
	pathData := strings.Split(relPath, "/")
	checkTypeMap := map[string]string{
		"tf":  "terraform",
		"k8s": "kubernetes",
		"cfn": "cloudformation",
		"arm": "arm",
	}
	providerMap := map[string]string{
		"aws":     "aws",
		"google":  "gcp",
		"azurerm": "azure",
	}

	checkType := pathData[0]
	p.CheckType = checkTypeMap[checkType]

	switch {
	case checkType == "tf":
		// tf/<provider>/<rsrcType>/<name>.rego
		p.Provider = providerMap[pathData[1]]
		p.RsrcType = pathData[2]
		p.Name = getName(pathData[3], p.RsrcType)

	case checkType == "k8s":
		// k8s/<name>.rego
		p.RsrcType = checkType
		p.Name = getName(pathData[1], p.RsrcType)

	case checkType == "cfn", checkType == "arm":
		// cfn/<rsrcType>/<name>.rego
		// arm/<rsrcType>/<name>.rego
		p.RsrcType = pathData[1]
		p.Name = getName(pathData[2], p.RsrcType)
	}
	if err := p.convertPolicy(regoFile, destPath); err != nil {
		return err
	}
	return nil
}

func (c *Converter) ConvertOpalBuiltIns() error {
	// lw dir structure: policies/opal/<policy_name>/<checkType>
	if err := setupBaseDirStructure(c.DestPath); err != nil {
		return err
	}

	regoFiles := Find(c.OpalRegoPath, ".rego")

	for i := len(regoFiles) - 1; i >= 0; i-- {
		fmt.Println("converting: ", regoFiles[i])
		p := Policy{Tool: "opal"}
		if err := p.Convert(regoFiles[i], c.DestPath); err != nil {
			return err
		}
	}

	if ManualCheck != nil {
		fmt.Printf(" Conversion not complete for the following %v policies: ", len(ManualCheck))
		fmt.Printf("Manually validate conversion of __rego__metadoc__'s for: \n %v ", ManualCheck)
	} else {
		fmt.Println("All policies converted with no issues")
	}
	return nil
}
