# 개선 작업 GitHub Issues

## 🎯 Milestone: v2.0.0 - Stability & Performance
**목표**: 현재 구현된 기능의 안정성, 성능, 사용성 개선
**예상 기간**: 4주

---

## Priority: HIGH (긴급)

### Issue #20: API 재시도 메커니즘 구현
**Labels**: `enhancement`, `priority:high`, `api`
**Milestone**: v2.0.0

**설명**: 
OpenAI/Claude API 호출 시 네트워크 오류나 일시적 장애에 대한 자동 재시도 메커니즘 구현

**작업 내용**:
- [ ] Exponential backoff 알고리즘 구현
- [ ] 최대 재시도 횟수 설정 (기본값: 3회)
- [ ] Rate limiting 감지 및 대응
- [ ] 재시도 가능한 에러 타입 정의
- [ ] 설정 파일에 재시도 옵션 추가

**구현 위치**:
- `internal/openai/client.go`
- `internal/claude/client.go`
- `internal/config/config.go`

**테스트**:
- [ ] 네트워크 오류 시뮬레이션 테스트
- [ ] Rate limiting 테스트
- [ ] Timeout 테스트

---

### Issue #21: 캐싱 시스템 구현
**Labels**: `enhancement`, `priority:high`, `performance`
**Milestone**: v2.0.0

**설명**:
AI 응답 및 문서 처리 결과를 캐싱하여 성능 향상

**작업 내용**:
- [ ] 캐시 인터페이스 설계
- [ ] 메모리 캐시 구현 (LRU)
- [ ] 디스크 캐시 구현 (선택적)
- [ ] AI 응답 캐싱 (prompt hash 기반)
- [ ] 템플릿 파싱 결과 캐싱
- [ ] 캐시 만료 정책 구현
- [ ] 캐시 통계 및 모니터링

**구현 위치**:
- `internal/cache/` (새 패키지)
- `internal/generate/generator.go`
- `internal/template/`

**테스트**:
- [ ] 캐시 히트/미스 테스트
- [ ] 동시성 테스트
- [ ] 메모리 사용량 테스트

---

### Issue #22: 에러 메시지 개선 및 다국어 지원
**Labels**: `enhancement`, `priority:high`, `i18n`, `ux`
**Milestone**: v2.0.0

**설명**:
사용자 친화적인 에러 메시지와 문제 해결 가이드 제공

**작업 내용**:
- [ ] 에러 코드 체계 도입
- [ ] 에러별 해결 방법 제공
- [ ] 다국어 에러 메시지 (한국어/영어)
- [ ] 컨텍스트 정보 포함 (파일명, 라인 등)
- [ ] 에러 레벨 구분 (ERROR, WARNING, INFO)

**구현 위치**:
- `internal/errors/` 확장
- `internal/i18n/messages.go`

**예시**:
```
Error [DOX001]: API key not found
Solution: Set OPENAI_API_KEY environment variable or use 'dox config --set openai.api_key'
```

---

### Issue #23: API 키 보안 강화
**Labels**: `security`, `priority:high`
**Milestone**: v2.0.0

**설명**:
API 키 저장 및 관리 보안 강화

**작업 내용**:
- [ ] 키체인/시크릿 매니저 통합 (macOS Keychain, Windows Credential Manager)
- [ ] API 키 검증 로직 강화
- [ ] 민감 정보 로깅 방지
- [ ] 설정 파일 권한 검사 (600)
- [ ] API 키 마스킹 개선 (로그, 에러 메시지)

**구현 위치**:
- `internal/secrets/` (새 패키지)
- `internal/config/config.go`
- `cmd/config.go`

---

## Priority: MEDIUM (중요)

### Issue #24: 테스트 커버리지 80% 달성
**Labels**: `test`, `priority:medium`, `quality`
**Milestone**: v2.0.0

**설명**:
현재 60-70% 수준의 테스트 커버리지를 80% 이상으로 향상

**작업 내용**:
- [ ] `internal/errors` 패키지 테스트 (현재 0%)
- [ ] `internal/ui` 패키지 테스트 (현재 0%)  
- [ ] `cmd` 패키지 테스트 강화 (현재 23.6%)
- [ ] `internal/template` 테스트 보완 (현재 46.2%)
- [ ] 통합 테스트 시나리오 추가
- [ ] 테스트 헬퍼 함수 개발

**목표 커버리지**:
- 전체: 80% 이상
- 핵심 패키지: 90% 이상

---

### Issue #25: 진행 상황 표시 개선
**Labels**: `enhancement`, `priority:medium`, `ux`
**Milestone**: v2.0.0

**설명**:
더 정확하고 유용한 진행 상황 정보 제공

