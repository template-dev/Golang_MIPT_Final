package ledger

import "strings"

func normalizeCategory(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}
