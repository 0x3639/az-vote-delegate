package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadNOMCTLProducer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"Producer":{"KeyFilePath":"producer","Index":3,"Password":"local-test-secret"}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := ReadProducer(path)
	if err != nil {
		t.Fatal(err)
	}
	if p.KeyFilePath != filepath.Join(dir, "wallet", "producer") || p.Index != 3 || p.Password != "local-test-secret" {
		t.Fatal("producer configuration not imported")
	}
}
func TestUnitQuoting(t *testing.T) {
	unit, err := Unit("/home/a user/bin/az-vote-delegate", "/home/a user/100%/config.json", true)
	if err != nil || !strings.Contains(unit, `"/home/a user/100%%/config.json"`) || !strings.Contains(unit, "WantedBy=default.target") {
		t.Fatal(unit, err)
	}
	if _, err = Unit("/binary", "/config\nExecStart=/bad", false); err == nil {
		t.Fatal("newline accepted")
	}
}
