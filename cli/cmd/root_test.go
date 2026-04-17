package cmd

import (
	"errors"
	"testing"

	"github.com/gh-xj/agent-repo-kit/cli/internal/appctx"
)

func TestUsageErrorCode(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil error", err: nil, want: 0},
		// Cobra-generated usage errors — must produce ExitUsage.
		{name: "unknown flag", err: errors.New(`unknown flag: --bogus`), want: appctx.ExitUsage},
		{name: "unknown command", err: errors.New(`unknown command "wat" for "ark"`), want: appctx.ExitUsage},
		{name: "invalid arg", err: errors.New(`invalid argument "x" for "--count" flag: strconv.ParseInt: parsing "x"`), want: appctx.ExitUsage},
		{name: "accepts exact args", err: errors.New(`accepts 2 arg(s), received 0`), want: appctx.ExitUsage},
		{name: "requires at least", err: errors.New(`requires at least 1 arg(s), only received 0`), want: appctx.ExitUsage},
		// Application errors — MUST NOT produce ExitUsage. These are the
		// false-positives the prior substring matcher caught.
		{name: "app error with 'requires' token", err: errors.New("skill requires justification"), want: 0},
		{name: "app error with 'accepts' token", err: errors.New("config accepts only bool values"), want: 0},
		{name: "generic app error", err: errors.New("disk full"), want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := usageErrorCode(tc.err)
			if got != tc.want {
				t.Fatalf("usageErrorCode(%q) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}
}
