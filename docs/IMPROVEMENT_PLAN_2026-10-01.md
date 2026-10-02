# План покращень: Sonnet 5.5, Opus 5.5, GPT 6.1, Grok 4.7

Початковий аудит: 1.28.1 (2026-10-01), коміт `8f0fd78`. План доповнено після аудиту
раннерів, хуків і метрик. Спочатку потрібні достовірні вимірювання й
результати хука; після цього можна оцінювати скорочення викликів і тексту.

База початкового аудиту після синхронізації `main`: **1.28.2**, коміт `997df43`.
Нові upstream-зміни збережені разом із локальними 0.1/G5/G6.1; конфлікт
`CHANGELOG.md` розв’язано зі збереженням release-блоку та Unreleased.

Історичні цифри Sonnet/Opus нижче перенесено з наданого плану. JSON-звіти й
траси цих прогонів не надані для аудиту, тому результати не перевірені
повторно. На момент аудиту нові модельні прогони не виконувались;
Sonnet A/B від 2026-10-02 наведено нижче. Назви моделей — цілі плану;
доступ до точних model IDs та фактично обслужену модель перевірити у preflight.

Codex CLI `0.159.3` уже встановлений, команда `debug prompt-input` доступна;
opencode у `PATH` не знайдено. CLI ще не підтверджує доступ до GPT 6.1.
У поточному `abrun` Codex і opencode не підключають хуки цього Claude-плагіна:
переносимість тексту залишається сліпою зоною. Це характеристика раннерів
паку, а не універсальне твердження про можливості хостів.

## Стан виконання

- [x] **0.1 — edit-hook:** причину мовчазного lint підтверджено A/B;
  фізичний cwd усуває неправильне diff-фільтрування через symlink.
  Вибрана конфігурація явно закріплена через `--config`; stderr, невдалі
  exit codes і порожній невдалий результат більше не приховуються.
  Native-регресії покривають symlink/HEAD, scope інших файлів та config
  над symlink із пробілами у шляхах; infrastructure cases покривають
  exit 1/3/124. RED → GREEN, повний race/shuffle suite, Bash syntax,
  lint-config verify та незалежні spec/quality review — PASS.
  Перевірений patch збережено разом із G5/G6.1 у поточному checkout на гілці
  `fix/claude-routing-evidence`; локальні коміти й модельні прогони не виконувались.
- [x] **0.2/G5 — докази читання Codex:** до CLI знімається snapshot
  staged-копії конкретного arm. `Skills` зараховується лише за successful
  completed event та точного повного captured output; JSON-поле
  `codex_skill_evidence` зберігає часткові, невдалі й непідтверджені випадки,
  byte coverage та доступність inventory, включно з repair turns.
  Literal `cat`/`head`/`sed`, quotes і native shell wrappers перевірені
  контрактами; unsupported forms не виконуються і лишаються `unverified`.
  Negative cases охоплюють path mentions, failed/started/partial output,
  custom executable, `.bak` та foreign-owner aliases, malformed `sed`;
  unusable read metadata не змінюють лічильник completion attempts.
  RED → GREEN, повний race/shuffle suite, vet, gofmt, Bash syntax,
  lint-config verify та незалежні spec/quality review — PASS.
  Це доказ captured output, не model-visible delivery чи retention.
- [x] **0.2/G6.1 — Claude routing:** `routing.confirmed.version=1`
  відокремлює correlated successful Skill results від legacy requests;
  Go edit attempts і successful results рахуються окремо. Pending Skill
  або непідтверджений Read fallback перед першою правкою дають partial
  evidence та nil router/style flags; пізній result не переписує snapshot.
  Routing blocks потребують PreToolUse, маркера routing gate та явного exit;
  generic blocks, cancelled/invalid metadata й дублікати не стають confirmed.
  `routing.turns` зберігає fresh context repair-сесій. Спільний підсумок
  implement/refactor/review розрізняє confirmed denominators, partial lower
  bounds, legacy та unmeasured; no-edit має N/A. RED → GREEN, independent
  spec/quality review та повний race/shuffle suite на `997df43` — PASS.
  Це tool-result evidence, не доказ retention чи фізичного часу запису файлів.
- [ ] **0.2/G6.2 — інші хости:** confirmed routing Codex/opencode та
  перевірка їхньої доступності метрик. Це наступний крок.
- [ ] **0.2/G6.3 — читання та команди:** references/whole-read coverage,
  script outcomes і категорії команд за профілями хостів.
- [ ] **0.2/G7–G8:** provenance, checkpoint і класифікація збоїв ще попереду.
- [ ] **0.3 — preflight:** доступ до цільових моделей та retained smoke
  traces ще не перевірені; потрібен окремо визначений бюджет прогонів.
- [x] **B1 — реалізація (2026-10-02):** edit-hook зберігає окремі атомарні
  JSON receipts кожного check: status, exit code, command, scope, config,
  tool/version, start/end і digests до/після. Verifier звіряє поточні входи;
  module/workspace/local replacements, tests, залежності, config, build env
  і tool binaries включені. Неповний snapshot забороняє reuse. `go-linting`
  володіє умовами зарахування; router/refactor/style синхронізовані.
  File-only/package/new-findings checks не підміняють ширший гейт чи race.
- [x] **B2 — реалізація (2026-10-02):** prompt-only lexer відокремлює
  comments/literals/imports, розпізнає aliases, grouped/dot/blank imports,
  context signatures, sync/logging types, error calls і bounded-worker
  contract у stub `pool`. Existing HTTP/SQL/security/resilience hints
  збережені за використанням; unused imports не обирають owner.
  Правила блокування edit-гейта не розширені.
- **Локальна валідація B1/B2:** focused hook/receipt/routing tests,
  повний `go test -count=1 -race -shuffle=on ./...` в `evals`, `go vet`,
  Bash syntax усіх hook scripts, lint-config verify, quick_validate чотирьох
  змінених skills і `git diff --check` — PASS. Незалежне read-only review
  перевірило FIFO, bounded config discovery і збереження HTTP/security hints;
  виявлені дефекти виправлені. Платних model calls не виконувалось.
