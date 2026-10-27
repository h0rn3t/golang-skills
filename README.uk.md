# Agent Skills для Go

**Українська** | [English](README.md)

Набір [Agent Skills](https://agentskills.io/) для ідіоматичного та
production-ready коду на Go 1.27. Пакет містить **24 модульних скіли**,
**73 довідкові файли**, **11 вбудованих скриптів** і **5 шаблонів-ассетів**.
Плагін Claude Code також містить агента `go-verify` та хуки маршрутизації й
перевірки після редагування.

Настанови походять із [Google Go Style Guide](https://google.github.io/styleguide/go/),
[Effective Go](https://go.dev/doc/effective_go), [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
і [Go Wiki CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments).

## Скіли

Пакет охоплює маршрутизацію, рефакторинг, рев'ю, HTTP, бази даних,
безпеку, стійкість, діагностику, стиль, імена, помилки, тестування,
конкурентність, generics, логування, документацію, продуктивність і структуру
пакетів. Актуальні назви — це директорії `skills/go-*/`.

## Вбудовані скрипти

**11 скриптів автоматизують** типові перевірки. Вони підтримують `--help`,
структурований вивід `--json` і задокументовані коди завершення; аналітичні
скрипти також мають `--limit`, а скрипти, що пишуть файли, вимагають `--force`.

Плагін містить `agents/go-verify.md`, хуки в `hooks/` і маніфести в
`.claude-plugin/`. У Go-проєкті хуки виводять драбину стриманості на старті
сесії та в кожного сабагента, крім `go-verify` і хостових Explore,
claude-code-guide та statusline-setup; `/go-code ultra <задача>` або `lite mode`
змінює її рівень до кінця сесії, `GOLANG_SKILLS_LADDER=lite|ultra` задає
рівень на старті сесії, а `off` вимикає її.

## Встановлення

### npx skills

```bash
npx skills add h0rn3t/golang-skills --all
```

### bun

`bunx` запускає той самий CLI без встановлення, або встановіть його один раз
глобально (`bun add -g skills` додає `skills` у `PATH`):

```bash
# запустити без встановлення
bunx skills add h0rn3t/golang-skills --all

# або встановити CLI один раз і користуватися ним
bun add -g skills
skills add h0rn3t/golang-skills --all
```

Оновлення працює так само: `bunx skills update` (додайте `-g` для глобальної
області).

### Codex, Copilot і Cursor

Той самий CLI встановлює скіли в кожен агент, який виявить; назвіть цілі
прапорцем `-a`, якщо встановлено кілька. Усі три читають проєктну директорію
`.agents/skills/`, а `-g` встановлює для користувача:

```bash
bunx skills add h0rn3t/golang-skills --all -a codex -a github-copilot -a cursor -g
```

| Агент | `--agent` | Глобальний шлях |
| --- | --- | --- |
| Codex | `codex` | `~/.codex/skills/` |
| GitHub Copilot | `github-copilot` | `~/.copilot/skills/` |
| Cursor | `cursor` | `~/.cursor/skills/` |

Після встановлення відкрийте нову сесію агента.

### Плагін Claude Code

```text
/plugin marketplace add h0rn3t/golang-skills
/plugin install golang-skills@golang-skills
```

### Ручне встановлення

Копіюйте повні директорії скілів разом із `references/`, `scripts/` і `assets/`:

```bash
cp -R skills/go-* ~/.claude/skills/
```

## Оновлення

### Claude Code

Оновіть знімок маркетплейсу, потім сам плагін і перезапустіть Claude Code, щоб
підхопити нову версію:

```bash
claude plugin marketplace update golang-skills
claude plugin update golang-skills@golang-skills
claude plugin list   # показує встановлену версію
```

### Codex, Copilot і Cursor

Оновіть глобальну інсталяцію і відкрийте нову сесію агента:

```bash
bunx skills update -g
```

Повторний `bunx skills add h0rn3t/golang-skills --all -g` теж працює і додає
скіли, що з'явилися після попереднього встановлення.

### Ручне встановлення

Оновіть checkout і замініть директорії скілів, а не копіюйте поверх них, щоб
видалені в апстрімі файли не залишалися:

```bash
git pull
rm -rf ~/.claude/skills/go-* && cp -R skills/go-* ~/.claude/skills/
```

Для Codex цільова директорія — `~/.codex/skills/` (глобальний шлях із таблиці вище) або `.agents/skills/` проєкту.

## Інструкції для проєкту

Скіл завантажується, коли його обирає matcher хоста або модель. Для робочого
проєкту, де кожна правка Go має починатися з router, вставте короткий блок із
[`docs/PROJECT_INSTRUCTIONS.md`](docs/PROJECT_INSTRUCTIONS.md) у його
`CLAUDE.md` або `AGENTS.md`: спершу router, потім `go-style-core` з його
карткою ідіом, потім лише ті owner-скіли, які потрібні задачі. У плагіні
Claude Code routing gate до того ж тримає кожну правку `.go` через Edit,
Write або MultiEdit, доки ці завантаження не зафіксовано; правку з shell
(`sed -i`, `gofmt -w`) він не тримає, а без `go` у PATH вимкнений і раз за
сесію про це повідомляє (`GOLANG_SKILLS_ROUTING_GATE=off` його вимикає). Хукам
не потрібно нічого, крім bash і Go: їхній парсер — невелика програма на
стандартній бібліотеці в `hooks/helper`, що збирається при першому запуску в
кеш-каталог користувача. Скіли, встановлені через
`npx skills` або скопійовані, хуків не мають, тож там працює лише інструкція;
хост, що завантажує `hooks/hooks.json` плагіна (Codex), запускає хуки з
вимкненим routing gate.

Після кожної правки `.go` post-edit хук без запиту запускає gofmt, `go vet`,
`go fix -diff`, тести пакета (`go test -short`) і golangci-lint з конфігом
репозиторію: виконується код репозиторію (`TestMain`, `init`). Неперевірену
зміну переглядайте з `GOLANG_SKILLS_EDIT_TESTS=off` і
`GOLANG_SKILLS_EDIT_LINT=off`. Конфіг лінтера, що завантажує плагіни
`custom:`, хук замінює бандловим baseline.

## Перевірка

`evals/` містить структурні Go-тести, fixtures, golden-тести та опційні
раннери `evalrun`/`abrun`. У конфігурації є **110 eval'ів на тригери** і
**75 eval'ів на якість**; `evalrun -kind quality -ids 65,66 -n 3 -no-skill`
повторює вибрані кейси без плагіна — це контроль, з яким звіряють новий кейс;
`-go-doc` дозволяє сесіям перевірити API через `go doc`.
Результати модельних прогонів — локальні тимчасові
артефакти й не зберігаються в репозиторії. Набір також перевіряє кожен
`SKILL.md` на відповідність специфікації Agent Skills
(`TestStructure`), тож для перевірок потрібен Go, а Node.js — ні.

Основні перевірки:

```bash
(cd evals && go test -count=1 -race -shuffle=on ./...)
bash -n hooks/*.sh
golangci-lint config verify --config skills/go-linting/assets/golangci.yml
```

Набір release-перевірок описаний у
[`docs/RELEASE_CHECKLIST.md`](docs/RELEASE_CHECKLIST.md).

`plugin-evals/` містить кейси для `claude plugin eval`: він завантажує плагін
в ізольованій сесії `claude -p` і порівнює прогони з плагіном і без нього.
Прогони витрачають токени й запускаються лише вручну; деталі в
[`plugin-evals/README.md`](plugin-evals/README.md).

## Бенчмарки

`abrun` міряє, що скіл робить із кодом, який модель пише, а не що вона про
нього говорить. Кожен fixture містить задокументовані, але нереалізовані
декларації та прихований golden-тест; сесія редагує справжні файли в
тимчасовому модулі, і harness оцінює результат. Дві arms відповідають на одне
питання: `baseline` завантажує це дерево скілів з ізольованого `HOME`, а
`no-skill` — порожній, тож різниця — це скіли і більше нічого.

Прогони 2026-10-24 нижче виконували corpus через CLI opencode, який віддає
обидві моделі з одного провайдера. Усі прогони зібралися й пройшли свій
golden-тест в обох arms, тож різниця — у формі, а не в коректності:

| Corpus | Модель | Δlines baseline → no-skill | Lint clean | Golden | $/run baseline → no-skill |
|---|---|---|---|---|---|
| implement (12) | DeepSeek V4.1 Flash | 58.2 → 77.5 (−25%) | 32/36 → 23/36 | 36/36 → 36/36 | 0.0136 → 0.0068 |
| implement (12) | Muse Spark 1.3 | 52.7 → 76.0 (−31%) | 29/36 → 21/36 | 36/36 → 36/36 | 0.0075 → 0.0034 |

Δlines — це кількість production-рядків, які сесія залишила, мінус stub-пакет,
з якого вона стартувала, тому менше число означає щільнішу реалізацію того
самого контракту. Lint clean рахує прогони без findings під бандловим
`golangci.yml`. Обидва прогони використовували `-corpus implement -arms
baseline,no-skill -n 3 -j 4`: 12 fixtures × 2 arms × 3 повтори = 72 сесії
кожен, на Go 1.27 і opencode 2.1.1 з `--standalone`.

Колонку `skill` у локальному звіті на цьому раннері читайте обережно:
opencode не емітить виклик `skill`, який рахує `abrun`, тож trigger rate і
routing там незмірені. Ці числа порівнюють home зі скілами і home без них —
вони не доводять, що модель викликала конкретний скіл. Виклики скілів
звітує раннер Claude Code.

Звіти — локальні тимчасові артефакти; відтворюйте їх тими самими флагами,
а не покладайтеся на збережену копію:

```bash
cd evals
go run ./cmd/abrun -corpus implement -runner opencode \
  -model opencode-go/muse-spark-1.3-contributor \
  -arms baseline,no-skill -n 3 -j 4 -out ab.json
```

## Структура проєкту

```text
skills/       Директорії runtime-скілів.
hooks/        Хуки маршрутизації та перевірки після редагування.
agents/       Опційні агенти плагіна.
evals/        Структурні тести, fixtures, golden-тести та раннери.
plugin-evals/ Кейси для `claude plugin eval`.
source/       Знімки upstream-настанов із provenance і ліцензіями.
docs/         Політика авторингу, власності правил, скриптів і release.
```

`COMPATIBILITY.md` — джерело правди для версієзалежних настанов Go.
`docs/RULE_OWNERSHIP.md` фіксує власника кожного правила, а
`docs/SCRIPT_JSON_CONTRACTS.md` — контракти виводу скриптів.

## Go 1.27

Скіли орієнтовані на Go 1.27 і підтримують Go 1.26 та 1.27. Перевіряйте
версієзалежні твердження встановленим toolchain і оновлюйте
[`COMPATIBILITY.md`](COMPATIBILITY.md) при зміні baseline.

## Ліцензія

Файли, написані для проєкту, мають ліцензію Apache-2.0. Upstream-знімки та
похідні настанови перелічені в
[`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md).
