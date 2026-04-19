// Package prompt provides tiny, stdlib-only terminal prompt helpers used by
// the `ark init` wizard. These are intentionally boring: Confirm accepts
// y/yes/n/no; Select asks for a single integer; MultiSelect accepts
// comma-separated integers. No TUI libraries.
package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// Confirm asks a yes/no question, using defaultYes when the user hits
// return with no input. Returns the boolean answer, or an error on read
// failure.
func Confirm(stdin io.Reader, stdout io.Writer, msg string, defaultYes bool) (bool, error) {
	suffix := " [y/N]: "
	if defaultYes {
		suffix = " [Y/n]: "
	}
	scanner := bufio.NewScanner(stdin)
	for {
		fmt.Fprint(stdout, msg+suffix)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return false, err
			}
			return defaultYes, nil
		}
		switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
		case "":
			return defaultYes, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(stdout, "please answer y or n")
		}
	}
}

// Select prints numbered options and reads a single integer selection.
// defaultIndex is used when the user hits return with no input. Indices
// are 1-based in the UI but 0-based in the returned int.
func Select(stdin io.Reader, stdout io.Writer, msg string, options []string, defaultIndex int) (int, error) {
	if len(options) == 0 {
		return 0, fmt.Errorf("prompt.Select: no options")
	}
	if defaultIndex < 0 || defaultIndex >= len(options) {
		defaultIndex = 0
	}
	scanner := bufio.NewScanner(stdin)
	for {
		fmt.Fprintln(stdout, msg)
		for i, opt := range options {
			marker := " "
			if i == defaultIndex {
				marker = "*"
			}
			fmt.Fprintf(stdout, "  %s %d) %s\n", marker, i+1, opt)
		}
		fmt.Fprintf(stdout, "select [default %d]: ", defaultIndex+1)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return 0, err
			}
			return defaultIndex, nil
		}
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			return defaultIndex, nil
		}
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > len(options) {
			fmt.Fprintf(stdout, "please enter a number between 1 and %d\n", len(options))
			continue
		}
		return n - 1, nil
	}
}

// MultiSelect reads a comma-separated list of 1-based integers.
// Empty input returns defaultSelected (already 0-based). Duplicates in
// user input are de-duplicated in the returned slice.
func MultiSelect(stdin io.Reader, stdout io.Writer, msg string, options []string, defaultSelected []int) ([]int, error) {
	if len(options) == 0 {
		return nil, fmt.Errorf("prompt.MultiSelect: no options")
	}
	defaultSet := map[int]bool{}
	for _, i := range defaultSelected {
		if i >= 0 && i < len(options) {
			defaultSet[i] = true
		}
	}
	scanner := bufio.NewScanner(stdin)
	for {
		fmt.Fprintln(stdout, msg)
		for i, opt := range options {
			marker := " "
			if defaultSet[i] {
				marker = "*"
			}
			fmt.Fprintf(stdout, "  %s %d) %s\n", marker, i+1, opt)
		}
		fmt.Fprint(stdout, "select (comma-separated, blank for default): ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, err
			}
			return sortedKeys(defaultSet), nil
		}
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			return sortedKeys(defaultSet), nil
		}
		parts := strings.Split(raw, ",")
		seen := map[int]bool{}
		result := make([]int, 0, len(parts))
		ok := true
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			n, err := strconv.Atoi(p)
			if err != nil || n < 1 || n > len(options) {
				fmt.Fprintf(stdout, "invalid entry %q; use numbers 1..%d\n", p, len(options))
				ok = false
				break
			}
			idx := n - 1
			if !seen[idx] {
				seen[idx] = true
				result = append(result, idx)
			}
		}
		if !ok {
			continue
		}
		return result, nil
	}
}

// IsInteractive reports whether the given file is attached to a terminal.
func IsInteractive(f *os.File) bool {
	if f == nil {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func sortedKeys(m map[int]bool) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// preserve ascending order for stable output
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j-1] > keys[j]; j-- {
			keys[j-1], keys[j] = keys[j], keys[j-1]
		}
	}
	return keys
}
