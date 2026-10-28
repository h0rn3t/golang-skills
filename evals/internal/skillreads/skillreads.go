// Package skillreads counts the Read tool calls a Claude Code session makes on
// skill files, so a run can show which references its sessions opened and
// which none of them did.
package skillreads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// Key names a skill file by its skill directory, "go-http/SKILL.md" or
// "go-http/references/JSON-V2.md", wherever the tree was staged. It returns ""
// for any other path.
func Key(p string) string {
	dir, file := path.Split(filepath.ToSlash(p))
	dir = strings.TrimSuffix(dir, "/")
	if file == "SKILL.md" {
		if skill := path.Base(dir); strings.HasPrefix(skill, "go-") {
			return skill + "/SKILL.md"
		}
		return ""
	}
	if path.Ext(file) != ".md" || path.Base(dir) != "references" {
		return ""
	}
	if skill := path.Base(path.Dir(dir)); strings.HasPrefix(skill, "go-") {
		return skill + "/references/" + file
	}
	return ""
}

// Count returns the Read calls per skill file in a Claude Code transcript,
// keyed as Key names them, or nil when the session read none. out is
// stream-json, one message per line, or the message array that
// --output-format json --verbose prints. Each content block arrives in its own
// assistant line, so a tool_use block is counted once by its id; one without
// an id is keyed by its message id and content, as the routing evidence is.
func Count(out []byte) map[string]int {
	var msgs []json.RawMessage
	if json.Unmarshal(out, &msgs) != nil {
		msgs = nil
		for line := range bytes.Lines(out) {
			msgs = append(msgs, line)
		}
	}
	var counts map[string]int
	seen := map[string]bool{}
	for _, raw := range msgs {
		var msg struct {
			Type    string `json:"type"`
			Message struct {
				ID      string `json:"id"`
				Content []struct {
					Type  string `json:"type"`
					ID    string `json:"id"`
					Name  string `json:"name"`
					Input struct {
						FilePath string `json:"file_path"`
					} `json:"input"`
				} `json:"content"`
			} `json:"message"`
		}
		if json.Unmarshal(raw, &msg) != nil || msg.Type != "assistant" {
			continue
		}
		for _, c := range msg.Message.Content {
			if c.Type != "tool_use" || c.Name != "Read" {
				continue
			}
			key := Key(c.Input.FilePath)
			if key == "" {
				continue
			}
			id := c.ID
			if id == "" {
				id = "message:" + msg.Message.ID + ":" + c.Input.FilePath
			}
			if seen[id] {
				continue
			}
			seen[id] = true
			if counts == nil {
				counts = map[string]int{}
			}
			counts[key]++
		}
	}
	return counts
}

// Files lists every skill file under root that Key names, sorted, so a total
// can carry a file no session read as zero.
func Files(root string) []string {
	var files []string
	for _, pattern := range []string{"skills/go-*/SKILL.md", "skills/go-*/references/*.md"} {
		// fs.Glob ignores I/O errors and fails only on a malformed pattern.
		matches, _ := fs.Glob(os.DirFS(root), pattern)
		for _, m := range matches {
			files = append(files, Key(m))
		}
	}
	slices.Sort(files)
	return files
}

// Summary renders a total as one line: the Read calls, and how many of the
// SKILL.md and reference files it lists any session read.
func Summary(total map[string]int) string {
	var calls, skills, skillsRead, refs, refsRead int
	for key, n := range total {
		calls += n
		if strings.HasSuffix(key, "/SKILL.md") {
			skills++
			if n > 0 {
				skillsRead++
			}
			continue
		}
		refs++
		if n > 0 {
			refsRead++
		}
	}
	return fmt.Sprintf("skill file reads: %d call(s); SKILL.md read %d/%d, references read %d/%d",
		calls, skillsRead, skills, refsRead, refs)
}