- [ ] **B1/B2 — модельна перевірка:** окремо B1, B2 і їхнє поєднання;
  Claude implement/refactor, `-shell go`, обидві моделі, n=2 після preflight
  та визначення бюджету. Зберігати `routing:`, actual command categories,
  gate blocks, зайві owner loads, owners до першої правки, wall time,
  tokens/cost. G6.3 command categories ще відкриті. Скорочення Bash-викликів
  і cost на Sonnet виміряне нижче; Opus single-run smoke наведений нижче,
  matched A/B на Opus не виконано. Historical Sonnet A/B не підтвердив
  final reuse без verifier; його uptake перевірений окремими final runs.
- [x] **Sonnet 5.5 low — A/B (2026-10-02):** виконано 88 сесій на чотирьох
  frozen arms, `-shell go`, n=2, j=2; ще 2 smoke. Build/golden — 88/88,
  але один B2 refactor є no-op, тому це 87 завершених змін.
  B1 не підтвердив економію 2–4 Bash calls. B2 допоміг `pool`, але втратив
  `go-resilience` із Retry-After контракту `fetch`; це відтворена прогалина,
  на момент прогону ще не була виправлена. Повторне зарахування receipts моделлю не підтверджене:
  10 implement-сесій заявили `pass (hook)`, verifier не викликали.
  Opus і підтвердження продуктової користі залишаються відкритими.
- [ ] **B3 — Opus edits до loads:** контролювати в кожному майбутньому
  прогоні; якщо повернуться — аналізувати prompt note. Перший рядок note
  у цій реалізації збережено.
- [x] **Виправлення після Sonnet A/B (2026-10-02):** actionable
  Retry-After/backoff contracts у stubs знову дають `go-resilience`; unrelated
  prose і literals не є сигналом. Додано Bash verifier із expected cwd/argv,
  `hook_credit=true` відокремлено від legacy state-only `valid=true`.
  Hook output позначає reuse як unverified і дає абсолютні команди без
  `CLAUDE_PLUGIN_ROOT` у shell. Explicit namespaced Skill/Bash action є
  і в successful stdout, і в error stderr. Router/refactor/style/gate та
  script contracts синхронізовані; edit-gate heuristic не розширений.
  Native regressions перевіряють wider scope, інший cwd, stale input,
  цитування шляхів та executable commands на обох output paths.
- **Post-fix live smoke:** чотири цільові Sonnet 5.5 low сесії під час
  уточнення інструкції, `fetch` та `ledger`, build/golden 4/4; CLI cost
  близько `$1.00`, cap `$3`. Fetch завантажив resilience до правки. Жодного
  неперевіреного `pass (hook)` більше не заявлено. Сам verifier модель
  ще не викликала; вона запускає частину прямих checks або прямо називає
  missing evidence. Uptake, повне виконання closing gate та економія
  залишаються непідтвердженими; це smoke, не новий matched A/B.
  Артефакти: `/Users/eugeneshershen/ab-results/sonnet55-low-b1-b2-fixes-20261002-173326`.
- [x] **B1 behavioral acceptance, Sonnet 5.5 low (2026-10-02):** primary
  prompt action називає всі selected owners/testing з exact Skill names і
  go-linting за наявності Bash; Skill results очікуються до edit message.
  Новий Bash workflow guard після Go edit вимагає gate owner та поточну
  спробу verifier. Batch `--gate` зараховує лише придатні package checks,
  решту повертає в `required_direct`; direct checks дозволені після failed
  attempt. Inventory/read-only/no-edit sessions не розширюють цей scope.
  Старі generations, trailing/conditional/background verifier mentions і
  `go -C` / env / timeout options перевірені negative regressions.
  Fingerprint виключає чотири live-confirmed host-only env fields;
  code/config/build/dependency/test inputs лишаються у digest.
  Stop перевіряє explicit hook-pass claims за current credits та fresh inputs;
  per-check credits не губляться в межах generation. Є opt-out без credit.
  Основний single run: go-linting **11/11** (раніше 5/11), actual verifier
  **11/11** (раніше 0/11), selected upfront owners **11/11**, build/golden
  **11/11**. Optional testing у refactor без test edits не є missing owner.
  Після фінального hardening окремі current-source gateway/fetch samples
  також мають owner/verifier/upfront/golden 2/2 і supported hook claims.
  JSON-only audit спочатку пропустив останній verifier gateway через grep;
  readback current-generation state підтвердив lint credit і виправив false flag.
  Ці прогони доводять спостережену поведінку, не стійку економію чи retention.
  Артефакти: `/Users/eugeneshershen/ab-results/sonnet55-low-b1-final-20261002-210219`.


- [x] **Opus 5.5 — release smoke (2026-10-02):** поточний checkout,
  `claude-opus-5-5`, requested effort `low`, `-shell go`, baseline-only,
  n=1, j=2, без repair/judge. Model ID підтверджено assistant events;
  effective effort CLI окремо не повідомляє. Implement build/golden 7/7,
  refactor 4/4, реальні зміни 11/11, go-linting і successful verifier 11/11,
  required router/style/selected owners до першої Go-правки 11/11.
  Unsupported hook-pass claims після current-generation readback — 0.
  Вартість CLI — $7.1395; refactor скоротив production code на 81 рядок.
  Старий JSON-only audit дав false flags fetch/gateway через tail/grep;
  successful tool results і latest/attempted credits перевірені окремо.
  Native routing aggregate лишається переважно partial; ці 11/11 отримані
  додатковим корельованим аудитом трасс. Optional go-testing у refactor
  не належить required owners до першої production edit.
  Неблокируюча проблема: pricing loc-diff відхилив /var → /private/var;
  модель повідомила про це, використала wc, independent runner підтвердив
  golden і скорочення. Existing roster legacy findings і revive naming у
  dispatch відокремлені; нових lint findings немає. Structural race/shuffle,
  go vet, Bash syntax, lint-config verify і diff checks — PASS.
  Один повтор доводить observed behavior; економія, repeated reliability,
  review/trigger corpus та інші хости цим прогоном не підтверджені.
  Артефакти: `/Users/eugeneshershen/ab-results/opus55-low-release-20261002-232529`.

