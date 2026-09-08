# Одиночный benchmark написания Go-кода — Sonnet 5

2026-09-08. Одна сессия `gateway` из корпуса `implement`, arm `baseline` с текущими скиллами и незакоммиченными исправлениями.

- Модель: `claude-sonnet-5`; CLI: `2.1.263 (Claude Code)`.
- Go: `go version go1.27.1 darwin/arm64`.
- Seed: 1; repetitions: 1; parallelism: 1.
- Время всего прогона: 92.2 с.
- Стоимость сессии по CLI: $0.21905.
- HEAD исходного репозитория: `1628440d35ce8457b0852fdf41a2d56fbf55c950`; working tree содержит незакоммиченные изменения.
- SHA-256 дерева плагина: `eefdafe8c65819a574f7f804e0553f07b13ee80d38003ddf4da4b06f32d330bc`.

```bash
go run ./cmd/abrun -corpus implement -tasks gateway -runner claude -model claude-sonnet-5 -arms baseline -n 1 -j 1 -seed 1 -timeout 10m -keep -verbose -out ../docs/evidence/2026-09-08-go-implement-single-sonnet-5.json
```

Команда выполнена из `evals` с `/Users/eugeneshershen/go/bin` в начале PATH. Harness ограничивает Claude инструментами `Skill,Read,Glob,Grep,Edit,Write`; shell внутри модельной сессии недоступен, PostToolUse hook плагина активен. Сборку и golden tests запускает harness после ответа модели.

| Проверка / метрика | Результат |
|---|---:|
| Изменение заготовки | Да |
| Ошибка CLI / зарегистрированный доступ к корпусу | Нет / нет |
| Сборка | PASS |
| Скрытые golden tests | PASS |
| Тесты, написанные моделью | Не созданы |
| Строки production Go, включая комментарии и пустые строки | 37 → 104 (+67) |
| Объявленные функции | 1 → 3 (+2) |
| Типы / интерфейсы | 1 / 0, без изменений |
| Публичные объявления | 2, без изменений |
| Ветвления по метрике harness | +7 |
| Зарегистрированные вызовы Skill | go-code |

Две добавленные функции: `filterActive` и `writeJSON`. Модель копирует и сортирует accounts, создаёт индекс по ID и задаёт четыре серверных timeout. Вызов Skill для go-http не зарегистрирован; это метрика вызовов инструмента, а не доказательство отсутствия чтения файла через Read.

## Дополнительные проверки контракта

После завершения измерения исходный scratch-каталог скопирован; дополнительные тесты запущены только в копии. Сгенерированный код и исходный результат benchmark сохранены без изменений.

Обнаружены два пропуска golden-набора:

1. `HEAD /healthz`, `HEAD /accounts`, `HEAD /accounts/a-1` возвращают 200 вместо требуемого документацией 405. Причина: GET-паттерны ServeMux также принимают HEAD.
2. При `accounts == nil` запрос `GET /accounts` возвращает `null` вместо JSON-массива `[]`. Для непустого и явно пустого ненулевого slice этот конкретный тест проходит.

Следовательно, PASS относится к существующему golden-набору. Полный задокументированный контракт модель в этой сессии не выполнила. Это единичное наблюдение без контрольной группы; оно не устанавливает влияние последних правок скиллов.

## Артефакты

- [Raw JSON](2026-09-08-go-implement-single-sonnet-5.json).
- [Исходный и сгенерированный код, golden tests и diff](2026-09-08-go-implement-single-sonnet-5.sources.json).
- [Команда, версии, SHA-256 и diff исходных скиллов](2026-09-08-go-implement-single-sonnet-5.metadata.json).
- [Вывод harness](2026-09-08-go-implement-single-sonnet-5.log).
- [Вывод дополнительных проверок](2026-09-08-go-implement-single-sonnet-5.edge-checks.log).
- Сгенерированный файл: `/var/folders/p2/6zhdz7hs66lc8055sxqydmj80000gn/T/abrun-work-2017038999/gateway/gateway.go`.
