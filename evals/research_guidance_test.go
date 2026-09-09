package evals_test

import "testing"

func TestJSONV2ExampleRejectsInvalidDocuments(t *testing.T) {
	code := exampleBlock(t, "skills/go-http/references/JSON-V2.md", "## One bounded request document")
	runExampleTest(t, `package example
import (json "encoding/json/v2"; "net/http"; "net/http/httptest"; "strings"; "testing")
func handle(w http.ResponseWriter, r *http.Request) {
`+code+`
w.WriteHeader(http.StatusNoContent)
}
func TestRequest(t *testing.T) {
 const valid = "{\"name\":\"ok\"}"
 for _, tt := range []struct { name, body string; want int }{
  {"object", valid, 204},
  {"empty object", "{}", 204},
  {"whitespace", valid + " \n\t", 204},
  {"exact limit", valid + strings.Repeat(" ", (1<<20)-len(valid)), 204},
  {"over limit", valid + strings.Repeat(" ", (1<<20)-len(valid)+1), 400},
  {"empty", "", 400},
  {"null", "null", 400},
  {"array", "[]", 400},
  {"scalar", "1", 400},
  {"truncated", "{", 400},
  {"unknown", "{\"other\":1}", 400},
  {"wrong case", "{\"Name\":\"ok\"}", 400},
  {"duplicate", "{\"name\":\"a\",\"name\":\"b\"}", 400},
  {"invalid UTF8", "{\"name\":\"\xff\"}", 400},
  {"second object", valid + " {}", 400},
  {"junk", valid + " junk", 400},
 } {
  t.Run(tt.name, func(t *testing.T) {
   w := httptest.NewRecorder()
   handle(w, httptest.NewRequest("POST", "/", strings.NewReader(tt.body)))
   if w.Code != tt.want { t.Errorf("handle(%q): status=%d, want %d", tt.name, w.Code, tt.want) }
  })
 }
}
`)
}

func TestAsTypeExampleMatchesBothCauses(t *testing.T) {
	code := exampleBlock(t, "skills/go-error-handling/SKILL.md", "### Matching multiple typed errors")
	runExampleTest(t, `package example
import ("errors"; "fmt"; "io/fs"; "os"; "testing")
func target(err error) string {
`+code+`
return "unmatched"
}
func TestTarget(t *testing.T) {
 for _, tt := range []struct { name string; err error; want string }{
  {"first", &fs.PathError{Path:"source"}, "source"},
  {"second", &os.LinkError{New:"destination"}, "destination"},
  {"wrapped second", fmt.Errorf("operation: %w", &os.LinkError{New:"destination"}), "destination"},
  {"unmatched", errors.New("other"), "unmatched"},
 } {
  t.Run(tt.name, func(t *testing.T) {
   if got := target(tt.err); got != tt.want { t.Errorf("target(%q) = %q, want %q", tt.name, got, tt.want) }
  })
 }
}
`)
}
