package watcher

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var sampleRecords = []transferRecord{
	{
		Block:  12_000_001,
		TxHash: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		From:   "0x1111111111111111111111111111111111111111",
		To:     "0x2222222222222222222222222222222222222222",
		Amount: 1500.5,
		Symbol: "USDC",
	},
	{
		Block:  12_000_002,
		TxHash: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		From:   "0x3333333333333333333333333333333333333333",
		To:     "0x4444444444444444444444444444444444444444",
		Amount: 0.000001,
		Symbol: "USDT",
	},
}

// --- CSV ---

func TestCSVWriter_HeaderAndRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")

	w, err := newCSVWriter(path)
	if err != nil {
		t.Fatalf("newCSVWriter: %v", err)
	}
	for _, r := range sampleRecords {
		if err := w.write(r); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	if err := w.close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("csv read: %v", err)
	}

	if len(rows) != 3 { // header + 2 data rows
		t.Fatalf("expected 3 rows (header + 2 data), got %d", len(rows))
	}

	// header
	wantHeader := []string{"block", "tx_hash", "from", "to", "amount", "token"}
	for i, col := range wantHeader {
		if rows[0][i] != col {
			t.Errorf("header[%d]: got %q, want %q", i, rows[0][i], col)
		}
	}

	// first data row
	r0 := rows[1]
	checks := []struct{ col, got, want string }{
		{"block", r0[0], "12000001"},
		{"tx_hash", r0[1], sampleRecords[0].TxHash},
		{"from", r0[2], sampleRecords[0].From},
		{"to", r0[3], sampleRecords[0].To},
		{"amount", r0[4], "1500.500000"},
		{"token", r0[5], "USDC"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("row1 %s: got %q, want %q", c.col, c.got, c.want)
		}
	}

	// second data row — spot-check amount precision for a tiny value
	if rows[2][4] != "0.000001" {
		t.Errorf("row2 amount: got %q, want %q", rows[2][4], "0.000001")
	}
}

func TestCSVWriter_DataFlushedOnClose(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")

	w, err := newCSVWriter(path)
	if err != nil {
		t.Fatalf("newCSVWriter: %v", err)
	}
	if err := w.write(sampleRecords[0]); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read before close — csv.Writer buffers, so data may not be on disk yet.
	raw, _ := os.ReadFile(path)
	linesBefore := strings.Count(string(raw), "\n")

	if err := w.close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	raw, _ = os.ReadFile(path)
	linesAfter := strings.Count(string(raw), "\n")

	if linesAfter <= linesBefore {
		t.Errorf("expected more lines after close (flush); before=%d after=%d", linesBefore, linesAfter)
	}
}

// --- Markdown ---

func TestMarkdownWriter_HeaderAndRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.md")

	w, err := newMarkdownWriter(path)
	if err != nil {
		t.Fatalf("newMarkdownWriter: %v", err)
	}
	for _, r := range sampleRecords {
		if err := w.write(r); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	if err := w.close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")

	if len(lines) != 4 { // header + separator + 2 data rows
		t.Fatalf("expected 4 lines, got %d:\n%s", len(lines), string(raw))
	}

	// header
	if !strings.HasPrefix(lines[0], "| Block |") {
		t.Errorf("line 0 (header): got %q", lines[0])
	}

	// separator — all cells should contain only dashes and colons
	for _, cell := range strings.Split(lines[1], "|") {
		trimmed := strings.TrimSpace(cell)
		if trimmed == "" {
			continue
		}
		for _, ch := range trimmed {
			if ch != '-' && ch != ':' {
				t.Errorf("separator line contains unexpected char %q in cell %q", ch, trimmed)
			}
		}
	}

	// first data row contains expected values
	row1 := lines[2]
	for _, want := range []string{"12000001", sampleRecords[0].TxHash, "1500.50", "USDC"} {
		if !strings.Contains(row1, want) {
			t.Errorf("row1 missing %q:\n  %s", want, row1)
		}
	}

	// second data row
	row2 := lines[3]
	if !strings.Contains(row2, "USDT") {
		t.Errorf("row2 missing token symbol:\n  %s", row2)
	}
}

func TestMarkdownWriter_RowFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.md")

	w, err := newMarkdownWriter(path)
	if err != nil {
		t.Fatalf("newMarkdownWriter: %v", err)
	}
	rec := sampleRecords[0]
	if err := w.write(rec); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	raw, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	dataLine := lines[2]

	// must start and end with pipe
	if !strings.HasPrefix(dataLine, "|") || !strings.HasSuffix(dataLine, "|") {
		t.Errorf("data line must be wrapped in pipes: %q", dataLine)
	}

	// tx hash and addresses must be backtick-quoted
	if !strings.Contains(dataLine, "`"+rec.TxHash+"`") {
		t.Errorf("tx hash not backtick-quoted in: %q", dataLine)
	}
	if !strings.Contains(dataLine, "`"+rec.From+"`") {
		t.Errorf("from address not backtick-quoted in: %q", dataLine)
	}
}
