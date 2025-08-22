# dox

[![Go Version](https://img.shields.io/badge/go-1.21-blue.svg)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/pyhub-kr/pyhub-documents-cli)](https://github.com/pyhub-kr/pyhub-documents-cli/releases)
[![HeadVer](https://img.shields.io/badge/versioning-HeadVer-blue)](https://github.com/line/headver)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

문서 자동화 및 AI 기반 콘텐츠 생성을 위한 강력한 CLI 도구입니다.

## 🎯 소개

`dox`는 반복적인 문서 작업을 자동화하고, 마크다운을 Office 문서로 변환하며, 템플릿 기반 문서 생성을 지원하는 Go 언어 기반 CLI 도구입니다.

### 왜 dox인가?

- 📝 **수작업 제거**: 수십, 수백 개의 문서에서 텍스트를 일괄 변경
- 🔄 **형식 변환**: 마크다운으로 작성하고 Word/PowerPoint로 자동 변환
- 📋 **템플릿 활용**: 계약서, 보고서 등 반복 문서를 템플릿으로 자동 생성
- 🌍 **한국어 지원**: 한국어 인터페이스 완벽 지원
- 🚀 **간단한 설치**: 단일 실행 파일, 별도 의존성 없음

## ✨ 주요 기능

### 구현 완료 ✅

- **문서 텍스트 일괄 치환**: Word/PowerPoint 문서의 텍스트를 YAML 규칙으로 한 번에 변경
- **마크다운 변환**: Markdown 파일을 Word(.docx) 또는 PowerPoint(.pptx)로 변환
- **템플릿 문서 처리**: 플레이스홀더가 포함된 템플릿 문서를 데이터로 채워 완성
- **국제화(i18n)**: 한국어/영어 인터페이스 자동 감지 및 선택
- **크로스 플랫폼**: Windows, macOS, Linux 모두 지원

### 개발 예정 🚧

- **AI 콘텐츠 생성**: OpenAI를 활용한 문서 내용 자동 생성 (Phase 2)
- **HWP 지원**: 한글(HWP) 파일 형식 지원 (Phase 3)

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

# Apple Silicon (M1/M2)
curl -L https://github.com/pyhub-kr/pyhub-documents-cli/releases/latest/download/dox-darwin-arm64 -o dox

# 실행 권한 부여 및 설치
chmod +x dox
sudo mv dox /usr/local/bin/
```

#### Linux
```bash
curl -L https://github.com/pyhub-kr/pyhub-documents-cli/releases/latest/download/dox-linux-amd64 -o dox
chmod +x dox
sudo mv dox /usr/local/bin/
```

### 소스에서 빌드

Go 1.21 이상이 설치되어 있어야 합니다.

```bash
# 저장소 클론
git clone https://github.com/pyhub-kr/pyhub-documents-cli.git
cd pyhub-documents-cli

# 빌드
make build

# 특정 플랫폼용 빌드
make build-windows  # Windows
make build-darwin   # macOS
make build-linux    # Linux
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
```

### 2. 마크다운을 Office 문서로 변환

마크다운으로 작성한 문서를 Word나 PowerPoint로 변환합니다.

#### Word 문서로 변환
```bash
# 기본 변환
dox create --from 주간보고서.md --output 주간보고서.docx

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
- `##` (H2): 섹션 첫 번째면 슬라이드 제목, 아니면 굵은 텍스트
- 리스트, 문단, 코드 블록 등 모두 지원

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

## 🌍 다국어 지원

### 언어 자동 감지

시스템 언어 설정에 따라 자동으로 한국어 또는 영어 인터페이스를 표시합니다.

**감지 우선순위:**
1. `--lang` 플래그
2. `LANG` 환경 변수
3. `LC_ALL` 환경 변수
4. 기본값 (영어)

### 한국어 인터페이스 사용

```bash
# 명시적으로 한국어 지정
dox --lang ko replace --rules rules.yml --path ./docs

# 시스템 언어가 한국어인 경우 자동 감지
$ echo $LANG
ko_KR.UTF-8

$ dox create --from 보고서.md --output 보고서.docx
보고서.md를 Word 문서로 변환 중...
✅ 보고서.docx 생성 완료
```

### 영어 인터페이스 사용

```bash
# 명시적으로 영어 지정
dox --lang en replace --rules rules.yml --path ./docs

# 결과
Converting report.md to Word document...
✅ Successfully created report.docx
```

## 📖 명령어 상세 가이드

### `replace` - 텍스트 일괄 치환

Word와 PowerPoint 문서의 텍스트를 YAML 규칙에 따라 일괄 변경합니다.

#### 옵션
- `--rules, -r`: YAML 규칙 파일 경로 (필수)
- `--path, -p`: 대상 파일 또는 디렉토리 경로 (필수)
- `--dry-run`: 실제 변경 없이 미리보기
- `--backup`: 원본 파일 백업 생성
- `--recursive`: 하위 디렉토리 포함 (기본값: true)
- `--exclude`: 제외할 파일 패턴

#### 규칙 파일 형식
```yaml
# 단순 치환
- old: "찾을 텍스트"
  new: "바꿀 텍스트"

# 여러 규칙
- old: "v1.0.0"
  new: "v2.0.0"
- old: "2024"
  new: "2025"
- old: "구버전"
  new: "신버전"
```

#### 실제 사용 예시
```bash
# 모든 문서의 연도 업데이트
cat > year-update.yml << EOF
- old: "2024년"
  new: "2025년"
- old: "2024-"
  new: "2025-"
- old: "FY2024"
  new: "FY2025"
EOF

dox replace --rules year-update.yml --path ./연간보고서 --backup

# 특정 파일 제외
dox replace --rules rules.yml --path . --exclude "*.backup"
```

### `create` - 마크다운 변환

마크다운 파일을 Word 또는 PowerPoint 문서로 변환합니다.

#### 옵션
- `--from, -f`: 입력 마크다운 파일 (필수)
- `--output, -o`: 출력 파일 경로 (필수)
- `--format`: 출력 형식 (docx/pptx, 확장자에서 자동 감지)
- `--force`: 기존 파일 덮어쓰기

#### Word 변환 특징
- 모든 마크다운 요소 지원
- 제목 계층 구조 유지
- 리스트, 코드 블록, 인용문 스타일 적용

#### PowerPoint 변환 규칙
- `#` (H1): 새 슬라이드 생성 및 제목
- `##` (H2): 슬라이드 내 섹션 제목
- `###`-`######`: 굵은 텍스트로 변환
- 리스트: 글머리 기호로 변환
- 코드 블록: 고정폭 폰트로 표시

#### 실제 사용 예시
```bash
# 주간 회의록 작성 워크플로우
echo "# 주간 회의록 - 2025년 1월 1주차

## 참석자
- 김부장
- 이과장
- 박대리

## 안건
1. 프로젝트 진행 상황
2. 다음 주 계획
3. 이슈 사항

## 결정 사항
- 마감일 1주 연장
- 추가 인력 투입" > 회의록.md

# Word 문서로 변환
dox create --from 회의록.md --output 회의록_20250101.docx
```

### `template` - 템플릿 문서 처리

플레이스홀더가 포함된 템플릿 문서를 데이터로 채워 완성합니다.

#### 옵션
- `--template, -t`: 템플릿 문서 파일 (필수)
- `--output, -o`: 출력 파일 경로 (필수)
- `--values, -v`: 값을 포함한 YAML/JSON 파일
- `--set, -s`: 개별 값 설정 (key=value 형식)
- `--force`: 기존 파일 덮어쓰기

#### 플레이스홀더 형식
템플릿 문서에서 `{{변수명}}` 형식을 사용합니다.

#### 실제 사용 예시

**견적서 자동 생성:**

템플릿 (견적서_템플릿.docx):
```
견적서

수신: {{고객사}}
담당자: {{담당자명}}

품목: {{제품명}}
수량: {{수량}}
단가: {{단가}}원
총액: {{총액}}원

유효기간: {{유효기간}}
```

값 파일 (견적_데이터.yml):
```yaml
고객사: "ABC 주식회사"
담당자명: "김철수 과장"
제품명: "소프트웨어 라이선스"
수량: 10
단가: "1,000,000"
총액: "10,000,000"
유효기간: "발행일로부터 30일"
```

실행:
```bash
dox template \
  --template 견적서_템플릿.docx \
  --values 견적_데이터.yml \
  --output 견적서_ABC_20250101.docx
```

## 🛠️ 개발

### 프로젝트 구조
```
pyhub-docs/
├── cmd/            # CLI 명령어 구현
├── internal/       # 내부 패키지
│   ├── document/   # 문서 처리 로직
│   ├── i18n/       # 국제화 지원
│   ├── markdown/   # 마크다운 변환
│   ├── replace/    # 텍스트 치환
│   └── template/   # 템플릿 처리
├── locales/        # 번역 파일
├── scripts/        # 빌드 스크립트
└── tests/          # 테스트 파일
```

### 테스트 실행
```bash
# 모든 테스트 실행
go test ./...

# 커버리지 확인
go test -cover ./...

# 특정 패키지 테스트
go test ./internal/replace
```

### 빌드
```bash
# 현재 플랫폼용 빌드
make build

# 모든 플랫폼용 빌드
make build-all

# 정리
make clean
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

### 버전 확인
```bash
dox version
# 출력: dox version 1.2534.23
```

## 🤝 기여하기

프로젝트 개선에 참여해 주세요!

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

## 📄 라이선스

MIT 라이선스 - [LICENSE](LICENSE) 파일 참조

## 🙏 감사의 말

- [Cobra](https://github.com/spf13/cobra) - CLI 프레임워크
- [goldmark](https://github.com/yuin/goldmark) - 마크다운 파서
- Go 커뮤니티의 훌륭한 오픈소스 라이브러리들

## 📞 지원

- 🐛 버그 리포트: [Issues](https://github.com/pyhub-kr/pyhub-documents-cli/issues)
- 💬 질문과 토론: [Discussions](https://github.com/pyhub-kr/pyhub-documents-cli/discussions)
- 📧 이메일: support@pyhub.kr (예정)

---

**참고**: 이 프로젝트는 활발히 개발 중입니다. 로드맵의 기능들이 순차적으로 추가될 예정입니다.