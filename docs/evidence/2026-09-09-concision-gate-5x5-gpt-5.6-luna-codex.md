# Concision Gate 5×5 — GPT-5.6-Luna (medium) on Codex CLI

2026-09-09. Пять фиксированных passing-реализаций проверены пять раз на каждом
из двух плеч: 25 baseline и 25 `Concision Gate` сессий. **Gate заметно повысил
вероятность получить нерастущий код, но не стал надёжным ограничением и пока не
готов для основного skill.**

- Нерастущие результаты: 7/25 baseline против 16/25 candidate.
- Медиана production delta: +5 против −2 строк.
- Среднее по всем результатам: +5.40 против +1.88 строк; эффект −3.52.
- Candidate всё же вырос в 9/25 сессий, вплоть до +33 строк и +5 функций.
- После исключения одного конфликта имени с helper-ом hidden golden реальный
  behavioral golden одинаков: 24/25 на каждом плече.

## Дизайн

Входами стали пять разных passing-реализаций `gateway`, сохранённых из
предыдущего implement A/B. Каждый вход запускался пять раз с одинаковым
refactor prompt на двух вариантах plugin:

- `baseline`: текущий `go-code-refactor`;
- `concision-gate`: тот же plugin плюс проверяемый критерий — итоговый
  production line count не растёт, пустой diff допустим, ясность и поведение
  сохраняются.

Модель не видела independent golden; harness добавлял его после завершения
review и после изоляции model-authored tests.

## Запуск

- Runner: Codex CLI 0.153.4.
- Model: `gpt-5.6-luna`; reasoning effort `medium`.
- Пять fixtures × два плеча × пять повторов = 50 сессий.
- Seed 1; `j=4`; timeout 10 минут; `-keep`.
- Изолированный experiment clone:
  `f12c73b899727fbad5ab2c73c92149895b5da7a7`.
- Baseline plugin SHA-256:
  `f208ef8bcb0d8ce38f6564a0d74dd50c43a64aeb1acdf4f7cda3606ba7442d4b`.
- Candidate plugin SHA-256:
  `61ebc0e44bb4d98f3b66986a48c12653f528483fbfa7903facf1498289ee35c3`.
- Raw report SHA-256:
  `9767b63ff14c902cf4a8337817ac998da5341c3a35831efeee9dc3d8358ab1e1`.

Во время прогона основной checkout перешёл с `f12c73b` на `cb48226` внешним
изменением. Это не затронуло эксперимент: runner, fixtures и plugins находились
в уже созданном изолированном clone на `f12c73b`.

## Результаты по входам

Значения включают все сессии, в том числе golden failures. `≤0` — число
результатов без роста production line count.

| Input | Baseline lines / ≤0 / golden | Candidate lines / ≤0 / golden | Candidate − baseline |
|---|---:|---:|---:|
| `0` | −2.0 / 4/5 / 5/5 | −5.0 / 5/5 / 4/5 | −3.0 |
| `1` | +1.8 / 2/5 / 5/5 | −3.0 / 5/5 / 5/5 | −4.8 |
| `2` | +5.8 / 0/5 / 4/5 | +6.2 / 1/5 / 5/5 | +0.4 |
| `3` | +12.2 / 0/5 / 4/5 | +9.6 / 2/5 / 5/5 | −2.6 |
| `4` | +9.2 / 1/5 / 5/5 | +1.6 / 3/5 / 5/5 | −7.6 |
| **Все** | **+5.40 / 7/25 / 23/25** | **+1.88 / 16/25 / 24/25** | **−3.52** |

Средние новые функции: 1.24 baseline против 0.76 candidate; branches: −0.28
против −0.68. Все 50 сессий собрали package. Model-authored tests появились и
прошли в 8/25 baseline и 6/25 candidate.

Exact randomization test со сменой labels только внутри каждого фиксированного
input дал:

- mean line effect −3.52: two-sided `p = 0.0883`;
- non-growth rate +36 percentage points: `p = 0.00452`;
- function effect −0.48: `p = 0.2168`;
- branch effect −0.40: `p = 0.1662`.

`p` для non-growth не корректировался за несколько просмотренных метрик и
остаётся exploratory. Вместе с распределением он всё же поддерживает узкий
вывод: текст повышает вероятность нерастущего результата. Он не подтверждает
гарантированный line cap и не доказывает улучшение читаемости.

## Исполнение gate

`go-code-refactor` прочитан в 25/25 сессий обоих плеч. Несмотря на это:

- candidate нарушил собственный `Δlines ≤ 0` в 9/25 сессий;
- положительные candidate deltas: `+2, +2, +4, +5, +5, +12, +16, +22, +33`;
- только 8/25 финальных сообщений candidate вообще упомянули измеренное
  сокращение или line count;
- пустого diff не было: все 25 candidate-сессий фактически редактировали код.

В больших нарушениях Luna снова извлекала `serveAccountList`, `findAccount`,
`newHandler`, `handleHealth`, `handleAccountList`, `requestHandler` и похожие
одноразовые helpers. То есть completion criterion сдвинул распределение, но
иногда полностью проигрывал prior «reads better = split into focused helpers».

## Golden failures

Raw golden rate — 23/25 baseline и 24/25 candidate, но эти три failures имеют
разную природу:

1. Baseline `input 3 / rep 2` добавил unexported helper `serve`, который
   столкнулся с helper того же имени в hidden `golden_test.go`. Production
   package до добавления golden собирался. Supplemental запуск с переименованным
   только golden helper прошёл; это harness namespace collision, а не
   подтверждённое изменение поведения.
2. Baseline `input 2 / rep 4` реально сломал method contract: POST и HEAD к
   `/accounts` вернули 200 вместо 405.
3. Candidate `input 0 / rep 3` реально изменил wire shape: nil accounts
   вернулись как `null` вместо `[]`. Финальное сообщение при этом ошибочно
   заявило сохранение nil behavior.

После классификации реальный observed behavior pass — 24/25 на каждом плече.
Разницы корректности здесь не установлено.

## Решение

Не переносить Concision Gate в основной `go-code-refactor` как безусловное
правило. Положительный эффект на частоту нерастущего кода достаточно силён для
следующего эксперимента, но 36% нарушений gate и один wire regression делают
текущий текст ненадёжным.

Следующая проверяемая гипотеза — не усиливать prose ещё раз, а дать модели
исполняемый line gate: скрипт фиксирует baseline, возвращает ошибку при росте и
показывает точный delta до завершения. Это проверит feedback loop, а не ещё одну
вариацию формулировки. Отдельно golden helper следует перенести во внешний test
package или дать ему collision-resistant имя.

## Артефакты

- [Raw JSON](2026-09-09-concision-gate-5x5-gpt-5.6-luna-codex.json).
- [Пять inputs, 50 outputs, 14 model tests и golden](2026-09-09-concision-gate-5x5-gpt-5.6-luna-codex.sources.json).
- [Точный candidate block](2026-09-09-concision-gate-smoke-gpt-5.6-luna-codex.variant.md),
  переиспользованный без изменений из smoke.
- [Предыдущий `n=1` smoke](2026-09-09-concision-gate-smoke-gpt-5.6-luna-codex.md).

