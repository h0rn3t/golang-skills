package evals_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativePluginManifests(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	var claude struct{ Name, Version string }
	readJSON(t, filepath.Join(root, ".claude-plugin/plugin.json"), &claude)
	for _, host := range []string{"codex", "cursor"} {
		t.Run(host, func(t *testing.T) {
			var manifest struct {
				Name, Version, Skills, Rules string
				Hooks                        json.RawMessage
				Agents                       []string
			}
			readJSON(t, filepath.Join(root, "."+host+"-plugin/plugin.json"), &manifest)
			if manifest.Name != claude.Name || manifest.Version != claude.Version {
				t.Errorf("%s identity = %s@%s, want %s@%s", host, manifest.Name, manifest.Version, claude.Name, claude.Version)
			}
			if manifest.Skills != "./skills/" {
				t.Errorf("%s skills = %q, want ./skills/", host, manifest.Skills)
			}
			if host == "codex" {
				var path string
				if err := json.Unmarshal(manifest.Hooks, &path); err != nil {
					t.Fatal(err)
				}
				if path != "./hooks/codex.json" {
					t.Fatalf("Codex hooks = %q, want its own hook configuration", path)
				}
				if _, err := os.Stat(filepath.Join(root, path)); err != nil {
					t.Fatal(err)
				}
				var config struct {
					Hooks map[string][]struct {
						Matcher string
						Hooks   []struct {
							Command string
							Timeout int
						}
					}
				}
				readJSON(t, filepath.Join(root, path), &config)
				for _, event := range []string{"SessionStart", "UserPromptSubmit", "SubagentStart", "PostToolUse"} {
					if len(config.Hooks[event]) == 0 {
						t.Errorf("Codex hook config missing %s", event)
					}
				}
				post := config.Hooks["PostToolUse"]
				if len(config.Hooks["PreToolUse"]) != 0 || len(post) != 1 || post[0].Matcher != "^apply_patch$" || len(post[0].Hooks) != 1 || !strings.HasSuffix(post[0].Hooks[0].Command, " codex-edit") || post[0].Hooks[0].Timeout < 150 {
					t.Fatalf("Codex must register its patch adapter without the Claude skill-load gate: %+v", config)
				}
				return
			}
			if manifest.Rules != "./rules/" || manifest.Agents == nil || len(manifest.Agents) != 0 {
				t.Errorf("Cursor must load rules and explicitly disable Claude agent discovery: %+v", manifest)
			}
			var hooks struct {
				Version int
				Hooks   map[string]json.RawMessage
			}
			if err := json.Unmarshal(manifest.Hooks, &hooks); err != nil || hooks.Version != 1 || hooks.Hooks == nil || len(hooks.Hooks) != 0 {
				t.Errorf("Cursor must explicitly replace Claude hook discovery with an empty native config: %s (%v)", manifest.Hooks, err)
			}
		})
	}
	t.Run("Codex marketplace", func(t *testing.T) {
		var marketplace struct {
			Name    string
			Plugins []struct {
				Name     string
				Source   struct{ Source, Path string }
				Policy   struct{ Installation, Authentication string }
				Category string
			}
		}
		readJSON(t, filepath.Join(root, ".agents/plugins/marketplace.json"), &marketplace)
		if marketplace.Name != claude.Name || len(marketplace.Plugins) != 1 {
			t.Fatalf("Codex marketplace = %+v, want one plugin", marketplace)
		}
		p := marketplace.Plugins[0]
		if p.Name != claude.Name || p.Source.Source != "local" || p.Source.Path != "./" || p.Policy.Installation != "AVAILABLE" || p.Policy.Authentication != "ON_INSTALL" || p.Category == "" {
			t.Errorf("invalid Codex marketplace entry: %+v", p)
		}
	})
	t.Run("Cursor marketplace", func(t *testing.T) {
		var marketplace struct {
			Name    string
			Plugins []struct{ Name, Source string }
		}
		readJSON(t, filepath.Join(root, ".cursor-plugin/marketplace.json"), &marketplace)
		if marketplace.Name != claude.Name || len(marketplace.Plugins) != 1 || marketplace.Plugins[0].Name != claude.Name || marketplace.Plugins[0].Source != "./" {
			t.Errorf("invalid Cursor marketplace: %+v", marketplace)
		}
	})
	t.Run("Cursor rule", func(t *testing.T) {
		rule := readFile(t, filepath.Join(root, "rules/go-skills.mdc"))
		for _, required := range []string{"alwaysApply: true", "go-code", "go-code-refactor", "go-code-review", "go-style-core", "SKILL.md", "not Go work"} {
			if !strings.Contains(rule, required) {
				t.Errorf("Cursor rule missing %q", required)
			}
		}
		if strings.Contains(rule, "~/.codex/") || strings.Contains(rule, "<skills>") {
			t.Error("Cursor plugin rule must resolve installed skills without a machine-specific directory or placeholder")
		}
	})
}
