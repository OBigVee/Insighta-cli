package display

import (
	"fmt"
	"strings"
)

// PrintTable renders a simple formatted table to stdout
func PrintTable(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Build separator
	sepParts := make([]string, len(widths))
	for i, w := range widths {
		sepParts[i] = strings.Repeat("─", w+2)
	}
	topSep := "┌" + strings.Join(sepParts, "┬") + "┐"
	midSep := "├" + strings.Join(sepParts, "┼") + "┤"
	botSep := "└" + strings.Join(sepParts, "┴") + "┘"

	// Print header
	fmt.Println(topSep)
	headerCells := make([]string, len(headers))
	for i, h := range headers {
		headerCells[i] = fmt.Sprintf(" %-*s ", widths[i], h)
	}
	fmt.Println("│" + strings.Join(headerCells, "│") + "│")
	fmt.Println(midSep)

	// Print rows
	for _, row := range rows {
		cells := make([]string, len(headers))
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			cells[i] = fmt.Sprintf(" %-*s ", widths[i], cell)
		}
		fmt.Println("│" + strings.Join(cells, "│") + "│")
	}
	fmt.Println(botSep)
}
