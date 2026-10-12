---
type: llm
weight: 3
---

PASS if the answer satisfies every item:

- Rejects InsecureSkipVerify even behind an environment check and points to RootCAs for a private CA
- Flags the explicit CurvePreferences as turning off the default hybrids (X25519MLKEM768 and the Go 1.26+/1.27 additions) and recommends deleting the field
- Removes cfg.Token from the wrapped error rather than relying on log filtering downstream
- Keeps %w so the caller can still match the underlying error
- Does not add a hand-written CipherSuites list or otherwise re-specify what the crypto/tls defaults already provide

FAIL if any item is missing or contradicted.
