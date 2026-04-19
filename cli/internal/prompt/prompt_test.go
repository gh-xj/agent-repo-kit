package prompt

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestConfirm_EmptyUsesDefault(t *testing.T) {
	var out bytes.Buffer
	got, err := Confirm(strings.NewReader("\n"), &out, "go?", true)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !got {
		t.Fatalf("expected default true, got false")
	}
}

func TestConfirm_YesNo(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"y\n", true},
		{"yes\n", true},
		{"YES\n", true},
		{"n\n", false},
		{"no\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			var out bytes.Buffer
			got, err := Confirm(strings.NewReader(tc.in), &out, "?", false)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestConfirm_ReprompsOnGarbage(t *testing.T) {
	var out bytes.Buffer
	got, err := Confirm(strings.NewReader("maybe\ny\n"), &out, "?", false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !got {
		t.Fatalf("expected true after re-prompt")
	}
	if !strings.Contains(out.String(), "please answer") {
		t.Fatalf("expected re-prompt message, got %q", out.String())
	}
}

func TestSelect_Default(t *testing.T) {
	var out bytes.Buffer
	got, err := Select(strings.NewReader("\n"), &out, "pick", []string{"a", "b", "c"}, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != 1 {
		t.Fatalf("got %d want 1", got)
	}
}

func TestSelect_ValidInput(t *testing.T) {
	var out bytes.Buffer
	got, err := Select(strings.NewReader("3\n"), &out, "pick", []string{"a", "b", "c"}, 0)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != 2 {
		t.Fatalf("got %d want 2", got)
	}
}

func TestSelect_RejectsOutOfRange(t *testing.T) {
	var out bytes.Buffer
	got, err := Select(strings.NewReader("9\n2\n"), &out, "pick", []string{"a", "b", "c"}, 0)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != 1 {
		t.Fatalf("got %d want 1", got)
	}
}

func TestMultiSelect_Default(t *testing.T) {
	var out bytes.Buffer
	got, err := MultiSelect(strings.NewReader("\n"), &out, "pick", []string{"a", "b", "c"}, []int{0, 2})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !reflect.DeepEqual(got, []int{0, 2}) {
		t.Fatalf("got %v want [0 2]", got)
	}
}

func TestMultiSelect_ParsesCSV(t *testing.T) {
	var out bytes.Buffer
	got, err := MultiSelect(strings.NewReader("1,3,3\n"), &out, "pick", []string{"a", "b", "c"}, nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !reflect.DeepEqual(got, []int{0, 2}) {
		t.Fatalf("got %v want [0 2]", got)
	}
}

func TestMultiSelect_RejectsGarbage(t *testing.T) {
	var out bytes.Buffer
	got, err := MultiSelect(strings.NewReader("x\n2\n"), &out, "pick", []string{"a", "b", "c"}, nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("got %v want [1]", got)
	}
}