Етап 0 загалом і модельна перевірка B1 ще не завершені. Система receipts
реалізована локально; мовчання не дозволяє `pass (hook)`.
Live schema/smoke Codex ще належить 0.3; G5 перевірений контрактними тестами
та fake CLI без платних запитів.

### Sonnet 5.5 low: результати B1/B2, 2026-10-02

Контроль — `dde7b5c` / 1.28.3, лише B1, лише B2 та B1+B2 на тій самій
базі; implementation-код під час експерименту не змінювали. Чотири повні
plugin arms у спільному randomized run кожного corpus, seed `20261002`.
В тимчасовій копії раннера додано лише два named arms; scoring/session
код не змінено. CLI `2.1.287`, Go `1.27.1`, golangci-lint `2.13.2`.
Model `claude-sonnet-5-5` підтверджено assistant events, effort `low`
задано CLI; немає evidence model fallback чи зміни plugin hashes.

Ліміт — $60 загалом / $1.50 на сесію. Облікований CLI list-price cost:
**$23.5161** разом із двома smoke, без repair/judge/retries.
88 evaluation sessions: implement 7 fixtures × 4 arms × 2,
refactor 4 fixtures × 4 arms × 2. Усі build/golden пройшли.
Implement model tests — 56 pass; refactor model tests — 32 skipped,
оскільки моделі не створили test files. B2 `report` rep 0 не змінив код;
cost і результат збережені, completed-refactor paired analysis його виключає.

| Corpus | Arm | n | Bash requests/run | Successful Bash results/run | Gate blocks/run | Wall s/run | Cost за arm |
|---|---|---:|---:|---:|---:|---:|---:|
| implement | контроль | 14 | 3.57 | 2.50 | 0.50 | 51.9 | $3.9446 |
| implement | B1 | 14 | 3.21 | 2.14 | 0.79 | 54.5 | $3.9591 |
| implement | B2 | 14 | 4.43 | 2.71 | 0.50 | 56.8 | $4.2114 |
| implement | B1+B2 | 14 | 3.57 | 2.21 | 0.93 | 56.0 | $3.9660 |
| refactor | контроль | 8 | 4.13 | 2.25 | 0.50 | 37.4 | $1.8904 |
| refactor | B1 | 8 | 4.38 | 2.50 | 0.38 | 36.2 | $1.8172 |
| refactor | B2 | 8 | 4.13 | 2.50 | 0.25 | 33.0 | $1.7617 |
| refactor | B1+B2 | 8 | 3.00 | 1.63 | 0.38 | 32.6 | $1.6924 |

Таблиця зберігає всі outcomes, включно з no-op. У семи completed B2
refactor pairs: ΔBash `0.00`, Δblocks `−0.143`, Δwall `−4.67 s`.
Це exploratory n=2; повтори однієї fixture не є незалежними задачами.
Командні категорії та input/output/cache tokens збережені в `analysis.json`.
Категорії рахують Bash calls із командою й успішним tool result;
compound/conditional shell та bundled scripts не доводять точну кількість
внутрішніх subprocesses. Робота edit-хука в Bash counter не входить.

**B1:** ΔBash `−0.36` для implement і `+0.25` для refactor.
Очікувані 2–4 виклики не заощаджені. Receipts видимі в live PostToolUse;
збережено 152 hook runs / 760 check records. Шість B1 і чотири B1+B2
implement-сесії заявили `pass (hook)`, але verifier не викликали.
Це непідтверджене зарахування за чинною політикою, а не доказ stale state
в кожному такому випадку. Shell-profile adaptation і конкретний load/verify
крок потребують окремого виправлення та порівняння.

**B2:** у `pool` обидва повтори завантажили context/concurrency до правки;
gate blocks `2 → 0` сумарно. Водночас `fetch` має відтворену регресію hints:
контроль називає HTTP/errors/resilience, B2 — context/HTTP/errors.
Retry-After у контрактному коментарі видалено lexical scan, а contract
heuristic відновлює лише bounded workers. Fetch blocks за два повтори:
контроль `1`, B1 `5`, B2 `4`, B1+B2 `10`; не всі ці блоки зумовлені
resilience — є пропущені named owners, testing і defensive. Не переносити
ці цифри як causal effect тільки одного scanner pattern.

**B1+B2:** implement Bash не змінився, blocks зросли `0.50 → 0.93`;
refactor Bash `4.13 → 3.00`, blocks `0.50 → 0.38`. Малий корпус не
підтверджує стійкої продуктової користі; B2 прогалину слід усунути до
наступного paid comparison.

**Routing/B3:** Go edit-tool attempts до confirmed Skill load — `0`.
Router/style підтверджені до першої Go edit у `87/87` edited sessions;
no-op має N/A. Це не перевірка фізичного часу Bash edits чи retention.
Native `routing:` залишив `partial` у `63/88` і `observed` у `25/88`:
службове message-string event CLI та silent successful PreToolUse не
сумісні з усіма припущеннями поточного parser. Окремий trace audit
корелює Skill success metadata / edit results / routing gate exit 2;
нативні прапорці не переписані. Opus B3 цим прогоном не перевірений.

Артефакти поза репозиторієм:
`/Users/eugeneshershen/ab-results/sonnet55-low-b1-b2-20261002-155745`.
`protocol.json`, plugin hashes, temporary runner diff, reports,
`analysis.json`, `summary.csv`, `validation.json`, raw launches/traces,
scratch sources і copies receipts збережені для readback.

## 0. Що вже відомо

Розміри чотирьох `SKILL.md` на 1.28.1: 19 789, 18 032,
20 550 і 12 932 байти відповідно. На поточній 1.28.2 повторний замір дає
20 445, 20 089, 20 670 і 13 884 байти. Значення `description` у 24 frontmatter
займають 6 905 байтів; це інший замір, ніж відрендерений listing із назвами,
шляхами й обмеженнями хоста. Байти й символи не є токенами. Поведінкові
твердження таблиці нижче лишаються історичними результатами початкового плану.

