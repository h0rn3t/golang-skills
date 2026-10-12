---
description: 'InsecureSkipVerify rejected regardless of environment, the pinned CurvePreferences identified as silently disabling the default post-quantum key exchanges, and the token removed from the error text.'
tags: [quality]
runs: 3
max_turns: 12
allowed_tools: [Read, Glob, Grep, Skill]
---

Review this Go 1.27 TLS setup for a service calling an internal API. It builds &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: os.Getenv("ENV") == "staging", CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256}} and returns fmt.Errorf("dial %s with token %s: %w", addr, cfg.Token, err) when the handshake fails. Say what you would change and why. Do not run commands.
