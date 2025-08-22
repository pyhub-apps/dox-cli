package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pyhub/pyhub-docs/internal/i18n"
	"github.com/spf13/cobra"
)

var (
	// Configuration flags
	cfgFile  string
	verbose  bool
	quiet    bool
	langFlag string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pyhub-docs",
	Short: "Document automation and AI-powered content generation CLI",
	Long: `pyhub-docs is a powerful CLI tool for document automation.

It provides:
  • Bulk text replacement across Word/PowerPoint documents
  • Markdown to Office document conversion
  • Template-based document generation
  • AI-powered content generation (coming soon)

Examples:
  # Replace text in documents
  pyhub-docs replace --rules rules.yml --path ./docs

  # Create document from markdown
  pyhub-docs create --from report.md --template template.docx --output final.docx

  # Generate content with AI (coming soon)
  pyhub-docs generate --type blog --prompt "Docker best practices" --output blog.md`,
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
	cobra.OnInitialize(initConfig, initI18n)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pyhub/config.yml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-error output")
	rootCmd.PersistentFlags().StringVar(&langFlag, "lang", "", i18n.T(i18n.MsgFlagLang))

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

// initI18n initializes the internationalization system
func initI18n() {
	// Try to load from external files first (for development)
	execPath, _ := os.Executable()
	localesDir := filepath.Join(filepath.Dir(execPath), "locales")
	
	if _, err := os.Stat(localesDir); err == nil {
		// External locale files exist
		i18n.InitWithFiles(localesDir, langFlag)
	} else {
		// Use embedded locale files
		i18n.Init(langFlag)
	}
	
	// Update command descriptions after i18n is initialized
	rootCmd.Short = i18n.T(i18n.MsgCmdRootShort)
	rootCmd.Long = i18n.T(i18n.MsgCmdRootLong)
}