| Факт | Звідки | Що з цього випливає |
|---|---|---|
| Гарячий шлях: `go-code` 19,8 KB, `go-style-core` 18,0 KB (з карткою), `go-code-refactor` 20,6 KB, `go-code-review` 12,9 KB; описи 24 скілів 7,2 KB у кожній сесії | `wc -c` | Описи платить кожна сесія на кожному хості; великі SKILL.md — кожна маршрутизована |
| Скорочення тексту на ~1–3K токенів не видно у вартості при n=2 | три прогони 2026-09-30 | Текст різати заради хостів без хуків і компакції, не заради грошей на Claude |
| Поведінкові зміни видно: `go-testing` без умови → блоки гейта 1,14 → 0,36 (Sonnet), 1,21 → 0,64 (Opus) | 1.27.1 | Шукати зайві ходи, повторні правки, зайві тести — це гроші |
| Opus 5.5 low виконує рядок роутера в note і пропускає наступні, якщо в них немає конкретної дії з інструментом | шість формулювань, 1.28.0 | Рядок роутера не чіпати; усе обов'язкове — в ньому, у формі «the X skill (Skill tool, name X)» |
| Contract Table коштує +22% (Sonnet) / +26% (Opus) за тієї ж якості; без неї тести не пишуться (0/35) | n=5, medium | Рішення: лишити, тести потрібні |
| Залишкові блоки гейта (~0,36–0,43 на implement-сесію) — owner-скіли, які гейт бачить лише в тексті правки | 1.28.0 | Їх можна назвати заздалегідь (етап B2) |
| У сесіях із shell моделі самі запускають `gofmt`/`vet`/`test`/`lint` — 3–5 викликів на сесію, поверх edit-хука, що вже все прогнав | `-shell go`, 1.28.0 | Етап B1 |
| Recall у review на Sonnet low гуляє ±0,04 на ідентичному контексті | 1.27.0 | Різницю менше 0,05 при n=2 не трактувати |
| Opus 5.5 low у review майже не вантажить `go-code-review` (0,1 Skill-ходу), recall при цьому ~0,90 | 2026-09-29 | Review-корпус на Opus low міряє саму модель |

## 1. Правила вимірювання для всіх етапів

- Дві гілки в одному прогоні: `reference` (worktree попереднього релізу або
  копія плагіна з варіантом) проти `baseline` (робоче дерево). Ніколи не
  порівнювати клітинки з різних прогонів напряму — шум між прогонами
  (блоки гейта 0,50 vs 1,14 у тій самій гілці) більший за більшість ефектів.
- `-shell go` для implement і refactor — так працюють реальні користувачі
  Claude із доступним shell. Прапорець підтримує лише Claude; shell і права
  інших раннерів фіксувати окремо. Claude review — без shell; для Codex shell
  потрібен також для читання, тому необхідний окремий read-only профіль.
- Читати рядок `routing:` у підсумку `abrun` (перший інструмент, перше
  завантаження, правка до завантаження, блоки гейта, shell-виклики) — він
  реагує на зміну тексту раніше, ніж golden чи вартість.
  Зараз `routing:` є Claude-метрикою, а не спільною метрикою всіх раннерів.
  Після G6.1 новий confirmed summary спирається на tool results; legacy
  attempts показуються окремо. Partial/unmeasured не входять у confirmed
  знаменники, тому старий рядок не слід порівнювати з новим без визначень.
- n=2 для перевірки «нічого не зламали», n=3 для вибору формулювання,
  n=5 для продуктового рішення. `low` для поведінки завантажень, `medium` для
  коректності.
  Це розміри exploratory runs, а не доказ відсутності регресій. Критерії
  рішення, парність, виключення й межі невизначеності визначено у розділі 6.
- Орієнтири ціни (low, n=2): Sonnet implement ≈ $8, Opus implement ≈ $14,
  refactor ≈ $3/$5, review ≈ $4/$6. Ліміт сесії акаунта: ~70 сесій Opus за
  кілька годин уже вперлися в нього 2026-10-01 — великі прогони розносити.

## 2. Етапи

### Етап 0. Надійність бази перед платними порівняннями

1. Розібрати падіння `TestVetHook/in_a_git_checkout_only_lint_issues_new_since_HEAD_count`
   на початковому `hook_test.go:759`: хук повертав 0 без нового `errcheck`.
   **Виконано:** logical/physical cwd через symlink спричиняв неправильне
   `--new-from-rev=HEAD` diff-фільтрування. `cd -P`, явний вибраний config
   і видимість помилок запуску lint перевірені регресіями; деталі вище.
2. Перевірити парсери читання та метрики хостів (G5–G6), підготувати
   provenance/checkpoint (G7–G8) перед дорогими прогонами.
3. Preflight: версії CLI, точні model IDs, effective effort, фактично
   доступні skills/tools і схема трас. Почати з однієї фікстури на раннер;
   її модельний виклик також входить у бюджет.

Критерій виходу: встановлена причина падіння, виправлення з регресійним
тестом, зелений structural suite та негативні parser cases без false positives.
Перший пункт, G5 та Claude G6.1 contracts виконані; G6.2/G6.3–G8
і live preflight ще відкриті.

### Етап A. Відкрити GPT 6.1 і Grok 4.7 — найбільший невідомий

Передумова: перевірити доступ Codex до GPT 6.1; встановити opencode та
перевірити доступ до Grok 4.7. Раннери в `abrun` уже є (`-runner codex`,
`-runner opencode`; opencode вимагає `-model`).

1. **Чи допомагає пак узагалі.** `-arms no-skill,baseline` на implement,
   refactor і review, n=2. Для Claude `no-skill` лише показував, що фікстура
   має пастку; для GPT/Grok це перша відповідь на питання, чи читають вони
   скіли.
2. **Чи читають SKILL.md.** Codex не має Skill tool: скіл «спрацьовує», лише
   коли модель сама відкриває файл через shell. `abrun` уже рахує це
   (`codexSkillPath`, `commands`). Після G5 парсер звіряє completed/status,
   exit code і повний captured output із staged snapshot; лише такий доказ
   потрапляє в `Skills`. Це не доводить retention. Мета A2 лишається:
   роутер і `go-style-core` до першої правки в кожній implement-сесії;
   ordering/routing метрики належать G6.
