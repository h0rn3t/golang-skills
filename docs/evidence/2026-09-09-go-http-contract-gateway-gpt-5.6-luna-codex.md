# `go-http` HTTP-contract workflow — GPT-5.6-Luna (medium) on Codex CLI

2026-09-09. **Переносить candidate в основной skill не следует:** текущий
`go-http` прошёл golden в 5/5 сессиях, а вариант с явным процессом
`contract → implement → verify` — в 4/5. Единственный провал candidate вернул
`null` вместо `[]` для nil-входа. Все десять реализаций собирались и правильно
возвращали 405 для HEAD.

## Что сравнивалось

- Runner: Codex CLI 0.153.4.
- Модель: `gpt-5.6-luna`, reasoning effort `medium`.
- Корпус: `implement`; fixture: `gateway`.
- Seed 1; по 5 повторов на плечо; `j=4`; `-keep`.
- Baseline: рабочее дерево на `f12c73b899727fbad5ab2c73c92149895b5da7a7`,
  plugin SHA-256 `f208ef8bcb0d8ce38f6564a0d74dd50c43a64aeb1acdf4f7cda3606ba7442d4b`.
- Reference: тот же commit с единственным изменением — 13 строк `Workflow` в
  `skills/go-http/SKILL.md`; plugin SHA-256
  `1462e1e43edd0cc811adc2725d574f2f516ac240916d753bd3bda03d753ca100`.
- Raw report SHA-256:
  `6e6f92fe37c0ec7361e63c00465775c54e2b7fce2347d6e3f81cb488efe6cef3`.

Candidate требует до правки вывести HTTP-контракт из методов, путей, статусов,
wire shape и лимитов, затем реализовать минимальный код и проверить применимые
результаты через `httptest`. Остальной текст plugin не менялся. Точный diff
сохранён рядом с отчётом.

## Результаты

Метрики размера ниже включают все пять сессий каждого плеча, в том числе
candidate-сессию, провалившую golden.

| Плечо | Build | Golden | HEAD → 405 | nil/empty/filtered → `[]` | Средние строки | Средние новые функции | Средние branches |
|---|---:|---:|---:|---:|---:|---:|---:|
| Текущий `go-http` | 5/5 | 5/5 | 5/5 | 5/5 | +96.0 | +3.0 | +13.4 |
| HTTP-contract workflow | 5/5 | 4/5 | 5/5 | 4/5 | +93.2 | +2.6 | +12.6 |

При `n=5` различия −2.8 строки и −0.4 функции слишком малы относительно
разброса, чтобы считать их эффектом workflow. Correctness не улучшилась:
baseline уже насыщен на HEAD, а candidate добавил один дефект wire shape.

Во всех сессиях загрузился `go-http`. Провалившаяся candidate-сессия также
загрузила `go-data-structures` и `go-defensive`, а в финальном сообщении заявила
«JSON responses with empty arrays» и выполненный HTTP-contract test. В
сохранённом результате модельских `_test.go` нет (`model_tests=skipped`), а
независимый golden обнаружил ответ `null`. Это снова показывает, что загрузка
правила и декларация проверки не доказывают его исполнения.

## Решение

Не применять candidate и не расширять этот эксперимент: на Luna исходный skill
уже дал 5/5 по проверяемым контрактам. Следующий обоснованный шаг для Sonnet 5 —
не переносить туда результат Luna, а отдельно проверить feedback/repair loop,
в котором модель получает конкретный провал contract test до завершения
сессии.

## Артефакты и воспроизведение

- [Raw JSON](2026-09-09-go-http-contract-gateway-gpt-5.6-luna-codex.json).
- [Точный candidate patch](2026-09-09-go-http-contract-gateway-gpt-5.6-luna-codex.patch).
- [Оба текста skill, golden и все 10 generated sources](2026-09-09-go-http-contract-gateway-gpt-5.6-luna-codex.sources.json).

Из `evals`, где `CANDIDATE_ROOT` — архив commit выше с применённым patch:

```sh
go run ./cmd/abrun -corpus implement -tasks gateway -runner codex \
  -model gpt-5.6-luna -effort medium -arms baseline,reference \
  -reference-root "$CANDIDATE_ROOT" -n 5 -j 4 -seed 1 -timeout 10m \
  -keep -out ../docs/evidence/2026-09-09-go-http-contract-gateway-gpt-5.6-luna-codex.json
```

