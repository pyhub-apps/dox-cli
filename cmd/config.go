package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pyhub/pyhub-docs/internal/config"
	"github.com/pyhub/pyhub-docs/internal/secrets"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	configPath   string
	configInit   bool
	configList   bool
	configSet    string
	configGet    string
	useKeyring   bool
	noKeyring    bool
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
  dox config --init

  # 설정 확인
  dox config --list

  # 특정 설정 확인
  dox config --get openai.api_key

  # 설정 변경
  dox config --set "openai.api_key=sk-..."
  dox config --set "global.lang=ko"`,
	RunE: runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().BoolVar(&configInit, "init", false, "기본 설정 파일 생성")
	configCmd.Flags().BoolVarP(&configList, "list", "l", false, "모든 설정 표시")
	configCmd.Flags().StringVar(&configSet, "set", "", "설정 값 변경 (key=value)")
	configCmd.Flags().StringVar(&configGet, "get", "", "특정 설정 값 확인")
	configCmd.Flags().StringVar(&configPath, "path", "", "설정 파일 경로 (기본: ~/.pyhub/config.yml)")
	configCmd.Flags().BoolVar(&useKeyring, "use-keyring", false, "시스템 키체인에 API 키 저장 (안전한 저장)")
	configCmd.Flags().BoolVar(&noKeyring, "no-keyring", false, "시스템 키체인 사용하지 않음")
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
	case "openai.api_key", "claude.api_key":
		provider := "openai"
		if key == "claude.api_key" {
			provider = "claude"
		}
		
		// 먼저 키체인에서 확인
		storage := secrets.NewSecureStorage()
		if apiKey, err := storage.RetrieveAPIKey(provider); err == nil && apiKey != "" {
			masked := secrets.MaskAPIKey(apiKey)
			fmt.Printf("%s: %s (stored in keychain)\n", key, masked)
		} else {
			// 설정 파일에서 확인
			var apiKey string
			if provider == "openai" {
				apiKey = cfg.OpenAI.APIKey
			} else {
				apiKey = cfg.Claude.APIKey
			}
			
			if apiKey != "" && apiKey != "<stored-in-keychain>" {
				masked := secrets.MaskAPIKey(apiKey)
				fmt.Printf("%s: %s\n", key, masked)
			} else if apiKey == "<stored-in-keychain>" {
				fmt.Printf("%s: (error reading from keychain)\n", key)
			} else {
				fmt.Printf("%s: (not set)\n", key)
			}
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
		// API 키 검증
		if err := secrets.ValidateAPIKey("openai", value); err != nil {
			return fmt.Errorf("API 키 검증 실패: %w", err)
		}
		
		// 키체인 사용 여부 확인
		if useKeyring && !noKeyring {
			storage := secrets.NewSecureStorage()
			if storage.IsSupported() {
				if err := storage.StoreAPIKey("openai", value); err != nil {
					secrets.Warnf("키체인 저장 실패 (설정 파일에 저장됩니다): %v", err)
					cfg.OpenAI.APIKey = value
				} else {
					// 키체인에 저장 성공 - 설정 파일에는 플레이스홀더 저장
					cfg.OpenAI.APIKey = "<stored-in-keychain>"
					fmt.Println("API 키가 시스템 키체인에 안전하게 저장되었습니다")
				}
			} else {
				secrets.Warnf("이 시스템에서는 키체인이 지원되지 않습니다")
				cfg.OpenAI.APIKey = value
			}
		} else {
			cfg.OpenAI.APIKey = value
		}
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
	case "claude.api_key":
		// API 키 검증
		if err := secrets.ValidateAPIKey("claude", value); err != nil {
			return fmt.Errorf("API 키 검증 실패: %w", err)
		}
		
		// 키체인 사용 여부 확인
		if useKeyring && !noKeyring {
			storage := secrets.NewSecureStorage()
			if storage.IsSupported() {
				if err := storage.StoreAPIKey("claude", value); err != nil {
					secrets.Warnf("키체인 저장 실패 (설정 파일에 저장됩니다): %v", err)
					cfg.Claude.APIKey = value
				} else {
					// 키체인에 저장 성공 - 설정 파일에는 플레이스홀더 저장
					cfg.Claude.APIKey = "<stored-in-keychain>"
					fmt.Println("API 키가 시스템 키체인에 안전하게 저장되었습니다")
				}
			} else {
				secrets.Warnf("이 시스템에서는 키체인이 지원되지 않습니다")
				cfg.Claude.APIKey = value
			}
		} else {
			cfg.Claude.APIKey = value
		}
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