3. **Перевірити головну обіцянку 1.28.0.** v1.27.1 проти 1.28.x: чи
   з'являються ідіоми з картки в коді (lint, `go fix` hunks) тепер, коли
   картка в `go-style-core`, а не в окремому файлі.
   Релізне порівняння міряє сумарний ефект змін між версіями; для висновку
   саме про перенесення картки потрібен варіант на однаковій базі.
4. **Довгі SKILL.md через shell.** Перевірити в трасах, чи не обрізає хост
   вивід `cat` для файлів 18–21 KB (`go-code`, `go-code-refactor`,
   `go-style-core`) і чи не читає модель їх кусками (`sed -n`, `head`). Якщо
   обрізає — найважливіше має стояти на початку файлу, або файл треба
   розбити.
   Відрізняти весь вміст у raw trace від того, що реально отримала модель;
   partial read з успішним exit code не є whole read.
5. **«Read when» буквально?** Порахувати читання references на сесію. Якщо
   GPT/Grok відкривають усе, що перелічено в Resource Routing, — етап D стає
   пріоритетом для них.
6. **Скрипти.** Чи підставляє модель шлях замість `<installed-skill-dir>`
   (Codex: `~/.agents/skills/...`) і скільки викликів падає з кодом 127.
7. **Блок для `AGENTS.md`.** `docs/PROJECT_INSTRUCTIONS.md` — єдина
   маршрутизація, яку ці хости мають без хуків. Потрібен прапорець `abrun`,
   що кладе блок у корінь scratch-модуля перед сесією. Порівняти `baseline`
   із блоком проти `baseline` без нього; `no-skill` без блоку — третій
   контроль. Не змушувати `no-skill` читати неіснуючі скіли. Підставляти шлях
   ізольованого arm home, зберігати хеш та rendered text інструкції.
8. **Review на інших хостах.** Задати окремий read-only tool profile;
   нинішні Codex/opencode сесії не отримують окремої review-політики.
   Перевірити digest дерева й відсутність доступу до hidden golden.
   Не називати умови однаковими, якщо shell, permissions або hooks різні.
9. **Opencode skill tool.** Парсер очікує події `skill`. Preflight має
   підтвердити їхню schema/status/result на фактичній версії CLI.
   Відсутня метрика — `unmeasured`, а не 0.

Ціна: встановити після preflight. A1, A3 й A7 — окремі порівняння.
За наявних 7 implement, 4 refactor і 6 review фікстур A1 із двома arms та
n=2 — 68 сесій на модель, ще до A3/A7. Кількість прогонів сама не є бюджетом.

### Етап B. Claude Code: сесії з shell

1. **Гейт зараховує перевірений результат edit-хука.** Спочатку потрібен
   явний запис кожної перевірки: `pass`, `fail`, `skipped`, `unavailable`,
   exit code, scope, config, версія інструмента й перевірений стан коду.
   Лише відповідний запис після останньої правки дозволяє `pass (hook)`.
   Package vet, lint лише нових findings і file-only gofmt не замінюють
   ширший gate репозиторію. `go build`, `go fix -diff`, потрібний race-check
   та `govulncheck` лишаються за гейтом, якщо немає придатного результату;
   test у хуку працює без `-race`. Очікування: −2–4 Bash-виклики на сесію
   лишається гіпотезою. Текст має лишатися
   умовним («де працює хук плагіна»), бо поточні GPT/Grok раннери його
   не підключають.
   Міряти: `-shell go`, обидві моделі, implement і refactor, n=2 — рядок
   `routing:`, категорії фактичних команд, wall time та вартість. Об’єднання
   команд зменшує кількість Bash-викликів без зменшення роботи; виклики
   самого хука у цьому лічильнику відсутні. Умови зарахування — у розділі 6.
2. **Відсутні owner-сигнали у стабах.** Prompt-хук уже сканує цілі файли;
   `net/http`, `database/sql` та `os/exec` уже розпізнаються з імпортів.
   Перевірити прогалини `sync`, `context`, `errors`, `log/slog`, сигнатури
   й контракти. Stub `pool` імпортує лише `context`, але контракт вимагає
   bounded workers; сам import-scanner не покриє `go-concurrency`.
   Імпорт — hint, а не безумовний owner: перевірити aliases, grouped,
   dot/blank imports, коментарі й рядки. Не розширювати edit-гейт автоматично
   через heuristic для prompt-hints. Міряти gate blocks/run, зайві loads,
   owners до першої правки та tokens/cost. B1 і B2 перевіряти окремо,
   потім їх поєднання.
3. **Правки до завантаження на Opus.** У 1.28.0 — 0 сесій, але раніше
   траплялося 2–7 на 14. Тримати під наглядом у кожному прогоні; якщо
   повернеться — шукати в note, не в гейті.

### Етап C. Текст, що вантажиться завжди (усі хости)

1. **Описи скілів** (7,2 KB у системному промпті кожної сесії). Найдовші:
   `go-code-refactor` 692 символи, `go-security` 559, `go-code` 540,
   `go-troubleshooting` 532, `go-logging` 473, `go-http` 438. Різати лише з
   `evalrun -kind trigger` і оновленням golden у
   `TestFrontmatterDescriptionsInvariant`.
   Цей golden перевіряє точний текст, а не trigger accuracy. Додати held-out
   негативні запити та перевірку без prompt-хука, який може приховати регресію.
