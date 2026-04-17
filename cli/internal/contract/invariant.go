package contract

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type invariantEntry struct {
	Fields map[string]string
	Index  int
}

func runInvariantChecks(root string, cfg InvariantContractConfig) []CheckResult {
	name := "invariants:" + cfg.File
	fullPath := filepath.Join(root, cfg.File)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		if cfg.Required {
			return []CheckResult{{Name: name, Passed: false, Detail: "cannot read " + cfg.File}}
		}
		return []CheckResult{{Name: name, Passed: true, Detail: "optional file missing"}}
	}

	entries := parseInvariantEntries(string(content))
	if len(entries) == 0 {
		return []CheckResult{{Name: name, Passed: false, Detail: "no invariant entries found"}}
	}

	results := make([]CheckResult, 0)
	enforceableCount := 0
	for _, entry := range entries {
		if normalizeFieldValue(entry.Fields["enforceability"]) != "enforceable" {
			continue
		}
		enforceableCount++

		entryID := entry.Fields["id"]
		if isEmptyFieldValue(entryID) {
			entryID = fmt.Sprintf("entry-%d", entry.Index)
		}

		missingRequired := missingFields(entry.Fields, cfg.RequiredFields)
		if len(missingRequired) > 0 {
			results = append(results, CheckResult{
				Name:   "invariants:" + entryID + ":required-fields",
				Passed: false,
				Detail: "missing field(s): " + strings.Join(missingRequired, ","),
			})
			continue
		}

		status := normalizeFieldValue(entry.Fields["status"])
		if status != "accepted" && status != "enforced" {
			results = append(results, CheckResult{
				Name:   "invariants:" + entryID + ":status",
				Passed: false,
				Detail: "invalid status: " + entry.Fields["status"],
			})
			continue
		}

		anyExceptionField := false
		for _, field := range cfg.ExceptionFields {
			if !isEmptyFieldValue(entry.Fields[field]) {
				anyExceptionField = true
				break
			}
		}

		if anyExceptionField {
			missingException := missingFields(entry.Fields, cfg.ExceptionFields)
			if len(missingException) > 0 {
				results = append(results, CheckResult{
					Name:   "invariants:" + entryID + ":exception-fields",
					Passed: false,
					Detail: "missing field(s): " + strings.Join(missingException, ","),
				})
				continue
			}

			expiryRaw := strings.TrimSpace(entry.Fields["exception_expires"])
			expiryAt, dateOnly, parseErr := parseExceptionExpiry(expiryRaw)
			if parseErr != nil {
				results = append(results, CheckResult{
					Name:   "invariants:" + entryID + ":exception-expiry",
					Passed: false,
					Detail: parseErr.Error(),
				})
				continue
			}

			if exceptionExpired(expiryAt, dateOnly) {
				results = append(results, CheckResult{
					Name:   "invariants:" + entryID + ":exception-expiry",
					Passed: false,
					Detail: "expired on " + expiryRaw,
				})
				continue
			}
		}

		results = append(results, CheckResult{Name: "invariants:" + entryID + ":contract", Passed: true})
	}

	if enforceableCount == 0 {
		results = append(results, CheckResult{Name: "invariants:enforceable", Passed: true, Detail: "no enforceable invariants"})
	}

	entryFailures := CountFailures(results)
	summary := CheckResult{Name: name, Passed: entryFailures == 0}
	if entryFailures > 0 {
		summary.Detail = fmt.Sprintf("%d check(s) failed", entryFailures)
	}
	return append([]CheckResult{summary}, results...)
}

func parseInvariantEntries(content string) []invariantEntry {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lines := make([]string, 0)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var entries []invariantEntry
	itemIndent := -1
	lastKey := ""
	var current invariantEntry
	active := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := leadingSpaces(line)
		if strings.HasPrefix(trimmed, "- ") {
			if itemIndent == -1 {
				itemIndent = indent
			}
			if indent == itemIndent {
				if active {
					entries = append(entries, current)
				}
				active = true
				lastKey = ""
				current = invariantEntry{Fields: map[string]string{}, Index: len(entries) + 1}

				key, value, ok := splitYAMLField(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
				if ok {
					current.Fields[key] = value
					lastKey = key
				}
				continue
			}

			if active && lastKey != "" && isEmptyFieldValue(current.Fields[lastKey]) {
				current.Fields[lastKey] = "(list)"
			}
			continue
		}

		if !active {
			continue
		}

		key, value, ok := splitYAMLField(trimmed)
		if !ok {
			continue
		}
		current.Fields[key] = value
		lastKey = key
	}

	if active {
		entries = append(entries, current)
	}

	return entries
}

func splitYAMLField(line string) (string, string, bool) {
	colon := strings.Index(line, ":")
	if colon <= 0 {
		return "", "", false
	}

	key := strings.TrimSpace(line[:colon])
	if key == "" || strings.Contains(key, " ") {
		return "", "", false
	}

	value := strings.TrimSpace(line[colon+1:])
	value = strings.Trim(value, `"'`)
	return key, value, true
}

func leadingSpaces(line string) int {
	spaces := 0
	for _, ch := range line {
		if ch != ' ' {
			break
		}
		spaces++
	}
	return spaces
}

func normalizeFieldValue(value string) string {
	normalized := strings.TrimSpace(value)
	normalized = strings.Trim(normalized, `"'`)
	return strings.ToLower(normalized)
}

func isEmptyFieldValue(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return true
	}
	if trimmed == "[]" || trimmed == "{}" || strings.EqualFold(trimmed, "null") {
		return true
	}
	return false
}

func missingFields(fields map[string]string, required []string) []string {
	missing := make([]string, 0)
	for _, name := range required {
		if isEmptyFieldValue(fields[name]) {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
}

func parseExceptionExpiry(raw string) (time.Time, bool, error) {
	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			return parsed, layout == "2006-01-02", nil
		}
	}
	return time.Time{}, false, fmt.Errorf("invalid exception_expires: %q", raw)
}

func exceptionExpired(expiryAt time.Time, dateOnly bool) bool {
	nowUTC := time.Now().UTC()
	if dateOnly {
		todayUTC := time.Date(nowUTC.Year(), nowUTC.Month(), nowUTC.Day(), 0, 0, 0, 0, time.UTC)
		return expiryAt.Before(todayUTC)
	}
	return expiryAt.UTC().Before(nowUTC)
}
