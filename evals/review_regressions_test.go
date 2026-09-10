package evals_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSortedKeysModernizationPreservesWireFormat(t *testing.T) {
	code := exampleBlock(t, "skills/go-code-refactor/references/MODERNIZATION.md", "### `slices` and `maps` over hand-rolled loops")
	before, rest, ok := strings.Cut(code, "// after — nil result permitted")
	if !ok {
		t.Fatal("приклад не містить варіанта, що допускає nil")
	}
	sorted, appended, ok := strings.Cut(rest, "// after — original non-nil empty result is observable")
	if !ok {
		t.Fatal("приклад не містить варіанта, що зберігає non-nil")
	}
	runExampleTest(t, `package example
import ("encoding/json"; "maps"; "slices"; "sort"; "testing")
func before(m map[string]int) []string {
`+before+`
return keys
}
func sortedKeys(m map[string]int) []string {
`+sorted+`
return keys
}
func appendedKeys(m map[string]int) []string {
`+appended+`
return keys
}
func TestWireFormat(t *testing.T) {
 for _, m := range []map[string]int{nil, {}, {"b":2,"a":1}} {
  want, err := json.Marshal(before(m)); if err != nil {t.Fatal(err)}
  got, err := json.Marshal(appendedKeys(m)); if err != nil {t.Fatal(err)}
  if string(got) != string(want) {t.Errorf("appendedKeys(%v): JSON=%s, want %s", m, got, want)}
  if !slices.Equal(sortedKeys(m), before(m)) {t.Errorf("sortedKeys(%v)=%v, want the same order as %v", m, sortedKeys(m), before(m))}
 }
 // The documented divergence: the short form drops the non-nil empty result.
 got, err := json.Marshal(sortedKeys(nil)); if err != nil {t.Fatal(err)}
 if string(got) != "null" {t.Errorf("sortedKeys(nil): JSON=%s, want null", got)}
}
`)
}

func TestEncryptionExampleRejectsInvalidKeys(t *testing.T) {
	code := exampleBlock(t, "skills/go-security/references/SECRETS-AND-CRYPTO.md", "## Encryption at rest")
	runExampleTest(t, `package example
import ("bytes"; "crypto/aes"; "crypto/cipher"; "testing")
func encrypt(key, plaintext, additionalData []byte) ([]byte, error) {
`+code+`
return ct, nil
}
func TestEncryption(t *testing.T) {
 for _, size := range []int{0, 15, 16, 24, 31, 32, 33} {
  key := make([]byte, size)
  ct, err := encrypt(key, []byte("message"), []byte("scope"))
  valid := size == 16 || size == 24 || size == 32
  if (err == nil) != valid {t.Fatalf("encrypt(key length %d): error=%v, want valid=%t", size, err, valid)}
  if !valid {continue}
  block, err := aes.NewCipher(key); if err != nil {t.Fatal(err)}
  aead, err := cipher.NewGCMWithRandomNonce(block); if err != nil {t.Fatal(err)}
  got, err := aead.Open(nil, nil, ct, []byte("scope"))
  if err != nil || !bytes.Equal(got, []byte("message")) {t.Fatalf("decrypt: got=%q error=%v", got, err)}
  if _, err := aead.Open(nil, nil, ct, []byte("other scope")); err == nil {t.Fatal("прийнято неправильні додаткові дані")}
 }
}
`)
}

func TestLeakModeDoesNotCertifyOrdinaryTests(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"go.mod":       "module leakexample\n\ngo 1.27\n",
		"leak_test.go": "package leakexample\nimport \"testing\"\nfunc TestLeak(t *testing.T) { go func(){ select{} }() }\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.CommandContext(t.Context(), "bash", filepath.Join(repoRoot(t), "skills/go-code-refactor/scripts/verify-refactor.sh"), "--json", "leaks", "./...")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("leaks повернув успіх без перевірки профілю: %s", out)
	}
	// 3 is "tests passed, leaks unverified"; 2 stays the usage/environment error.
	if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != 3 {
		t.Fatalf("leaks: error=%v output=%s, очікував код 3", err, out)
	}
	var result struct {
		Passed       bool `json:"passed"`
		TestsPassed  bool `json:"tests_passed"`
		LeaksChecked bool `json:"leaks_checked"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("leaks: вивід не є JSON (%v): %s", err, out)
	}
	if result.Passed || !result.TestsPassed || result.LeaksChecked {
		t.Fatalf("leaks: невірний стан перевірки: %s", out)
	}
}
