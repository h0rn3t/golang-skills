# Этап 1 — согласование prompt-а с Concision Gate, GPT-5.6-Luna (medium) на Codex

2026-09-09. Пять фиксированных gateway-входов, по одному прогону на вход на
каждом из двух плеч prompt-а: 5 control и 5 candidate сессий. **Candidate прошёл
все четыре Go-критерия этапа 1, включая первый настоящий empty diff за всю
серию concision-экспериментов. Разница в строках между prompt-ами при этом не
установлена: p = 0.625.**

- Golden: 5/5 на обоих плечах, ни одного behavior и ни одного harness failure.
- Line gate (`Δlines ≤ 0`): 4/5 control против **5/5** candidate.
- Empty diff: 0 control против **1** candidate, намеренный и проверенный.
- Средняя Δlines по всем пяти: −1.60 control против −2.80 candidate.
- Парный эффект −1.20 строки, exact randomization two-sided **p = 0.6250**.

## Дизайн

Этап 1 держит плагин постоянным и меняет только верхнеуровневый prompt. Повод
из плана: текущий prompt заканчивается `Apply the changes to the files`, тогда
как Concision Gate разрешает пустой diff, и в 25/25 candidate-сессий прогона
5×5 модель что-нибудь меняла.

- `control`: текущий `refactorPrompt` из `abrun` без изменений.
- `candidate`: текст этапа 1 плана. Единственная адаптация — `./%s` вместо
  «the package», потому что `abrun` подставляет имя фикстуры в шаблон:

```text
Review the Go package in ./%s for a shorter, equally clear implementation.
Change it only when production code becomes no longer and behavior remains
identical. If no such improvement exists, leave the files unchanged.
```

Входы — те же пять passing-реализаций `gateway` из прогона 5×5, взятые из его
[sources.json](2026-09-09-concision-gate-5x5-gpt-5.6-luna-codex.sources.json)
(`inputs[i].source`, 135/145/127/126/132 строки) и разложенные как фикстуры
refactor-корпуса `gw0`…`gw4`. Все пять проходят hidden golden **до** правки
модели, поэтому любое падение после неё принадлежит правке. Golden взят из
`evals/ab/_implement/_golden/gateway/golden_test.go` в исправленном после
этапа 0 виде — с `goldenServe` вместо `serve`.

Воспроизведение корпуса из закоммиченного sources.json:

```python
import json, os, pathlib
d = json.load(open('docs/evidence/2026-09-09-concision-gate-5x5-gpt-5.6-luna-codex.sources.json'))
golden = pathlib.Path('evals/ab/_implement/_golden/gateway/golden_test.go').read_text()
for i, inp in enumerate(d['inputs']):
    for sub in (f'evals/ab/gw{i}', f'evals/ab/_golden/gw{i}'):
        os.makedirs(sub, exist_ok=True)
    pathlib.Path(f'evals/ab/gw{i}/gateway.go').write_text(inp['source'])
    pathlib.Path(f'evals/ab/_golden/gw{i}/golden_test.go').write_text(golden)
```

Фикстуры удалены из рабочего дерева после прогона: оставленные в `evals/ab`, они
попадали бы в каждый последующий refactor-прогон по умолчанию и молча меняли бы
его корпус.

## Запуск

- Runner: Codex CLI 0.153.4; model `gpt-5.6-luna`; reasoning effort `medium`.
- Пять входов × два плеча × один повтор = 10 сессий; seed 1; `j=4`;
  timeout 10 минут; `-keep`.
- Плечо `baseline` в обоих прогонах; плагин из рабочего дерева на `1dfa990`.
- Plugin SHA-256 **обоих** прогонов:
  `5509f19786639774c1bc6e6bc822d5f8fd7ae56efab7836976fbff63a6b89148` —
  плагин держался постоянным, менялся только prompt.
- Дерево `skills/`, `agents/`, `hooks/`, `.claude-plugin` на `1dfa990`
  байт-в-байт равно `18e701e`, то есть ровно тому плагину, который называет
  план: `e21e614` добавил только evidence-файлы.
- Незакоммиченные изменения в рабочем дереве на момент прогона затрагивали
  только `evals/`, `docs/` и README — ни одной строки плагина.

## Результаты по входам

| Вход | Control Δlines / gate | Candidate Δlines / gate | Candidate − control |
|---|---:|---:|---:|
| `gw0` | +2 / ✗ | −5 / ✓ | −7 |
| `gw1` | −5 / ✓ | 0 / ✓ (empty diff) | +5 |
| `gw2` | 0 / ✓ | −2 / ✓ | −2 |
| `gw3` | −2 / ✓ | −1 / ✓ | +1 |
| `gw4` | −3 / ✓ | −6 / ✓ | −3 |
| **Все** | **−1.60 / 4/5** | **−2.80 / 5/5** | **−1.20** |

Golden 5/5 на каждом плече. `behavior_failure` и `harness_failure` — ноль во
всех десяти прогонах. Δtypes, Δiface, Δfuncs и Δpattern — ноль на candidate;
control добавил 0.20 функции на прогон. Δbranch −0.60 control против −1.25
candidate.

Exact paired randomization по пяти парам (2⁵ = 32 расстановки знаков) для
парной суммы −6: **two-sided p = 20/32 = 0.6250**. Разница в строках между
prompt-ами не установлена, и при `n=1` на вход установлена быть не могла.
Критерии этапа 1 — screening gate, а не тест значимости.

