# Аудит Go skills для Claude Opus 5 и GPT 5.6

2026-09-09. Checkout: `1c9c7ef`, Go 1.27.1, darwin/arm64. Ниже сохранён первоначальный аудит инструкций и проверок до правок. Результат последующих исправлений — в разделе [Применённые исправления](#применённые-исправления). Новый модельный A/B не проводился.

Основной резерв качества — устранить противоречия между владельцами правил, исправить опасные примеры и измерять результат выполнения задач. Повторное усиление требований «меньше кода» и массовое сокращение описаний без эксперимента сейчас не обоснованы.

**Объём проверки.** Полностью прочитаны все 24 `SKILL.md`; проведена инвентаризация всех ресурсов и выборочная углублённая проверка справочников, шаблонов, конфигурации lint и реализации eval runner. Это не полный повторный аудит каждого Go-сниппета во всех справочниках.

| Объект | Объём |
|---|---:|
| Основные навыки | 24 |
| Слова в `SKILL.md`, включая frontmatter и примеры | 28 927 |
| Markdown-файлы во всём `skills/` | 93 |
| Слова во всех этих Markdown-файлах | 80 836 |
| Скрипты / Go-файлы / YAML | 10 / 6 / 1 |
| Trigger-сценарии | 105: 64 train, 41 validation |
| Quality-сценарии | 59: 21 train, 38 validation |

Слова посчитаны через whitespace split, это не токены модели. Весь каталог не загружается автоматически. Но примерный набор для HTTP-реализации с тестами — `go-code`, `go-style-core`, `go-http`, `go-error-handling`, `go-testing`, `go-linting` — уже содержит 7 280 слов до чтения references. Оптимизировать нужно фактически загруженный набор, а не только размер отдельных файлов.

**Приоритетные находки.** P1 — исправить до следующего цикла оптимизации промптов; P2 — уточнить и проверить отдельной группой изменений. Приоритет относится к пакету инструкций; он не означает наличие подтверждённого инцидента в пользовательском сервисе.

1. **P1: неправильная проверка конца JSON в security-reference.** В [INJECTION.md](../skills/go-security/references/INJECTION.md#decoding-untrusted-structures), строка 180, предлагается `dec.More()` после значения. Воспроизведение: для `{}]` первый `Decode` успешен, `More()` возвращает false, но второй `Decode` обнаруживает ошибку. Модель может принять повреждённый запрос и выполнить side effect. В [go-http](../skills/go-http/SKILL.md#handler-shape) правильная проверка EOF уже есть. Исправление: один владелец single-document правила, ссылка из security и исполняемые негативные случаи, включая лишний закрывающий разделитель. `More` предназначен для элементов текущего массива/объекта. [Документация Go](https://pkg.go.dev/encoding/json#Decoder.More).

2. **P1: таблица упрощений подменяет глубокое копирование поверхностным.** В [OVER-ENGINEERING.md](../skills/go-code-refactor/references/OVER-ENGINEERING.md#reach-for-what-go-ships), строка 129, `Deep-copy helpers` заменяются на `slices.Clone`/`maps.Clone`. В [BOUNDARY-COPYING.md](../skills/go-defensive/references/BOUNDARY-COPYING.md#clone-is-shallow) верно написано обратное: вложенные ссылки сохраняются. Проба `[]*int` подтверждает изменение оригинала через элемент клона. Исправление: таблица должна различать shallow copy и требуемую глубину владения; удалять deep-copy helper можно только после проверки типов элементов и контракта. Для performance-исключения нельзя разрешать отказ от необходимого владения только потому, что копия дорогая.

3. **P1: concurrency-пример противоречит ограничению fan-out.** [go-concurrency](../skills/go-concurrency/SKILL.md#goroutine-lifetimes), строки 38–47: после запрета неограниченного fan-out показано `for item := range queue { wg.Go(...) }`. `WaitGroup` ожидает завершения, но не ограничивает число задач. Проба с 64 элементами одновременно блокирует все 64 обработчика. Исправление: либо явно ограниченная небольшая группа для примера join, либо bounded admission до запуска goroutine. Дополнительно routing обещает errgroup/pipelines в [ADVANCED-PATTERNS.md](../skills/go-concurrency/references/ADVANCED-PATTERNS.md), но reference не раскрывает errgroup. Нужен короткий выбор `WaitGroup` / `errgroup.WithContext`, с различием join, sibling cancellation и лимита параллелизма.

4. **P1: review-инструкции конфликтуют с read-only scope и общим gate.** [go-code-review](../skills/go-code-review/SKILL.md#review-procedure), строки 28–35 и 210–216, требует `./...`, затем «Fix everything», а также переносит все недоказанные findings в отчёт. [go-style-core](../skills/go-style-core/SKILL.md#house-style-wins) сохраняет review-only режим; [go-linting](../skills/go-linting/SKILL.md#verification-gate) выбирает область и переиспользует проверки. Исправление: review сообщает о дефектах; исправления выполняются в режиме fix; область и команды берутся из одного gate. Искать проблемы всех серьёзностей полезно, но неподтверждённую гипотезу надо отличать от обязательной правки. Не следует запрещать статическое доказательство: [review-template](../skills/go-code-review/assets/review-template.md), строки 13–14, ошибочно сводит `verified` к выполненному запуску, а любое чтение — к `plausible`.

5. **P1: eval runner наказывает разрешённый результат.** [go-code-refactor](../skills/go-code-refactor/SKILL.md#concision-gate) разрешает корректный empty diff. Однако `resultStatus` в `evals/cmd/abrun/main.go:1571` возвращает `ERR` при `!Edited`, а `summarizeArm` в строках 1633–1700 исключает такой результат из valid-средних. Это искажает сравнение моделей, которые умеют остановиться без ненужной правки. Исправление: независимо учитывать выполнение задания, correctness, изменение файлов, соблюдение LOC-политики и ошибки harness. No-op допустим для завершённого refactor-аудита, но не для оставленного implementation stub. В [сохранённом повторе](evidence/2026-09-09-refactor-repeat-luna-sonnet-medium.md) проблема уже наблюдалась; она остаётся в текущем коде.

6. **P2: противоречивые рекомендации по контексту и точке исправления.** [go-context](../skills/go-context/SKILL.md#where-to-put-application-data), строка 86, относит deadlines/cancellation к context values. Это механизмы `Deadline`/`Done` и производных контекстов, не `WithValue`. В [PATTERNS.md](../skills/go-context/references/PATTERNS.md#when-to-use-contextbackground), строки 73–75, предлагается принимать неиспользуемый `ctx` ради будущего, хотя остальной пакет отвергает спекулятивные API. В [OVER-ENGINEERING.md](../skills/go-code-refactor/references/OVER-ENGINEERING.md#the-restraint-ladder), строки 74–79, любой fix направляется в общий callee, а [go-troubleshooting](../skills/go-troubleshooting/SKILL.md#fix-and-regression-proof) правильно различает ошибку callee и нарушение контракта одним caller. Исправление: сохранить правила их владельцев без повторного универсального рецепта.

7. **P2: частные предпочтения записаны как универсальные ограничения.** [go-defensive](../skills/go-defensive/SKILL.md#defensive-checklist-priority) требует zero=unset и копий на входе/выходе; исключения находятся глубже. [IOTA.md](../skills/go-style-core/references/IOTA.md#string-representation), строка 84, требует `String()` для каждого enum. [go-performance](../skills/go-performance/SKILL.md#pass-values), строка 95, допускает pointer ради возможного будущего роста struct. [go-logging](../skills/go-logging/SKILL.md#production-observability-checklist), строка 171, превращает dashboard/alerts в критерий готовности любой service feature. Исправление: выражать нужный контракт и условие применения прямо рядом с правилом. Сохранять обязательные требования проекта, но не добавлять API, копии и инфраструктуру только по наличию enum, slice или service.

8. **P2: справочники и короткие таблицы отстали от Go 1.27 и соседних правил.** В [INJECTION.md](../skills/go-security/references/INJECTION.md#sql), строки 27–29, `LIMIT` смешан с идентификаторами, хотя параметризовать значение лимита можно; `go-database` сам показывает `LIMIT $2`. [PostgreSQL SELECT](https://www.postgresql.org/docs/18/sql-select.html#SQL-LIMIT). В [go-data-structures](../skills/go-data-structures/SKILL.md#declaring-empty-slices) nil→null не ограничено `encoding/json` v1: проба v2 даёт `[]`. В [WEB-SERVER.md](../skills/go-http/references/WEB-SERVER.md#structure), строка 40, `NewServer` требует вызвать отсутствующий у этого типа `Shutdown`. В [go-code-review](../skills/go-code-review/SKILL.md#testing) `NewTestServer` назван real transport, хотя его default — in-memory network. [Документация httptest](https://pkg.go.dev/net/http/httptest#NewTestServer). Исправления должны обновлять references и сводные таблицы вместе с entrypoint.

**Две дополнительные границы примеров.** Contiguous 2D slice из `go-data-structures:104–110` пригоден для фиксированных строк, но `append` в одну строку может перезаписать следующую; это воспроизведено. Это не дефект фиксированного размера сам по себе: пример должен явно задавать этот контракт либо ограничивать capacity через full slice expression. В `go-testing:112` совет избегать serialized comparisons нужен для внутренних структур, но контракт HTTP/JSON требует проверять и wire representation; это условие следует назвать, чтобы не потерять nil/empty, headers и error-text regression checks.

**Решение по каждому навыку.** Объём — слова в полном `SKILL.md`. Число quality validation — записи `evals.json`, а не число успешных модельных прогонов и не гарантия независимого holdout.

| Навык | Слова | Quality validation | Что улучшить или сохранить |
|---|---:|---:|---|
| [go-code](../skills/go-code/SKILL.md) | 1 408 | 5 | Сохранить routing по изменяемому решению. Проверить согласование description с условием «обычный local identifier не требует отдельного owner». Сжать повтор restraint ladder после устранения конфликтов. |
| [go-code-refactor](../skills/go-code-refactor/SKILL.md) | 2 964 | 3 | Сохранить текущую LOC-политику и реальный счётчик; исправить deep-copy/shared-fix reference и оценку empty diff. Проверить, что gate не стимулирует удаление полезных комментариев. |
| [go-code-review](../skills/go-code-review/SKILL.md) | 1 926 | 1 | P1: read-only, единый gate, статическое доказательство и фильтрация гипотез. Checklist должен указывать критерии риска, а не переписывать всех owners. |
| [go-concurrency](../skills/go-concurrency/SKILL.md) | 1 109 | 3 | P1: bounded пример и честный reference-routing; отделить ожидание, отмену и admission. |
| [go-context](../skills/go-context/SKILL.md) | 629 | 3 | Исправить context values; не навязывать параметры «на будущее». Thread-safe Context не делает thread-safe объект в `Value`. |
| [go-data-structures](../skills/go-data-structures/SKILL.md) | 919 | 5 | Указать v1/v2 для nil JSON и фиксированный размер contiguous rows. Убрать blanket-правило «любой ручной loop — дефект», сохранить проверку семантики stdlib API. |
| [go-database](../skills/go-database/SKILL.md) | 1 172 | 1 | Сохранить rows/tx lifecycle и PostgreSQL-reference. Уточнить pool sizing с учётом числа экземпляров; смягчить «unit-test only row mapping» и обязательную упаковку migrations через embed до project defaults. |
| [go-defensive](../skills/go-defensive/SKILL.md) | 1 266 | 0 | Условия copying/enum/recover/epsilon должны стоять рядом с правилом. Добавить сценарии владения и корректного zero default. |
| [go-documentation](../skills/go-documentation/SKILL.md) | 752 | 0 | Ограничить doc-check изменённым API; не чинить документацию всего проекта ради одного комментария. Проверять contract/cleanup/errors, а не наличие точки само по себе. |
| [go-error-handling](../skills/go-error-handling/SKILL.md) | 1 029 | 3 | Сохранить удачную таблицу `%w` / sentinel / custom type. Свести повтор log-or-return к одной норме; явно ограничить `log.Fatal` точкой завершения программы. |
| [go-functions](../skills/go-functions/SKILL.md) | 749 | 1 | Сохранить выбор configuration API по callers. Правила порядка функций и замены bool custom type сделать условными; не порождать типы ради стиля. |
| [go-generics](../skills/go-generics/SKILL.md) | 864 | 1 | Сохранить проверенную версионность. «Второй тип» и «alias только для migration» оформить как defaults с учётом запрошенного public API; нужны positive/negative пары. |
| [go-http](../skills/go-http/SKILL.md) | 1 057 | 1 | Сохранить EOF, HEAD, array contract и max+1 response limit. Исправить server-reference, различать API/full server/streaming scopes; не загружать и не применять весь server checklist для client-only задачи. |
| [go-interfaces](../skills/go-interfaces/SKILL.md) | 892 | 1 | Сохранить правило реального consumer и heuristic-only assertions. Объединить «return concrete» и «return interface» в явный выбор; не менять method sets массово ради receiver consistency. |
| [go-linting](../skills/go-linting/SKILL.md) | 1 654 | 4 | Сделать единственным владельцем gate. Проверить, что denylist допускает документированные исключения (`google/uuid` v5, измеренно нужный zap), вместо текста-исключения при безусловном deny. Setup/CI детали — кандидаты на references. |
| [go-logging](../skills/go-logging/SKILL.md) | 1 183 | 4 | Сохранить точное объяснение enriched logger/Handler. Вынести расширенный observability checklist в условный reference, убрать автоматическое расширение scope. |
| [go-naming](../skills/go-naming/SKILL.md) | 932 | 1 | Сузить discovery до naming decisions/API rename; синхронизировать исключение `_` prefix с категорическим MixedCaps. Удалить повтор decision tree/table, если eval сохраняет routing. |
| [go-packages](../skills/go-packages/SKILL.md) | 1 165 | 2 | Сохранить трёхступенчатую dependency ladder, удалить устаревший порядок `x/` в restraint-reference. Требовать fallback только для поддерживаемых GOOS, не «любой платформы». |
| [go-performance](../skills/go-performance/SKILL.md) | 914 | 2 | Сохранить measured-before/after и retention guidance. Числа 2×/7×/12× снабдить исходным benchmark/toolchain либо убрать как универсальные обещания. Исключить спекулятивные pointer API. |
| [go-resilience](../skills/go-resilience/SKILL.md) | 1 263 | 6 | Сильный образец outcome/contract guidance. Сохранить replay safety, budget, admission, ambiguous outcomes; добавить исполнительные проверки вместо дальнейшего роста общих правил. |
| [go-security](../skills/go-security/SKILL.md) | 1 165 | 0 | P1: исправить JSON-reference, добавить quality cases. Затем проверить LIMIT, parser limits, DNS/redirect binding и аргументы subprocess; не считать наличие API полной защитой sink. |
| [go-style-core](../skills/go-style-core/SKILL.md) | 994 | 4 | Сохранить house-style/scope/communication ownership; убрать противоречащие дубли из других skills. `String()` для enum оставить по потребности, не всегда. |
| [go-testing](../skills/go-testing/SKILL.md) | 1 138 | 0 | Добавить wire-contract исключение; различать in-memory и socket transport. Добавить validation на test design и отсутствие лишних helpers/dependencies. |
| [go-troubleshooting](../skills/go-troubleshooting/SKILL.md) | 1 783 | 6 | Сохранить evidence chain, live-capture границы и fix-at-responsible-boundary. Сжатие runtime-каталога пробовать отдельно: его объём содержит полезные неочевидные различия. |

**Что адаптировать под модели.** Предлагается одно общее Go-ядро. Отдельные копии всех skills под каждую модель сейчас удвоят обслуживание без доказанной пользы.

- **Opus 5:** явно обозначать scope и нужный объём результата; убрать повторные универсальные self-check/verification предписания, сохранив обязательные проверки проекта и поведения. Effort и delegation limits задавать в host/runner. Руководство Anthropic отдельно предупреждает об over-verification и избыточной делегации. Это основание для эксперимента, не доказательство экономии именно на этом пакете. [Prompting Claude Opus 5](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/prompting-claude-opus-5).
- **GPT 5.6:** сохранять outcome, constraints и completion criteria; устранять противоречия и повторы по одной группе. Не усиливать общий запрет на длинные ответы без проверки полноты. Не переносить результаты Luna на Sol/Terra: для семейства нужна отдельная строка результатов на каждый точный model ID. [OpenAI GPT-5.6 prompting guidance](https://developers.openai.com/api/docs/guides/prompt-guidance-gpt-5p6).
- **Обе модели:** entrypoint содержит область применения, важные инварианты, решение по неоднозначным случаям и условные ссылки. Tutorial-примеры и настройка инструментов читаются по необходимости. Удаление текста оценивается по выполнению задачи, а не по достижению универсального лимита слов. [Anthropic skill authoring](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices).

**Что уже показывают сохранённые эксперименты.** Эти результаты прочитаны заново из файлов репозитория; это исторические snapshots, а не fresh validation текущей версии.

- [Opus 5 medium, 2026-09-08](evidence/2026-09-08-go-refactor-control-opus-5-medium.md): 20 no-skill + 20 baseline; build/golden 20/20 на каждом плече. Один model-authored test baseline не прошёл. Цена из raw JSON: $3.209648 против $9.4434915, примерно 2.94×. Отчёт не устанавливает устойчивого эффекта размера кода; долю цены, вызванную чтением skills, эти данные не выделяют. SHA-256 raw JSON перепроверен и совпадает с отчётом.
- [GPT-5.6-Luna medium implement, 2026-09-08](evidence/2026-09-08-go-implement-control-gpt-5.6-luna-medium.uk.md): raw JSON подтверждает 20/20 build/golden в обоих плечах. Такой насыщенный correctness-набор плохо различает улучшения. Нулевые cost поля Codex не означают бесплатный запуск.
- [Luna description compression](evidence/2026-09-07-description-compression-codex-luna-medium.md): опубликованный отчёт показывает меньший catalog/input и 88/105 против 84/105 trigger pass; при одном повторе изменение качества не установлено. Это не основание ни обещать улучшение от сжатия, ни массово откатывать его.
- [Текущий skill против no-skill на gw3](evidence/2026-09-09-refactor-current-vs-none-luna-medium.md): одна положительная пара, не доказательство общего эффекта. В этом же отчёте отмечено ложное `skill fired` из упоминания отсутствующего пути: проверять надо успешное чтение правильного snapshot, а не regex по команде.

**Порядок улучшений и критерии приёмки.**

1. Исправить фактические дефекты и owner-conflicts: JSON/Clone/fan-out, review-only, context, reference drift. Для каждой правки сначала сохранить воспроизводящий случай. Не переписывать одновременно descriptions, routing, примеры и model settings.
2. Исправить оценку empty diff и происхождения загруженного skill. В `evals/cmd/evalrun/main.go:370–445` quality проверяется текстовым judge после `Skill,Read,Glob,Grep`; такой pass не означает, что сгенерированный код компилируется или выполняет контракт. Сохранить этот дешёвый слой для решений, но дополнить сложные случаи независимыми исполняемыми тестами.
3. Закрыть покрытие: `go-security` — 0 quality вообще; `go-defensive`, `go-documentation`, `go-testing` — 0 quality validation. Добавить хотя бы по positive, near-miss и scope-negative случаю в проблемных зонах. Пять навыков также не имеют положительных trigger validation: review, data-structures, documentation, naming, performance. Название set не делает кейсы независимым holdout, если по ним уже редактировали skill.
4. Сохранить текущий plugin как `reference`; сравнить `no-skill`, `reference`, `candidate` отдельно на Opus 5 и выбранных GPT-5.6 model IDs. Фиксировать runner/version, effort, доступные tools, prompt, fixture/plugin hashes, успешное чтение references, outputs и diff. Порядок jobs перемешивать; seed runner не гарантирует детерминированность модели. Для сравнений внутри модели tool permissions должны совпадать.
5. Начать с дешёвого screening дефектов и смешанных задач. Затем повторить отобранные изменения на нескольких независимых fixtures; например, 5 повторов на fixture/arm как пилот, с дальнейшим размером по дисперсии и цене. n=5 само по себе не доказывает улучшения. Во время оптимизации оставить новые неиспользованные случаи для итоговой проверки.
6. Измерять correctness/task completion, scope compliance, quality/precision review findings, сохранность полезных комментариев и whole-code readability; отдельно physical/code LOC, Go tokens, named functions и closures. Затем сравнивать input/cached/output tokens, tool calls, время и стоимость. Не выбрасывать неудачные модельные outputs из знаменателя; инфраструктурные сбои и retry attempts сохранять отдельно. Только после этого сокращать entrypoints и descriptions по одной группе.

Существующий Concision Gate — явная политика этого пакета. Автоматически ослаблять его ради абстрактной «читабельности» не предлагается. Его исполнение и побочные эффекты нужно измерять; более короткий source не доказывает лучшего контракта или меньшей сложности.

**Проверено в этом аудите.**

- `quick_validate.py` — PASS для 24/24 навыков.
- `PATH=/Users/eugeneshershen/go/bin:$PATH go test -count=1 ./...` из `evals/` — PASS. Основной пакет 17.995 s, `abrun` 11.646 s. Это локальные repository tests, без нового платного модельного прогона.
- `go doc` сверил `Decoder.More`, `WaitGroup.Go`, `context.Context`, `httptest.NewTestServer` с установленным Go 1.27.1.
- [Изолированные воспроизведения: исходник и вывод](evidence/2026-09-09-all-skills-opus5-gpt56-audit.probes.txt) — exit 0; наблюдения ниже являются воспроизведением ограничений/ошибок рекомендаций, а не pass новых навыков:

```text
JSON "{}]": first=<nil> More=false second=invalid character ']' looking for beginning of value
JSON "{} {}": first=<nil> More=true second=<nil>
JSON "{} ": first=<nil> More=false second=EOF
WaitGroup queue: 64 tasks simultaneously blocked, no admission limit
2D contiguous slice: appending row0 changes row1[0] from 7 to 9
slices.Clone([]*int): editing cloned element changes original from 7 to 9
nil []string: json/v1=null json/v2=[]
```

Свежие Opus/GPT A/B, полный executable-аудит всех сниппетов, реальный PostgreSQL и полный CI с npm-валидатором в этом анализе не запускались. Подтверждены дефекты инструкций и ограничения измерения; величина улучшения поведения моделей пока не измерена.


## Применённые исправления

2026-09-09, по запросу «исправь находки ревью». Исправлены конкретные дефекты
в 21 каталоге навыков, связанных references/шаблонах и eval runner. `go-code`,
`go-resilience`, `go-troubleshooting` сохранены: аудит рекомендовал оставить
их основные контракты. `go-code-refactor/SKILL.md`, включая Concision Gate,
также не изменён; исправлены его справочники.

| Находка | Результат |
|---|---|
| JSON `More` вместо EOF | Security ссылается на HTTP single-document owner; добавлены исполняемые случаи с хвостами `]` и `}` |
| Clone вместо deep copy | Исправлены restraint table, boundary copying и BEHAVIOR-TRAPS; сохранены глубина копии и nil/empty policy |
| Неограниченная очередь | Основной пример использует фиксированное число workers с отменой и join; reference различает WaitGroup, errgroup и admission |
| Review расширяет scope | Review-only, статическое доказательство, гипотезы и один repository gate согласованы; поправлены сводные checklist rows |
| Empty diff и skill loading | Завершённый refactor no-op входит в средние; implementation/stub без правки остаётся ошибкой. Codex требует успешный завершённый command и соответствующий frontmatter в output |
| Context / место исправления | Чистым helpers не навязывается ctx; cancellation не называется Value; shared fix применяется только к ошибочному shared contract |
| Универсальные preferences | Уточнены enum, copy, receiver, alias, signatures, migration, pool и observability conditions; исключения zap/UUID v5 не запрещены blanket denylist |
| Reference drift | Исправлены SQL LIMIT, nil JSON v1/v2, несуществующий Shutdown и transport claims; contiguous rows ограничены capacity; performance multipliers перестали быть универсальными обещаниями |

Добавлено 20 quality-сценариев (6 train, 14 validation) и 5 trigger validation.
Всего теперь **79 quality** (27 train / 52 validation) и **110 trigger**
(64 train / 46 validation). У security, defensive, documentation и testing
теперь по три quality validation. Это созданные регрессионные сценарии,
а не результаты их модельного выполнения или новый статистически независимый holdout.
README-счётчики и документация runner синхронизированы.

**Наблюдавшийся red → green:** новые извлечённые из Markdown тесты сначала
показали 64 активных задачи вместо 8, незавершение idle queue после отмены
и перезапись соседней строки после append; затем прошли. Runner-регрессии
сначала воспроизвели false skill reads, исключение no-op из среднего и отсутствие
completion evidence; исправления прошли соответствующие тесты.

**Финальная проверка:**

- `quick_validate.py`: 24/24 PASS.
- `agentskills-validate@1.0.1`: 24/24 PASS.
- `go test -count=1 ./...` из `evals/`: PASS; итоговый основной пакет 20.950 s,
  `cmd/abrun` 13.653 s. Markdown-примеры исполняются с `-race` через harness.
- `go build ./...`, `go vet ./...`: PASS.
- `go test -race -count=1 ./cmd/abrun`: PASS после lint-cleanup, 13.209 s.
- `golangci-lint run --config ../skills/go-linting/assets/golangci.yml
  --new-from-rev=HEAD . ./cmd/abrun`: PASS, 0 issues в изменениях.
- Проверка конфигурации golangci-lint и относительных ссылок: PASS.
- Независимое ревью runner и исправленных owner/summary правил завершено
  без оставшихся actionable findings. `graft build` обновил граф.

[Полные независимые пробы применения навыков](evidence/2026-09-09-all-skills-review-fixes.probes.md)
сохранены отдельно. В обоих проходах агент выбрал подходы, сохраняющие контракты
шести сценариев; это качественная проверка инструкций текущей моделью сессии,
не Opus/GPT A/B. Candidate-проба нашла оставшуюся ошибку nil/Clone в
BEHAVIOR-TRAPS; она исправлена и проверена независимым review.

Массовое сжатие всех entrypoints, расширение benchmark corpus и платные
сравнения моделей остаются отдельным экспериментальным этапом из аудита.
Новые Go-зависимости не добавлены. Исторические benchmark results не пересчитаны:
старые записи без completion evidence консервативно сохраняют прежнее исключение.
Partial/truncated skill reads без frontmatter остаются неподтверждёнными;
совпадение transcript не доказывает hash установленного файла или применение правила.

Новый concurrency regression размещён в отслеживаемом `evals/go127_examples_test.go`:
на момент создания отдельного файла `.gitignore` исключал новые файлы в `evals/`.
После проверки пользователь удалил это правило; исключения `evals/` больше нет.
Тест сохранён в существующем файле. После переноса целевые тесты прошли за 4.057 s,
проверка lint изменённых строк снова завершилась с 0 issues; `git diff --check` чист.
