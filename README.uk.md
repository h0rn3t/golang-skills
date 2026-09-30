# Agent Skills для Go

**Українська** | [English](README.md)

Набір [Agent Skills](https://agentskills.io/) для ідіоматичного та
production-ready коду на Go 1.27. Пакет містить **24 модульних скіли**,
**70 довідкових файлів**, **11 вбудованих скриптів** і **5 шаблонів-ассетів**.
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

Для Codex цільова директорія — `~/.agents/skills/`.

## Інструкції для проєкту

Скіл завантажується, коли його обирає matcher хоста або модель. Для робочого
проєкту, де кожна правка Go має починатися з router, вставте короткий блок із
[`docs/PROJECT_INSTRUCTIONS.md`](docs/PROJECT_INSTRUCTIONS.md) у його
`CLAUDE.md` або `AGENTS.md`: спершу router, потім `go-style-core` і його
`CURRENT-GO.md`, потім лише ті owner-скіли, які потрібні задачі. У плагіні
Claude Code routing gate до того ж тримає кожну правку `.go`, доки ці
завантаження не зафіксовано (`GOLANG_SKILLS_ROUTING_GATE=off` його вимикає);
Codex, Copilot і Cursor хуків не запускають, тож там працює лише інструкція.

## Перевірка

`evals/` містить структурні Go-тести, fixtures, golden-тести та опційні
раннери `evalrun`/`abrun`. У конфігурації є **108 eval'ів на тригери** і
**62 eval'и на якість**. Результати модельних прогонів — локальні тимчасові
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

## Структура проєкту

```text
skills/       Директорії runtime-скілів.
hooks/        Хуки маршрутизації та перевірки після редагування.
agents/       Опційні агенти плагіна.
evals/        Структурні тести, fixtures, golden-тести та раннери.
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
