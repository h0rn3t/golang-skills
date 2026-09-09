# Paired concision review — GPT-5.6-Luna (medium) on Codex CLI

2026-09-09. **Отдельный review-проход не сделал уже корректный HTTP-код
лаконичнее.** Текущий `go-code-refactor` добавил в среднем 9.2 production
строки, а вариант с `deletion test`, `Middle Man` и
`Speculative Generality` — 7.8. Candidate уменьшил рост относительно control
на 1.4 строки, но добавил на 0.4 функции больше. Оба плеча прошли независимый
golden 5/5.

## Дизайн

Входами стали пять разных passing-реализаций `gateway`, созданных baseline
плечом предыдущего implement A/B. Каждый вход скопирован в два независимых
scratch-каталога:

- `baseline`: обычный refactor prompt и текущий plugin;
- `deletion-review`: тот же prompt и plugin плюс один блок перед `Workflow` в
  `go-code-refactor`.

Пять входов — это пять fixtures, по одному прогону каждого плеча на fixture.
Так сравниваются две review-инструкции на одинаковом исходном коде без шума от
повторной первоначальной реализации.

Candidate переносит три идеи из `mattpocock/skills` на commit
`3cca18b368ae95cdbdebbff572ccafa662551015`: deletion test из
`codebase-design` и запахи `Middle Man`/`Speculative Generality` из
`code-review`. Он требует проверить каждый unexported helper/type/layer,
встроить shallow pass-through код и оставить абстракции, за которыми остаётся
реальная общая сложность.

## Запуск

- Runner: Codex CLI 0.153.4.
- Model: `gpt-5.6-luna`; reasoning effort `medium`.
- Seed 1; `j=4`; timeout 10 минут; `-keep`.
- Репозиторий: `f12c73b899727fbad5ab2c73c92149895b5da7a7`.
- Baseline plugin SHA-256:
  `f208ef8bcb0d8ce38f6564a0d74dd50c43a64aeb1acdf4f7cda3606ba7442d4b`.
- Candidate plugin SHA-256:
  `0d2b1bad423568ed755ae77a339fa8040bebd5cfdf1944e567a053efcebee38e`.
- Raw report SHA-256:
  `f429c366621d95f6b57bd7e313b03f548b9414450d339a434da70af75fcfb3b0`.

Одинаковый prompt для обоих плеч:

```text
Refactor the Go package in ./<fixture> so it reads better. Keep observable
behavior identical: the exported API, error texts, and rendered output must
not change. Apply the changes to the files.
```

## Результаты

| Вход | Baseline lines / funcs / types | Candidate lines / funcs / types | Candidate − baseline lines / funcs |
|---|---:|---:|---:|
| `0` | −5 / 0 / 0 | −5 / 0 / 0 | 0 / 0 |
| `1` | +12 / +4 / +1 | +11 / +5 / 0 | −1 / +1 |
| `2` | +8 / +2 / 0 | +7 / +1 / 0 | −1 / −1 |
| `3` | +18 / +2 / 0 | +12 / +3 / 0 | −6 / +1 |
| `4` | +13 / +3 / 0 | +14 / +4 / 0 | +1 / +1 |
| **Среднее** | **+9.2 / +2.2 / +0.2** | **+7.8 / +2.6 / 0** | **−1.4 / +0.4** |

- Build: 5/5 в обоих плечах.
- Independent golden: 5/5 в обоих плечах.
- Model-authored tests: 2/5 baseline и 3/5 candidate; все пять прошли до того,
  как harness спрятал их и добавил independent golden.
- `go-code-refactor` был прочитан во всех 10 сессиях.

Candidate короче control на трёх входах, равен на одном и длиннее на одном.
Почти весь средний выигрыш даёт одна пара `#3` (−6 строк); при пяти парах это
направление, а не установленный эффект.

## Что модель фактически сделала

На единственном входе, где production-код уменьшился, оба плеча независимо
заменили ручной string comparator на `cmp.Compare` и получили одинаковые
−5 строк. Deletion Review не добавил к этому сокращению ничего.

На остальных четырёх входах оба плеча в основном выносили существующие ветки
маршрутизации в `serveHealth`, `serveAccounts`, `serveAccount`,
`handleRequest`, `accountID` и похожие helpers. Candidate не распознал эти
одноразовые извлечения как shallow middle men: он добавил от одной до пяти
функций на каждом из четырёх растущих входов.

То есть название запаха и deletion test немного сдвинули line count, но не
победили более сильный prior модели: «reads better» означает «split into focused
helpers». Сохранение поведения не компенсирует провал основной цели этого
эксперимента — лаконичности.

## Решение и следующий тест

Не добавлять Deletion Review в основные skills и не вводить отдельный
concision-review в этом виде.

Следующая более узкая гипотеза — сделать лаконичность проверяемым критерием
завершения, а не review-запахом: для concision-only прохода итоговый production
line count не растёт; пустой diff является успешным результатом; растущую
трансформацию модель откатывает. Это следует проверять на тех же пяти входах
отдельным A/B, потому что такой жёсткий gate может вызвать code golf или скрыть
полезную локальность.

## Артефакты

- [Raw JSON](2026-09-09-paired-concision-review-gpt-5.6-luna-codex.json).
- [Точный candidate block](2026-09-09-paired-concision-review-gpt-5.6-luna-codex.variant.md).
- [Golden, пять входов, десять outputs и пять model tests](2026-09-09-paired-concision-review-gpt-5.6-luna-codex.sources.json).

