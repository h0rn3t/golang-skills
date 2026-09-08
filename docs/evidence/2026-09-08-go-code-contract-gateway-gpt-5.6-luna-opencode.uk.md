# go-code «Read the Contract Before Routing» — `gateway` на GPT-5.6 luna через opencode

Короткий контрольний прогін нової секції в `skills/go-code/SKILL.md`: чи змінює
вона маршрутизацію до `go-http` на фікстурі `gateway`. Три повтори на плече, тому
це напрямок, а не результат.

## Прогін

- Завершено: 2026-09-08
- Корпус: `implement`, фікстура `gateway`
- Runner: `opencode` 1.18.29, модель `opencode-go/gpt-5.6-luna`
- Плечі: `reference` — дерево скілів на коміті `6d32157` без секції;
  `baseline` — робоче дерево з секцією (некомічена правка, +24 рядки в `go-code/SKILL.md`)
- Повтори: 3 на плече, 6 сесій, seed 1
- Сирий звіт: [`2026-09-08-go-code-contract-gateway-gpt-5.6-luna-opencode.json`](2026-09-08-go-code-contract-gateway-gpt-5.6-luna-opencode.json), SHA-256 `bbe6c3720f0530f2b57039d82a60a8e1038c067c6e0fb995f56e17585263d131`
- Вартість: $0.041 reference, $0.055 baseline

```bash
go run ./cmd/abrun -corpus implement -runner opencode -model opencode-go/gpt-5.6-luna \
  -reference-root ../before -arms reference,baseline -tasks gateway \
  -n 3 -j 3 -seed 1 -keep -verbose -out ../docs/evidence/2026-09-08-go-code-contract-gateway-gpt-5.6-luna-opencode.json
```

Усі 6 сесій зібралися і пройшли golden-тест. Коректність на цій моделі насичена
(20/20 в обох плечах у [контрольному прогоні](2026-09-07-go-implement-control-gpt-5.6-luna-medium.md)),
тому вимірюваним лишається маршрутизація і розмір коду.

## Результати

| Сесія | `go-http` дійшов | Скіли | Δрядків | Δfuncs | $ |
|---|---|---|---:|---:|---:|
| reference 0 | ні | go-code | +74 | 0 | 0.012 |
| reference 1 | ні | go-code | +99 | 4 | 0.013 |
| reference 2 | так | go-code, go-http, go-style-core | +81 | 2 | 0.015 |
| baseline 0 | ні | go-code | +80 | 3 | 0.013 |
| baseline 1 | так | go-code, go-data-structures, go-defensive, go-http, go-testing | +85 | 1 | 0.020 |
| baseline 2 | так | go-code, go-data-structures, go-defensive, go-http, go-security | +77 | 1 | 0.022 |

| Плече | `go-http` | Δрядків, середнє | Розкид |
|---|---:|---:|---:|
| reference | 1/3 | 84.7 | 74–99 |
| baseline | 2/3 | 80.7 | 77–85 |

Для порівняння: у [вчорашньому прогоні](2026-09-07-go-implement-multirunner-luna-opencode.json)
на тому самому runner і моделі дерево без секції довело `go-http` до моделі в 3 з 5
сесій `gateway`.

## Інтерпретація

- Різниця 1/3 проти 2/3 у маршрутизації при n=3 лежить у межах шуму: вчорашній
  baseline без секції дав 3/5. Прогін не доводить, що секція покращує маршрутизацію.
- Сесії з секцією, які взагалі маршрутизували, завантажили 4–5 власників
  (data-structures, defensive, http, security/testing) замість 1–2; це узгоджується
  з «кожен пункт списку обирає свій рядок», але коштує на ~35% дорожче за сесію.
- Розмір коду: −4 рядки, розкид звузився з 25 до 8 рядків. При n=3 це не ефект.
- Жодне фінальне повідомлення не містить виписаного контракту; чи писала модель
  список під час сесії, з цього звіту не видно — abrun зберігає лише фінальний
  вивід.

Секція лишається неміряною. Щоб отримати результат, потрібна модель, де
`gateway` реально падає без `go-http`, і n≥10 на плече. Жодна з наявних у
корпусі моделей цієї умови не виконує: контрольне плече проходить golden у всіх
сесіях.