2. **Компакція.** Claude Code після автокомпакції повертає перші ~5K токенів
   кожного скіла; `go-style-core` (≈6,9K) і `go-code` (≈7,6K) довші.
   Розставити секції так, щоб у перші 5K потрапляло те, що потрібне в кінці
   довгої сесії (Delete Pass, звіт, перші рядки картки). Для заміру потрібна
   довга фікстура — у поточних корпусах компакції не буває.
   Є також спільний бюджет 25K токенів для повторно прикріплених скілів:
   він заповнюється від найновіших, тому старий роутер може зникнути повністю.
   G4 має перевіряти і обрізаний скіл, і витіснений роутер із підтвердженою
   подією компакції. Джерело: [Claude Code skills](https://code.claude.com/docs/en/skills#how-skill-content-stays-in-context).
3. **Не робити без заміру:** винести «Writing New Code» (9,8 KB з 19,8 KB
   `go-code`) у reference. Fix-сесії платять за неї даремно, але Opus low
   пропускає непершорядкові інструкції — винесений текст може не читатися
   зовсім.

### Етап D. Холодні references (~70 KB)

Цінність — для хостів, що виконують «Read when» буквально (див. A5); на Claude
їх майже не читають (за 1.27.0 жодна сесія не відкрила жодного з п'яти
видалених файлів). Кандидати з аудиту:

| Файл / скіл | Економія | Зміст скорочення |
|---|---|---|
| `go-naming` references | −9 KB | `REPETITION.md`, `VARIABLES.md` дублюють SKILL.md |
| `go-functions/references/PRINTF-STRINGER.md` | −8,5 KB | переказ документації `fmt` |
| `go-documentation` | −5,5 KB | таблиця розділів, основи, що перевіряє revive |
| `go-context/references/PATTERNS.md` | −4,6 KB | Contents, immutability, Quick Reference |
| `go-functions/references/OPTIONS-VS-STRUCTS.md` | −4,3 KB | третя копія дерева рішень |
| `go-concurrency` | −4 KB | злити три references у `ADVANCED-PATTERNS.md` |
| `go-defensive` | −6,5 KB | `MUST-FUNCTIONS`, `PANIC-RECOVER`, `GLOBAL-STATE` — база Effective Go/Uber |

Кожне видалення файлу — оновити лічильники в `README.md`, `README.uk.md`,
`.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json`. Перевірка —
структурні тести і `evalrun` quality для зачеплених скілів.
Додатково оновити `docs/RULE_OWNERSHIP.md`, Resource Routing, міжскілові
посилання, fixtures та припущення тестів про видалені шляхи. Нормативне
правило має лишитися у своєму owner. Рідке читання не доводить безпечності
видалення рідкісного контракту; додати відповідні edge cases.

### Етап E. Review

1. **Компактний формат знахідок у `go-code-review`.** Вивід моделі — близько
   половини вартості review-сесії. Формат «один рядок: `file:line — дефект —
   fix (verified|plausible)`». Зберегти severity на рядку або заголовки
   Must Fix / Should Fix / Nits: scorer бере severity саме звідти.
   Наприклад: `Must Fix · file.go:42 — дефект і доказ — fix (verified)`.
   Перевірити parser формату до модельного A/B. Міряти на Sonnet, review,
   n=3 (історична оцінка ≈ $13): recall, must, as-must, baits/unkeyed,
   точність доказу/fix, output tokens і вартість. Scorer зараховує близьку
   file:line цитату без перевірки змісту дефекту; recall сам не доводить якість.
2. **Opus low не вантажить `go-code-review`.** Recall і так ~0,90, тож
   спершу з'ясувати, чи це взагалі коштує грошей або якості; лише тоді
   пробувати форму рядка роутера з 1.28.0 для review-note.

### Етап F. Рішення продукту (за вами)

1. **Характеризаційні тести в refactor.** Коли Opus їх пише (2 з 8 сесій),
   сесія дорожча на 30–50%. Contract Table ви лишили, бо потрібні тести —
   ймовірно, тут та сама логіка; підтвердити.
   До рішення лишити тестову політику незмінною. Оцінювати незалежність
   тестів і здатність ловити поведінкову мутацію; passing model tests можуть
   повторювати помилку реалізації. Розділяти початкові characterization tests
   та їх переписування після refactor.

### Етап G. Інфраструктура вимірювань

1. `-shell go` за замовчуванням для Claude implement/refactor-прогонів,
   що ухвалюють рішення; review — окремий профіль. Зберегти явний режим
   без shell для діагностики та історичних відтворень. Записувати effective
   profile; зміна default не робить старі звіти порівнянними.
2. Розклад токенів по гілках (`cache_w`, `cache_r`, `out`, ходи) прямо в
   підсумку `abrun` — зараз це робить скретч-скрипт.
3. `abrun` розпізнає «session limit» і зупиняє прогін або позначає сесію як
   невраховану, а не як помилку моделі (2 сесії втрачено 2026-10-01).
4. Довга фікстура для ефектів компакції (етап C2). Поточний `maxSteps=40`
   може завадити її досягти: потрібні окремі ліміт і бюджет. Траса має
   доводити подію компакції; довжина файла цього не гарантує.
5. **Виконано локально:** достовірний captured-output evidence Codex
   (статус вище, contracts у `codex_evidence_test.go`). Native schema/smoke
   ще перевірити у 0.3; host-specific routing лишається G6.
6. **Частково виконано:** Claude confirmed routing і availability (G6.1).
   Далі G6.2 для інших хостів та G6.3 для references, scripts, whole reads
   і категорій команд. Повний G6 ще не завершений.
7. Provenance: наявні plugin hashes доповнити станом fixture/golden/scorer,
   CLI/Go/linter versions, requested/effective model/effort, tools та instructions.
8. Типи збоїв, checkpoint після кожної сесії, припинення нових jobs при
   квоті та resume без повтору завершених jobs. JSON зараз пишеться наприкінці.
9. Usage і wall time: unavailable USD не дорівнює $0; Codex зараз повертає
   0. Окремо input/cached/cache write/read/output/turns, elapsed time й джерело
   usage. Значення runner-specific; API estimate і subscription quota різні.
   Включати в бюджет помилкові сесії, judge та repair.
10. Розділити implicit discovery runs та explicit router runs: перші міряють
    trigger, другі — поведінку після завантаження. Discovery preflight робити
    із фактичного cwd/model, не лише з arm home/default model.
11. Перевірити rendered listing із паком і зі сторонніми скілами: хости
    можуть скорочувати descriptions; Codex може також пропускати skills.
    Front-load scope/trigger words; обсяг 7,2 KB сам не доводить доставку.
    Джерела: [Codex skills](https://developers.openai.com/codex/skills),
    [Claude Code skills](https://code.claude.com/docs/en/skills).

## 3. Що враховувати для кожної моделі

- **Sonnet 5.5.** Вантажить скіли пакетом, слідує note; читає references
  рідко. Добрий детектор якості review (Opus там насичується). Ціна implement
  ≈ $0,25–0,29 на сесію (low, з Contract Table).
- **Opus 5.5.** На low виконує лише рядок роутера в note; паралельно шле
  кілька правок, і кожна блокується окремо — пропущене завантаження
  коштує повторного надсилання всіх правок. У review на low майже не вантажить
  скіл. Ціна implement ≈ $0,49–0,51 на сесію (low).
- **GPT 6.1 (Codex).** Поточний раннер не має Skill tool і не підключає
  хуки цього плагіна; скіл читається як файл через shell. Маршрутизацію
  мають забезпечити текст і блок `AGENTS.md`, а читання — підтвердити G5.
  Поведінку цільової моделі не виміряно.
- **Grok 4.7 (opencode).** Поточний раннер не підключає хуки Claude Code;
  парсер очікує skill tool. Фактичну механіку й поведінку цільової моделі
  ще треба перевірити після встановлення CLI.

## 4. Порядок і бюджет

| Порядок | Етап | Передумова | Ціна (оцінка) | Головна метрика |
|---|---|---|---|---|
| 0 | Structural baseline, G5–G8, preflight | — | локальні перевірки $0; smoke оплачується | достовірність вимірювань |
| 1 | B1 та B2 окремо, потім разом | етап 0; явний результат хука для B1 | перерахувати окремі порівняння | calls, owners, latency, quality |
| 2 | A1–A9 (GPT, Grok) | preflight раннерів і доступ до моделей | 68 сесій/модель лише для A1 | успішні reads до edit; незалежна якість |
| 3 | E1 (формат review) | етап 0; parser формату й семантичний контроль | ≈ $13 | severity, false positives, fix, вартість |
| 4 | C1 (описи) | етап 0; rendered listing і held-out запити | ≈ $5–10 (evalrun trigger) | trigger accuracy |
| 5 | D (холодні references) | результат A5 | ≈ $10–20 (evalrun quality) | без регресії quality |
| за потребою | G1–G4, G9–G11 | перед відповідним рішенням | розробка без model calls; валідаційні runs оплачуються | provenance, usage, compaction |
| — | F1 | ваше рішення | — | — |

B1 або B2 на двох Claude-моделях, implement/refactor, двох arms та n=2 —
88 сесій. Два окремі порівняння — 176; ще одне повне підтвердження поєднання
дає 264. Початкові ≈ $40 не покривають автоматично всі ці експерименти;
бюджет рахувати за corpus-specific ціною й обраним subset, не за самим n.

## 5. Чого не робити

- Не переформульовувати рядок роутера в prompt-note «на око»: три з чотирьох
  нових формулювань погіршили Opus 5.5 low.
- Не прибирати Contract Table заради економії — рішення ухвалене.
- Не зливати дрібні скіли (`go-naming`, `go-context`): ходу не економить,
  а обсяг завантаження росте.
- Не судити про зміну за golden при n=2: читати `routing:`, lint і `go fix`
  hunks; golden на implement у 5.5 майже завжди 100%.
- Не різати текст «для економії на Claude»: за трьома прогонами це не видно
  у вартості; різати заради хостів без хуків і компакції — і міряти там.

## 6. Додані вимоги до доказів і критерії рішення

### Що підтверджено кодом під час аудиту

Таблиця фіксує початковий аудит коміту `8f0fd78`; номери рядків стосуються
цього стану. Виправлення й поточний статус наведені в «Стан виконання».

| Знахідка | Статус і джерело | Зміна плану |
|---|---|---|
| Мовчання хука може означати skip або приховану помилку lint; lint scope — нові issues від HEAD | verified у початковому аудиті: [go-vet-on-edit.sh](../hooks/go-vet-on-edit.sh), рядки 77–84, 123–137; першопричина падіння згодом підтверджена й виправлена у 0.1 | B1 ще потребує явного результату й scope |
| Codex шукає шлях у будь-якій `command_execution`; schema не містить exit code/output | verified: [codex.go](../evals/cmd/abrun/codex.go), рядки 213–255 | G5 до A2/A4 |
| `commands`/`routing` для opencode не заповнюються; `routing` для Codex відсутній | verified: [main.go](../evals/cmd/abrun/main.go), рядки 1065–1087 | G6 до A5/A6 |
| Claude `loaded` стає true від Skill-виклику, а blocks рахує кожен PreToolUse exit 2 | verified: [routing.go](../evals/cmd/abrun/routing.go), рядки 26–83 | відділити attempted/successful loads та routing block |
| Owners скануються з цілого файла; HTTP/SQL/exec імпорти вже впливають на hints | verified: [go-code-routing.sh](../hooks/go-code-routing.sh), рядки 119–154 | B2 додає лише відсутні сигнали |
| Review severity береться з текстових labels; recall визначається близькістю цитати | verified: [review.go](../evals/cmd/abrun/review.go), рядки 204–282 | E1 зберігає severity й семантичний контроль |
| JSON-звіт пишеться тільки після всіх jobs; помилка appendTrace не піднімає error | verified: [main.go](../evals/cmd/abrun/main.go), рядки 530–570, 979–981 | G7–G8: checkpoint і видима відсутність trace |
| Codex setup перевіряє listing у home без target model, skills копіює в `.codex/skills` | verified: [codex.go](../evals/cmd/abrun/codex.go), рядки 93–122 | перевірити discovery з реального cwd/model та документованим `.agents/skills` |

Наявність парсера чи проходження його unit tests не підтверджує schema
поточної версії CLI. Для нового хоста потрібен один retained smoke trace
і перевірка його подій перед основним корпусом. Вихід CLI з кодом 0 також
не доводить завершення модельної роботи або наявності usage.

### Результати хука для B1

- Для кожного check зберігати команду, effective config/toolchain, scope,
  start/end, status, exit code й діагностику; findings відрізняти від
  infrastructure failure. `gofmt -l` потребує порожнього виводу, не лише exit 0.
- Зафіксувати digest перевірених входів до й після check. Нова правка,
  зміна конфігурації/залежностей або зміна пакета під час check робить
  попередній запис непридатним. Запис пізнього завершення старого хука
  не має перекривати результат для новішого коду.
- Окремі негативні cases: відсутній linter/timeout, disabled lint/tests,
  відсутній module, config/typecheck failure, timeout, кілька правок одного
  пакета, зміна іншого файла пакета, старий lint debt і новий finding.
- Придатний package check можна перевикористати для того самого стану
  пакета. New-issues-only lint позначати саме так; репозиторний full gate
  не вважається виконаним на його основі. Власник семантики гейта —
  `go-linting`; синхронізувати `go-code`, `go-code-refactor` та розділ
  The Edit Hook Record у `go-style-core`, щоб не залишити суперечливі правила.

### Достовірність host metrics для G5–G6

- Для Codex негативні fixtures: `ls`, `echo` або пошук шляху без читання,
  failed `cat`, started без completed, duplicate events, `head`/`sed`,
  truncated output, читання іншого файла з cross-links на SKILL.md.
  Позитивні fixtures мають підтверджувати вміст саме потрібного скіла.
- Successful read і whole read — різні метрики. Зберігати returned coverage
  і ознаки truncation; довжина raw trace не доводить model-visible delivery.
  Якщо schema не дає потрібних доказів, позначити невиміряне поле та
  перевірити траси вручну, не виводити whole-read процент.
- «Правка до завантаження» має розрізняти спробу й успішну правку та
  перевіряти конкретний router/style/owners, а не будь-який Skill-виклик.
  Slash/implicit activation потребує власного host evidence.
- Reference reads рахувати за успішністю, файлами, обсягом і релевантністю
  task. Script failures групувати за причинами; exit 127 — лише один
  випадок, ще є permissions, timeout, неправильний target та toolchain.
- Кількість tool calls, виконаних Go-команд і роботи edit-хука — окремі
  величини. Розбивка `gofmt`/vet/test/lint потрібна для висновку про B1.

### Протокол порівняння та захист від хибного висновку

1. До запуску визначити основну метрику, мінімальний корисний ефект,
   допустиму межу якості, fixtures, arms, model/effort/tools, n, seed та
   правила виключення/retry. Одна поведінкова зміна на порівняння.
2. Порівнювати результати парами за fixture/rep. `buildJobs` уже перемішує
   jobs, але repeated sessions однієї fixture не є різними незалежними
   задачами. Показувати per-fixture дельти й raw counts; невизначеність
   оцінювати з урахуванням fixture, не лише кількості сесій.
3. n=2/3 — smoke і вибір кандидата; n=5 — наступний evidence step, не
   автоматичне продуктове схвалення. Різниця <0,05 у наведеному Sonnet
   review є історичним шумом конкретного корпусу, не універсальним порогом.
   Якщо даних замало щодо визначеної межі, рішення — `inconclusive`.
4. Вартість показувати для всіх billed attempts; quality — із явними
   denominators для completed/valid/excluded. Квоту/auth/timeout/harness
   не зараховувати до model defects. Неповна пара не стає доказом переваги
   іншого arm; не retry лише гіршу гілку. Критерії exclusion не змінювати
   після перегляду результатів.
5. Прогони для discovery не примушують router. Explicit runs перевіряють
   поведінку після завантаження окремо. Claude `no-skill` проти плагіна
   міряє також ефект hooks/tools; він не ізолює сам текст скілів.
6. `-repair` лишити вимкненим для основних routing/cost рішень. З ним
   раннер запускає ще одну сесію й агрегує статистику; не трактувати repair
   як той самий початковий контекст. Rescore review зберігає hash нових keys;
   не переписувати вихідні метрики без позначення зміни scorer.
7. Ідіоми оцінювати за конкретними можливостями fixture й семантичною
   доречністю. Нуль `go fix` hunks не покриває всі ідіоми; passing lint і
   golden не доводять дотримання тексту. Для compact review вручну звірити
   дефект, доказ, severity та fix, включно з false positives.
8. Усі decision runs: `-keep`, `-out` за межами репозиторію, provenance й
   збережені traces/generated source/model tests. Не зберігати auth або
   секрети в артефактах. Відсутні дані позначати, не заміняти нулями.

### Критерії виходу з етапів

| Етап | Критерій для рішення |
|---|---|
| A | підтверджені host isolation та delivery; per-fixture якість проти контролю; окремий результат AGENTS-блоку |
| B1 | придатні результати check після останньої правки; менше повторних команд без втрати required gate і якості |
| B2 | менше routing blocks без систематичних зайвих owners або погіршення якості/вартості |
| C1 | exact-description tests і held-out trigger cases; rendered listing не втрачає потрібні сигнали |
| C2/G4 | підтверджена компакція та коректна поведінка після обрізання/витіснення; короткий корпус не заміняє цю перевірку |
| D | правила/links/counts узгоджені; edge-case quality для зачеплених owners збережена |
| E1 | severity, recall/must, false positives і правильність fix збережені; output/вартість оцінені окремо |
| F1 | тестову політику визначено окремим рішенням на підставі користі та вартості |

До платного запуску зафіксувати cap у сесіях і грошах: fixtures × arms × n,
далі окремо smoke, підтвердження, repair/judge та retries. Оцінки початкової
таблиці — історичні орієнтири, не гарантія ціни. Інфраструктурна розробка
без model calls не має inference cost; її модельна валідація має.

Для змін скілів/хуків/раннерів потрібні changelog під Unreleased і повний
race/shuffle structural suite; README/manifests змінювати при зміні counts.
Версію не піднімати для звичайного PR. План є документом планування;
реалізовані зміни та їхні перевірки відмічати у «Стан виконання».