## Go-критерии

| Критерий | Результат |
|---|---|
| Candidate даёт golden 5/5 | ✓ 5/5 |
| `Δlines ≤ 0` минимум в 4/5 | ✓ 5/5 |
| Хотя бы один настоящий empty diff | ✓ `gw1` |
| Ни одного случая code golf | ✓ с одной оговоркой, см. ниже |

Ни один Stop-критерий не сработал: candidate не проиграл control по golden
(5/5 против 5/5), не вырос ни в одной сессии (control вырос в одной), и не
сократил validation, error или security behavior — `gw2` сохранил 400 на
неизвестном значении `active`, `gw4` сохранил 405 на всех трёх путях, а golden
покрывает status codes, фильтрацию, wire shape пустого списка и таймауты.

### Empty diff настоящий

`gw1` — первый empty diff за серию: в прогоне 5×5 их не было ни одного из 25.
Финальное сообщение: «Reviewed `./gw1`; no safe shortening was found that
preserves both clarity and behavior. Files were left unchanged.» Trace
подтверждает, что это не отказ от работы: сессия выполнила четыре команды,
включая `go test ./...`, `go vet ./...` и `gofmt -d gw1/gateway.go`. Именно для
этой проверки этап 0 сохраняет transcript — заявление в финальном сообщении
здесь совпало с фактически выполненными командами.

### Что именно сокращалось

- `gw0` (−5): трёхветочный компаратор `if a.ID < b.ID … return 0` заменён на
  `cmp.Compare(a.ID, b.ID)`. Stdlib вместо ручной ветки.
- `gw2` (−2): вложенные `if present { if invalid { … } }` слиты в одно условие.
  Ответ 400 сохранён.
- `gw3` (−1): убран избыточный `w.WriteHeader(http.StatusOK)` перед `Write`.
- `gw4` (−6): проверка метода вынесена из трёх кейсов `switch` в один guard.
  **Оговорка:** список трёх путей теперь дублируется между guard и `switch`, то
  есть добавление пути требует правки в двух местах. Поведение идентично и
  golden зелёный, но «reads at least as clearly» здесь спорно. Это не code
  golf — ни сжатия, ни односимвольных имён, ни снятой обработки ошибок, — но
  случай ровно того класса, для которого план предусматривает Concision-ось
  этапа 5.

### Стоимость сессии

Candidate выполнил заметно меньше команд: 9/4/6/7/9 против 11/7/12/11/23, то
есть 7.0 против 12.8 в среднем. Codex не отдаёт стоимость в долларах, поэтому
это единственный доступный прокси, и он не входит в критерии этапа. Направление
тем не менее в пользу candidate.

## Методологическая находка

Средняя Δlines в сводке `abrun` считается по **valid** прогонам, а прогон с
пустым diff признаётся невалидным guard-ом `!Edited`. Этот guard написан для
prompt-а, который требует правок: там «ничего не изменил» значит «прогон не
состоялся». Под prompt-ом, который пустой diff прямо разрешает, исключение
такого прогона завышает плечо: candidate показывает −3.5 по четырём валидным
против −2.80 по всем пяти. В заголовке этого отчёта стоит −2.80.

Решать, считать ли `empty_diff` валидным Δ0, должен этап 4 — там же, где
решается, быть ли гейту условным. До того сравнивать плечи следует по всем
прогонам, а не по valid.

## Контекст: гейт уже в main

Прогон 5×5 постановил не переносить Concision Gate в основной
`go-code-refactor` как безусловное правило. Тем не менее merge PR #2
(`1dfa990`) внёс его в main, и на момент этого прогона гейт живёт в
`skills/go-code-refactor/SKILL.md` как безусловное требование. Поэтому оба плеча
этого этапа несут гейт, и этап измеряет **prompt поверх уже отгруженного
гейта**, а не сам гейт. Этап 4 плана — «отделить concision от общего refactor» —
из будущей работы стал просроченной.

## Решение

Этап 1 пройден: candidate-prompt выполнил все четыре Go-критерия и не задел ни
одного Stop-критерия. Переходить к этапу 2 (feedback repair) допустимо.

Чего этот прогон **не** показывает: что candidate-prompt короче control-а по
строкам. При p = 0.625 и `n=1` на вход это неизмеримо, и переносить prompt в
репозиторий на основании −1.20 строки нельзя. Значение прогона в другом — в
том, что prompt, разрешающий пустой diff, впервые дал пустой diff, при 5/5
golden и 5/5 line gate. Это снимает противоречие между prompt-ом и гейтом,
которое план назвал первым в очереди.

## Артефакты

- [Raw JSON, control](2026-09-09-stage1-prompt-alignment-gpt-5.6-luna-codex.control.json),
  SHA-256 `3de40d0c43233425d91cf3459859c3d1a9360259533a923884a9382bdd4520be`.
- [Raw JSON, candidate](2026-09-09-stage1-prompt-alignment-gpt-5.6-luna-codex.candidate.json),
  SHA-256 `03cd8fa211ba23eb6d6c65341531480ff12cb8f4eaace686438116ea85287ca4`.
- Входы и golden: [sources.json прогона 5×5](2026-09-09-concision-gate-5x5-gpt-5.6-luna-codex.sources.json).
- [План этапов](2026-09-09-pocock-concision-next-experiments-plan.md).
