# Исследование: лаконичный, читаемый Go 1.27 из go-скилов — 2026-09-09

Вопрос: какими проверяемыми способами улучшить 24 скила репозитория, чтобы
код, который модель пишет по ним, был короче, понятнее, соответствовал
принятым стандартам и использовал возможности Go 1.27.

## Как получено

- Workflow `deep-research`: 5 направлений поиска, 20 источников, 100 извлечённых
  утверждений, 25 отправлены на трёхголосую проверку. Итог: 15 подтверждены
  (3-0 или 2-1), 4 опровергнуты, 6 не проверены — агенты проверки упёрлись в
  лимит сессии. Шаг синтеза тоже не выполнился, поэтому сводка ниже собрана
  вручную из подтверждённых утверждений и цитат источников.
- Локальная проверка на установленном toolchain: `go1.27.1`, `api/go1.27.txt`
  и `api/go1.26.txt`, `go tool fix help`, `go tool vet help`,
  `go doc encoding/json/v2`, `golangci-lint 2.13.2`. Утверждения о Go 1.27,
  которые не удалось сверить ни с `go.dev/doc/go1.27`, ни с toolchain, помечены
  **[не сверено]**.

Главный вывод: покрытие Go 1.27 в репозитории уже почти полное. Таблица API в
`COMPATIBILITY.md` совпадает с `api/go1.27.txt` по всем пакетам, которые скилы
рекомендуют. Резерв улучшения лежит не в перечислении новых API, а в четырёх
местах: механизм соблюдения правил внутри сессии, полнота каталога
модернизаторов, глубина guidance по `encoding/json/v2` и метрика «использует
1.27» в evals.

## Рекомендации по ожидаемому эффекту

### 1. Переносить правила из прозы в инструменты, срабатывающие во время сессии

Что говорят данные:

