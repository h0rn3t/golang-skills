# Повтор go-code-refactor: 10 Luna + 10 Sonnet 5 medium

2026-09-09. Виконано по **10 сесій із підтвердженою навичкою** для GPT-5.6-Luna medium та Claude Sonnet 5 medium. Build і independent golden пройдено **20/20**. У цій вибірці Luna не збільшила рядки коду; Sonnet частіше додавав helpers. Водночас Luna один раз видалила корисну документацію, тому числовий виграш не є повним успіхом якості.

## Протокол

- Поточний `skills/go-code-refactor/SKILL.md` із виправленнями пояснень baseline/diff/tests/paths, **без** неприйнятого token-aware candidate block.
- Skill SHA-256: `1ee8827c4ca7e69705cc164b134cb4adf05443b290a1cb76974645031be61414`.
- Повний plugin SHA-256 у всіх початкових і повторних запусках: `9e63881398c39cdb6c7340c69ec8fa5b98f327b77df167be219220de119b492b`.
- Snapshot: `/tmp/go-refactor-dual-20260909/repo`, базовий commit `ebc00ae`; поточні незакоммічені зміни runner скопійовано та зафіксовано patch/manifest. Робочий код runner і skill цим повтором не змінено.
- Ті самі п'ять passing gateway inputs `gw0`–`gw4`, два повтори на вхід, seed 1, лише baseline arm, без repair; j=2 на початковий runner. Повтори стартових збоїв Sonnet виконувалися з паралельністю 2.
- GPT-5.6-Luna: Codex CLI 0.153.4, `-effort medium`. Claude Sonnet 5: Claude CLI 2.1.263, `-effort medium`. Go 1.27.1.
- Перед запуском: 113 runner tests пройшли, усі 5 вихідних fixtures пройшли independent golden.

Завдання в обох випадках: `Refactor the Go package in ./%s so it reads better. Keep observable behavior identical: the exported API, error texts, and rendered output must not change. Apply the changes to the files.` Додано явне правило читання тільки призначеної копії навички. Точні початковий і уточнений prompts збережені окремо.

## Результати 20 сесій із навичкою

Середні враховують усі десять результатів моделі, зокрема обґрунтовані empty diff. Це важливо: поточний raw runner не включає незмінені файли у свої valid-середні.

| Метрика | Luna medium | Sonnet 5 medium |
|---|---:|---:|
| Прочитано призначений skill | 10/10 | 10/10 |
| Build / independent golden | 10/10 | 10/10 |
| Physical LOC не зростає | 9/10 | 5/10 |
| Code LOC не зростає | 10/10 | 6/10 |
| Середня Δphysical | −3.2 | +3.9 |
| Середня Δcode | −2.8 | +2.6 |
| Середня ΔGo tokens | −8.5 | +23.5 |
| Середня Δфункцій | 0 | +1.4 |
| Середня Δтипів | 0 | +0.1 |
| Справжній empty diff | 0/10 | 2/10 |
| Модель додала тести, які пройшли у harness | 5/10 | 1/10 |
| Виявлена втрата корисної документації | 1/10 | 0/10 |

Physical LOC включає всі рядки production `.go`, Code LOC — лише рядки з Go-токенами, включно з рядками multiline literals; test files виключено. Go token count виключає коментарі, крапки з комою та коми. Він допомагає відрізняти скорочення коду від перенесення рядків.

### Δphysical за входом і повтором

| Вхід | Luna: rep 0, rep 1 | Sonnet: rep 0, rep 1 |
|---|---|---|
| gw0 | -5, -7 | -5, -5 |
| gw1 | -5, -4 | +2, -3 |
| gw2 | +0, -2 | +2, +0 |
| gw3 | +1, -1 | +28, +18 |
| gw4 | -6, -3 | +2, +0 |

Зростання Sonnet найбільше на `gw3`: +28/+18 рядків і +7/+3 функції. У першому випадку він замінив closure окремим `accountsHandler`, виніс paths у константи й розділив маршрути на методи. Частина цих меж змістовна, однак це порушує безумовний gate поточної навички. Число функцій саме по собі не доводить гіршу читабельність.

Sonnet `gw4/1` справді відновив початковий файл після оцінки невигідного helper extraction; `gw2/1` залишив файл без змін після аудиту. Обидва прочитали правильний skill і є прийнятними empty-diff результатами. Raw runner позначає їх `[ERR]` через `Edited=false`, хоча транспортної помилки й регресії немає. Це не причина повторювати такі сесії або викидати їх із середнього.

### Недолік результату Luna

