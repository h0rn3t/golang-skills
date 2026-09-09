# evals/cmd/abrun/codex.go

- codexHomes · function · L45-L71 — func codexHomes(arms []arm) (func(), error)
- codexAuth · function · L76-L89 — func codexAuth() ([]byte, error)
- codexHomeDir · function · L92-L92 — func codexHomeDir(home string) string
- writeCodexHome · function · L97-L111 — func writeCodexHome(home, armDir string, auth []byte) error
- checkCodexSkills · function · L120-L126 — func checkCodexSkills(home, armName, armDir string) error
- codexSession · function · L138-L156 — func codexSession(o options, home, work, prompt string) ([]byte, error)
- codexCmd · function · L161-L193 — func codexCmd(timeout time.Duration, dir, home string, args ...string) ([]byte, error)
- codexEnv · function · L203-L214 — func codexEnv(home, dir string) []string
- codexEvent · struct · L217-L226 — codexEvent
- parseCodexStream · function · L240-L270 — func parseCodexStream(out []byte) (skills []string, final string, cost float64)
- codexCommands · function · L275-L292 — func codexCommands(out []byte) int
- skillsInPaths · function · L295-L303 — func skillsInPaths(text string) []string
