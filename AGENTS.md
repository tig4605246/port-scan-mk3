# port-scan-mk3 Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-03-02

## Active Technologies
- Go 1.24.x (`go 1.24.0`, toolchain `go1.24.4`) + Go standard library (`encoding/csv`, `encoding/json`, `strings`, `net`, `strconv`, `context`, `time`, `os`), existing internal packages (`pkg/input`, `pkg/task`, `pkg/scanapp`, `pkg/writer`) (001-enhance-input-parser)
- Local filesystem (input CSV, output CSV, resume JSON) (001-enhance-input-parser)

- Go 1.24.x (`go 1.24.0`, toolchain `go1.24.4`) + Go standard library (`flag`, `net`, `encoding/csv`, `encoding/json`, `context`, `os`, `time`), `golang.org/x/term`, `golang.org/x/sys` (001-ip-aware-baseline-spec)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.24.x (`go 1.24.0`, toolchain `go1.24.4`)

## Code Style

Go 1.24.x (`go 1.24.0`, toolchain `go1.24.4`): Follow standard conventions

## Recent Changes
- 001-enhance-input-parser: Added Go 1.24.x (`go 1.24.0`, toolchain `go1.24.4`) + Go standard library (`encoding/csv`, `encoding/json`, `strings`, `net`, `strconv`, `context`, `time`, `os`), existing internal packages (`pkg/input`, `pkg/task`, `pkg/scanapp`, `pkg/writer`)

- 001-ip-aware-baseline-spec: Added Go 1.24.x (`go 1.24.0`, toolchain `go1.24.4`) + Go standard library (`flag`, `net`, `encoding/csv`, `encoding/json`, `context`, `os`, `time`), `golang.org/x/term`, `golang.org/x/sys`

<!-- MANUAL ADDITIONS START -->
## Architecture Guardrails

- 程式碼結構必須符合 SOLID 設計原則。
- `cmd/port-scan` 只負責 CLI 組裝、參數解析與使用者 I/O；可重用邏輯必須放在 `pkg/`。
- 介面必須精簡且由消費端擁有；禁止 god struct、god interface 與循環依賴。
<!-- MANUAL ADDITIONS END -->
