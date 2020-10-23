package imagescancmd

import (
	"bytes"

	"github.com/soluble-ai/soluble-cli/pkg/client"
	"github.com/soluble-ai/soluble-cli/pkg/imagescan"
	"github.com/soluble-ai/soluble-cli/pkg/log"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	config := imagescan.Config{}
	opts := options.PrintClientOpts{}
	c := &cobra.Command{
		Use:   "image-scan",
		Short: "Run an image scanner",
		RunE: func(cmd *cobra.Command, args []string) error {
			config.APIClient = opts.GetAPIClient()
			config.Organizaton = opts.GetOrganization()
			scanner := imagescan.New(config)
			result, err := scanner.Run()
			if err != nil {
				return err
			}

			if config.ReportEnabled {
				rr := bytes.NewReader([]byte(result.N.String()))
				log.Infof("Uploading image scan results")
				values := map[string]string{
					"image":       config.Image,
					"scannerType": scanner.Name(),
				}
				err = config.APIClient.XCPPost(config.Organizaton, scanner.Name(), nil, values,
					client.XCPWithCIEnv, client.XCPWithReader("results_json", "results.json", rr))
				if err != nil {
					return err
				}
			}
			opts.Path = result.PrintPath
			opts.Columns = result.PrintColumns
			log.Infof("Vulnerability Report....")
			opts.PrintResult(result.N)
			return nil
		},
	}
	opts.Register(c)
	flags := c.Flags()
	flags.StringVarP(&config.Image, "image-name", "i", "", "Name of the image to scan")
	flags.BoolVarP(&config.ReportEnabled, "report", "r", false, "Upload scan results to soluble")
	_ = c.MarkFlagRequired("image")
	return c
}
