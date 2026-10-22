package digest

import "testing"

func TestGoldenName(t *testing.T) {
	for in, want := range map[string]string{
		"":    "sha256-e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		"abc": "sha256-ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
	} {
		if got := Name([]byte(in)); got != want {
			t.Errorf("Name(%q) = %q, want %q", in, got, want)
		}
	}
}
