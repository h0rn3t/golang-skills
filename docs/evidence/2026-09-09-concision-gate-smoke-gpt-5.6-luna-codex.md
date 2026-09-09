# Concision Gate smoke — GPT-5.6-Luna (medium) on Codex CLI

2026-09-09. Один фиксированный passing-вход проверен двумя независимыми
review-сессиями. **Concision Gate дал сильное положительное направление:**
обычный review вырос на 14 production-строк и четыре функции, candidate
сократил тот же исходник на 6 строк без новых функций. Оба результата прошли
build и independent golden.

Это `n=1`, поэтому результат обосновывает расширенный тест, но ещё не перенос
правила в основной skill.

## Гипотеза

Предыдущий Deletion Review называл `Middle Man`, `Speculative Generality` и
deletion test, но Luna всё равно интерпретировала «reads better» как извлечение
одноразовых helpers. Новая гипотеза превращает лаконичность в проверяемый
completion criterion:

- итоговый production line count не превышает исходный;
- каждая оставленная трансформация как минимум так же понятна;
- если оба условия нельзя выполнить, пустой diff считается успехом;
- validation, failure behavior, security controls и полезные границы остаются.

## Запуск

- Runner: Codex CLI 0.153.4.
- Model: `gpt-5.6-luna`; reasoning effort `medium`.
- Fixture: `gateway-concision-4` из предыдущего paired review.
- Один прогон на плечо; `j=2`; seed 1; timeout 10 минут; `-keep`.
- Репозиторий: `f12c73b899727fbad5ab2c73c92149895b5da7a7`.
- Baseline plugin SHA-256:
  `f208ef8bcb0d8ce38f6564a0d74dd50c43a64aeb1acdf4f7cda3606ba7442d4b`.
- Candidate plugin SHA-256:
  `61ebc0e44bb4d98f3b66986a48c12653f528483fbfa7903facf1498289ee35c3`.
- Raw report SHA-256:
  `181c08d6d4566b7866b53d2a4024f7d5bb479ccf571a7e84a2a4f79be0a01b8a`.

## Результат

| Плечо | Lines | Functions | Branches | Build | Golden |
|---|---:|---:|---:|---:|---:|
| Текущий `go-code-refactor` | +14 | +4 | 0 | pass | pass |
| `Concision Gate` | **−6** | **0** | **−2** | pass | pass |

Оба плеча прочитали одинаковые `go-code`, `go-code-refactor` и
`go-style-core`. Model-authored test files отсутствовали; harness добавил
golden только после окончания модели.

## Проверка generated code

Baseline повторил прежний нежелательный паттерн: вынес dispatcher в
`handleRequest`, `handleHealth`, `handleAccounts` и `handleAccount`, сохранив
ту же логику за дополнительными переходами.

Candidate не добавил функций. Он:

1. один раз вычислил известность пути и один раз проверил метод вместо трёх
   одинаковых проверок;
2. заменил `strconv.ParseBool` на сравнение с `"true"` после уже существующей
   строгой проверки допустимых значений;
3. сохранил сортировку, snapshot входа, таймауты, wire shape, статусы и ошибки,
   проверяемые golden.

Diff не похож на code golf: имена не сокращены, независимые действия не слиты,
validation и HTTP checks сохранены. Однако golden доказывает только покрытые
публичные сценарии, а один удачный output не доказывает устойчивость модели.

## Решение

Основной skill пока не менять. Следующий шаг — повторить тот же A/B на всех
пяти фиксированных входах. Принимать Concision Gate можно только если candidate
остаётся не длиннее исходника во всех пяти сессиях, golden остаётся 5/5, а
ручной просмотр не обнаруживает code golf или потерю полезной локальности.

## Артефакты

- [Raw JSON](2026-09-09-concision-gate-smoke-gpt-5.6-luna-codex.json).
- [Точный candidate block](2026-09-09-concision-gate-smoke-gpt-5.6-luna-codex.variant.md).
- [Исходник, два outputs и independent golden](2026-09-09-concision-gate-smoke-gpt-5.6-luna-codex.sources.json).