`gw4/0`: Δphysical −6, Δcode 0, Δtokens 0. Увесь виграш у рядках — видалений абзац doc comment про відкритий інтернет без proxy, анонімні клієнти, повільні mobile connections та утримання з'єднань. Це пояснення умов роботи й server timeouts. Решта diff лише перейменовує `filtered` на `matchingAccounts`.

Отже, physical gate пройдений ціною корисної документації. У цьому випадку подвійний physical/code gate також би не відхилив зміну: код не виріс. Потрібна окрема вимога й перевірка збереження змістовних пояснень. Golden такого дефекту не бачить. У решті 19 outputs порівняння коментарів не виявило аналогічної втрати.

## П'ять стартових збоїв і їх повторення

Фактично запущено **25 сесій: 10 Luna та 15 Sonnet**. П'ять початкових Sonnet-сесій не завантажили skill і не редагували source. Причина — моя початкова інструкція знайти каталог `.eval-plugin-*` через Glob: інструмент шукає файли, тому повертав порожній результат, хоча staged plugin існував. Сесії далі пробували зовнішній `.codex` шлях, недоступний у restricted mode, і зупинялися.

Повторено лише ці п'ять slots: `gw0/1`, `gw3/1`, `gw2/0`, `gw1/1`, `gw2/1`. Уточнення вимагало штатний `Skill(skill="golang-skills:go-code-refactor")`, який завантажує ту саму призначену копію. Текст завдання і skill не змінено. Повторів через великий diff, behavioral failure або empty diff не було.

Raw початковий report, усі п'ять retry reports і явне slot mapping збережені; відкинуті спроби не приховані й не враховані як незмінений коректний код. Результати після уточнення включені тому, що skill завантажився, незалежно від обсягу змін.

## Перевірка походження та межі порівняння

Для Luna live-файл у виділеному `.codex/skills` перевірено за SHA-256, а в кожному trace підтверджено успішне читання цього шляху та актуальних фрагментів gate/baseline. Для Sonnet перевірено registered plugin у `system.init`, SHA-256 staged SKILL.md і успішний `Skill` або прямий `Read`. Один початковий Sonnet-прогін читав правильний файл через Read, тому raw список `skills=[]` не означав відсутність навички. Сторонніх читань замість призначеного skill у фінальній двадцятці не виявлено.

Це **порівняння model+runner**, а не чистий ефект моделей: Claude runner має `Skill,Read,Glob,Grep,Edit,Write` без Bash; Codex має shell. Sonnet не міг виконати Go checks сам, але незалежний harness перевірив його результати. Native Skill invocation у повторах також відрізняється від початкового filesystem-loading prompt. Без no-skill/control плеча тут не можна оцінити користь skill саму по собі. Два повтори на кожен із п'яти gateway inputs не встановлюють загальну перевагу на інших Go-задачах.

## Витрати

CLI Sonnet повідомив $3.2822208 за десять включених сесій. П'ять стартових відмов додали $0.3086682, разом $3.5908890 за 15 спроб. Codex не повернув dollar cost; нулі у raw summary не означають безплатний запуск. Token usage та execution timing збережені в raw traces/метаданих; додаткову оцінку ціни не робили.

## Артефакти

- [Маніфест і hashes](2026-09-09-refactor-repeat-luna-sonnet-medium.manifest.json), [snapshot patch](2026-09-09-refactor-repeat-luna-sonnet-medium.snapshot.patch).
- [Luna raw](2026-09-09-refactor-repeat-luna-sonnet-medium.luna.json), [Sonnet initial raw](2026-09-09-refactor-repeat-luna-sonnet-medium.sonnet.json).
- [Когорта 20 сесій із посиланням на raw slots](2026-09-09-refactor-repeat-luna-sonnet-medium.cohort.json), [summary](2026-09-09-refactor-repeat-luna-sonnet-medium.summary.json).
- [Inputs, outputs, model tests, provenance та usage](2026-09-09-refactor-repeat-luna-sonnet-medium.sources.json), [повні 25 traces](2026-09-09-refactor-repeat-luna-sonnet-medium.traces.tar.gz).
- [Початковий prompt](2026-09-09-refactor-repeat-luna-sonnet-medium.prompt.txt), [native-Skill уточнення](2026-09-09-refactor-repeat-luna-sonnet-medium.retry-prompt.txt), [retry slots](2026-09-09-refactor-repeat-luna-sonnet-medium.retry-slots.json).

Імена raw reports у cohort/sources відносні до experiment directory; у цій папці до них додано префікс `2026-09-09-refactor-repeat-luna-sonnet-medium.`. П'ять retry reports, команди запуску й аудит збережені поруч. Зміни навички або runner за результатами цього повтору не застосовувалися.
