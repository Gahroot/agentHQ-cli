package output

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = old
	return buf.String()
}

func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	f()
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = old
	return buf.String()
}

func TestPrintJSON(t *testing.T) {
	data := map[string]string{"key": "value", "foo": "bar"}
	got := captureStdout(func() {
		PrintJSON(data)
	})

	// Verify it's valid JSON
	var parsed map[string]string
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("PrintJSON output is not valid JSON: %v\nGot: %s", err, got)
	}

	// Verify indentation (pretty-printed)
	if !strings.Contains(got, "  ") {
		t.Errorf("expected indented JSON output, got: %s", got)
	}

	// Verify values
	if parsed["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", parsed["key"])
	}
	if parsed["foo"] != "bar" {
		t.Errorf("expected foo=bar, got foo=%s", parsed["foo"])
	}
}

func TestPrintSuccess_TextMode(t *testing.T) {
	JSONMode = false
	defer func() { JSONMode = false }()

	got := captureStdout(func() {
		PrintSuccess("operation completed")
	})

	if !strings.HasPrefix(got, "✓") {
		t.Errorf("expected text mode output to start with ✓, got: %s", got)
	}
	if !strings.Contains(got, "operation completed") {
		t.Errorf("expected output to contain message, got: %s", got)
	}
}

func TestPrintSuccess_JSONMode(t *testing.T) {
	JSONMode = true
	defer func() { JSONMode = false }()

	got := captureStdout(func() {
		PrintSuccess("operation completed")
	})

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\nOutput: %s", err, got)
	}

	if parsed["status"] != "success" {
		t.Errorf("expected status=success, got status=%v", parsed["status"])
	}
	if parsed["message"] != "operation completed" {
		t.Errorf("expected message='operation completed', got message=%v", parsed["message"])
	}
}

func TestPrintError_TextMode(t *testing.T) {
	JSONMode = false
	defer func() { JSONMode = false }()

	got := captureStderr(func() {
		PrintError("something went wrong")
	})

	if !strings.HasPrefix(got, "✗") {
		t.Errorf("expected text mode error to start with ✗, got: %s", got)
	}
	if !strings.Contains(got, "something went wrong") {
		t.Errorf("expected output to contain error message, got: %s", got)
	}
}

func TestPrintError_JSONMode(t *testing.T) {
	JSONMode = true
	defer func() { JSONMode = false }()

	// In JSON mode, PrintError calls PrintJSON which writes to os.Stdout
	got := captureStdout(func() {
		PrintError("something went wrong")
	})

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("expected valid JSON, got error: %v\nOutput: %s", err, got)
	}

	if parsed["status"] != "error" {
		t.Errorf("expected status=error, got status=%v", parsed["status"])
	}
	if parsed["message"] != "something went wrong" {
		t.Errorf("expected message='something went wrong', got message=%v", parsed["message"])
	}
}

func TestPrintTable(t *testing.T) {
	headers := []string{"ID", "NAME", "STATUS"}
	rows := [][]string{
		{"1", "Alice", "active"},
		{"2", "Bob", "inactive"},
	}

	got := captureStdout(func() {
		PrintTable(headers, rows)
	})

	// Verify headers are present
	if !strings.Contains(got, "ID") {
		t.Errorf("expected output to contain header 'ID', got: %s", got)
	}
	if !strings.Contains(got, "NAME") {
		t.Errorf("expected output to contain header 'NAME', got: %s", got)
	}
	if !strings.Contains(got, "STATUS") {
		t.Errorf("expected output to contain header 'STATUS', got: %s", got)
	}

	// Verify row data is present
	if !strings.Contains(got, "Alice") {
		t.Errorf("expected output to contain 'Alice', got: %s", got)
	}
	if !strings.Contains(got, "Bob") {
		t.Errorf("expected output to contain 'Bob', got: %s", got)
	}
	if !strings.Contains(got, "active") {
		t.Errorf("expected output to contain 'active', got: %s", got)
	}
	if !strings.Contains(got, "inactive") {
		t.Errorf("expected output to contain 'inactive', got: %s", got)
	}

	// Verify multiple lines (header + 2 rows)
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines (1 header + 2 rows), got %d lines:\n%s", len(lines), got)
	}
}
