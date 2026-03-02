package testkit

import "testing"

func TestParseBatchOutputName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    BatchOutputName
		wantErr bool
	}{
		{name: "main", in: "/tmp/scan_results-20260302T123456Z.csv", want: BatchOutputName{Prefix: "scan_results", Timestamp: "20260302T123456Z", Sequence: 0}},
		{name: "open-with-seq", in: "opened_results-20260302T123456Z-2.csv", want: BatchOutputName{Prefix: "opened_results", Timestamp: "20260302T123456Z", Sequence: 2}},
		{name: "bad", in: "scan_results.csv", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBatchOutputName(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected parsed value: got=%+v want=%+v", got, tt.want)
			}
		})
	}
}

func TestAssertBatchPair(t *testing.T) {
	if err := AssertBatchPair("scan_results-20260302T123456Z-1.csv", "opened_results-20260302T123456Z-1.csv"); err != nil {
		t.Fatalf("expected valid pair, got err=%v", err)
	}
	if err := AssertBatchPair("scan_results-20260302T123456Z-1.csv", "opened_results-20260302T123457Z-1.csv"); err == nil {
		t.Fatalf("expected mismatch error")
	}
}
