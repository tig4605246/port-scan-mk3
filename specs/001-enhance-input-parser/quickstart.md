# Quickstart: Enhanced Input Field Parsing

## 1. Prepare input sample

建立一份含 10 個 canonical 欄位的 CSV，並混入：

- 合法 `accept` 列
- 合法 `deny` 列
- 缺欄列
- 非 `tcp` 列
- `port` 非法列
- `src_ip` 或 `dst_ip` 不在各自 network segment 內的列
- 與其他列共享同一 `dst_ip+port+protocol` 但背景欄位不同的列

## 2. Run targeted tests first (TDD gate)

```bash
go test ./pkg/input ./pkg/task ./pkg/scanapp -count=1
```

## 3. Run full unit/integration regression

```bash
go test ./... -count=1
```

## 4. Run coverage gate

```bash
bash scripts/coverage_gate.sh
```

## 5. Run e2e gate (mandatory for this feature)

```bash
bash e2e/run_e2e.sh
```

## 6. Validate expected behavior

- parser 可識別大小寫/空白差異欄位名
- 10 欄缺一即列級拒絕
- `decision=accept|deny` 皆可進入可執行目標
- 掃描執行按 `dst_ip+port+protocol` 只執行一次
- 列級背景資訊可回溯到該次執行結果
- 解析摘要含 total/accepted/rejected/error buckets

## 7. Verification Evidence (2026-03-03)

- `go test ./... -count=1`: PASS
- `bash scripts/coverage_gate.sh`: PASS (`coverage gate passed: 86.3%`)
- `bash e2e/run_e2e.sh`: PASS (`e2e report generated at e2e/out`)

## 8. Success Criteria Measurement Samples

- **SC-003** (`invalid_rows.csv` correction cycle): 3/3 rows corrected within 10 minutes -> `100%`.
- **SC-004** (mixed dataset execution-key completion): 2 successful keys / 2 expected keys -> `100%`.
