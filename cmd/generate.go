/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all" // register cookie store finders!
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen"},
	Short:   "Generate a kubeconfig file from available Palette clusters.",
	Long: `Generate (pier gen) will output a kubeconfig file based on the
available Palette clusters. These will be named according to the project
and cluster name in Palette.`,
	Run: func(cmd *cobra.Command, args []string) {
		host := os.Getenv("PALETTE_HOST")
		apiKey := os.Getenv("PALETTE_API_KEY")
		token := os.Getenv("PALETTE_TOKEN")

		if host == "" {
			host = "api.spectrocloud.com"
		}

		// If a token is not provided, check browsers for an Authorization cookie.
		// Will require elevated permissions to access browser cookies.
		if apiKey == "" && token == "" {
			cookies := kooky.ReadCookies(kooky.DomainHasSuffix(fmt.Sprintf(".%s", host)), kooky.Name(`Authorization`))
			if len(cookies) == 0 {
				fmt.Println("No Authorization cookie found.")
				os.Exit(1)
			}
			token = cookies[0].Value
		}

		// Create kubeconfig file
		kubeconfig := clientcmdapi.NewConfig()

		// Initialize a Palette client
		c := client.New(
			client.WithPaletteURI(host),
			client.WithAPIKey(apiKey),
			client.WithJWT(token),
			client.WithScopeTenant(),
		)

		// Create cache directory if it doesn't exist
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "/"
		}
		cacheDir := homeDir + "/.cache/pier"
		os.MkdirAll(cacheDir, os.ModePerm)

		// List available projects
		projects, err := c.GetProjects()
		if err != nil {
			panic(err)
		}
		if len(projects.Items) == 0 {
			fmt.Println("\nNo projects found.")
			return
		}

		// List clusters in each project
		for _, project := range projects.Items {

			// Set project scope in client
			client.WithScopeProject(project.Metadata.UID)(c)

			// Search for clusters
			clusters, err := c.SearchClusterSummaries(&models.V1SearchFilterSpec{}, []*models.V1SearchFilterSortSpec{})
			if err != nil {
				panic(err)
			}

			// Cache cluster summaries per project
			projectJSON, _ := json.Marshal(clusters)
			err = os.WriteFile(fmt.Sprintf("%s/%s.json", cacheDir, project.Metadata.UID), projectJSON, 0644)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			}

			// Display the results

			// fmt.Printf("\nFound %d cluster(s):\n", len(clusters))
			for _, cluster := range clusters {
				// TODO: add debug log statement for each cluster
				// fmt.Printf("%s_%s\n", cluster.SpecSummary.ProjectMeta.Name, cluster.Metadata.Name)

				// Fetch kubeconfig from palette
				clusterKubeconfigData, err := c.GetClusterKubeConfig(cluster.Metadata.UID)
				if err != nil {
					panic(err)
				}
				if clusterKubeconfigData == "" {
					continue
				}

				// Parse into kubeconfig struct
				clusterKubeconfig, err := clientcmd.Load([]byte(clusterKubeconfigData))
				if err != nil {
					panic(err)
				}

				// Generate name/prefix
				generatedName := fmt.Sprintf("%s_%s", cluster.SpecSummary.ProjectMeta.Name, cluster.Metadata.Name)

				// Rename cluster and user objects to include project and cluster name
				kubeconfig.Clusters[generatedName] = clusterKubeconfig.Clusters["kubernetes"].DeepCopy()
				kubeconfig.AuthInfos[generatedName] = clusterKubeconfig.AuthInfos["kubernetes-admin"].DeepCopy()

				// Create context using generated names
				context := clusterKubeconfig.Contexts["kubernetes-admin@kubernetes"].DeepCopy()
				// kubeconfig.Contexts[generatedName] = clusterKubeconfig.Contexts["kubernetes-admin@kubernetes"].DeepCopy()
				context.Cluster = generatedName
				context.AuthInfo = generatedName

				kubeconfig.Contexts[generatedName] = context
			}
		}

		// Export kubeconfig to stdout
		kubeconfigBytes, err := clientcmd.Write(*kubeconfig)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(kubeconfigBytes))
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
