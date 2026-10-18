package delegate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateRoundTripAndLostFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "state.json")
	s, err := LoadState(path, 1)
	if err != nil || len(s.Pillars) != 0 {
		t.Fatal(s, err)
	}
	s.Entries("A")[testID] = &Entry{Cast: ptr(Yes), BlockHash: testHash, OwnerOverride: true}
	if err = SaveState(path, s); err != nil {
		t.Fatal(err)
	}
	restored, err := LoadState(path, 1)
	if err != nil || !restored.Entries("A")[testID].OwnerOverride || *restored.Entries("A")[testID].Cast != Yes {
		t.Fatal(restored, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("state permissions", err)
	}
	if _, err = LoadState(path, 2); err == nil {
		t.Fatal("cross-chain state accepted")
	}
	if err = os.WriteFile(path, []byte(`{"version":1`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = LoadState(path, 1); err == nil {
		t.Fatal("corrupt state accepted")
	}
}
func TestStateLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	unlock, err := LockState(path)
	if err != nil {
		t.Fatal(err)
	}
	if other, err := LockState(path); err == nil {
		other()
		t.Fatal("second lock succeeded")
	}
	unlock()
	unlock, err = LockState(path)
	if err != nil {
		t.Fatal(err)
	}
	unlock()
}
