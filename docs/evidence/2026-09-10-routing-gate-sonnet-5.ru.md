# Sonnet 5: одиночные прогоны маршрутизации go-code после 1.6.0

2026-09-10. Вопрос: загружает ли модель после `go-code` скилы, которые роутер
требует до первой правки, и срабатывает ли новый хук `go-code-routing.sh`.

**Итог по четырём сессиям:** на дереве 1.6.0 (`7e78e3b`) обе сессии загрузили
`go-style-core` и владельца до первой правки `.go`; на референсе v1.5.0
`go-style-core` не загрузила ни одна, только `go-code` и владельца. Хук в этих
четырёх сессиях ни разу не блокировал, и это был дефект: в плагине Skill
получает имя вида `golang-skills:go-code`, а хук принимал только `go-*`, поэтому
не узнавал, что `go-code` загружен. Исправлено в этом же коммите и проверено
отдельной живой сессией, где хук заблокировал первую правку один раз и модель
после этого загрузила `go-style-core` и повторила правку.

По n=1 на арм и фикстуру нельзя утверждать, что порядок загрузки объясняется
текстом, а не разбросом. Прямое наблюдение годится только для одного: новая
формулировка совместима с тем, чтобы Sonnet 5 её выполнял, а старая в этих же
условиях выполнена не была, что совпадает с
[экспериментом 2026-09-08](2026-09-08-sonnet-routing-execution-pocock.md).

## Проведение

Claude Code 2.1.263, `-model claude-sonnet-5`, effort не задан, `--restricted`,
инструменты `Skill,Read,Glob,Grep,Edit,Write`, плагин арма через `--plugin-dir`.
Корпус implement, фикстуры `gateway` (владелец `go-http`) и `catalog`
(владелец `go-error-handling`), обычный prompt корпуса без просьбы перечислять
скилы. Baseline — рабочее дерево на `7e78e3b` (release 1.6.0); reference —
worktree на теге `v1.5.0`. Сессии запускались из вложенного Claude Code с
снятой переменной `CLAUDECODE`.

```sh
go run ./cmd/abrun -corpus implement -tasks gateway,catalog -runner claude \
  -model claude-sonnet-5 -arms baseline,reference -reference-root "$REF_1_5_0" \
  -n 1 -j 2 -seed 1 -timeout 8m -keep -out routing-sonnet5.json
```

## Результаты

Порядок вызовов до первой правки `.go` взят из `trace.jsonl` каждой сессии.

| Арм | Фикстура | Загружено до первой правки | Golden | Δlines | $ |
|---|---|---|---:|---:|---:|
| 1.6.0 | gateway | go-code, go-style-core, go-http | pass | +94 | 0.24 |
| 1.6.0 | catalog | go-code, go-style-core, go-error-handling | pass | +37 | 0.24 |
| v1.5.0 | gateway | go-code, go-http | **fail** (HEAD → 200, want 405) | +78 | 0.17 |
| v1.5.0 | catalog | go-code, go-error-handling | pass | +36 | 0.17 |

Столбец `$` — среднее по арму из отчёта harness. В сессии 1.6.0/gateway после
первой правки дополнительно загружен `go-linting`. Блокировок хука: 0 из 4.
Провал golden на v1.5.0/gateway — правило `GET` также матчит `HEAD` из
`go-http`; при n=1 это наблюдение, не сравнение армов.

## Дефект хука и его проверка

В `.eval-plugin-*/hooks/hooks.json` хук был зарегистрирован, и модель в финальном
сообщении ссылалась на результаты gofmt/vet-хука, то есть хуки плагина под
`--restricted` работают. Но состояние `loaded` не появилось ни в одной сессии:
`record()` отбрасывал имя `golang-skills:go-code`. Исправление — брать часть
после последнего `:`; `TestRoutingGate` теперь подаёт имя с префиксом.

Живая проверка после исправления, те же флаги, что у harness, модуль из одного
`main.go`, prompt: загрузить `golang-skills:go-code` и сразу править `main.go`,
не загружая других скилов, пока результат инструмента не заставит.

```
Skill:golang-skills:go-code
Edit:main.go                       ← PreToolUse: exit 2
  go-code routing gate: this session loaded go-code but not: go-style-core
Skill:golang-skills:go-style-core
Edit:main.go                       ← успешно
```

Состояние записано в `$CLAUDE_PLUGIN_DATA`, который Claude Code задаёт сам:
`~/.claude/plugins/data/golang-skills-inline/routing/<session_id>/` с файлами
`loaded` (`go-code go-style-core`) и `reminded` (`go-style-core`). Стоимость
сессии $0.11.

## Границы

- n=1 на арм и фикстуру; никаких средних и p-значений. Разброс между двумя
  сессиями одного арма в эксперименте 2026-09-08 уже был 1/2 по `go-style-core`.
- Хук проверен на плагине; установка через `npx skills` хуков не имеет, там
  действует только текст.
- Живая сессия проверяет механику блокировки, не качество кода.
- Codex не запускался.

## Артефакты

- [JSON-отчёт harness](2026-09-10-routing-gate-sonnet-5.json) — четыре сессии,
  пути рабочих деревьев, стоимость, вывод golden.
- [Пять transcript'ов](2026-09-10-routing-gate-sonnet-5.traces.tar.gz):
  `baseline-*`, `reference-*` и `live-hook-plumbing.jsonl`.

Суммарная стоимость: $0.94.
