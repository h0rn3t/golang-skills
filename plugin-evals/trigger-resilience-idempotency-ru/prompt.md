---
description: 'Duplicate effects and multi-replica delivery semantics'
tags: [trigger]
runs: 3
max_turns: 10
allowed_tools: [Read, Glob, Grep, Skill]
---

Наш Go consumer повторно создаёт отправку после таймаута. Нужна идемпотентная обработка повторных событий между репликами.
