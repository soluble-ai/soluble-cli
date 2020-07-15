package elevate

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/soluble-ai/go-jnode"
	"github.com/soluble-ai/soluble-cli/pkg/model"
	"github.com/soluble-ai/soluble-cli/pkg/options"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func getKubernetesUser(name string) (string, error) {
	apiConfig, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return "", err
	}
	ctx := apiConfig.Contexts[apiConfig.CurrentContext]
	authInfo := apiConfig.AuthInfos[ctx.AuthInfo]
	if authInfo.Username != "" {
		return authInfo.Username, nil
	}
	if len(authInfo.ClientCertificateData) == 0 {
		return "", fmt.Errorf("only client certificate identities are currently supported")
	}
	block, _ := pem.Decode(authInfo.ClientCertificateData)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	return cert.Subject.CommonName, nil
}

func whoAmICommand() *cobra.Command {
	opts := &options.PrintOpts{}
	c := &cobra.Command{
		Use:   "who-am-i",
		Short: "Display the name of the current kuberentes user",
		RunE: func(c *cobra.Command, args []string) error {
			user, err := getKubernetesUser("")
			if err != nil {
				return err
			}
			opts.PrintResult(jnode.NewObjectNode().Put("user", user))
			return nil
		},
	}
	opts.Register(c)
	return c
}

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:   "elevate",
		Short: "",
	}
	c.AddCommand(whoAmICommand())
	return c
}

func init() {
	model.AddContextValueSupplier("kubernetes_user", getKubernetesUser)
}
