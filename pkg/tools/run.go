package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/exit"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/print"
	"github.com/soluble-ai/soluble-cli/pkg/util"
)

const AssessmentDirectoryValue = "ASSESSMENT_DIRECTORY"

func RunSingleAssessment(tool Single) (*Result, error) {
	if err := tool.Validate(); err != nil {
		return nil, err
	}
	r, err := tool.Run()
	if err != nil {
		return nil, err
	}
	r.Tool = tool
	if err := processResult(r); err != nil {
		return nil, err
	}
	return r, nil
}

func RunConsoliatedAssessments(tool Consolidated) (Results, error) {
	if err := tool.Validate(); err != nil {
		return nil, err
	}
	results, err := tool.RunAll()
	if err != nil {
		return nil, err
	}
	for _, result := range results {
		rerr := processResult(result)
		if rerr != nil {
			// processResult only fails if the upload failed, and if that
			// fais then it's likely that nothing is going to work
			return nil, rerr
		}
	}
	return results, err
}

func processResult(result *Result) error {
	o := result.Tool.GetAssessmentOptions()
	result.AddValues(result.Tool.GetToolOptions().GetStandardXCPValues())
	if result.Directory != "" {
		result.UpdateFileFingerprints()
		if result.Values[AssessmentDirectoryValue] == "" {
			if o.RepoRoot != "" {
				reldir, err := filepath.Rel(o.RepoRoot, result.Directory)
				if err == nil && !strings.HasPrefix(reldir, "..") {
					result.AddValue(AssessmentDirectoryValue, reldir)
				}
			}
		}
	}
	if o.PrintFingerprints || o.SaveFingerprints != "" {
		d, err := json.Marshal(result.FileFingerprints)
		util.Must(err)
		n, err := jnode.FromJSON(d)
		util.Must(err)
		if o.PrintFingerprints {
			p := &print.JSONPrinter{}
			p.PrintResult(os.Stderr, n)
		}
		if o.SaveFingerprints != "" {
			p := &print.JSONPrinter{}
			f, err := os.Create(o.SaveFingerprints)
			if err != nil {
				log.Warnf("Could not save fingerprints: {warning:%s}", err)
			} else {
				p.PrintResult(f, n)
				_ = f.Close()
			}
		}
	}
	if o.PrintResultOpt {
		p := &print.JSONPrinter{}
		p.PrintResult(os.Stderr, result.Data)
	}
	if o.SaveResult != "" {
		f, err := os.Create(o.SaveResult)
		if err != nil {
			return err
		}
		p := &print.JSONPrinter{}
		p.PrintResult(f, result.Data)
		_ = f.Close()
	}
	if o.PrintResultValues {
		writeResultValues(os.Stderr, result)
	}
	if o.SaveResultValues != "" {
		f, err := os.Create(o.SaveResultValues)
		if err != nil {
			return err
		}
		writeResultValues(f, result)
		_ = f.Close()
	}
	if o.UploadEnabled {
		if err := result.Upload(o.GetAPIClient(), o.GetOrganization(), o.Tool.Name()); err != nil {
			return err
		}
		if result.Assessment != nil && len(o.parsedFailThresholds) > 0 {
			result.Assessment.EvaluateFailures(o.parsedFailThresholds)
			if result.Assessment.Failed {
				exit.Code = 2
				a := result.Assessment
				exit.AddFunc(func() {
					log.Errorf("{warning:%s} has {danger:%d %s findings}",
						a.Title, a.FailedCount, a.FailedSeverity)
				})
			}
		}
	}
	return nil
}

func writeResultValues(w io.Writer, result *Result) {
	for k, v := range result.Values {
		fmt.Fprintf(w, "%s=%s\n", k, v)
	}
}
