---
type: llm
weight: 3
---

PASS if the answer satisfies every item:

- Rejects InsecureSkipVerify even behind an environment check and points to RootCAs for a private CA
- Flags the explicit CurvePreferences as turning off the default hybrids, naming SecP256r1MLKEM768 or SecP384r1MLKEM1024 (Go 1.26) and not only X25519MLKEM768, and recommends deleting the field; reordering the list is not a fix (crypto/tls ignores the order)
- Removes cfg.Token from the wrapped error rather than relying on log filtering downstream
- Keeps %w so the caller can still match the underlying error
- Does not add a hand-written CipherSuites list or otherwise re-specify what the crypto/tls defaults already provide

FAIL if any item is missing or contradicted.
