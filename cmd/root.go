package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pyhub/pyhub-docs/internal/config"
	"github.com/pyhub/pyhub-docs/internal/i18n"
	"github.com/spf13/cobra"
)

var (
	// Configuration flags
	cfgFile  string
	verbose  bool
	quiet    bool
	langFlag string
	
	// Global configuration instance
	appConfig *config.Config
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
	// 설정 파일 경로 결정
	configPath := cfgFile
	if configPath == "" {
		configPath = config.GetConfigPath()
	}
	
	// 설정 파일 로드
	cfg, err := config.Load(configPath)
	if err != nil {
		// 에러가 있어도 기본 설정으로 계속 진행
		cfg = config.DefaultConfig()
	}
	
	// 전역 설정 인스턴스 저장
	appConfig = cfg
	
	// CLI 플래그가 설정 파일보다 우선순위가 높음
	// verbose 플래그가 명시적으로 설정되었는지 확인
	if rootCmd.PersistentFlags().Changed("verbose") {
		// CLI 플래그가 우선
	} else {
		// 설정 파일의 값 사용
		verbose = cfg.Global.Verbose
	}
	
	if rootCmd.PersistentFlags().Changed("quiet") {
		// CLI 플래그가 우선
	} else {
		quiet = cfg.Global.Quiet
	}
	
	if rootCmd.PersistentFlags().Changed("lang") {
		// CLI 플래그가 우선
	} else if cfg.Global.Lang != "" {
		langFlag = cfg.Global.Lang
	}
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