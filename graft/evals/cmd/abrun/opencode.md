# evals/cmd/abrun/opencode.go

- opencodeHomes · function · L37-L63 — func opencodeHomes(arms []arm) (func(), error)
- opencodeAuth · function · L68-L81 — func opencodeAuth() ([]byte, error)
- writeOpencodeHome · function · L86-L107 — func writeOpencodeHome(home, armDir string, auth []byte) error
- checkOpencodeSkills · function · L114-L132 — func checkOpencodeSkills(home, armName, armDir string) error
- opencodeSession · function · L141-L148 — func opencodeSession(o options, home, work, prompt string) ([]byte, error)
- opencodeCmd · function · L159-L191 — func opencodeCmd(timeout time.Duration, dir, home string, args ...string) ([]byte, error)
- opencodeEnv · function · L203-L214 — func opencodeEnv(home, dir string) []string
- opencodeEvent · struct · L219-L229 — opencodeEvent
- parseOpencodeStream · function · L235-L268 — func parseOpencodeStream(out []byte) (skills []string, final string, cost float64)
