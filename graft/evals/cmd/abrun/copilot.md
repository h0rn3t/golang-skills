# evals/cmd/abrun/copilot.go

- copilotHomes · function · L49-L71 — func copilotHomes(arms []arm) (func(), error)
- writeCopilotHome · function · L80-L85 — func writeCopilotHome(home, armDir string) error
- checkCopilotSkills · function · L95-L113 — func checkCopilotSkills(home, armName, armDir string) error
- copilotSession · function · L126-L153 — func copilotSession(o options, a arm, work, prompt string) ([]byte, error)
- copilotCmd · function · L157-L160 — func copilotCmd(timeout time.Duration, dir, home string, args ...string) ([]byte, error)
- copilotOutput · function · L171-L203 — func copilotOutput(timeout time.Duration, dir, home string, args ...string) ([]byte, string, error)
- copilotEnv · function · L211-L217 — func copilotEnv(home, dir string) []string
- copilotEvent · struct · L222-L229 — copilotEvent
- parseCopilotStream · function · L239-L270 — func parseCopilotStream(out []byte) (skills []string, final string, cost float64)
