# Sonnet 5: одиночные implement-прогоны против no-skill

2026-09-08. По одному наблюдению на каждую пару задача/вариант: 4 задачи × 2 варианта = 8 результатов. В этом запросе выполнены 7 новых сессий; gateway/baseline повторно взят из непосредственно предшествующего одиночного прогона.

Модель `claude-sonnet-5`, CLI `2.1.263 (Claude Code)`, `go version go1.27.1 darwin/arm64`. Seed 1; `n=1`; три новые пары запущены с `j=2`, отдельный gateway/no-skill — с `j=1`. Reasoning effort не задавался, используется default CLI.

Дерево плагина перед запуском совпало с предыдущим gateway/baseline: SHA-256 `eefdafe8c65819a574f7f804e0553f07b13ee80d38003ddf4da4b06f32d330bc`. Оба новых raw-отчёта используют тот же prompt и корпус implement. Изменений скиллов или benchmark-harness в ходе сравнения не было.

## Результаты

Сборка и существующие скрытые golden tests прошли в 8/8 сессий. Ни одна модельная сессия не создала собственных тестов. CLI-ошибок, неизменённых заготовок и зарегистрированного выхода к исходному корпусу нет.

Строки — итоговый размер production Go-файлов, включая комментарии и пустые строки. Разница = со скиллами минус no-skill. Новые функции включают именованные методы, но не анонимные функции.

| Задача | Строки no-skill | Строки со скиллами | Разница | Новые функции/методы, без → со | Стоимость, без → со |
|---|---:|---:|---:|---:|---:|
| catalog | 73 | 57 | -16 | 2 → 0 | $0.045 → $0.128 |
| feed | 82 | 79 | -3 | 0 → 0 | $0.045 → $0.095 |
| ledger | 91 | 87 | -4 | 0 → 0 | $0.052 → $0.096 |
| gateway | 103 | 104 | +1 | 0 → 2 | $0.080 → $0.219 |

### Детали структуры и вызовов Skill

| Задача | Добавленные строки, без → со | Новые типы верхнего уровня, без → со | Ветвления, без → со | Зарегистрированные Skill-вызовы baseline |
|---|---:|---:|---:|---|
| catalog | +37 → +21 | 1 → 0 | 4 → 4 | `go-code`, `go-error-handling` |
| feed | +47 → +44 | 0 → 0 | 4 → 4 | `go-code` |
| ledger | +41 → +37 | 0 → 0 | 6 → 6 | `go-code` |
| gateway | +66 → +67 | 0 → 0 | 7 → 7 | `go-code` |

В no-skill вызовов Skill нет: этот инструмент исключён. В baseline таблица отражает вызовы Skill, а не чтение SKILL.md через Read. Новых интерфейсов и публичных объявлений нет ни в одном варианте; метрика типов harness не учитывает локальные типы внутри функций.

## Что видно в сохранённом коде

- **catalog:** no-skill создаёт `resolveError` с `Error` и `Unwrap`; baseline использует `fmt.Errorf("catalog: resolve %q: %w", sku, err)`. Проверки идентичности причины и дедупликации SKU проходят в обоих вариантах; разница в объёме реализации составляет 16 строк.
- **feed:** no-skill отдельно ведёт `seenKinds`; baseline получает список kinds из уже построенного counts. В обоих вариантах пустые массивы и объект имеют нужные JSON-типы.
- **ledger:** алгоритмы практически одинаковы. Разница в четыре строки — форматирование возвращаемого литерала Ledger в конструкторе, а не удалённая логика.
- **gateway:** baseline добавляет `filterActive` и `writeJSON`; no-skill оставляет операции в обработчиках. Итоговая разница в размере — одна строка, но поведение на nil-входе различается.

## Дополнительные проверки gateway

Одни и те же два теста из предыдущего одиночного прогона запущены на копиях обеих реализаций. Эти проверки выполнены после benchmark и не были показаны модели. Сохранённые реализации и исходные golden-наборы не изменены.

| Проверка | no-skill | Со скиллами |
|---|---|---|
| HEAD на известном маршруте должен вернуть 405 | FAIL: 200 | FAIL: 200 |
| GET /accounts при accounts == nil должен вернуть [] | PASS | FAIL: null |

HEAD-провал есть на /healthz, /accounts и /accounts/a-1. Ненулевой пустой slice возвращает [] в обеих реализациях. Таким образом, golden PASS не означает полного выполнения документации gateway.

## Стоимость и границы вывода

- Семь новых сессий: **$0.54067**.
- Все восемь результатов, включая предыдущий gateway/baseline: **$0.75972**.
- Сумма по no-skill: $0.22199; по baseline: $0.53773 — **2.42×**.
- n=1 показывает отдельные реализации. Оценки устойчивого эффекта, доверительных интервалов и статистической значимости здесь нет.
- Разница no-skill/baseline относится ко всему плагину. Она не изолирует эффект последних исправлений: для такого сравнения нужен reference со старой версией скиллов.
- Claude ограничен инструментами Skill/Read/Glob/Grep/Edit/Write; в no-skill исключён Skill, а в baseline загружен плагин с PostToolUse gofmt/vet hook. Shell самой модели недоступен; сборку и golden tests запускает harness.

## Команды

```bash
go run ./cmd/abrun -corpus implement -tasks catalog,feed,ledger -runner claude -model claude-sonnet-5 -arms no-skill,baseline -n 1 -j 2 -seed 1 -timeout 10m -keep -verbose -out ../docs/evidence/2026-09-08-go-implement-single-comparison-sonnet-5.corpus.json
```

```bash
go run ./cmd/abrun -corpus implement -tasks gateway -runner claude -model claude-sonnet-5 -arms no-skill -n 1 -j 1 -seed 1 -timeout 10m -keep -verbose -out ../docs/evidence/2026-09-08-go-implement-single-comparison-sonnet-5.gateway-control.json
```

Команды выполнены из evals; `/Users/eugeneshershen/go/bin` стоял в начале PATH.

## Артефакты

- [Сводные данные](2026-09-08-go-implement-single-comparison-sonnet-5.summary.json).
- [Raw: catalog/feed/ledger](2026-09-08-go-implement-single-comparison-sonnet-5.corpus.json).
- [Raw: gateway/no-skill](2026-09-08-go-implement-single-comparison-sonnet-5.gateway-control.json).
- [Предыдущий gateway/baseline](2026-09-08-go-implement-single-sonnet-5.json).
- [Все исходники, diff и дополнительные тесты](2026-09-08-go-implement-single-comparison-sonnet-5.sources.json).
- [Метаданные, версии, hashes и снимок исходных условий](2026-09-08-go-implement-single-comparison-sonnet-5.metadata.json).
- [Вывод новых пар](2026-09-08-go-implement-single-comparison-sonnet-5.corpus.log).
- [Вывод gateway/no-skill](2026-09-08-go-implement-single-comparison-sonnet-5.gateway-control.log).
