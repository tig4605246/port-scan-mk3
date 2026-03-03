package scanapp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xuxiping/port-scan-mk3/pkg/config"
	"github.com/xuxiping/port-scan-mk3/pkg/input"
	"github.com/xuxiping/port-scan-mk3/pkg/speedctrl"
	"github.com/xuxiping/port-scan-mk3/pkg/task"
)

func TestIndexToRuntimeTarget_WhenInputsInvalid_ReturnsErrors(t *testing.T) {
	targets := []scanTarget{{ip: "10.0.0.1"}}
	ports := []int{80}

	if _, _, err := indexToRuntimeTarget(nil, ports, 0); err == nil {
		t.Fatal("expected empty targets error")
	}
	if _, _, err := indexToRuntimeTarget(targets, nil, 0); err == nil {
		t.Fatal("expected empty ports error")
	}
	if _, _, err := indexToRuntimeTarget(targets, ports, -1); err == nil {
		t.Fatal("expected negative index error")
	}
	if _, _, err := indexToRuntimeTarget(targets, ports, 2); err == nil {
		t.Fatal("expected out of range error")
	}

	gotTarget, gotPort, err := indexToRuntimeTarget(targets, ports, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotTarget.ip != "10.0.0.1" || gotPort != 80 {
		t.Fatalf("unexpected mapping: %+v port=%d", gotTarget, gotPort)
	}
}

func TestBuildCIDRGroups_WhenInputsVary_ReturnsErrorsAndSortedTargets(t *testing.T) {
	if _, err := buildCIDRGroups([]input.CIDRRecord{{IPRaw: "10.0.0.1"}}); err == nil {
		t.Fatal("expected missing ip_cidr error")
	}

	_, ipNet, _ := net.ParseCIDR("10.0.0.0/24")
	if _, err := buildCIDRGroups([]input.CIDRRecord{{CIDR: "10.0.0.0/24", Net: ipNet}}); err != nil {
		t.Fatalf("expected fallback selector from net, got err=%v", err)
	}

	if _, err := buildCIDRGroups([]input.CIDRRecord{{CIDR: "10.0.0.0/24", IPRaw: "bad-selector"}}); err == nil {
		t.Fatal("expected expand selector error")
	}

	rows := []input.CIDRRecord{
		{CIDR: "10.0.0.0/24", Selector: mustSelectorNet(t, "10.0.0.3/32"), FabName: "fab", CIDRName: "name"},
		{CIDR: "10.0.0.0/24", Selector: mustSelectorNet(t, "10.0.0.1/32"), FabName: "fab", CIDRName: "name"},
	}
	groups, err := buildCIDRGroups(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := groups["10.0.0.0/24"].targets
	if len(got) != 2 || got[0].ip != "10.0.0.1" || got[1].ip != "10.0.0.3" {
		t.Fatalf("unexpected sorted targets: %#v", got)
	}
}

func TestBuildRuntime_WhenTotalCountMismatch_ReturnsError(t *testing.T) {
	rows := []input.CIDRRecord{
		{CIDR: "10.0.0.0/24", Selector: mustSelectorNet(t, "10.0.0.1/32")},
	}
	chunks := []task.Chunk{{
		CIDR:       "10.0.0.0/24",
		Ports:      []string{"80/tcp"},
		TotalCount: 2, // expected should be 1
	}}
	_, err := buildRuntime(chunks, rows, nil, config.Config{BucketRate: 1, BucketCapacity: 1})
	if err == nil {
		t.Fatal("expected total_count mismatch error")
	}
}

func TestBuildRichGroups_WhenDuplicateExecutionKey_PreservesMergedContext(t *testing.T) {
	rows := []input.CIDRRecord{
		{
			IsRich:            true,
			IsValid:           true,
			ExecutionKey:      "127.0.0.1:8080/tcp",
			DstIP:             "127.0.0.1",
			DstNetworkSegment: "127.0.0.0/24",
			Port:              8080,
			FabName:           "10.0.0.10",
			CIDRName:          "web",
			ServiceLabel:      "web",
			Decision:          "accept",
			PolicyID:          "P-1",
			Reason:            "allow",
			SrcIP:             "10.0.0.10",
			SrcNetworkSegment: "10.0.0.0/24",
		},
		{
			IsRich:            true,
			IsValid:           true,
			ExecutionKey:      "127.0.0.1:8080/tcp",
			DstIP:             "127.0.0.1",
			DstNetworkSegment: "127.0.0.0/24",
			Port:              8080,
			FabName:           "10.0.0.11",
			CIDRName:          "web",
			ServiceLabel:      "web",
			Decision:          "deny",
			PolicyID:          "P-2",
			Reason:            "audit",
			SrcIP:             "10.0.0.11",
			SrcNetworkSegment: "10.0.0.0/24",
		},
	}
	groups, err := buildRichGroups(rows)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	got := groups["127.0.0.1:8080/tcp"]
	if got.port != 8080 {
		t.Fatalf("unexpected group port: %d", got.port)
	}
	if len(got.targets) != 1 {
		t.Fatalf("expected single runtime target, got %d", len(got.targets))
	}
	if got.targets[0].policyID != "P-1|P-2" {
		t.Fatalf("unexpected merged policy id: %s", got.targets[0].policyID)
	}
	if got.targets[0].decision != "accept|deny" {
		t.Fatalf("unexpected merged decision: %s", got.targets[0].decision)
	}
}

func TestLoadOrBuildChunks_WhenRichRecordsProvided_BuildsExecutionKeyChunks(t *testing.T) {
	rows := []input.CIDRRecord{
		{IsRich: true, IsValid: true, ExecutionKey: "127.0.0.1:8080/tcp", Port: 8080, CIDRName: "web", DstIP: "127.0.0.1", DstNetworkSegment: "127.0.0.0/24"},
		{IsRich: true, IsValid: true, ExecutionKey: "127.0.0.1:8080/tcp", Port: 8080, CIDRName: "web", DstIP: "127.0.0.1", DstNetworkSegment: "127.0.0.0/24"},
		{IsRich: true, IsValid: true, ExecutionKey: "127.0.0.1:1/tcp", Port: 1, CIDRName: "web", DstIP: "127.0.0.1", DstNetworkSegment: "127.0.0.0/24"},
	}
	chunks, err := loadOrBuildChunks(config.Config{}, rows, nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(chunks) != 2 {
		t.Fatalf("expected 2 dedup chunks, got %d", len(chunks))
	}
	if chunks[0].TotalCount != 1 || chunks[1].TotalCount != 1 {
		t.Fatalf("expected each rich chunk total_count=1, got %#v", chunks)
	}
}

func TestBuildRichChunks_WhenNoUsableRows_ReturnsError(t *testing.T) {
	_, err := buildRichChunks([]input.CIDRRecord{{IsRich: true, IsValid: false}})
	if err == nil {
		t.Fatal("expected no usable row error")
	}
}

func TestDefaultString_WhenPrimaryEmpty_UsesFallback(t *testing.T) {
	if got := defaultString("", "x"); got != "x" {
		t.Fatalf("unexpected value: %s", got)
	}
	if got := defaultString(" y ", "x"); got != " y " {
		t.Fatalf("unexpected primary-preserved value: %s", got)
	}
}

func TestReadCIDRFileAndReadPortFile_WhenFileMissing_ReturnsError(t *testing.T) {
	if _, err := readCIDRFile("/not-exist", "ip", "ip_cidr"); err == nil {
		t.Fatal("expected read cidr file error")
	}
	if _, err := readPortFile("/not-exist"); err == nil {
		t.Fatal("expected read port file error")
	}
}

func TestLoadOrBuildChunks_WhenResumePathProvided_LoadsStateFromFile(t *testing.T) {
	dir := t.TempDir()
	resume := filepath.Join(dir, "resume.json")
	if err := os.WriteFile(resume, []byte(`[{"cidr":"10.0.0.0/24","next_index":1,"total_count":1}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	chunks, err := loadOrBuildChunks(config.Config{Resume: resume}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 1 || chunks[0].CIDR != "10.0.0.0/24" {
		t.Fatalf("unexpected chunks: %#v", chunks)
	}
}

func TestResumePath_WhenMultipleSourcesProvided_UsesPriorityOrder(t *testing.T) {
	if got := resumePath(config.Config{Resume: "cfg.json"}, RunOptions{ResumeStatePath: "opt.json"}); got != "opt.json" {
		t.Fatalf("unexpected resume path: %s", got)
	}
	if got := resumePath(config.Config{Resume: "cfg.json"}, RunOptions{}); got != "cfg.json" {
		t.Fatalf("unexpected resume path: %s", got)
	}
	if got := resumePath(config.Config{Output: "/tmp/scan_results.csv"}, RunOptions{}); got != "/tmp/"+defaultResumeStateFile {
		t.Fatalf("unexpected default resume path: %s", got)
	}
}

func TestEnsureFDLimit_WhenWorkersExceedLimit_ReturnsError(t *testing.T) {
	err := ensureFDLimit(1_000_000_000)
	if err == nil {
		t.Fatal("expected fd limit error for huge workers")
	}
}

func TestFetchPressure_WhenFieldMissingOrTypeUnsupported_ReturnsError(t *testing.T) {
	missingFieldAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, `{"x":1}`)
	}))
	defer missingFieldAPI.Close()
	if _, err := fetchPressure(&http.Client{Timeout: time.Second}, missingFieldAPI.URL); err == nil || !strings.Contains(err.Error(), "pressure field missing") {
		t.Fatalf("expected missing pressure field error, got %v", err)
	}

	unsupportedTypeAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintln(w, `{"pressure":true}`)
	}))
	defer unsupportedTypeAPI.Close()
	if _, err := fetchPressure(&http.Client{Timeout: time.Second}, unsupportedTypeAPI.URL); err == nil || !strings.Contains(err.Error(), "unsupported pressure field type") {
		t.Fatalf("expected unsupported type error, got %v", err)
	}
}

func TestPollPressureAPI_WhenFirstTwoRequestsFail_DoesNotReturnFatalError(t *testing.T) {
	steps := []int{500, 500, 200, 200}
	var mu sync.Mutex
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		status := steps[0]
		if len(steps) > 1 {
			steps = steps[1:]
		}
		mu.Unlock()
		if status >= 400 {
			http.Error(w, "fail", status)
			return
		}
		_, _ = fmt.Fprintln(w, `{"pressure":10}`)
	}))
	defer api.Close()

	cfg := config.Config{
		PressureAPI:      api.URL,
		PressureInterval: 5 * time.Millisecond,
	}
	ctrl := speedctrl.NewController()
	logOut := &lockedBuffer{}
	logger := newLogger("info", false, logOut)
	errCh := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go pollPressureAPI(ctx, cfg, RunOptions{PressureLimit: 90, PressureHTTP: &http.Client{Timeout: time.Second}}, ctrl, logger, errCh)

	time.Sleep(60 * time.Millisecond)
	cancel()
	time.Sleep(10 * time.Millisecond)

	select {
	case err := <-errCh:
		t.Fatalf("expected no fatal error before 3rd failure, got %v", err)
	default:
	}
	if ctrl.IsPaused() {
		t.Fatal("expected not paused at low pressure")
	}
	logs := logOut.String()
	if !strings.Contains(logs, "(1/3)") || !strings.Contains(logs, "(2/3)") {
		t.Fatalf("expected first two failure logs, got: %s", logs)
	}
}

func mustSelectorNet(t *testing.T, raw string) *net.IPNet {
	t.Helper()
	_, n, err := net.ParseCIDR(raw)
	if err != nil {
		t.Fatalf("parse cidr failed: %v", err)
	}
	return n
}
