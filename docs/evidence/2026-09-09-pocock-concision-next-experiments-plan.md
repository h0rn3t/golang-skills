# План следующих concision-экспериментов по подходам Matt Pocock

Дата: 2026-09-09.

## Цель

Сделать refactor-результаты GPT-5.6-Luna `medium` короче без ухудшения
читаемости, публичного поведения и проверяемых контрактов.

Ветка [`1.2.0`](https://github.com/h0rn3t/golang-skills/tree/1.2.0) на commit
`18e701e3725dcb14cf4105932817794924c7e190` содержит проверенный prose-вариант
`Concision Gate`. Его пока следует считать экспериментальным:

- non-growth вырос с 7/25 до 16/25;
- медиана production delta изменилась с +5 до −2 строк;
- candidate всё равно вырос в 9/25 сессий, максимум на 33 строки;
- один candidate изменил `nil → []` на `nil → null`;
- после исключения конфликта helper-а hidden-теста behavioral golden одинаков:
  24/25 на каждом плече.

## Принципы

План переносит следующие идеи Pocock:

- ясные и проверяемые completion criteria;
- tight red/green feedback loop;
- model-invoked и user-invoked skills как разные бюджеты;
- разделение implementation и review;
- независимые Spec и Standards/Concision оси review;
- pruning по наблюдаемому поведению, а не по длине документа.

Источники:

- [writing-for-agents](https://github.com/mattpocock/skills/blob/3cca18b368ae95cdbdebbff572ccafa662551015/skills/productivity/writing-for-agents/SKILL.md);
- [skill mechanics](https://github.com/mattpocock/skills/blob/3cca18b368ae95cdbdebbff572ccafa662551015/skills/productivity/writing-for-agents/SKILL-MECHANICS.md);
- [diagnosing-bugs](https://github.com/mattpocock/skills/blob/3cca18b368ae95cdbdebbff572ccafa662551015/skills/engineering/diagnosing-bugs/SKILL.md);
- [code-review](https://github.com/mattpocock/skills/blob/3cca18b368ae95cdbdebbff572ccafa662551015/skills/engineering/code-review/SKILL.md);
- [codebase-design](https://github.com/mattpocock/skills/blob/3cca18b368ae95cdbdebbff572ccafa662551015/skills/engineering/codebase-design/SKILL.md).

## Этап 0. Исправить измерительный стенд

Модельные сессии не требуются. **Выполнено 2026-09-09.**

- [x] Использовать collision-resistant имена helper-ов в hidden golden. Не
      считать столкновение generated helper `serve` с helper-ом теста
      behavioral regression. → `serve` → `goldenServe`, `members` →
      `goldenMembers`, `assertJSONEqual` → `goldenAssertJSONEqual`,
      `fakeStore` → `goldenStore`, `errTransport` → `goldenErrTransport`.
      Повторение блокирует `TestGoldenHelpersAreCollisionResistant`: каждое
      package-level имя в golden обязано содержать `golden`.
- [x] Зафиксировать определение production LOC: физические строки в production
      `*.go`, без `_test.go`, как считает `abrun`. → комментарий у
      `metrics.Lines` и раздел «Production LOC» в `evals/ab/README.md`.
- [x] Добавить в отчёт отдельные поля:
      `line_gate_pass`, `empty_diff`, `reported_counts`, `behavior_failure` и
      `harness_failure`. → плюс `commands` и `trace_path`. Harness failure
      печатается как `HRN`, а не `ERR`, и не попадает ни в одно среднее.
- [x] Сохранять Codex JSONL traces, чтобы отличать выполненное измерение от
      декларации в финальном сообщении. → `trace.jsonl` в корне scratch-дерева
      для любого runner-а, сохраняется с `-keep`; `codexCommands` считает
      фактические shell-вызовы.
- [x] Покрыть новые классификации unit-тестами `evals/cmd/abrun`. →
      `TestClassifyGolden`, `TestReportedCounts`, `TestCodexCommands`,
      `TestGoldenHelpersAreCollisionResistant`,
      `TestRunOneSeparatesHarnessCollisionFromRegression`.

**Done when:** один fixture с намеренной коллизией классифицируется как harness
failure, реальный HTTP-регресс — как behavior failure, а сохранённый trace
показывает фактические команды модели.

**Проверено:** `TestRunOneSeparatesHarnessCollisionFromRegression` прогоняет три
случая через реальный `runOne` и toolchain — production helper, столкнувшийся с
именем overlay, даёт `harness_failure` и статус `HRN`; сломанный контракт даёт
`behavior_failure`; сессия без правок даёт `empty_diff`. В каждом случае
`trace.jsonl` сохранён и читается. Точный текст HTTP-регресса из прогона 5×5
(`POST /accounts = 200, want 405`) классифицируется в `TestClassifyGolden`,
там же `undefined: NewServer` остаётся behavior failure: сломанный публичный
контракт принадлежит модели, а не стенду.

## Этап 1. Согласовать верхнеуровневый prompt с Concision Gate

Текущий prompt заканчивается `Apply the changes to the files`, тогда как skill
разрешает пустой diff. В 25/25 candidate-сессиях модель что-нибудь изменила.

Сравнить на пяти фиксированных gateway inputs:

- `control`: текущий refactor prompt;
- `candidate`: prompt ниже с тем же plugin из ветки `1.2.0`.

```text
Review the package for a shorter, equally clear implementation.
Change it only when production code becomes no longer and behavior remains
identical. If no such improvement exists, leave the files unchanged.
```

Параметры: GPT-5.6-Luna, Codex, effort `medium`, seed 1, один прогон на input,
два плеча — 10 сессий.

**Go:** candidate даёт golden 5/5, `Δlines <= 0` минимум в 4/5, хотя бы один
настоящий empty diff и ни одного случая code golf.

**Stop:** candidate проигрывает control по golden, растёт минимум в двух
сессиях или сокращает validation/error/security behavior.

**Пройден 2026-09-09.**
[Отчёт](2026-09-09-stage1-prompt-alignment-gpt-5.6-luna-codex.md): golden 5/5 на
обоих плечах, line gate 4/5 control против 5/5 candidate, первый настоящий empty
diff за серию (`gw1`, подтверждён trace-ом: `go test`, `go vet`, `gofmt -d`),
ни одного Stop-критерия. Средняя Δlines −1.60 против −2.80; парный эффект
−1.20 строки при exact randomization **p = 0.6250**, то есть разница в строках
между prompt-ами не установлена и переносу в репозиторий не подлежит. Один
случай на границе code golf: `gw4` дал −6 строк, продублировав список путей
между guard и `switch` — материал для Concision-оси этапа 5. Плагин обоих плеч
идентичен по SHA-256, менялось только prompt.

## Этап 2. Проверить feedback repair

Использовать десять уже сохранённых проблемных Concision Gate outputs:

- девять результатов с `Δlines > 0`;
- один результат с `nil → null`.

Для каждого результата запустить два repair-плеча:

- `control`: обычное `Re-review for concision while preserving behavior`;
- `candidate`: точный feedback с исходным/current LOC и independent golden
  failure.

Пример candidate feedback:

```text
Starting production LOC: 132
Current production LOC: 148
Gate: FAIL, growth +16

Independent contract failure: nil accounts encode as null, want [].

Repair the implementation or restore the starting version.
Done when LOC <= 132 and the contract test passes.
```

Итого: 10 outputs × 2 плеча = 20 Luna medium сессий.

**Go:** candidate исправляет минимум 8/10, не создаёт новых golden failures и
заметно превосходит generic repair.

**Stop:** feedback не лучше control либо после repair остаётся новый behavioral
regression.

**Пройден 2026-09-09.**
[Отчёт](2026-09-09-stage2-feedback-repair-gpt-5.6-luna-codex.md): **10/10**
починено против **1/10** у generic repair, exact paired sign test
**p = 0.00391**; ноль новых golden failures против одного у control (`rp1`
сломал `POST`/`HEAD /accounts` на 200 вместо 405); регрессию `nil → null`
candidate исправил, control нет. Ни одного чистого откката — все десять
результатов отличаются от стартовой реализации. Candidate дешевле: 8.4 команды
на сессию против 13.8.

Главная находка не в счёте починок. **Определение LOC несущее:** `rp5` содержит
138 физических строк и 101 строку без пустых и комментариев; generic repair
ответил «Lean already… Production code: 101 lines» и не изменил ничего. Модель
не выдумала число — она выбрала другую защитимую конвенцию и по ней была права,
пока харнесс писал рост +12. Это retroactively объясняет часть роста в прогоне
5×5: гейт требует «no more production lines», не определяя строку. В плече с
числами в терминах харнесса заявленное моделью итоговое число совпало с
харнессом в 10/10.

Что из этого **не** следует: правка скилла. Плечи различаются содержанием
информации, а не формулировкой, и никакой текст в `SKILL.md` не может сообщить
модели измеренный LOC её собственного результата. Это может только петля —
её и проверяет этап 3.

## Этап 3. Проверить автоматический однократный repair loop

Выполнять только после успешного этапа 2.

```text
model edit
→ measure production LOC
→ run independent golden
→ return exact failures
→ one repair turn
→ measure and test again
```

Сравнить:

- prose-only `Concision Gate` из `1.2.0`;
- тот же prose плюс один автоматический repair-turn.

Параметры первого раунда: пять inputs × три повтора × два плеча = 30 сессий.

**Go:** golden 15/15, `Δlines <= 0` минимум в 14/15, медиана не выше нуля,
после repair нет результата больше `+5`, ручной review не находит code golf.

**Expand:** только после выполнения всех критериев повторить с `n=5`.

**Stop:** repair loop маскирует ошибки, меняет golden или сокращает код ценой
понятности/полезной локальности.

**Выполнен частично 2026-09-09, Expand заблокирован.**
[Отчёт](2026-09-09-stage3-repair-loop-gpt-5.6-luna-codex.md): четыре критерия из
пяти выполнены с отрывом — гейт **8/15 → 15/15**, медиана **+0 → −3**, максимум
**+17 → +0**, golden 15/15, эффект −3.80 строки при within-input randomization
**p = 0.01572**, ни одного откката. Петля сработала в 6 прогонах и во всех шести
вывела в зелёное, включая `gw0 #0`, где она поймала **ту же регрессию
`nil → null`**, которую прогон 5×5 отгрузил как behavior failure.

Пятый критерий провален: `gw2 #0` и `gw2 #1` взяли гейт, удалив из doc-коммента
`NewServer` абзац про открытый интернет — обоснование таймаутов — и добавив
взамен хелперы (Δкомментариев −8, Δкода +6). Причина в определении LOC: гейт
считает физические строки вместе с комментариями, поэтому документация легально
оплачивает код. Системным это не является (13/15 сократили именно код, −1.73 из
−2.80 эффекта — код), но Stop-критерий назван прямо.

Перед Expand закрыть дыру и перепрогнать те же 30 сессий:

- [x] Считать строки кода отдельно от комментариев; запретить рост кода,
      оплаченный удалением документации. Новое поле `Δcode`; `Δlines` не менять,
      иначе рвётся сравнимость с прошлыми отчётами. → `metrics.Code` считает
      физические строки, содержащие хотя бы один Go-токен: пустая строка и
      строка только с комментарием не считаются, строка кода с хвостовым
      комментарием считается один раз, многострочный литерал — на каждой своей
      строке. Новый флаг `code_gate_pass` (`Δcode <= 0`); гейт взят только при
      обоих флагах, `line_gate_pass` сохраняет прежнее определение. На пяти
      реальных фикстурах токенный счёт совпал с наивным non-blank non-comment
      строка в строку (`dispatch` 55, `pricing` 115, `report` 43, `store` 52,
      `gateway` 11), а блочные комментарии и raw-строки, где эти два способа
      расходятся, закрыты `TestCodeCount`.
- [x] Назвать это ограничение в тексте feedback-а. → feedback печатает оба числа
      (`Starting production LOC: 132 physical, 96 code`), вердикт с обеими
      дельтами (`Gate: FAIL, physical -1, code +6`), определение каждого счёта и
      строку «Both counts gate: deleting a comment does not pay for a line of
      code, and documentation that explains a decision stays». Условие срабатывания
      петли — `line_gate_pass && code_gate_pass && probePass`, поэтому обмен
      документации на код запускает repair-turn сам по себе.
- [ ] Перезапустить этап 3; Expand только после чистого пятого критерия.
      **Заблокирован не дырой, а единицей измерения — см. ниже.**

Дыра закрыта в харнессе, а не в тексте скилла. `TestRunOneCodeGateBlocksDocsPayingForCode`
проводит через `runOne` ровно случай `gw2`: файл стал физически короче на 1
строку, кода в нём стало на 3 строки больше. Физический гейт этот обмен
пропускает — и обязан пропускать, иначе прошлые отчёты нечем читать; отказывает
code-гейт, и с `-repair` петля срабатывает именно на нём (`pre_repair_golden`
остаётся `true`).

### Смоук на `gw2` и переоценка 30 сессий: единица измерения

[Отчёт](2026-09-09-stage3-codegate-smoke-gpt-5.6-luna-codex.md). Шесть сессий на
`gw2` (единственный вход, где обмен случился в этапе 3) плюс ретроспективная
переоценка всех 30 сохранённых деревьев этапа 3 — без новых сессий.

Code-гейт работает: в 2 из 3 прогонов петли он поймал ровно тот обмен
(Δlines −2 при Δcode +3 и Δlines −1 при Δcode +3) и отказал при зелёном
physical. По всем 30 прогонам этапа 3 он отказывает **четырём** прогонам петли,
а не двум, которые нашёл ручной review: заявленное `line gate 15/15` — это
`code gate 11/15`.

Но смоук нашёл вторую, более общую дыру. Третий прогон петли прошёл **оба**
line-гейта (Δlines +0, Δcode −2), собрав литерал `&http.Server{…}` с восьмью
полями в одну строку на 201 символ и потратив освободившиеся строки на три
хелпера: **+47 токенов**. `gofmt` такой результат не трогает, то есть
форматтером line-гейт не защищается.

Переоценка это подтверждает по всей выборке. В токенах (все токены production
кроме `;` и `,` — единица, которую не двигают ни переносы, ни комментарии):

| Плечо | Δlines | Δcode | Δtokens | Line gate | Code gate | Токены не выросли |
|---|---:|---:|---:|---:|---:|---:|
| prose | +1.00 | +0.40 | +10.67 | 8/15 | 8/15 | 9/15 |
| с петлёй | −2.80 | −1.73 | **+4.13** | 15/15 | **11/15** | **8/15** |

Эффект loop-минус-prose в токенах **−6.53 при p = 0.598** против −3.80 строки
при p = 0.01572. То есть измеренный эффект этапа 3 **имеет форму строки и не
переживает переход к единице, которую нельзя переупаковать**. Худший случай —
`gw1 #1`: Δlines +0, Δcode +1, Δtokens +61, один обработчик разложен на три
функции внутри того же физического размера, оба line-гейта зелёные.

Отсюда пятый критерий провален шире, чем считалось: удаление документации было
частным способом оплаты, code golf — второй, и он проходит оба line-гейта.
Перед перепрогоном 30 сессий нужно решение по единице:

1. гейтовать и токены (`Δtokens <= 0` третьей осью) — **рекомендация**;
2. оставить строки и назвать golf в тексте feedback-а;
3. сменить единицу на statements/declarations, переопределив всю серию.

`metrics.Tokens` уже считается и печатается как отчётная метрика — гейтом она
не является, потому что что третья ось сделает с поведением модели, пока не
измерено.

Отдельным пунктом остаётся правка формулировки гейта в `go-code-refactor`:
отгруженный в main текст требует «no more production lines», не исключая
комментарии, то есть поощряет у пользователей тот же обмен. Это изменение
текста скилла и по правилу репозитория идёт вариантом и измерением, а не
вместе с этой правкой стенда.

Побочная находка, обесценивающая точность этапов 1 и 2: разброс внутри одного
входа велик — `gw1` на prose-плече дал −5, −5 и **+17** при одинаковых prompt,
плагине и seed. `n=1` на вход был слишком мал, и −1.20 строки этапа 1 лежит
внутри собственного шума одной фикстуры.

## Этап 4. Отделить concision от общего refactor

Безусловный zero-growth gate не подходит каждому behavior-preserving refactor:
полезная читаемость или safety иногда законно увеличивают код.

После подтверждения feedback loop сравнить:

1. условный раздел в `go-code-refactor`, активный только при явном запросе
   `shorter`, `less code`, `concise`, `лаконичнее`;
2. отдельный user-invoked `go-concision`, который не платит description context
   load в обычных Go-сессиях.

**Предпочтение:** отдельный skill только при доказанной пользе автоматического
gate. До этого новый skill является лишней сущностью.

**Done when:** явный concision-запрос надёжно активирует режим, обычный refactor
не получает zero-growth ограничение, а context cost не растёт без запроса.

## Этап 5. Разделить проверку на две оси

Использовать отдельного read-only verifier вместо второго свободно
редактирующего review:

- `Behavior/Spec`: публичный контракт, nil/empty wire shape, методы, ошибки;
- `Concision`: shallow helpers, Middle Man, Speculative Generality, лишние
  функции/типы.

Verifier возвращает только конкретные findings. Исходная сессия получает их в
одном repair-turn.

**Smoke:** пять фиксированных inputs.

**Stop:** verifier не обнаруживает известный `nil → null` либо пропускает
очевидный helper-growth `+22/+33`.

## Этап 6. Проверить обобщаемость

HTTP gateway недостаточен для общего изменения `go-code-refactor`. Лучший
вариант проверить на существующем refactor-корпусе:

- `dispatch`;
- `pricing`;
- `report`;
- `store`.

Сначала `n=3`, затем `n=5` только при положительном результате.

**Go:** correctness не ухудшается ни на одном fixture, минимум три fixture не
растут, aggregate line effect отрицательный, сокращение не переносит сложность
в дополнительные функции и типы.

**Stop:** любой воспроизводимый behavioral regression либо выигрыш только на
одном fixture после просмотра всех четырёх.

## Этап 7. Pruning самого skill

Выполнять последним, используя сохранённые traces:

- [ ] удалить model-relative no-op инструкции;
- [ ] оставить критические решения в `SKILL.md`;
- [ ] вынести branch-specific детали за точные context pointers;
- [ ] убрать дублирование между `go-code`, `go-code-refactor` и
      `OVER-ENGINEERING.md`;
- [ ] сравнить полный plugin с reference arm на одинаковых model, effort, seed
      и fixtures.

Не принимать сокращение только по числу слов. Оно должно сохранить или
улучшить golden, non-growth rate, размер generated code и стоимость сессии.

## Порядок выполнения

```text
стенд
→ согласованный prompt
→ feedback repair
→ автоматический gate
→ отдельный concision-режим
→ двухосевая проверка
→ другие fixtures
→ pruning
```

Этапы 0, 1 и 2 закрыты 2026-09-09; этап 3 выполнен частично в тот же день. Дыра
с оплатой кода документацией закрыта в стенде тем же днём (`Δcode`,
`code_gate_pass`, оба числа в feedback-е) и подтверждена смоуком на `gw2`.
Ближайший шаг — **не** перепрогон, а решение по единице измерения: переоценка 30
сессий этапа 3 показала, что его эффект имеет форму строки и в токенах не
установлен (−6.53, p = 0.598). Перепрогон 30 сессий запускается уже на выбранной
единице; Expand до `n=5` — только после чистого пятого критерия. Следующий этап
не начинается, пока предыдущий не прошёл свой `Go` criterion.

Пункты, которые прогоны этапов 1 и 2 добавили к последующим этапам:

- Любой feedback о строках обязан называть конвенцию подсчёта. Несогласованность
  конвенций, а не нежелание модели, объясняет как минимум один из десяти
  провалов generic repair на этапе 2. Петля этапа 3 должна отдавать модели то же
  число, которое считает харнесс, и говорить, что это физические строки
  production `*.go`.
- Этап 4 получил прямой аргумент из данных: `rp0` вырос на +3 строки, чтобы
  починить `nil → []`. Безусловный zero-growth гейт запретил бы эту починку.

- Этап 4 больше не будущая работа. Concision Gate внесён в main merge-ем PR #2
  (`1dfa990`) как безусловное правило, тогда как отчёт 5×5 постановил этого не
  делать. Условность гейта теперь исправление, а не улучшение.
- Этап 4 должен решить и то, считать ли прогон с `empty_diff` валидным `Δ0`.
  Guard `!Edited` признаёт его невалидным, что под prompt-ом, разрешающим пустой
  diff, завышает плечо: −3.5 по valid против −2.80 по всем пяти.