**작업 내용**:
- [ ] ETA (예상 완료 시간) 계산 및 표시
- [ ] 작업 취소 기능 (Ctrl+C graceful shutdown)
- [ ] 처리 속도 표시 (files/sec, MB/sec)
- [ ] 상세 진행률 (현재 파일명 표시)
- [ ] 로그 레벨 옵션 (--log-level debug|info|warn|error)

**구현 위치**:
- `internal/ui/progress.go`
- `cmd/` 각 명령어 파일

---

### Issue #26: 대용량 파일 처리 최적화
**Labels**: `enhancement`, `priority:medium`, `performance`
**Milestone**: v2.0.0

**설명**:
메모리 효율적인 대용량 파일 처리

**작업 내용**:
- [ ] 스트리밍 방식 문서 처리
- [ ] 청크 단위 읽기/쓰기
- [ ] 메모리 풀 사용
- [ ] 파일 크기별 처리 전략
- [ ] 메모리 사용량 모니터링

**구현 위치**:
- `internal/document/`
- `internal/replace/`

---

### Issue #27: 드라이런 모드 확장
**Labels**: `enhancement`, `priority:medium`, `feature`
**Milestone**: v2.0.0

**설명**:
모든 명령에 --dry-run 지원 및 변경 사항 미리보기

**작업 내용**:
- [ ] `generate` 명령 드라이런 (토큰 수, 예상 비용 표시)
- [ ] `template` 명령 드라이런 (변경될 변수 표시)
- [ ] `create` 명령 드라이런 (생성될 파일 정보)
- [ ] 변경 사항 diff 형식 출력
- [ ] JSON 형식 출력 옵션

**구현 위치**:
- `cmd/` 각 명령어 파일

---

## Priority: LOW (개선)

### Issue #28: 코드 리팩토링 - AI 클라이언트 통합
**Labels**: `refactor`, `priority:low`, `code-quality`
**Milestone**: v2.0.0

**설명**:
OpenAI와 Claude 클라이언트의 중복 코드 제거

**작업 내용**:
- [ ] AI 클라이언트 인터페이스 정의
- [ ] 공통 로직 추출 (retry, error handling)
- [ ] 팩토리 패턴 적용
- [ ] 의존성 주입 구조 개선

**구현 위치**:
- `internal/ai/` (새 패키지)
- `internal/openai/` 리팩토링
- `internal/claude/` 리팩토링

---

### Issue #29: 대화형 설정 마법사
**Labels**: `enhancement`, `priority:low`, `ux`
**Milestone**: v2.0.0

**설명**:
초보자를 위한 대화형 설정 마법사

**작업 내용**:
- [ ] `dox config --wizard` 명령 추가
- [ ] API 키 설정 가이드
- [ ] 기본값 설정 안내
- [ ] 설정 검증 및 테스트

**구현 위치**:
- `cmd/config.go`
- `internal/wizard/` (새 패키지)

---

### Issue #30: CI/CD 파이프라인 개선
**Labels**: `ci/cd`, `priority:low`, `infrastructure`
**Milestone**: v2.0.0

**설명**:
GitHub Actions 워크플로우 개선

**작업 내용**:
- [ ] 테스트 자동화 (모든 PR에서)
- [ ] golangci-lint 통합
- [ ] gosec 보안 스캔
- [ ] 코드 커버리지 리포트
- [ ] 바이너리 이름 통일 (pyhub-docs → dox)
- [ ] Docker 이미지 빌드 및 배포

**구현 위치**:
- `.github/workflows/`

---

### Issue #31: 벤치마크 및 성능 모니터링
**Labels**: `test`, `priority:low`, `performance`
**Milestone**: v2.0.0

**설명**:
성능 측정 및 회귀 방지를 위한 벤치마크

**작업 내용**:
- [ ] 핵심 기능 벤치마크 테스트
- [ ] 성능 회귀 감지
- [ ] 메모리 프로파일링
- [ ] CPU 프로파일링
- [ ] 벤치마크 결과 트래킹

**구현 위치**:
- `*_bench_test.go` 파일들

---

## 구현 순서 및 일정

### Week 1 (긴급)
- Issue #20: API 재시도 메커니즘
- Issue #22: 에러 메시지 개선
- Issue #23: API 키 보안

### Week 2 (성능)
- Issue #21: 캐싱 시스템
- Issue #26: 대용량 파일 처리

### Week 3 (품질)
- Issue #24: 테스트 커버리지
- Issue #31: 벤치마크

### Week 4 (사용성)
- Issue #25: 진행 상황 표시
- Issue #27: 드라이런 모드
- Issue #29: 설정 마법사

### Ongoing (지속적)
- Issue #28: 코드 리팩토링
- Issue #30: CI/CD 개선

## 성공 지표

- ✅ API 실패율 < 1%
- ✅ 테스트 커버리지 > 80%
- ✅ 평균 응답 시간 30% 개선
- ✅ 메모리 사용량 50% 감소
- ✅ 사용자 만족도 향상