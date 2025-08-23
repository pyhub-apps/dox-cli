# dox - Document Automation CLI 🚀

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/pyhub-kr/pyhub-documents-cli)](https://github.com/pyhub-kr/pyhub-documents-cli/releases)
[![HeadVer](https://img.shields.io/badge/versioning-HeadVer-blue)](https://github.com/line/headver)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Issues](https://img.shields.io/github/issues/pyhub-kr/pyhub-documents-cli)](https://github.com/pyhub-kr/pyhub-documents-cli/issues)

문서 자동화 및 AI 기반 콘텐츠 생성을 위한 강력한 CLI 도구입니다. 아름다운 프로그레스 바와 색상 출력으로 Word/PowerPoint 문서를 효율적으로 처리합니다.

한국어 | [English](README.en.md)

## 🎯 소개

`dox`는 반복적인 문서 작업을 자동화하고, 마크다운을 Office 문서로 변환하며, 템플릿 기반 문서 생성과 AI 콘텐츠 생성을 지원하는 Go 언어 기반 CLI 도구입니다.

### 왜 dox인가?

- 📝 **수작업 제거**: 수십, 수백 개의 문서에서 텍스트를 일괄 변경
- 🔄 **형식 변환**: 마크다운으로 작성하고 Word/PowerPoint로 자동 변환
- 📋 **템플릿 활용**: 계약서, 보고서 등 반복 문서를 템플릿으로 자동 생성
- 🤖 **AI 통합**: OpenAI를 활용한 콘텐츠 자동 생성
- 🌍 **한국어 지원**: 한국어 인터페이스 완벽 지원
- 🚀 **간단한 설치**: 단일 실행 파일, 별도 의존성 없음
- 🎨 **아름다운 UI**: 색상 출력과 프로그레스 바로 시각적 피드백 제공

## ✨ 주요 기능

### 🔄 문서 텍스트 일괄 치환
- Word(.docx)와 PowerPoint(.pptx) 파일의 텍스트 일괄 변경
- YAML 기반 규칙 파일로 쉬운 관리
- 재귀적 디렉토리 처리 및 패턴 제외 기능
- 동시 처리로 40-70% 성능 향상
- 자동 백업 생성 기능
- 프로그레스 바와 색상 출력으로 진행 상황 추적

### 📝 문서 생성
- 마크다운을 Word 또는 PowerPoint로 변환
- 템플릿 기반 문서 생성
- 스타일과 포맷 보존
- 복잡한 문서 구조 지원
- 코드 블록, 리스트, 테이블 등 모든 마크다운 요소 지원

### 🤖 AI 콘텐츠 생성
- OpenAI를 활용한 블로그, 보고서, 요약 생성
- 다양한 콘텐츠 타입과 커스터마이징 가능한 파라미터
- Temperature와 토큰 제어로 출력 미세 조정
- GPT-3.5와 GPT-4 모델 지원
- 설정 파일을 통한 API 키 관리

### 📋 템플릿 처리
- 플레이스홀더가 있는 Word/PowerPoint 템플릿 처리
- YAML/JSON 기반 데이터 주입
- 복잡한 데이터 구조 지원
- 누락된 플레이스홀더 검증 및 감지
- 배치 처리 기능

### 🎨 아름다운 UI
- 더 나은 가독성을 위한 색상 출력
- 긴 작업을 위한 프로그레스 바
- AI 작업을 위한 로딩 스피너
- 파일 타입별 색상 구분
- 시각적 서식이 있는 요약 통계
- NO_COLOR 환경 변수 지원

### 🌍 국제화
- 한국어와 영어 인터페이스 완벽 지원
- 시스템 로케일 기반 자동 언어 감지
- --lang 플래그로 쉬운 언어 전환

### ⚙️ 설정 관리
- YAML 기반 설정 파일 시스템
- 환경 변수 지원
- 우선순위: CLI 플래그 > 설정 파일 > 환경 변수
- 전역 설정과 명령별 설정

## 📦 설치

### 빠른 설치 (권장)

#### Windows
```powershell
# PowerShell에서 실행
Invoke-WebRequest -Uri "https://github.com/pyhub-kr/pyhub-documents-cli/releases/latest/download/dox-windows-amd64.exe" -OutFile "dox.exe"

# PATH에 추가하거나 원하는 위치로 이동
Move-Item dox.exe C:\Windows\System32\
```

#### macOS
```bash
# Intel Mac
curl -L https://github.com/pyhub-kr/pyhub-documents-cli/releases/latest/download/dox-darwin-amd64 -o dox

# Apple Silicon (M1/M2/M3)
curl -L https://github.com/pyhub-kr/pyhub-documents-cli/releases/latest/download/dox-darwin-arm64 -o dox

# 실행 권한 부여 및 설치
chmod +x dox
sudo mv dox /usr/local/bin/
```

#### Linux
```bash
# AMD64
curl -L https://github.com/pyhub-kr/pyhub-documents-cli/releases/latest/download/dox-linux-amd64 -o dox

# ARM64
curl -L https://github.com/pyhub-kr/pyhub-documents-cli/releases/latest/download/dox-linux-arm64 -o dox

chmod +x dox
sudo mv dox /usr/local/bin/
```

### 설치 확인
```bash
# 버전 확인
dox version

# 도움말 확인
dox --help
```

### 소스에서 빌드

Go 1.21 이상이 설치되어 있어야 합니다.

```bash
# 저장소 클론
git clone https://github.com/pyhub-kr/pyhub-documents-cli.git
cd pyhub-documents-cli

# 빌드
go build -o dox

# 또는 전역 설치
go install

# 특정 플랫폼용 빌드
GOOS=windows GOARCH=amd64 go build -o dox.exe
GOOS=darwin GOARCH=arm64 go build -o dox
GOOS=linux GOARCH=amd64 go build -o dox
```

## 🚀 빠른 시작

### 1. 문서 텍스트 일괄 치환

여러 Word/PowerPoint 문서에서 텍스트를 한 번에 변경합니다.

#### 규칙 파일 작성 (rules.yml)
```yaml
# 버전 업데이트
- old: "v1.0.0"
  new: "v2.0.0"

# 연도 변경
- old: "2024년"
  new: "2025년"

# 회사명 변경
- old: "구 회사명"
  new: "신 회사명"
```

#### 실행 명령
```bash
# 단일 파일 처리
dox replace --rules rules.yml --path 보고서.docx

# 디렉토리 내 모든 문서 처리
dox replace --rules rules.yml --path ./문서폴더

# 미리보기 (실제 변경하지 않음)
dox replace --rules rules.yml --path ./문서폴더 --dry-run

# 백업 생성 후 처리
dox replace --rules rules.yml --path ./문서폴더 --backup

# 동시 처리로 성능 향상
dox replace --rules rules.yml --path ./문서폴더 --concurrent --max-workers 8

# 특정 파일 제외
dox replace --rules rules.yml --path . --exclude "*.backup"
```

### 2. 마크다운을 Office 문서로 변환

마크다운으로 작성한 문서를 Word나 PowerPoint로 변환합니다.

#### Word 문서로 변환
```bash
# 기본 변환
dox create --from 주간보고서.md --output 주간보고서.docx

# 템플릿 사용
dox create --from 내용.md --template 회사템플릿.docx --output 최종보고서.docx

# 기존 파일 덮어쓰기
dox create --from 월간보고서.md --output 월간보고서.docx --force
```

#### PowerPoint 프레젠테이션으로 변환
```bash
# 마크다운을 프레젠테이션으로 변환
dox create --from 발표자료.md --output 발표자료.pptx
```

**PowerPoint 변환 규칙:**
- `#` (H1): 새 슬라이드 생성
- `##` (H2): 슬라이드 제목 또는 굵은 텍스트
- `###`-`######`: 굵은 텍스트로 변환
- 리스트: 글머리 기호로 변환
- 코드 블록: 고정폭 폰트로 표시

**예시 마크다운 (발표자료.md):**
```markdown
# 2025년 사업 계획

## 목표
- 매출 200% 성장
- 신규 고객 1,000명 확보
- 해외 진출

# 실행 전략

## 1분기 계획
- 제품 개선
- 마케팅 강화

## 2분기 계획
- 파트너십 확대
- 신규 기능 출시
```

### 3. 템플릿 문서 처리

플레이스홀더가 있는 템플릿 문서를 데이터로 채웁니다.

#### 템플릿 문서 준비
Word/PowerPoint 문서에 `{{변수명}}` 형식의 플레이스홀더를 삽입합니다.

예시 (계약서_템플릿.docx):
```
계약서

갑: {{회사명}}
을: {{고객명}}
계약일: {{계약일}}
금액: {{금액}}원
```

#### 값 파일 작성 (values.yml)
```yaml
회사명: "파이허브 주식회사"
고객명: "김철수"
계약일: "2025년 1월 1일"
금액: "10,000,000"
```

#### 실행 명령
```bash
# YAML 파일로 값 제공
dox template --template 계약서_템플릿.docx --values values.yml --output 계약서_최종.docx

# 명령줄에서 직접 값 설정
dox template --template 보고서_템플릿.pptx --output 보고서_202501.pptx \
  --set 제목="월간 보고서" \
  --set 작성자="홍길동" \
  --set 날짜="2025년 1월"

# JSON 파일도 지원
dox template --template 템플릿.docx --values data.json --output 결과.docx
```

### 4. AI 콘텐츠 생성

OpenAI (GPT) 또는 Claude를 활용하여 다양한 콘텐츠를 생성합니다.

#### API 키 설정

**OpenAI 사용 시:**
```bash
# 환경 변수 설정
export OPENAI_API_KEY="your-openai-api-key"

# 또는 설정 파일 사용
dox config --set openai.api_key "your-openai-api-key"
```

**Claude 사용 시:**
```bash
# 환경 변수 설정
export ANTHROPIC_API_KEY="your-anthropic-api-key"
# 또는
export CLAUDE_API_KEY="your-anthropic-api-key"

# 또는 설정 파일 사용
dox config --set claude.api_key "your-anthropic-api-key"
```

#### 콘텐츠 생성

**OpenAI (GPT) 사용:**
```bash
# 블로그 포스트 생성 (기본: GPT-3.5)
dox generate --type blog --prompt "Go 테스팅 베스트 프랙티스" --output blog.md

# GPT-4로 보고서 생성
dox generate --type report --prompt "3분기 매출 분석" --model gpt-4 --output report.md

# 커스텀 파라미터로 생성
dox generate --type custom \
  --prompt "Docker에 대한 기술 튜토리얼 작성" \
  --temperature 0.7 \
  --max-tokens 2000 \
  --output tutorial.md
```

**Claude 사용:**
```bash
# Claude로 블로그 생성 (모델 이름으로 자동 감지)
dox generate --type blog --prompt "AI 윤리 가이드라인" \
  --model claude-3-sonnet-20240229 --output blog.md

# Claude Opus로 복잡한 분석
dox generate --provider claude --model claude-3-opus-20240229 \
  --prompt "대규모 시스템 아키텍처 분석" \
  --max-tokens 4000 --output analysis.md

# Claude Haiku로 빠른 요약
dox generate --provider claude --model claude-3-haiku-20240307 \
  --type summary --prompt "$(cat long-document.md)" \
  --output summary.md

# 이메일 작성
dox generate --provider claude --type email \
  --prompt "프로젝트 지연에 대한 사과 메일" \
  --output email.md
```

**지원하는 AI 모델:**
- **OpenAI**: GPT-3.5-Turbo, GPT-4, GPT-4-Turbo
- **Claude**: Claude 3 Opus (최고 성능), Claude 3 Sonnet (균형), Claude 3 Haiku (빠른 응답)

### 5. 설정 관리

```bash
# 설정 파일 초기화
dox config --init

# 모든 설정 보기
dox config --list

# 설정값 지정
dox config --set openai.api_key "your-key"
dox config --set global.lang "ko"
dox config --set replace.concurrent true

# 설정값 조회
dox config --get openai.model
```

## ⚙️ 설정 파일

dox는 명령줄 플래그와 설정 파일을 모두 지원합니다. 우선순위:
1. 명령줄 플래그 (최우선)
2. 설정 파일
3. 환경 변수 (최하위)

### 설정 파일 위치

`~/.pyhub/config.yml`:

```yaml
# OpenAI 설정
openai:
  api_key: "your-openai-api-key"  # 또는 OPENAI_API_KEY 환경 변수 사용
  model: "gpt-3.5-turbo"
  max_tokens: 2000
  temperature: 0.7

# Claude 설정
claude:
  api_key: "your-anthropic-api-key"  # 또는 ANTHROPIC_API_KEY 환경 변수 사용
  model: "claude-3-sonnet-20240229"
  max_tokens: 2000
  temperature: 0.7

# 문서 치환 설정
replace:
  backup: true
  recursive: true
  concurrent: true
  max_workers: 8

# 콘텐츠 생성 설정
generate:
  model: "gpt-3.5-turbo"  # 또는 claude 모델명
  max_tokens: 2000
  temperature: 0.7
  content_type: "blog"

# 전역 설정
global:
  verbose: false
  quiet: false
  lang: "ko"  # 또는 "en" (영어)
```

## 🌍 다국어 지원

### 언어 자동 감지

시스템 언어 설정에 따라 자동으로 한국어 또는 영어 인터페이스를 표시합니다.

**감지 우선순위:**
1. `--lang` 플래그
2. 설정 파일의 `global.lang`
3. `LANG` 환경 변수
4. `LC_ALL` 환경 변수
5. 기본값 (영어)

### 사용 예시

```bash
# 명시적으로 한국어 지정
dox --lang ko replace --rules rules.yml --path ./docs

# 설정 파일에서 기본 언어 지정
dox config --set global.lang ko

# 시스템 언어가 한국어인 경우 자동 감지
$ echo $LANG
ko_KR.UTF-8

$ dox create --from 보고서.md --output 보고서.docx
보고서.md를 Word 문서로 변환 중...
✅ 보고서.docx 생성 완료
```

## 📖 명령어 상세 가이드

### 전역 플래그
- `--config` - 설정 파일 경로 지정
- `--verbose, -v` - 자세한 출력
- `--quiet, -q` - 조용한 모드 (에러만 출력)
- `--no-color` - 색상 출력 비활성화
- `--lang` - 인터페이스 언어 (ko, en)

### `replace` - 텍스트 일괄 치환

Word와 PowerPoint 문서의 텍스트를 YAML 규칙에 따라 일괄 변경합니다.

#### 옵션
- `--rules, -r`: YAML 규칙 파일 경로 (필수)
- `--path, -p`: 대상 파일 또는 디렉토리 경로 (필수)
- `--dry-run`: 실제 변경 없이 미리보기
- `--backup`: 원본 파일 백업 생성
- `--recursive`: 하위 디렉토리 포함 (기본값: true)
- `--exclude`: 제외할 파일 패턴
- `--concurrent`: 동시 처리 활성화
- `--max-workers`: 워커 수 (기본값: CPU 코어 수)

### `create` - 마크다운 변환

마크다운 파일을 Word 또는 PowerPoint 문서로 변환합니다.

#### 옵션
- `--from, -f`: 입력 마크다운 파일 (필수)
- `--output, -o`: 출력 파일 경로 (필수)
- `--template, -t`: 스타일링을 위한 템플릿 문서
- `--format`: 출력 형식 (docx/pptx, 확장자에서 자동 감지)
- `--force`: 기존 파일 덮어쓰기

### `template` - 템플릿 문서 처리

플레이스홀더가 포함된 템플릿 문서를 데이터로 채워 완성합니다.

#### 옵션
- `--template, -t`: 템플릿 문서 파일 (필수)
- `--output, -o`: 출력 파일 경로 (필수)
- `--values`: 값을 포함한 YAML/JSON 파일
- `--set`: 개별 값 설정 (key=value 형식)
- `--force`: 기존 파일 덮어쓰기

### `generate` - AI 콘텐츠 생성

OpenAI를 활용하여 다양한 콘텐츠를 생성합니다.

#### 옵션
- `--prompt, -p`: 생성 프롬프트 (필수)
- `--type, -t`: 콘텐츠 타입 (blog, report, summary, custom)
- `--output, -o`: 출력 파일 경로
- `--model`: AI 모델 (gpt-3.5-turbo, gpt-4)
- `--max-tokens`: 최대 응답 토큰 수
- `--temperature`: 창의성 레벨 (0.0-1.0)
- `--api-key`: OpenAI API 키

### `config` - 설정 관리

설정 파일을 관리합니다.

#### 옵션
- `--init`: 설정 파일 초기화
- `--list`: 모든 설정값 나열
- `--get <key>`: 특정 값 조회
- `--set <key=value>`: 설정값 지정

### `version` - 버전 정보

```bash
dox version
# 출력:
# dox version 1.2534.28
#   Commit: abc123
#   Built:  2025-01-01
```

## 📁 예제

### 실제 사용 시나리오

#### 시나리오 1: 연말 문서 업데이트
```bash
# 1. 규칙 파일 생성
cat > year-end-update.yml << EOF
- old: "2024년"
  new: "2025년"
- old: "4분기"
  new: "1분기"
- old: "연말"
  new: "연초"
EOF

# 2. 모든 문서 백업 및 업데이트
dox replace --rules year-end-update.yml \
  --path ./company-docs \
  --backup \
  --concurrent

# 3. 변경 보고서 생성
dox generate --type report \
  --prompt "2025년 문서 업데이트 완료 보고서 작성" \
  --output update-report.md

# 4. Word로 변환
dox create --from update-report.md --output update-report.docx
```

#### 시나리오 2: 월간 보고서 자동화
```bash
#!/bin/bash
# monthly-report.sh

# 1. AI로 보고서 초안 생성
dox generate --type report \
  --prompt "$(cat metrics.txt) 기반 월간 성과 보고서 작성" \
  --output draft.md

# 2. 템플릿에 데이터 삽입
dox template \
  --template report-template.docx \
  --values monthly-data.yml \
  --output monthly-report.docx \
  --set month="$(date +%B)" \
  --set year="$(date +%Y)"

# 3. 프레젠테이션 생성
dox create --from draft.md --output presentation.pptx
```

더 많은 예제는 [examples/](examples/) 디렉토리를 참조하세요.

## 🔧 고급 사용법

### 성능 최적화

대량 문서 처리 시 동시 처리 사용:
```bash
# 16개 워커로 처리
dox replace --rules rules.yml --path ./large-docs \
  --concurrent --max-workers 16

# 진행 상황 모니터링
dox replace --rules rules.yml --path ./docs \
  --concurrent --verbose
```

### CI/CD 통합

```yaml
# .github/workflows/docs.yml
name: Document Processing
on:
  push:
    paths:
      - 'docs/**'
jobs:
  process:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Install dox
        run: |
          curl -L https://github.com/pyhub-kr/pyhub-documents-cli/releases/latest/download/dox-Linux-x86_64 -o dox
          chmod +x dox
      - name: Process documents
        run: |
          ./dox replace --rules ci-rules.yml --path docs/
          ./dox create --from CHANGELOG.md --output CHANGELOG.docx
```

## 🛠️ 개발

### 프로젝트 구조
```
dox/
├── cmd/            # CLI 명령어 구현
├── internal/       # 내부 패키지
│   ├── config/     # 설정 관리
│   ├── errors/     # 커스텀 에러 타입
│   ├── generate/   # AI 콘텐츠 생성
│   ├── i18n/       # 국제화 지원
│   ├── markdown/   # 마크다운 변환
│   ├── openai/     # OpenAI API 클라이언트
│   ├── replace/    # 텍스트 치환
│   ├── template/   # 템플릿 처리
│   └── ui/         # UI 컴포넌트 (프로그레스 바, 색상)
├── locales/        # 번역 파일
├── examples/       # 사용 예제
├── scripts/        # 빌드 스크립트
└── tests/          # 테스트 파일
```

### 테스트 실행
```bash
# 모든 테스트 실행
go test ./...

# 커버리지 확인
go test -cover ./...

# 레이스 조건 검사
go test -race ./...

# 특정 패키지 테스트
go test ./internal/replace
```

### 빌드
```bash
# 현재 플랫폼용 빌드
go build -o dox

# 크로스 컴파일
make build-all

# 릴리스 빌드 (최적화)
go build -ldflags="-s -w" -o dox
```

## 🔢 버저닝 (HeadVer)

이 프로젝트는 [HeadVer](https://github.com/line/headver) 버저닝 시스템을 사용합니다.

### 버전 형식
```
{head}.{yearweek}.{build}
```

- **head**: 주요 버전 (수동 관리, Breaking Change 시 증가)
- **yearweek**: 연도(2자리) + 주차(2자리) - 자동 생성
- **build**: 해당 주의 빌드 번호 - 자동 생성

### 예시
- `1.2534.0`: 버전 1, 2025년 34주차, 첫 번째 빌드
- `1.2534.5`: 같은 주의 5번째 빌드
- `2.2601.0`: 버전 2 (Breaking Change), 2026년 1주차

## 🤝 기여하기

프로젝트 개선에 참여해 주세요! 자세한 내용은 [CONTRIBUTING.md](CONTRIBUTING.md)를 참조하세요.

### 기여 방법
1. 이슈를 먼저 생성하여 논의
2. 저장소 포크
3. 기능 브랜치 생성 (`feature/기능명`)
4. 테스트 작성 (TDD)
5. 구현 및 커밋
6. Pull Request 제출

### 개발 가이드라인
- Go 1.21+ 사용
- `gofmt`로 코드 포맷팅
- 테스트 커버리지 80% 이상 유지
- 한국어/영어 i18n 지원 필수
- Conventional Commits 사용

## 🗺️ 로드맵

- [ ] Excel 파일 지원 (.xlsx)
- [ ] PDF 생성 및 처리
- [ ] HWP (한글) 포맷 지원
- [ ] 클라우드 스토리지 통합 (S3, Google Drive)
- [ ] 웹 UI 인터페이스
- [ ] 플러그인 시스템
- [ ] 더 많은 AI 제공자 (Claude, Gemini, Local LLMs)
- [ ] 문서 비교 및 diff 기능
- [ ] 배치 처리 개선
- [ ] Docker 컨테이너 지원

## 📄 라이선스

MIT 라이선스 - [LICENSE](LICENSE) 파일 참조

## 🙏 감사의 말

- [Cobra](https://github.com/spf13/cobra) - CLI 프레임워크
- [unioffice](https://github.com/unidoc/unioffice) - Office 문서 처리
- [goldmark](https://github.com/yuin/goldmark) - 마크다운 파서
- [progressbar](https://github.com/schollz/progressbar) - 프로그레스 표시
- [color](https://github.com/fatih/color) - 터미널 색상
- Go 커뮤니티의 훌륭한 오픈소스 라이브러리들

## 📞 지원

- 🐛 버그 리포트: [Issues](https://github.com/pyhub-kr/pyhub-documents-cli/issues)
- 💬 질문과 토론: [Discussions](https://github.com/pyhub-kr/pyhub-documents-cli/discussions)
- 📧 이메일: support@pyhub.kr
- 📚 문서: [Wiki](https://github.com/pyhub-kr/pyhub-documents-cli/wiki)

---

Made with ❤️ by [PyHub Korea](https://pyhub.kr)