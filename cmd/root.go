package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Configuration flags
	cfgFile string
	verbose bool
	quiet   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pyhub-documents-cli",
	Short: "Document automation and AI-powered content generation CLI",
	Long: `pyhub-documents-cli is a powerful CLI tool for document automation.

It provides:
  • Bulk text replacement across Word/PowerPoint documents
  • Markdown to Office document conversion
  • Template-based document generation
  • AI-powered content generation (coming soon)

Examples:
  # Replace text in documents
  pyhub-documents-cli replace --rules rules.yml --path ./docs

  # Create document from markdown
  pyhub-documents-cli create --from report.md --template template.docx --output final.docx

  # Generate content with AI (coming soon)
  pyhub-documents-cli generate --type blog --prompt "Docker best practices" --output blog.md`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pyhub/config.yml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-error output")

	// Version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(`{{with .Name}}{{printf "%%s version information:\n" .}}{{end}}
  Version:    %s
  Commit:     %s
  Built:      %s
`, Version, Commit, BuildDate))
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	// TODO: Implement configuration loading
	// This will be implemented when we add the config package
}