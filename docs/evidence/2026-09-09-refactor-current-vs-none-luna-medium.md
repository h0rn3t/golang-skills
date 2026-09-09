# Один парний прогін: поточний Go skill проти no-skill

2026-09-09. GPT-5.6-Luna medium, Codex CLI 0.153.4, один `gw3` input × два плеча × один повтор = **2 сесії**. У цій парі поточний skill дав компактніший результат за збережених перевірених контрактів. Це один позитивний приклад, не доказ стійкого або загального ефекту.

## Протокол

- Той самий вихідний код, prompt, runner, model, effort, seed 1, j=2, timeout 10m, без repair та без повторів за результатами.
- `baseline`: поточний повний Go-plugin із `go-code-refactor`; `no-skill`: ізольований HOME без Go-скі́лів. Це не абляція одного файла при незмінному наборі решти skills.
- Skill SHA-256: `1ee8827c4ca7e69705cc164b134cb4adf05443b290a1cb76974645031be61414`.
- Використано перевірений snapshot `/tmp/go-refactor-dual-20260909/repo`: всі файли plugin та runner перед запуском побайтово звірено з поточним checkout. Попередні 113 runner tests і input golden залишаються застосовними до незміненого snapshot; їх не запускали повторно.
- `gw3` обрано до запуску як дослідницький приклад із помітними попередніми відмінностями структури. Він не є репрезентативною вибіркою всього Go-коду. Seed задає порядок jobs, не гарантує детермінованого декодування моделі.

Обидва плеча отримали однаковий refactor prompt. Додаткова умова однакова: якщо `$HOME/.codex/skills/go-code-refactor/SKILL.md` існує, прочитати його; якщо ні — працювати без skills. Пошук альтернативних інсталяцій заборонено. Зміст інструкцій скорочення не додано до no-skill prompt.

## Результати

| Метрика | Без Go-скі́лів | Поточний skill |
|---|---:|---:|
| Build | pass | pass |
| Independent golden | pass | pass |
| Δphysical LOC | +21 | −2 |
| Δcode LOC | +16 | −2 |
| ΔGo tokens | +88 | −17 |
| Нові функції | +3 | 0 |
| Δbranches | +1 | −1 |
| Втрата наявних doc-коментарів | немає | немає |
| Нові model-authored tests | немає | немає |

Результат зі skill має на **23 фізичні рядки, 18 рядків коду та 105 Go-токенів менше**, ніж результат без skill. Існуючі doc-коментарі збережені в обох outputs.

Без skill модель винесла `accountsHandler`, `isAccountPath`, `filterAccounts`, додала `accountsPath` та окремий sentinel `errInvalidActiveQuery`. Частина helpers відокремлює змістовні операції; сам їх приріст не доводить гіршої читабельності.

Зі skill модель об'єднала перевірки `active` в одну умову й прибрала вкладену повторну перевірку, не додаючи helpers. Також перейменувала route booleans. Позитивний висновок стосується компактності та відсутності регресій у наявному golden; сліпої оцінки читабельності не проводили.

## Походження та помилка raw-лічильника

Preflight runner підтвердив відсутність Go-skills у no-skill HOME. У trace control перша перевірка шляху повернула `ABSENT`; модель явно продовжила без skill, наступні команди читали тільки source/module та запускали Go checks. У baseline trace успішно прочитано поточний файл: його перші 240 рядків точно збігаються зі snapshot, включно з актуальним gate і виправленим baseline workflow.

Raw JSON помилково показує `skills=[go-code-refactor]` навіть у no-skill: parser зараховує згадку шляху в умовній shell-команді, хоча файл відсутній і гілка читання не виконалася. Тому raw `skill 100%` у control не використовується як доказ завантаження. Перевірено фактичний stdout усіх п'яти команд control: жодного прочитаного skill немає. Команди та повні traces збережені; raw JSON не виправляли заднім числом.

## Межі

n=1 на одному вибраному input не дає статистичного висновку, не оцінює середню користь skill і не доводить повну поведінкову еквівалентність за межами independent golden. Немає грошової вартості від Codex CLI; raw нуль не означає безплатний запуск. Usage збережено в sources/traces. Код навички, runner та цільового репозиторію цим прогоном не змінювався.

## Артефакти

- [Raw JSON](2026-09-09-refactor-current-vs-none-luna-medium.raw.json).
- [Input, outputs, diff, commands, provenance та usage](2026-09-09-refactor-current-vs-none-luna-medium.sources.json).
- [Повні traces двох сесій](2026-09-09-refactor-current-vs-none-luna-medium.traces.tar.gz).
- [Маніфест і точна команда](2026-09-09-refactor-current-vs-none-luna-medium.manifest.json), [prompt](2026-09-09-refactor-current-vs-none-luna-medium.prompt.txt), [execution timing](2026-09-09-refactor-current-vs-none-luna-medium.execution.json).
