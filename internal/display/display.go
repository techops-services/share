package display

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/techops-services/share/internal/client"
)

// HumanSize formats a byte count into a human-readable string.
func HumanSize(bytes int) string {
	const (
		kb = 1024
		mb = kb * 1024
	)
	switch {
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// HumanTime formats an ISO timestamp into a short, human-readable string.
// Falls back to the raw string if parsing fails.
func HumanTime(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}

	now := time.Now().UTC()
	diff := now.Sub(t)

	switch {
	case diff < 0:
		// Future date (expiry) -- show relative
		absDiff := -diff
		return "in " + relativeDuration(absDiff)
	case diff < time.Minute:
		return "just now"
	default:
		return relativeDuration(diff) + " ago"
	}
}

// relativeDuration returns a concise human-readable duration.
func relativeDuration(d time.Duration) string {
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		days := int(d.Hours()) / 24
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
}

// PrintShareTable writes a formatted table of shares to the given writer.
func PrintShareTable(w io.Writer, shares []client.Share) {
	if len(shares) == 0 {
		fmt.Fprintln(w, "No shares found.")
		return
	}

	// Column headers
	headers := []string{"ID", "URL", "Created", "Expires", "Views", "Size"}
	widths := []int{10, 42, 14, 14, 6, 10}

	// Print header
	printRow(w, headers, widths)
	printSeparator(w, widths)

	// Print rows
	for _, s := range shares {
		row := []string{
			s.ID,
			s.URL,
			HumanTime(s.CreatedAt),
			HumanTime(s.ExpiresAt),
			fmt.Sprintf("%d", s.Views),
			HumanSize(s.Size),
		}
		printRow(w, row, widths)
	}
}

// PrintVerboseUpload writes detailed upload information.
func PrintVerboseUpload(w io.Writer, resp *client.UploadResponse, rl *client.RateLimitInfo) {
	fmt.Fprintf(w, "  ID:       %s\n", resp.ID)
	fmt.Fprintf(w, "  URL:      %s\n", resp.URL)
	fmt.Fprintf(w, "  Size:     %s\n", HumanSize(resp.Size))
	fmt.Fprintf(w, "  Expires:  %s\n", HumanTime(resp.ExpiresAt))
	if rl != nil {
		fmt.Fprintf(w, "  Rate:     %d/%d remaining\n", rl.Remaining, rl.Limit)
	}
}

// printRow writes a single row with fixed-width columns.
func printRow(w io.Writer, cols []string, widths []int) {
	for i, col := range cols {
		width := widths[i]
		if len(col) > width {
			col = col[:width]
		}
		if i < len(cols)-1 {
			fmt.Fprintf(w, "%-*s  ", width, col)
		} else {
			fmt.Fprintf(w, "%-*s", width, col)
		}
	}
	fmt.Fprintln(w)
}

// printSeparator writes a dashed separator line matching column widths.
func printSeparator(w io.Writer, widths []int) {
	parts := make([]string, len(widths))
	for i, width := range widths {
		parts[i] = strings.Repeat("-", width)
	}
	fmt.Fprintln(w, strings.Join(parts, "  "))
}
