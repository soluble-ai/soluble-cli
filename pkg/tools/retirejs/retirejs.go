package retirejs

import (
	"os"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/tools"
	"github.com/spf13/cobra"
)

type Tool struct {
	tools.DirectoryBasedToolOpts
}

var _ tools.Interface = &Tool{}

func (t *Tool) Name() string {
	return "retirejs"
}

func (t *Tool) CommandTemplate() *cobra.Command {
	return &cobra.Command{
		Use:   "retirejs",
		Short: "Scan with retirejs to identify vulnerable JavaScript libraries",
	}
}

func (t *Tool) Run() (*tools.Result, error) {
	dat, err := t.RunDocker(&tools.DockerTool{
		Name:      "retirejs",
		Image:     "gcr.io/soluble-repo/soluble-retirejs:latest",
		Directory: t.GetDirectory(),
	})
	if err != nil {
		if dat != nil {
			_, _ = os.Stderr.Write(dat)
		}
		return nil, err
	}
	n, err := jnode.FromJSON(dat)
	if err != nil {
		_, _ = os.Stderr.Write(dat)
		return nil, err
	}

	// Example raw output:
	/*
		{
		  "version": "2.2.4",
		  "start": "2021-01-14T17:41:38.426Z",
		  "data": [
		    {
		      "file": "node_modules/lodash/package.json",
		      "results": [
		        {
		          "component": "lodash",
		          "version": "4.17.15",
		          "vulnerabilities": [
		            {
		              "info": [
		                "https://snyk.io/vuln/SNYK-JS-LODASH-590103"
		              ],
		              "below": "4.17.20",
		              "severity": "low",
		              "identifiers": {
				"CVE": [
				  "CVE-2020-11111"
				],
		                "summary": "Prototype pollution attack"
		              }
		            }
		          ]
		        }
		      ]
		    }
		  ],
		  "messages": [],
		  "errors": [
		    "Could not follow symlink: /src/resources",
		    "Could not follow symlink: /src/START_HERE/features-gherkin",
		    "Could not follow symlink: /src/START_HERE/features-code.js",
		    "Could not follow symlink: /src/START_HERE/tests"
		  ],
		  "time": 20.42
		}
	*/
	retirejsVersion := n.Path("version").AsText()
	result := &tools.Result{
		Directory: t.Directory,
		Values: map[string]string{
			"RETIREJS_VERSION": retirejsVersion,
		},
		Data:      n.Path("data"), // TODO: include errors?
		PrintData: createPrintData(n.Path("data")),
		PrintPath: []string{},
		PrintColumns: []string{
			"Component", "File", "InstalledVersion", "FixedVersion", "Severity", "CVE", "Summary",
		},
	}
	return result, nil
}

func createPrintData(n *jnode.Node) *jnode.Node {
	result := jnode.NewArrayNode()
	for _, e := range n.Elements() {
		file := e.Path("file").AsText()
		for _, r := range e.Path("results").Elements() {
			component := r.Path("component").AsText()
			version := r.Path("version").AsText()
			for _, v := range r.Path("vulnerabilities").Elements() {
				vc := result.AppendObject().Put("File", file)
				vc.Put("Component", component)
				vc.Put("InstalledVersion", version)
				vc.Put("Severity", v.Path("severity").AsText())
				vc.Put("FixedVersion", v.Path("below").AsText())
				vc.Put("CVE", v.Path("identifiers").Path("CVE").AsText()) // BUG: returns jnode slice - want JSON slice.
				vc.Put("Summary", v.Path("identifiers").Path("summary").AsText())
			}
		}
	}
	return result
}
