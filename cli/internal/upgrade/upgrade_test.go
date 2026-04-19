package upgrade

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFlavor_Clone(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(tmp, "cli", "bin", "ark")
	if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bin, []byte{}, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := DetectFlavor(bin); got != FlavorClone {
		t.Fatalf("got %v want FlavorClone", got)
	}
}

func TestDetectFlavor_Prebuilt(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "bin", "ark")
	if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bin, []byte{}, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := DetectFlavor(bin); got != FlavorPrebuilt {
		t.Fatalf("got %v want FlavorPrebuilt", got)
	}
}
