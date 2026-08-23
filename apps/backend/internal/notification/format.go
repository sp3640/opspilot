package notification

import (
	"fmt"
	"sort"
	"strings"
)

// sortedFieldLines renders msg.Fields as "Key: Value" lines in a
// deterministic (alphabetical-by-key) order, since Go map iteration order
// is random and every provider needs stable output for both users and tests.
func sortedFieldLines(fields map[string]string) []string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("%s: %s", k, fields[k]))
	}
	return lines
}

// plainTextBody renders a Message as a simple multi-line plain-text block,
// shared by the Email provider and used as the generic fallback body.
func plainTextBody(msg Message) string {
	var b strings.Builder
	b.WriteString(msg.Body)
	for _, line := range sortedFieldLines(msg.Fields) {
		b.WriteString("\n")
		b.WriteString(line)
	}
	return b.String()
}
