package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pyhub/pyhub-docs/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	configPath string
	configInit bool
	configList bool
	configSet  string
	configGet  string
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "설정 파일 관리",
	Long: `설정 파일을 관리합니다. 기본 설정 파일 위치는 ~/.pyhub/config.yml 입니다.

설정 파일을 통해 다음을 구성할 수 있습니다:
  • OpenAI API 키 및 기본 모델
  • 각 명령어의 기본 옵션
  • 전역 설정 (verbose, quiet, language)

예제:
  # 기본 설정 파일 생성
  pyhub-docs config --init

  # 설정 확인
  pyhub-docs config --list

  # 특정 설정 확인
  pyhub-docs config --get openai.api_key

  # 설정 변경
  pyhub-docs config --set "openai.api_key=sk-..."
  pyhub-docs config --set "global.lang=ko"`,
	RunE: runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().BoolVar(&configInit, "init", false, "기본 설정 파일 생성")
	configCmd.Flags().BoolVarP(&configList, "list", "l", false, "모든 설정 표시")
	configCmd.Flags().StringVar(&configSet, "set", "", "설정 값 변경 (key=value)")
	configCmd.Flags().StringVar(&configGet, "get", "", "특정 설정 값 확인")
	configCmd.Flags().StringVar(&configPath, "path", "", "설정 파일 경로 (기본: ~/.pyhub/config.yml)")
}

func runConfig(cmd *cobra.Command, args []string) error {
	// 설정 파일 경로 결정
	if configPath == "" {
		configPath = config.GetConfigPath()
	}

	// --init: 기본 설정 파일 생성
	if configInit {
		return initConfigFile(configPath)
	}

	// --list: 모든 설정 표시
	if configList {
		return listConfig(configPath)
	}

	// --get: 특정 설정 값 확인
	if configGet != "" {
		return getConfig(configPath, configGet)
	}

	// --set: 설정 값 변경
	if configSet != "" {
		return setConfig(configPath, configSet)
	}

	// 플래그가 없으면 도움말 표시
	cmd.Help()
	return nil
}

func initConfigFile(path string) error {
	// 파일이 이미 존재하는지 확인
	if _, err := os.Stat(path); err == nil {
		if !force {
			return fmt.Errorf("설정 파일이 이미 존재합니다: %s (덮어쓰려면 --force 사용)", path)
		}
	}

	// 기본 설정 생성
	cfg := config.DefaultConfig()

	// 환경 변수에서 OpenAI API 키 가져오기
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		cfg.OpenAI.APIKey = apiKey
	}

	// 설정 파일 저장
	if err := cfg.Save(path); err != nil {
		return fmt.Errorf("설정 파일 생성 실패: %w", err)
	}

	fmt.Printf("설정 파일이 생성되었습니다: %s\n", path)
	return nil
}

func listConfig(path string) error {
	// 설정 파일 로드
	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("설정 파일 로드 실패: %w", err)
	}

	// YAML로 출력
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("설정 출력 실패: %w", err)
	}

	fmt.Println("현재 설정:")
	fmt.Println("---")
	fmt.Print(string(data))
	return nil
}

func getConfig(path string, key string) error {
	// 설정 파일 로드
	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("설정 파일 로드 실패: %w", err)
	}

	// 키에 따라 값 출력
	// 간단한 구현 - 실제로는 더 복잡한 키 탐색이 필요
	switch key {
	case "openai.api_key":
		if cfg.OpenAI.APIKey != "" {
			// API 키는 일부만 표시
			if len(cfg.OpenAI.APIKey) > 8 {
				fmt.Printf("%s: sk-...%s\n", key, cfg.OpenAI.APIKey[len(cfg.OpenAI.APIKey)-4:])
			} else {
				fmt.Printf("%s: %s\n", key, cfg.OpenAI.APIKey)
			}
		} else {
			fmt.Printf("%s: (not set)\n", key)
		}
	case "openai.model":
		fmt.Printf("%s: %s\n", key, cfg.OpenAI.Model)
	case "openai.max_tokens":
		fmt.Printf("%s: %d\n", key, cfg.OpenAI.MaxTokens)
	case "openai.temperature":
		fmt.Printf("%s: %.2f\n", key, cfg.OpenAI.Temperature)
	case "global.verbose":
		fmt.Printf("%s: %v\n", key, cfg.Global.Verbose)
	case "global.quiet":
		fmt.Printf("%s: %v\n", key, cfg.Global.Quiet)
	case "global.lang":
		fmt.Printf("%s: %s\n", key, cfg.Global.Lang)
	case "replace.backup":
		fmt.Printf("%s: %v\n", key, cfg.Replace.Backup)
	case "replace.recursive":
		fmt.Printf("%s: %v\n", key, cfg.Replace.Recursive)
	case "replace.concurrent":
		fmt.Printf("%s: %v\n", key, cfg.Replace.Concurrent)
	default:
		return fmt.Errorf("알 수 없는 설정 키: %s", key)
	}

	return nil
}

func setConfig(path string, keyValue string) error {
	// key=value 형식 파싱
	parts := strings.SplitN(keyValue, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("잘못된 형식: %s (key=value 형식이어야 함)", keyValue)
	}
	
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// 설정 파일 로드
	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("설정 파일 로드 실패: %w", err)
	}

	// 키에 따라 값 설정
	switch key {
	case "openai.api_key":
		cfg.OpenAI.APIKey = value
	case "openai.model":
		cfg.OpenAI.Model = value
	case "openai.max_tokens":
		var tokens int
		fmt.Sscanf(value, "%d", &tokens)
		cfg.OpenAI.MaxTokens = tokens
	case "openai.temperature":
		var temp float64
		fmt.Sscanf(value, "%f", &temp)
		cfg.OpenAI.Temperature = temp
	case "global.verbose":
		cfg.Global.Verbose = (value == "true")
	case "global.quiet":
		cfg.Global.Quiet = (value == "true")
	case "global.lang":
		cfg.Global.Lang = value
	case "replace.backup":
		cfg.Replace.Backup = (value == "true")
	case "replace.recursive":
		cfg.Replace.Recursive = (value == "true")
	case "replace.concurrent":
		cfg.Replace.Concurrent = (value == "true")
	default:
		return fmt.Errorf("알 수 없는 설정 키: %s", key)
	}

	// 설정 검증
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("설정 검증 실패: %w", err)
	}

	// 설정 파일 저장
	if err := cfg.Save(path); err != nil {
		return fmt.Errorf("설정 파일 저장 실패: %w", err)
	}

	fmt.Printf("설정이 업데이트되었습니다: %s = %s\n", key, value)
	return nil
}