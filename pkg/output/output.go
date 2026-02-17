package output

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
)

var JSONMode bool

func PrintJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func PrintSuccess(msg string) {
	if JSONMode {
		PrintJSON(map[string]interface{}{"status": "success", "message": msg})
		return
	}
	fmt.Printf("✓ %s\n", msg)
}

func PrintError(msg string) {
	if JSONMode {
		PrintJSON(map[string]interface{}{"status": "error", "message": msg})
		return
	}
	fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
}

func PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	// Print headers
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, h)
	}
	fmt.Fprintln(w)
	// Print rows
	for _, row := range rows {
		for i, col := range row {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, col)
		}
		fmt.Fprintln(w)
	}
	w.Flush()
}