- Факторный эксперимент на 1 650 сессиях Claude Code (Sonnet 4.6, 16 050
  функций): длина файла правил 25–500 строк, позиция инструкции, разбиение
  на несколько файлов и даже противоречие между файлами **не дали**
  статистически различимого эффекта на соблюдение правил. Соблюдение падает
  внутри сессии: каждая следующая сгенерированная функция снижает шансы
  соблюдения примерно на 5,6 % (OR = 0,944). Авторы рекомендуют hooks
  (PreToolUse/PostToolUse/Stop), напоминания в task-prompt и линтеры, а не
  перестройку файлов ([arXiv 2605.10039](https://arxiv.org/abs/2605.10039);
  цифры о decay и Bayes-факторы — **[не сверено]**, голосование не завершилось).
- Контекстные файлы AGENTS.md не повышают долю решённых задач и увеличивают
  стоимость инференса более чем на 20 %; при этом *инструкции* из них модели
  выполняют, а *обзоры репозитория* — бесполезны
  ([arXiv 2602.11988](https://arxiv.org/abs/2602.11988)). Независимая абляция
  на Claude Code и Codex (288 прогонов) тоже не нашла эффекта
  ([arXiv 2607.27250](https://arxiv.org/abs/2607.27250), **[не сверено]**).
  Оба исследования на TypeScript/Python; на Go не переносятся напрямую.
- Cursor: «Не копируйте style guide в правила — для этого есть линтер»; правила
  держать минимальными: команды, паттерны, ссылки на канонические примеры
  ([cursor.com/blog/agent-best-practices](https://cursor.com/blog/agent-best-practices)).
- Anthropic: цикл «validator → fix → repeat» «greatly improves output quality»
  ([Skill best practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices)).

Что это значит для репозитория:

- `hooks/go-vet-on-edit.sh` сейчас запускает только `gofmt -l` и `go vet` на
  пакете. Добавить туда `go fix -diff .` на том же пакете, только как отчёт
  (без применения). Это единственная точка, где правило «пиши современный Go»
  напоминается модели после каждой правки, а не один раз при загрузке скила.
  Область — пакет отредактированного файла, чтобы не нарушать правило scope из
  `go-style-core` «Write Current Go».
- Самая низкая дисциплина у задач рефакторинга: в том же эксперименте ~45 %
  против ~84 % на новом коде (**[не сверено]**, цифра из обзора статьи). Это
  совпадает с наблюдением из `evals/ab/README.md`, что fixture-эффект в
  `go-code-refactor` неустойчив между моделями. Для рефакторинга ставка на
  `verify-refactor.sh` и линтер важнее любой формулировки в SKILL.md.
- Скилы не удлинять. Все SKILL.md уже ≤ 317 строк при рекомендованном
  Anthropic пределе 500, а данные выше говорят, что длина сама по себе не
  двигает соблюдение.

### 2. Дополнить каталог модернизаторов `go fix`

Официальные release notes Go 1.27 (подтверждено 3-0,
[go.dev/doc/go1.27](https://go.dev/doc/go1.27)): добавлены `atomictypes`,
`embedlit`, `slicesbackward`, `unsafefuncs`; `fmtappendf` удалён «по
стилистическим причинам»; `waitgroup` переименован в `waitgroupgo`.

Локально `go tool fix help` в go1.27.1 показывает 27 анализаторов. В скилах и
`COMPATIBILITY.md` ни разу не названы шесть: `atomictypes`, `reflecttypefor`,
`slicesbackward`, `unsafefuncs`, `buildtag`, `plusbuild`. Ещё девять
(`mapsloop`, `slicescontains`, `slicessort`, `stditerators`, `stringsbuilder`,
`stringscut`, `stringscutprefix`, `stringsseq`, `hostport`) упомянуты только в
сводной строке таблицы `go-linting/SKILL.md`.

Куда вносить:

- `COMPATIBILITY.md`, раздел «Language features»: у `atomic.Int64` и друзей
  указать `go fix` analyzer `atomictypes`; у `slices.Backward` — `slicesbackward`.
- `go-linting/SKILL.md`, таблица модернизаторов: три новых 1.27 плюс
  `reflecttypefor`; отдельной строкой — что `appendclipped` и `slicesdelete`
  выключены по умолчанию, потому что меняют nil-ность базового слайса и
  зануляют хвост ([pkg.go.dev modernize](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize),
  2-1). Модель не должна предлагать эти переписывания как безопасные.
- Там же: `bloop` и `fmtappendf` включены в gopls, но выключены в `go fix`,
  поэтому `go fix -diff` их не покажет ([go.dev/gopls/analyzers](https://go.dev/gopls/analyzers),
  **[не сверено]**, gopls локально не установлен).
- `MODERNIZATION.md`, «Start with go fix»: после `go fix` обязателен
  `go build` — исправления могут оставить неиспользуемые импорты и переменные
  (подтверждено 3-0, документация пакета modernize).
- Мотивация от команды Go пригодится в `go-style-core` «Write Current Go» одной
  фразой: модернизаторы созданы именно потому, что LLM воспроизводят устаревшие
  идиомы из обучающего корпуса ([go.dev/blog/gofix](https://go.dev/blog/gofix), 3-0).

### 3. Ловушка `go` directive: без `go 1.27` в go.mod ничего из 1.27 не сработает

Три подтверждённых факта складываются в одну ловушку:

- `go fix` предлагает исправление только там, где `go.mod` или `//go:build`
  гарантирует нужную версию ([go.dev/blog/gofix](https://go.dev/blog/gofix), 3-0).
- С Go 1.26 `go mod init` пишет `go 1.(N-1).0`, то есть на toolchain 1.27 новый
  модуль получает `go 1.26.0` ([go.dev/doc/go1.26](https://go.dev/doc/go1.26), 3-0).
- С Go 1.27 `go test` запускает `stdversion` по умолчанию и отметит `uuid`,
  `strings.CutLast`, `url.URL.Clone` как «слишком новые» ([go.dev/doc/go1.27](https://go.dev/doc/go1.27), 3-0).

Итог: модель, следующая скилам, на свежем модуле получит либо ошибки vet, либо
молчаливое отсутствие модернизаций 1.27. `COMPATIBILITY.md` упоминает
`stdversion`, но не связку с `go mod init`. Добавить в `go-packages/SKILL.md`
(раздел о новом проекте) шаг «после `go mod init` выставить `go 1.27`» и в
`go-linting` строку о том, что пустой `go fix -diff` на модуле с `go 1.26`
ничего не доказывает.

### 4. `encoding/json/v2`: сейчас одна строка в ladder, а нужны конкретные формы

Сейчас скилы говорят о v2 только «быстрее и строже» (`go-packages`,
`go-performance`, `MODERNIZATION.md`). Ни один файл не называет
`MarshalWrite`, `UnmarshalRead`, `DefaultOptionsV1`, `FormatNilSliceAsNull`,
`MatchCaseInsensitiveNames`, `RejectUnknownMembers`, `Deterministic`.

Подтверждено (3-0, [go.dev/doc/jsonv2-migration](https://go.dev/doc/jsonv2-migration)
и [go.dev/doc/go1.27](https://go.dev/doc/go1.27)):

- Для простых вызовов миграция — смена импорта; сложность в поведении, не в API.
- `MarshalWrite(w, v)` заменяет создание `json.NewEncoder(w).Encode(v)` —
  короче на строку и на переменную. `UnmarshalRead(r, &v)` — симметрично.
- Каждое расхождение v1/v2 — это `Options`; `jsonv1.DefaultOptionsV1()` даёт
  поведение v1 через API v2, поздние опции перекрывают ранние, так что v2
  включается по одному поведению за раз.
- Финальные имена тегов в 1.27: `inline` переименован в `embed`; теги `format`
  и `unknown`, опция `DiscardUnknownMembers` и `SkipFunc` удалены. Любой пример
  с `inline` устарел.
- v1 теперь реализован поверх v2: текст ошибок может отличаться. Это относится к
  разделу `MODERNIZATION.md` «Toolchain shifts that break tests on their own»:
  тесты, сравнивающие строку ошибки `json`, ломаются при переходе на 1.27 без
  единой правки кода.

Проверено локально через `go doc encoding/json/v2`: nil-слайс маршалится как
`[]`, nil-map как `{}`; дубликаты имён и невалидный UTF-8 — ошибка;
сопоставление имён регистрозависимое. Соответствующее утверждение workflow
опроверг 1-2, но три из четырёх его частей документация toolchain
подтверждает; неподтверждённой осталась часть о «структурно невалидных типах».
Отдельно: v2 не сортирует ключи map по умолчанию, нужен `json.Deterministic`
для golden-тестов ([victoriametrics.com/blog/go-1-27](https://victoriametrics.com/blog/go-1-27/),
**[не сверено]** по release notes).

Владельцы: `go-defensive` (теги и nil-коллекции на границе), `go-packages`
(ladder), `go-testing` (golden-файлы и `Deterministic`), `MODERNIZATION.md`.
В `COMPATIBILITY.md` 1.27 добавить алиасы в v1: `json.Options`,
`json.RawMessage = jsontext.Value`, `UnmarshalTypeError.Unwrap` — они есть в
`api/go1.27.txt`.

### 5. Изменения поведения runtime в 1.27, которые меняют идиомы

Из release notes и обзоров ([go.dev/doc/go1.27](https://go.dev/doc/go1.27),
[victoriametrics.com/blog/go-1-27](https://victoriametrics.com/blog/go-1-27/));
эти конкретные пункты не входили в 25 проверенных утверждений, поэтому все
**[не сверено]** и требуют прочтения release notes перед правкой:

- Каналы `time.After`/`NewTimer`/`NewTicker` всегда небуферизованные,
  `asynctimerchan` удалён. Затрагивает советы о «дренировании» таймеров в
  `go-concurrency`; `time.After` сейчас упомянут только в SYMPTOM-CATALOG.
- Закрытие тела HTTP/1-ответа само дочитывает остаток, чтобы соединение
  вернулось в пул. Идиома `io.Copy(io.Discard, resp.Body)` перед `Close`
  становится лишней — `go-http` должен об этом сказать, иначе модель продолжит
  её писать.
- Профиль `goroutineleak` GA. Он уже используется в `verify-refactor.sh`, но
  не назван ни в `go-concurrency`, ни в `go-troubleshooting/SKILL.md`. Одна
  строка в SYMPTOM-CATALOG: «утечка горутин — `/debug/pprof/goroutineleak`».
- Для модулей `go 1.27` трейсбеки включают pprof-метки горутин — полезно в
  `go-troubleshooting` как источник данных.
- `go doc pkg@version` и слияние `require`-блоков в `go mod tidy` — мелочи для
  `go-packages`.

### 6. Пропуски Go 1.26

- Снято ограничение на самоссылочные ограничения типов:
  `type Adder[A Adder[A]] interface { Add(A) A }` (3-0,
  [go.dev/doc/go1.26](https://go.dev/doc/go1.26)). В `go-generics` и
  `COMPATIBILITY.md` этого нет; это редкий, но именно «1.26+» приём.
- Итераторы `reflect.Type.Fields/Methods/Ins/Outs` и `reflect.Value.Fields`
  (есть в `api/go1.26.txt`). По политике `COMPATIBILITY.md` вносить только если
  скил рекомендует; кандидат — `go-defensive` или `go-generics` там, где
  перечисляются поля через `NumField`/`Field(i)`.
- В 1.27 удалены GODEBUG `tls10server`, `tlsrsakex`, `tls3des`, `tlsunsafeekm`,
  `x509keypairleaf` (анонс в notes 1.26, 3-0). Для `go-security` это значит:
  TLS 1.2 минимум по умолчанию, `Certificate.Leaf` всегда заполнен — советы
  «выставьте MinVersion» можно сократить до проверки, что никто не понижает.
- `ReverseProxy.Director` → `Rewrite` уже есть в `MODERNIZATION.md`; трогать не
  нужно.

### 7. Новый анализатор gopls `errorsastypeshadow`

Скилы настойчиво рекомендуют `errors.AsType[T]`. gopls добавил анализатор,
который ловит типовую ошибку этой идиомы: в цепочке `if/else` второй вызов
`AsType` получает не исходную ошибку, а нулевое значение из первого `if`
([go.dev/gopls/analyzers](https://go.dev/gopls/analyzers), **[не сверено]**).
В `go-error-handling/SKILL.md` есть положительный пример, но нет
отрицательного. Добавить пару «так нельзя / так нужно» из трёх строк — это
ровно тот случай, когда пример убирает ошибку, которую правило не ловит.

### 8. Базовый `.golangci.yml`

`golangci-lint 2.13.2` локально предоставляет `modernize` с авто-исправлением.
В `skills/go-linting/assets/golangci.yml` его нет, хотя проза `go-linting`
называет его как опцию. Изменения к базовой конфигурации:

- Включить `modernize` — тогда gate ловит устаревшие идиомы даже там, где
  `go fix -diff` не запускали или scope его исключил. `intrange` и
  `copyloopvar` при этом избыточны (их покрывают `rangeint` и `forvar`).
- `usestdlibvars` — заменяет `"GET"` и `200` на `http.MethodGet` и
  `http.StatusOK`; прямой вклад в читаемость, авто-исправление есть.
- `nolintlint` с `require-explanation` и `require-specific` — чтобы модель не
  глушила линтер голым `//nolint` (практика из
  [конфигурации maratori](https://gist.github.com/maratori/47a4d00457a92aa426dbd48a18776322)).
- Не включать `exhaustruct`, `gomodguard`, `wsl` — помечены deprecated в v2
  ([golangci-lint.run/docs/linters](https://golangci-lint.run/docs/linters/)).
- `gofumpt` как formatter — спорно. Он механически убирает то, что иначе
  приходится описывать словами (пустые строки в начале блока, группировка
  `var`), но в базовой конфигурации maratori и в конфигурации репозитория
  выбран `goimports`. Предложение: оставить `goimports` по умолчанию, а в
  `FORMATTING.md` дать `gofumpt` как «строже, включайте в новых проектах»;
  сейчас он там только строкой таблицы.
- Пороги сложности (`gocognit 20`, `funlen 100/50` у maratori) как «ворота
  лаконичности» — не рекомендуются в базу. Единственное количественное
  свидетельство связи объёма кода и архитектурных smell'ов (ρ = 0,94) получено
  на Python и не прошло проверку ([arXiv 2605.02741](https://arxiv.org/abs/2605.02741),
  **[не сверено]**); restraint ladder в `go-code` решает ту же задачу без
  ложных срабатываний на длинных, но линейных функциях.

### 9. Итераторы видны только в references

`iter.Seq` встречается лишь в `CONTROL-FLOW.md` и `OVER-ENGINEERING.md`. В телах
`go-data-structures`, `go-functions`, `go-generics` нет ни одного примера
функции, возвращающей `iter.Seq[T]`, хотя именно они решают «вернуть слайс или
итератор». Добавить одну строку в таблицу решений `go-functions` («результат
только для обхода → `iter.Seq[T]`, для хранения → слайс») и
отрицательный пример `for k := range maps.Keys(m)` → `for k := range m`
(анализатор gopls `maprange`, **[не сверено]**).

### 10. Как писать сами скилы: что подтверждено, что уже сделано

- Anthropic: < 500 строк, references на один уровень, один default вместо
  списка альтернатив, примеры передают стиль лучше описаний, evals до текста,
  проверка на Haiku/Sonnet/Opus — Opus нужно меньше объяснений. Репозиторий
  уже соответствует всем пунктам, кроме, возможно, последнего: A/B-прогоны
  зафиксированы для Sonnet 5 и Opus 5, для Haiku — нет.
- 1 540 обновлений правил в 83 проектах: после правки правил соблюдение
  артефактов растёт с 49 % до 72 %; работают **конкретные правила с
  условиями**, абстрактные («пишите безопасно») — нет; 78 % разработчиков
  правят правила в ответ на конкретную ошибку ИИ
  ([arXiv 2606.12231](https://arxiv.org/pdf/2606.12231)). Это подтверждает
  формат «gate с проверяемыми условиями» из `evals/ab/variants/pattern-gate.md`.
- Предостережение об примерах: few-shot-примеры **увеличили** число Long
  Method у самых сильных моделей (Qwen-coder-480B 11 → 13, Gemini 2.5 Pro
  5 → 8), а «более способные модели производят больше процедурного раздувания»
  ([arXiv 2605.02741](https://arxiv.org/abs/2605.02741), **[не сверено]**,
  Python). Следствие: примеры в скилах держать короткими и парными
  (до/после), не давать полных программ, которые модель начнёт копировать по
  объёму.
- В 2 303 контекстных файлах из GitHub правила о безопасности и
  производительности встречаются в 14–15 % файлов против 60–76 % о тестах и
  сборке ([arXiv 2511.12884](https://arxiv.org/abs/2511.12884)). У репозитория
  есть `go-security`, `go-performance`, `go-resilience` — это уже редкое
  преимущество; не сокращать их ради краткости.

### 11. Измерять «использует 1.27» объективно

Сейчас `abrun` считает строки, функции и типы. Для критерия (d) есть дешёвая
метрика: число hunks в `go fix -diff ./...` на коде, который написала модель.
Ноль означает, что модернизаторов для написанного не осталось. Метрика не
покрывает то, для чего нет анализатора (`cmp.Or`, `errors.Join`, `iter.Seq`),
но она детерминирована, бесплатна и напрямую связана с пунктами 2 и 3.
Добавить её в отчёт `evals/ab/report` и в `evals/eval_test.go` как
регрессию на примеры из скилов: `go fix -diff` на извлечённых примерах должен
быть пустым.

Дополнительный источник для сверки: JetBrains `go-modern-guidelines` — 54
правила для Go 1.27 с версионными воротами, ставится как plugin в Claude Code
([dev.to/gde](https://dev.to/gde/go-in-practice-writing-modern-go-with-ai-testing-jetbrains-go-modern-guidelines-and-refactoring-151o)).
Разовая сверка их идентификаторов с таблицей «Reach For What Go Ships» покажет
пропуски без ручного перебора release notes.

## Что опровергнуто или не подтверждено

- «`go fix ./...` применяет всю suite modernize, а `cmd/modernize` нужен только
  для новых анализаторов» — опровергнуто 0-3: часть анализаторов (`bloop`,
  `fmtappendf`, `appendclipped`, `slicesdelete`) в `go fix` выключена.
- Все количественные утверждения из arXiv 2605.10039, 2607.27250, 2605.02741
  остались без голосов из-за лимита сессии. Направление выводов согласуется с
  подтверждёнными источниками (2602.11988, Cursor, Anthropic), но конкретные
  проценты цитировать без оговорки нельзя.
- Пункты о runtime 1.27 (таймеры, дренаж тела ответа, `go doc @version`)
  взяты из обзора VictoriaMetrics и не прошли отдельную проверку по
  `go.dev/doc/go1.27`.

## Сводка: куда вносить

| # | Файл | Правка |
|---|---|---|
| 1 | `hooks/go-vet-on-edit.sh` | `go fix -diff` на пакете, только отчёт |
| 2 | `COMPATIBILITY.md`, `go-linting/SKILL.md`, `MODERNIZATION.md` | 4 модернизатора 1.27, выключенные по умолчанию, `go build` после `go fix` |
| 3 | `go-packages/SKILL.md`, `go-linting/SKILL.md` | `go mod init` → `go 1.26.0`; выставить `go 1.27` явно |
| 4 | `go-defensive`, `go-packages`, `go-testing`, `MODERNIZATION.md`, `COMPATIBILITY.md` | `MarshalWrite`, `Options`, `embed`, nil → `[]`, `Deterministic`, текст ошибок v1 |
| 5 | `go-concurrency`, `go-http`, `go-troubleshooting/SYMPTOM-CATALOG.md` | таймеры, дренаж тела, `goroutineleak` |
| 6 | `go-generics`, `go-security`, `COMPATIBILITY.md` | самоссылочные constraints, удалённые GODEBUG |
| 7 | `go-error-handling/SKILL.md` | отрицательный пример shadowing `AsType` в `if/else` |
| 8 | `go-linting/assets/golangci.yml`, `FORMATTING.md` | `modernize`, `usestdlibvars`, `nolintlint`; `gofumpt` как опция |
| 9 | `go-functions`, `go-data-structures` | `iter.Seq` в таблице решений, `maps.Keys` в `range` |
| 10 | `docs/SKILL_AUTHORING_TEMPLATE.md` | короткие парные примеры; прогон на Haiku |
| 11 | `evals/ab/report`, `evals/eval_test.go` | метрика `go fix -diff` = 0 |

## Источники

Первичные: [go.dev/doc/go1.27](https://go.dev/doc/go1.27),
[go.dev/blog/go1.27](https://go.dev/blog/go1.27),
[go.dev/doc/go1.26](https://go.dev/doc/go1.26),
[go.dev/blog/gofix](https://go.dev/blog/gofix),
[go.dev/doc/jsonv2-migration](https://go.dev/doc/jsonv2-migration),
[pkg.go.dev/.../modernize](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize),
[go.dev/gopls/analyzers](https://go.dev/gopls/analyzers),
[golangci-lint.run/docs/linters](https://golangci-lint.run/docs/linters/),
[Anthropic skill best practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices),
[arXiv 2602.11988](https://arxiv.org/abs/2602.11988),
[arXiv 2511.12884](https://arxiv.org/abs/2511.12884),
[arXiv 2606.12231](https://arxiv.org/pdf/2606.12231),
[arXiv 2605.10039](https://arxiv.org/abs/2605.10039),
[arXiv 2607.27250](https://arxiv.org/abs/2607.27250),
[arXiv 2605.02741](https://arxiv.org/abs/2605.02741).
Блоги: [cursor.com/blog/agent-best-practices](https://cursor.com/blog/agent-best-practices),
[victoriametrics.com/blog/go-1-27](https://victoriametrics.com/blog/go-1-27/),
[dev.to/gde — JetBrains go-modern-guidelines](https://dev.to/gde/go-in-practice-writing-modern-go-with-ai-testing-jetbrains-go-modern-guidelines-and-refactoring-151o),
[maratori golangci config](https://gist.github.com/maratori/47a4d00457a92aa426dbd48a18776322).
Локально: `go1.27.1`, `api/go1.27.txt`, `api/go1.26.txt`, `go tool fix help`,
`go doc encoding/json/v2`, `golangci-lint 2.13.2`.
