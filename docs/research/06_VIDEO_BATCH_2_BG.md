# Видео пакет 2 — шест Shorts и едно дълго видео, кадър по кадър

**Дата:** 24 септември 2026
**Продължава:** [02_SOCIAL_GROWTH_BG.md](./02_SOCIAL_GROWTH_BG.md) §10 (Short A–E, Дълго 1) и [05_FIRST_VIDEO_MICROTASKS_BG.md](./05_FIRST_VIDEO_MICROTASKS_BG.md) (първото видео). Тук няма повторение на тях — само нови клипове.
**Менютата и бутоните** на OBS, Kdenlive и качването са в [03_PRODUCTION_MANUAL_BG.md](./03_PRODUCTION_MANUAL_BG.md); тук се сочи „→ Р 2.3“ и т.н.
**Всички цени и факти по-долу** са прочетени от живия API на unsolero.com на 24 септ. 2026. Всяка цена се проверява отново на страницата на вендора **в деня на записа** — точно какво и къде е написано при всеки клип.

---

## 0. СТОП — две неща преди да снимаш каквото и да е

### 0.1 Офертите изгасват утре. Всяко видео ще води към страници без бутони.

`OFFER_MAXIMUM_AGE=720h` (30 дни) от последния seed. Измерено днес от production API:

| Изгасва (българско време) | Оферти |
|---|---|
| **25 септ., 13:48** | Zoho Invoice, Zoho Books, Bigin, Zoho CRM, Zoho Campaigns, Zoho Projects, Zoho Bookings, MailerLite, Kit, Cal.com — **10 от 15** |
| 27 септ., 17:18 | Pipedrive |
| 28 септ., 12:07 | ActiveCampaign |
| 2 окт., 12:29 / 14:13 | Teachable, SE Ranking, monday.com |

След 2 октомври **нито една** страница не печели. И дори сега, всяка оферта показва на екрана „**Price last read Aug 26 · not re-verified since**“ — в кадър това казва „тези цени са стари“, точно обратното на обещанието във всеки скрипт.

**Поправката (готова, 24 септ.):** цените са препрочетени от страниците на вендорите и съвпадат с каталога. Seed-ът `backend/seeds/offer_reverification_2026_09_24.sql` обновява `last_checked_at` за **13 оферти и 3 промоции**. Ако цена в базата се различава от прочетената, seed-ът спира и не променя нищо. Пуска се на сървъра с командата от отговора в чата.

**Не са препрочетени, изтичат, ако не ги провериш ръчно:**
- **Pipedrive** (изтича 27 септ.): pipedrive.com блокира автоматичното четене.
- **SE Ranking** (изтича 2 окт.): от България страницата показва само евро (€109 / €87.20).

Отвори двете ценови страници в браузъра и кажи какво пише. Ако са в евро, ползвай безплатен VPN с US сървър.

**Открито при проверката:** Zoho Campaigns ($7 месечно) и Zoho Projects ($5 месечно) вече публикуват **месечна** цена. Според правилото на сайта сравняваната цена трябва да мине на месечната. Тогава стекът за 3 души става **$62** ($59 при годишно плащане за Projects, $47 с безплатния план на Projects). **Преди да снимаш F, J, K и Дълго 2 реши кое число ще кажеш на камера.**

**Готово, когато:** `curl -s https://unsolero.com/api/catalog/offers` връща 15 оферти и при 13 от тях пише `"freshness_status":"fresh"`.

### 0.2 Калкулаторът за места греши за monday.com при 1, 2, 4, 6–9 души

monday продава места в пакети (минимум 3; после 5, 10, 15…). Калкулаторът на `/categories/project-management` умножава линейно: при 6 души показва $72, реалната сметка е за 10 места. **В Short K снимай само при team size = 5** (точен пакет). Поправката в кода е отделна задача.

---

## 1. Кое първо и защо — по пари, не по гледания

Shorts линковете не са кликаеми (YouTube от 2023; TikTok преди 1,000 последователи). **Парите идват от дългото видео** — Shorts са, за да те намерят. Затова всеки Short по-долу води към страница, в която има поне един печеливш продукт, и се свързва като „Related video“ към Дълго 2.

| Клип | Hook (първото изречение) | Страница | Печели в нея | Оценка |
|---|---|---|---|---|
| **Дълго 2** | „$59 a month. That's the whole software stack for a three-person agency.“ | `/stacks/agency-3-people-under-150` | Bigin, Zoho Books, Zoho Projects (3/3) | **отлично** — единственото с кликаем линк |
| **I** | „Nine-dollar CRM. Watch what happens when I add people.“ | `/categories/crm` | Bigin, Zoho CRM, Pipedrive (3/6) | **отлично** — живо действие на екрана, собствен инструмент |
| **H** | „Three SEO tools. One pays me. Most of you shouldn't buy it.“ | `/compare/ahrefs-vs-semrush` | SE Ranking ($129 × 30% ≈ $38.70 на продажба — най-високото еднократно) | **отлично** — hook-ът на доверието |
| **G** | „Sell $2,000 a month — and the free one costs the most.“ | `/compare/teachable-vs-thinkific-vs-gumroad` | Teachable (1/3) | **добро** — силен обрат, но страницата честно препоръчва Thinkific над $1,000 |
| **F** | „This is the entire software stack for a three-person agency. $59 a month.“ | `/stacks/agency-3-people-under-150` | 3/3 | **добро** — захранва Дълго 2 |
| **K** | „Six people on monday.com? You pay for ten seats.“ | `/categories/project-management` | monday.com, Zoho Projects (2/5) | **добро** — само ако пакетите се потвърдят (виж K) |
| **J** | „If you work alone, your business software can cost zero dollars.“ | `/stacks/agency-3-people-under-150` | безплатни планове — ~$0 комисиона | **обхват, не пари** — прави го, защото „free“ hook-овете носят най-много гледания (Kevin 8.6M) |

**Ненужно — не снимай:** Kit vs MailerLite (Short A и Дълго 1 го покриват); „Brevo charges per send“ (Brevo не печели); всичко към Slack, Google Analytics, Zapier, Canva, Webflow, Stripe, Shopify, help desk — там печелят **0** продукта.

**Ред на публикуване** (след като 0.1 е готово): Дълго 2 → I → F → H → G → K → J, по един на ден. Ако Short A–E още не са публикувани, A остава първи.

---

## 2. Общи правила за всеки клип (прочети веднъж)

### 2.1 Hook-ът — трите слоя в кадър 0

Всеки hook е **три неща едновременно, от първия кадър**: (1) картина, която вече се движи или е zoom-ната върху число; (2) голям текст отгоре (≤ 6 думи, главни букви); (3) гласът започва на 0.0 сек. Без лого, без „Hey guys“, без пауза. 63% от най-добрите TikTok видеа казват основното в първите 3 сек. (→ 02 §4).

След hook-а веднага **foreshadow** — изречение, което обещава обрат („…and the smartest part is what's not on it“). После тялото с „**but**“ обрати на всеки 5–6 сек. Край: CTA + въпрос за коментар.

### 2.2 Мишката и скролът

- Мишката **влиза бавно, спира върху целта, стои 2 сек., не трепери**. Пауза 1 сек. преди и след всяко движение — място за рязане.
- Скрол: **колелцето, по едно щракване** на всеки ~0.5 сек. — никога тракпад (прескача). В Kdenlive после се ускорява, ако е бавно.
- **Никога не кликвай „View at …“ или бутон с етикет „Affiliate link“ на запис.** Записаният клик е реален клик; програмите забраняват изкуствени кликове. Само hover.
- Преди всеки запис: Do Not Disturb включено, consent банерът приет/затворен, само нужните табове отворени.

### 2.3 Екранът

Chrome в 576×1024 през DevTools device mode, zoom **125%** (→ Р 1.4, 05 M08). OBS профил `Vertical 9-16`, сцена `Unsolero vertical` (→ Р 1.1–1.5). **Пробен запис 20 сек. всяка сесия** (→ Р 1.8).

Всеки кадър е **отделен файл**. Имената: `<клип>-S<номер>.mp4`, например `I-S2.mp4`. Преименувай веднага след стоп.

### 2.4 Гласът

Записва се **отделно, първо**, в Audacity (→ Р 1.6), по един файл на клип: `<клип>-voice.wav`. Чети колоната „Глас“ дословно, 2.5–2.8 думи в секунда. Всеки ред — отделен take с 1 сек. тишина между тях (лесно рязане). Ако сгрешиш — кажи реда отново, не спирай записа. Числата се казват както са написани в „Глас“ („one twenty-nine“, не „$129“).

### 2.5 Монтажът (Kdenlive, → Р 2.1–2.7)

Ред: гласът на пътечка A1 → кадрите под него, нарязани по гласа → zoom-ове (→ Р 2.3) → надписите (→ Р 2.4) → субтитри с Whisper (→ Р 2.5) → звук −14 LUFS (→ Р 2.6) → render 1080×1920 (→ Р 2.7).
- **Надпис отгоре:** бял текст, черен контур 6 px, горната трета (извън safe зоната долу → Р 5.5).
- **Субтитри:** средата на екрана, по 2–4 думи.
- **Zoom:** 100% → 160% за 0.3 сек. върху числото, за което се говори.
- **Реже се на всеки 2–3 сек.** Ако кадър стои повече от 3 сек. без промяна — сложи zoom или смени надписа.

### 2.6 Проверка преди публикуване (всеки клип)

1. Гледай го на телефона, звук изключен: разбира ли се от надписите? 2. Звук включен: чува ли се всяка цифра? 3. Първите 2 сек.: ясно ли е за какво е, без контекст? 4. Цените на екрана = цените, прочетени днес? 5. Дисклоузърът е казан с глас и написан в описанието? Ако едно е „не“ — не качвай.

### 2.7 UTM

`?utm_source=<youtube|tiktok|instagram|linkedin>&utm_medium=<shorts|reels|video|post>&utm_campaign=2026-10-<клип>`
Кампаниите: `2026-10-stack59`, `2026-10-crm-seats`, `2026-10-seo-cliff`, `2026-10-course-math`, `2026-10-pm-seats`, `2026-10-solo-zero`, `2026-10-long-stack`. Отчетът по кампании е в admin analytics.

---

## Short I — „Nine-dollar CRM. Watch what happens when I add people.“

**Дължина:** 34–36 сек. · **Страница:** `https://unsolero.com/categories/crm` · **Hook модел:** №22 ценова котва + живо действие · **Печели:** Bigin, Zoho CRM, Pipedrive.

### Провери в деня на записа (10 мин.)

| Какво | Къде | Днес в сайта | Ако се различава |
|---|---|---|---|
| Bigin Express, месечно на потребител | `bigin.com/pricing` | $9 (годишно $7) | Спри. Сайтът се поправя първо, после снимаш. |
| Zoho CRM Standard | `zoho.com/crm/zohocrm-pricing.html` | $20 (годишно $14) | същото |
| Pipedrive Lite | `pipedrive.com/pricing` | $19.90 (годишно $14) | същото |
| Salesflare Growth | `salesflare.com/pricing` | $39 (годишно $29) | същото |
| Bigin Express: 3 pipelines лимит | Bigin pricing, сравнение на плановете | „caps three pipelines“ | махни реда 17–23 сек. |

Ако вендорската страница е в EUR (геолокация) — не превръщай в USD. Отвори я в прозорец Incognito; ако пак е EUR, пиши в описанието, че цените са прочетени в USD на [дата] от сайта на UNSOLERO и не показвай вендорската страница в кадър.

### Табове (в този ред)

1. `unsolero.com/categories/crm`

### Кадри

| Файл | Какво правиш | Колко записваш |
|---|---|---|
| **I-S1** | Страницата скролната така, че секцията **„Monthly total at your team size“** и полето **Team size** (показва **5**) са в кадър. Мишката влиза от долния десен ъгъл и спира върху бутона **+** („More seats“). Стои 2 сек. | 6 сек. |
| **I-S2** | Същият кадър. Таблицата показва 5 места. Мишката бавно минава отгоре надолу по колоната **„Per month, 5 seats“** — спира 1 сек. на всеки ред. | 12 сек. |
| **I-S3** | Мишката върху **+**. Кликвай **+ пет пъти, по един клик в секунда** (5 → 10). Заглавието на колоната става „Per month, 10 seats“, числата се удвояват. След последния клик стой 3 сек. | 10 сек. |
| **I-S4** | Скрол нагоре до продуктовата карта на **Bigin Express**; мишката спира върху цената. | 6 сек. |
| **I-S5** | Кликни **−** пет пъти (обратно на 5). Скролни до таблицата, мишката спира върху реда **Pipedrive Lite**, после **Zoho CRM Standard**. | 10 сек. |
| **I-S6** | Мишката върху полето **Team size**, стои неподвижно (фон за CTA). | 8 сек. |

Очаквани числа (провери ги на екрана): при 5 — Bigin $45, Freshsales $55, Pipedrive $99.50, HubSpot $100, Zoho CRM $100, Salesflare $195. При 10 — $90, $110, $199, $200, $200, $390.

### Сценарий

| Сек. | Кадър | Глас (английски, дословно) | Надпис отгоре | Монтаж |
|---|---|---|---|---|
| 0–3 | I-S3 (първите 2 клика) | „Nine-dollar CRM. Watch what happens when I add people.“ | `$9 CRM × YOUR TEAM =` | Zoom 150% върху полето Team size от кадър 0 |
| 3–8 | I-S2 | „Five people. Bigin: forty-five a month. Pipedrive and Zoho CRM: about a hundred.“ | `5 PEOPLE` | Zoom върху всеки ред, докато се казва |
| 8–11 | I-S2 (реда Salesflare) | „Salesflare: one ninety-five.“ | `$195/mo` | Zoom 170% |
| 11–17 | I-S3 (кликовете до 10) | „Ten people. Everything doubles. That's per-seat pricing — the sticker is never the bill.“ | `10 PEOPLE → ×2` | Ускори S3 ×1.5, числата да мигат |
| 17–22 | I-S4 | „**But** cheapest isn't the answer. Bigin Express caps you at three pipelines.“ | `CHEAPEST ≠ RIGHT` | Zoom върху цената на Bigin |
| 22–29 | I-S5 | „One pipeline and an inbox? Bigin. A real sales team? Pipedrive. Already on Zoho? Zoho CRM.“ | `SIMPLE → BIGIN` / `SALES TEAM → PIPEDRIVE` / `ON ZOHO → ZOHO CRM` (по един на 2 сек.) | Надписите се сменят с гласа |
| 29–35 | I-S6 | „Put in your own team size — it's free at unsolero dot com, link in bio. Some links pay us; the math doesn't care. What's your team size?“ | `unsolero.com/categories/crm` / `TEAM SIZE? ↓` | Цени прочетени [дата] — малък текст долу вляво |

**Cover:** кадърът от 11–17 с числата при 10 и надписа `$9 CRM × YOUR TEAM =`.
**Заглавие (YouTube):** `Your $9 CRM is not $9 #shorts` · **TikTok/IG caption първи ред:** `Per-seat pricing, live: 5 people vs 10 people on 6 CRMs.`
**Описание:**
```
What 6 CRMs cost a team of 5 and a team of 10, per month — Bigin, Freshsales, Pipedrive, HubSpot, Zoho CRM, Salesflare. Prices read from each vendor's page on [date], monthly billing.

Free calculator with your own team size: https://unsolero.com/categories/crm?utm_source=youtube&utm_medium=shorts&utm_campaign=2026-10-crm-seats

Some links on UNSOLERO pay a commission. Commission never changes the ranking or the math.

#crm #smallbusiness #saas
```
**Закачен коментар:** `What's your team size? I'll tell you which one fits.`
**Hook B (за повторен тест след 48 ч., → Р 6.3):** „‚Sick of paying per user pricing.' Real Reddit post. Here's why.“ + надпис `SICK OF PER-USER PRICING?`

---

## Short F — „$59 a month. The whole stack.“

**Дължина:** 34–36 сек. · **Страница:** `https://unsolero.com/stacks/agency-3-people-under-150` · **Hook модел:** №22 ценова котва + foreshadow · **Печели:** Bigin, Zoho Books, Zoho Projects.

### Провери в деня на записа

| Какво | Къде | Днес в сайта |
|---|---|---|
| Bigin Express | `bigin.com/pricing` | $9/потребител месечно |
| Zoho Books Standard | `zoho.com/books/pricing` (US изглед) | $20/месец, фиксирано |
| Zoho Projects Premium | `zoho.com/projects/pricing.html` | $4/потребител, **само годишно** |

### Табове

1. `unsolero.com/stacks/agency-3-people-under-150`

### Кадри

| Файл | Какво правиш | Колко |
|---|---|---|
| **F-S1** | Скролни до абзаца под „**The stack**“, който започва с „27 + 20 + 12 = 59 USD a month“. Мишката спира под „59 USD“. | 6 сек. |
| **F-S2** | Скролни до заглавието „**The stack**“. Бавно надолу през трите точки (Bigin → Zoho Books → Zoho Projects); мишката спира на цената във всяка точка 2 сек. | 15 сек. |
| **F-S3** | Скролни до „**Where to get them**“ — трите оферт карти. Бавно през тях; hover над цената (**не** над бутона „View at Zoho“). | 10 сек. |
| **F-S4** | Скролни до „**What we left out, and why**“. Бавно през списъка; спри на реда „Team chat — Slack Pro…“ и на „Scheduling — …Calendly…“. | 12 сек. |
| **F-S5** | Скролни горе до заглавието на страницата; мишката неподвижна. | 6 сек. |

### Сценарий

| Сек. | Кадър | Глас | Надпис | Монтаж |
|---|---|---|---|---|
| 0–3 | F-S1 | „This is the entire software stack for a three-person agency. Fifty-nine dollars a month.“ | `3 PEOPLE. WHOLE STACK. $59/mo` | Zoom 170% върху „59 USD“ от кадър 0 |
| 3–5 | F-S1 | „And the smartest part is what's *not* on it.“ | `…WHAT'S NOT ON IT` | лек zoom out |
| 5–10 | F-S2 (Bigin) | „Client records: Bigin, by Zoho. Nine a seat, three seats — twenty-seven.“ | `CRM · BIGIN · 3 × $9 = $27` | zoom на цената |
| 10–15 | F-S2 (Books) | „Invoicing: Zoho Books, twenty flat — with the bank feed that shows who actually paid.“ | `INVOICING · ZOHO BOOKS · $20` | |
| 15–19 | F-S2 (Projects) | „Projects: Zoho Projects, four a seat, billed yearly. Twelve.“ | `PROJECTS · 3 × $4 = $12 · yearly` | |
| 19–25 | F-S4 | „Budget was one-fifty. Ninety-one left — **on purpose**. No Slack. No Calendly. No big CRM.“ | `$91 LEFT. ON PURPOSE.` → `NO SLACK · NO CALENDLY` | червено зачертаване върху думите Slack / Calendly |
| 25–29 | F-S4 | „Chat and booking don't stop work or stop payment. Add them the day they do.“ | `BUY WHEN IT BREAKS` | |
| 29–35 | F-S5 | „Every price, and every tool we rejected, is at unsolero dot com — link in bio. Some links pay us; the stack doesn't change. What's your team size?“ | `unsolero.com/stacks` / `TEAM SIZE? ↓` | Цени [дата] долу вляво |

**Cover:** 0–3 сек. с `$59/mo`.
**Заглавие:** `A 3-person agency's whole software stack: $59/month #shorts`
**Описание:** като I, със страницата `/stacks/agency-3-people-under-150` и кампания `2026-10-stack59`. Добави: `All three are Zoho — one vendor means no integration project. The page says when that's the wrong call.`
**Закачен коментар:** `"Why all Zoho?" — one login, one client record, no integration work. When you already own a CRM or invoicing tool, keep it; the page explains.` (Това е възражението, което ще дойде първо — отговори го предварително.)
**Hook B:** „‚Death by twenty-nine dollars a month.' That's how one agency owner described their software bill.“ + `DEATH BY $29/MONTH`.
**Related video:** Дълго 2.

---

## Short H — „Three SEO tools. One pays me. Most of you shouldn't buy it.“

**Дължина:** 35–37 сек. · **Страница:** `https://unsolero.com/compare/ahrefs-vs-semrush` · **Hook модел:** №4 „Do NOT…“ + прозрачност (никой в нишата не казва това) · **Печели:** SE Ranking.

### Провери в деня на записа

| Какво | Къде | Днес в сайта |
|---|---|---|
| Ahrefs Starter | `ahrefs.com/pricing` | $29/месец |
| SE Ranking Core | `seranking.com/subscription.html` | $129 месечно, $103.20 годишно |
| Semrush SEO (най-евтиният) | `semrush.com/prices` | $139 месечно |
| Ahrefs Starter няма Site Audit и Rank Tracker | Ahrefs pricing, таблицата | „withholds site audit and rank tracking“ |

Semrush преопакова плановете си през 2026 — ако най-евтиният вече не е $139, **сайтът се поправя преди записа**.

### Табове

1. `unsolero.com/compare/ahrefs-vs-semrush`
2. `ahrefs.com/pricing` (само ако цената съвпада)

### Кадри

| Файл | Какво правиш | Колко |
|---|---|---|
| **H-S1** | Скролни до „**SE Ranking sits in the gap**“. Изречението „It is also, we should say plainly, the one product in this comparison whose link earns us a commission — and it is still not the one most readers here should buy.“ в кадър. Мишката бавно подчертава реда с hover движение отляво надясно. | 8 сек. |
| **H-S2** | Скролни до „**The prices**“. Мишката спира на всяка от трите цени по 2 сек. (Ahrefs → SE Ranking → Semrush). | 10 сек. |
| **H-S3** | Скролни до callout-а „**The test that decides it**“. Мишката неподвижна до него. | 8 сек. |
| **H-S4** | Скролни до „**What Ahrefs Starter withholds**“. Спри. | 6 сек. |
| **H-S5** | Таб 2: Ahrefs pricing, колоната Starter; мишката спира на цената. (Ако е EUR/геолокирано — пропусни кадъра, ползвай H-S2.) | 6 сек. |
| **H-S6** | Скролни до продуктовите карти долу; hover над цената на SE Ranking (не над бутона). | 6 сек. |

### Сценарий

| Сек. | Кадър | Глас | Надпис | Монтаж |
|---|---|---|---|---|
| 0–3 | H-S1 | „Three SEO tools. One of them pays me a commission. Most of you shouldn't buy it.“ | `DON'T BUY THE ONE THAT PAYS ME` | Zoom 160% върху „earns us a commission“, жълт highlight |
| 3–7 | H-S2 / H-S5 | „Ahrefs Starter: twenty-nine a month. Then — a cliff.“ | `AHREFS STARTER $29` | |
| 7–11 | H-S2 | „SE Ranking: one twenty-nine. Semrush: one thirty-nine.“ | `$29 → $129 → $139` | Числата се появяват едно по едно |
| 11–16 | H-S3 | „The question isn't which is best. It's this: is SEO your *job* — or a task on your list?“ | `JOB OR TASK?` | Zoom върху callout-а |
| 16–21 | H-S3 | „A task — a few keywords, a competitor's backlinks? Ahrefs Starter. That's all you need.“ | `TASK → AHREFS $29` | |
| 21–26 | H-S4 | „**But** Starter has no site audit and no rank tracking. If SEO is your job, you need a full platform.“ | `NO SITE AUDIT · NO RANK TRACKING` | |
| 26–31 | H-S6 | „Then SE Ranking — ten projects, a bit less than Semrush. That's the one that pays me.“ | `JOB → SE RANKING $129` | |
| 31–37 | H-S1 (широко) | „Prices read on [date]. Full comparison at unsolero dot com — link in bio. Job or task? Tell me below.“ | `unsolero.com/compare` / `JOB OR TASK? ↓` | |

**Cover:** `DON'T BUY THE ONE THAT PAYS ME`.
**Заглавие:** `Ahrefs $29 vs Semrush $139 — the SEO price cliff #shorts`
**Описание:** страницата `/compare/ahrefs-vs-semrush`, кампания `2026-10-seo-cliff`; добави: `Index size and data quality can't be compared from a pricing page — anyone ranking them on that from public info is guessing.`
**Закачен коментар:** `If SEO is a task for you: Ahrefs Starter. If it's your job: tell me how many sites you manage.`
**Hook B:** „Semrush is one thirty-nine. Ahrefs is twenty-nine. That's not a comparison — it's a cliff.“ + `$29 → CLIFF → $139`.
**LinkedIn версия (текст):** ред 1 `I run a software comparison site. In our SEO comparison, the only tool that pays us a commission is not the one most readers should buy. Here's the page saying so.` + скрийншот на H-S1.

---

## Short G — „Sell $2,000 a month — and the free one costs the most.“

**Дължина:** 35–37 сек. · **Страница:** `https://unsolero.com/compare/teachable-vs-thinkific-vs-gumroad` · **Hook модел:** №22 + парадокс · **Печели:** Teachable.

### Провери в деня на записа

| Какво | Къде | Днес в сайта |
|---|---|---|
| Gumroad такса | `gumroad.com/pricing` | 10% + 50¢ на продажба (30% през Discover) |
| Teachable Starter | `teachable.com/pricing` (**прочети цената на страницата, не структурираните данни** — те показват други числа) | $39 месечно, $29 годишно, **7.5% такса** |
| Thinkific Basic | `thinkific.com/pricing` | $40, годишно, 0% такса |

Сметката при $2,000/месец (напр. 40 продажби по $50): Gumroad $200 + 40 × $0.50 = **$220**; Teachable $39 + 7.5% × 2,000 = **$189**; Thinkific **$40**. Ако някоя цена се е сменила — смени и сметката.

### Табове

1. `unsolero.com/compare/teachable-vs-thinkific-vs-gumroad`

### Кадри

| Файл | Какво правиш | Колко |
|---|---|---|
| **G-S1** | Началото на страницата: заглавието и първия абзац („You cannot compare these on the monthly price…“). Мишката неподвижна. | 6 сек. |
| **G-S2** | Скролни до „**The prices, and the cut**“; спри на всяка от трите точки по 2 сек. | 10 сек. |
| **G-S3** | Скролни до „**Where the lines cross**“; бавно през абзаца, спри на „At 2,000, Thinkific's flat 40…“. | 10 сек. |
| **G-S4** | Скролни до callout-а „**The test that decides it**“. | 6 сек. |
| **G-S5** | Скролни до „**They are also not the same product**“ („Gumroad sells a file…“). | 6 сек. |
| **G-S6** | Продуктовите карти долу; hover над цената на Teachable (не над бутона). | 6 сек. |

**Карта в Kdenlive (title clip, → Р 2.4):** `G-CARD` — три колони: `GUMROAD $220` / `TEACHABLE $189` / `THINKIFIC $40`, под тях малко `at $2,000/month in sales`. Червената колона е Gumroad.

### Сценарий

| Сек. | Кадър | Глас | Надпис | Монтаж |
|---|---|---|---|---|
| 0–3 | G-S2 | „Gumroad is free. Teachable's thirty-nine. Thinkific's forty.“ | `FREE · $39 · $40` | Zoom върху трите цени последователно |
| 3–6 | G-CARD | „Sell two thousand dollars a month — and the free one costs the most.“ | `THE FREE ONE COSTS THE MOST` | Gumroad колоната светва червено |
| 6–11 | G-S2 (Gumroad) | „Gumroad takes ten percent plus fifty cents a sale. That's over two hundred dollars.“ | `10% + 50¢ = $220` | |
| 11–16 | G-S2 (Teachable) | „Teachable Starter: thirty-nine, plus a seven-and-a-half percent fee. One eighty-nine.“ | `$39 + 7.5% = $189` | |
| 16–19 | G-S3 | „Thinkific: forty, billed yearly. No cut.“ | `$40 FLAT` | Zoom на „At 2,000…“ |
| 19–25 | G-S4 | „**But** at a few hundred a month, flip it — Gumroad wins, because you pay nothing until you sell.“ | `UNDER A FEW $100 → GUMROAD` | |
| 25–29 | G-S5 | „And if you're selling one PDF, you don't need a course platform at all.“ | `1 PDF ≠ COURSE PLATFORM` | |
| 29–37 | G-S6 → G-S1 | „The crossover math is at unsolero dot com — link in bio. Only Teachable pays us, and it still loses at two thousand. What do you sell a month?“ | `unsolero.com/compare` / `MONTHLY SALES? ↓` | |

**Cover:** G-CARD с `THE FREE ONE COSTS THE MOST`.
**Заглавие:** `Gumroad vs Teachable vs Thinkific: the "free" one costs the most #shorts`
**Описание:** страницата, кампания `2026-10-course-math`; добави: `Payment processing fees sit on top of all three and are roughly equal, so none of these numbers is what lands in your account.`
**Закачен коментар:** `Tell me your monthly sales and I'll do the math for your number.` — **и после наистина я прави** в отговорите (всеки отговор е нов сигнал за алгоритъма, → 02 §8.6 т.10).
**Hook B:** „Teachable's own website shows two different prices for the same plan.“ — **само ако в деня на записа** `view-source:teachable.com/pricing` → `Ctrl+F` `"price"` още показва различни числа от видимите на страницата. Кадър: view-source с highlight-нато число, после видимата цена. Ако вече съвпадат — не ползвай този hook.

---

## Short K — „Six people on monday.com? You pay for ten seats.“

**Дължина:** 33–35 сек. · **Страница:** `https://unsolero.com/categories/project-management` · **Hook модел:** №21 предупреждение · **Печели:** monday.com, Zoho Projects.

### Провери в деня на записа — **задължително, hook-ът зависи от това**

| Какво | Къде | Какво търсиш |
|---|---|---|
| Пакетите места на monday | `monday.com/pricing` → падащото меню за брой места | Има ли 6? Ако опциите са 3, 5, 10… — hook A е верен |
| monday Basic цена | същото | $12 месечно / $9 годишно на място |
| Zoho Projects безплатно до 5 | `zoho.com/projects/pricing.html` | Free план, брой потребители |
| ClickUp, Notion, Teamwork | техните pricing страници | $10 / $12 / $12.99 месечно |

**Ако monday вече позволява 6 места:** hook A отпада. Ползвай hook B: „Project management for five people: zero, twenty, or sixty-five dollars.“ + `5 PEOPLE: $0 / $20 / $65`.

### Табове

1. `monday.com/pricing`
2. `unsolero.com/categories/project-management`

### Кадри

| Файл | Какво правиш | Колко |
|---|---|---|
| **K-S1** | Таб 1: отвори падащото меню за брой места (seats); мишката бавно минава по опциите 3 → 5 → 10. Задръж на 10. | 8 сек. |
| **K-S2** | Таб 2: калкулаторът „Monthly total at your team size“, Team size = **5** (по подразбиране). Мишката бавно по колоната „Per month, 5 seats“, по 1 сек. на ред. **Не променяй числото.** | 12 сек. |
| **K-S3** | Скролни до продуктовата карта на monday.com Basic; hover над описанието (там пише „three-seat minimum“ и „omits the timeline and calendar views“). | 8 сек. |
| **K-S4** | Картата на Zoho Projects Premium; hover над цената. | 6 сек. |
| **K-S5** | Обратно на калкулатора; мишката върху Team size, неподвижна. | 8 сек. |

Очаквани числа при 5: Zoho Projects $20 (billed yearly), ClickUp $50, monday $60, Notion $60, Teamwork $64.95.

### Сценарий

| Сек. | Кадър | Глас | Надпис | Монтаж |
|---|---|---|---|---|
| 0–3 | K-S1 | „Six people on monday dot com? You pay for ten seats.“ | `6 PEOPLE = 10 SEATS` | Zoom върху „10“ в менюто |
| 3–7 | K-S1 | „It sells seats in buckets — three, five, ten. So here's what exactly five people pay on five tools.“ | `3 · 5 · 10` | |
| 7–13 | K-S2 | „Zoho Projects: twenty, billed yearly — or zero, it's free up to five. ClickUp: fifty.“ | `ZOHO $20 (or $0) · CLICKUP $50` | Zoom на редовете |
| 13–17 | K-S2 | „monday Basic: sixty. Notion: sixty. Teamwork: sixty-five.“ | `MONDAY $60 · NOTION $60 · TEAMWORK $65` | |
| 17–23 | K-S3 | „**But** monday's Basic plan leaves out timeline and calendar views — the ones most teams end up wanting.“ | `NO TIMELINE · NO CALENDAR` | Zoom върху изречението в описанието |
| 23–28 | K-S4 | „Real question: who's doing what, what's late, what's waiting on the client. Anything that answers those three is enough.“ | `WHO? LATE? WAITING?` | |
| 28–34 | K-S5 | „Per-seat math for your team size, free, at unsolero dot com — link in bio. Some links pay us; the order doesn't change. How many people?“ | `unsolero.com/categories` / `HOW MANY? ↓` | |

**Cover:** `6 PEOPLE = 10 SEATS`.
**Заглавие:** `monday.com seat buckets: 6 people, 10 seats #shorts`
**Описание:** страницата, кампания `2026-10-pm-seats`.
**Закачен коментар:** `Team of 6–9 on monday? Check the seat dropdown before you pay.`

---

## Short J — „If you work alone, your business software can cost zero dollars.“

**Дължина:** 30–32 сек. · **Страница:** `https://unsolero.com/stacks/agency-3-people-under-150` (секцията „When this stack is wrong“) · **Hook модел:** №1 „for FREE“ · **Роля:** обхват и последователи; почти не печели.

### Провери в деня на записа

| Какво | Къде | Какво трябва да е вярно |
|---|---|---|
| Bigin безплатно за 1 потребител | `bigin.com/pricing` | Free план, 1 user |
| Zoho Invoice без платен план | `zoho.com/invoice/pricing.html` | „Free“ |
| Zoho Projects безплатно до 5 | `zoho.com/projects/pricing.html` | Free план, до 5 users |

Ако едно от трите вече не е безплатно — **не снимай**; hook-ът става лъжа.

### Табове

1. `unsolero.com/stacks/agency-3-people-under-150`
2. `bigin.com/pricing` 3. `zoho.com/invoice/pricing.html` 4. `zoho.com/projects/pricing.html`

### Кадри

| Файл | Какво правиш | Колко |
|---|---|---|
| **J-S1** | Таб 1: скролни до „**When this stack is wrong**“; изречението „a solo consultant can run this whole shape at 0 USD until the second seat“ в кадър. Мишката под „0 USD“. | 6 сек. |
| **J-S2** | Таб 2: Free колоната на Bigin; мишката спира върху „Free“. | 5 сек. |
| **J-S3** | Таб 3: „Free“ на Zoho Invoice. | 5 сек. |
| **J-S4** | Таб 4: Free планът на Zoho Projects и броя потребители. | 5 сек. |
| **J-S5** | Таб 1: скролни до „**The stack**“ (цените при 3 души). | 8 сек. |

### Сценарий

| Сек. | Кадър | Глас | Надпис |
|---|---|---|---|
| 0–3 | J-S1 | „If you work alone, your business software can cost zero dollars.“ | `SOLO BUSINESS SOFTWARE: $0` |
| 3–6 | J-S1 | „Not a trial. Free plans — with one catch at the end.“ | `NOT A TRIAL` |
| 6–10 | J-S2 | „Client records: Bigin — free for one user.“ | `CRM · BIGIN · $0` |
| 10–15 | J-S3 | „Invoices, estimates, expenses: Zoho Invoice. Free. There's no paid version.“ | `INVOICING · $0` |
| 15–19 | J-S4 | „Projects: Zoho Projects — free up to five people.“ | `PROJECTS · $0` |
| 19–25 | J-S5 | „**The catch:** hire person two, and Bigin becomes nine a seat. And Invoice isn't accounting — no bank reconciliation.“ | `THE CATCH: PERSON #2` |
| 25–31 | J-S1 | „Plan for that day, not today. What it costs at three people is at unsolero dot com — link in bio. Some links pay us.“ | `unsolero.com/stacks` / `SOLO? ↓` |

**Cover:** `SOLO BUSINESS SOFTWARE: $0`. **Заглавие:** `A $0 software stack for a one-person business #shorts`. **Кампания:** `2026-10-solo-zero`. **Related video:** Дълго 2.

---

## Дълго 2 — „A 3-Person Agency's Entire Software Stack for $59 a Month (2026 — Every Price Read This Week)“

**Дължина:** 8–9 мин. · **Формат:** 1920×1080, запис на екрана + глас · **Страница:** `/stacks/agency-3-people-under-150` · **Това е видеото, което носи кликове** — линкът в описанието е кликаем.

### Подготовка на екрана (различна от Shorts)

OBS: профил за хоризонтално (1920×1080, → Р 1.1). Chrome на цял екран (`F11`), zoom **110%**. Табове в този ред:
1. `unsolero.com/stacks/agency-3-people-under-150`
2. `bigin.com/pricing` 3. `zoho.com/books/pricing` 4. `zoho.com/projects/pricing.html`
5. `unsolero.com/categories/crm` 6. `unsolero.com/build`

**Builder-ът (глава 7):** пусни го веднъж **без запис** със същите входове. Ако резултатът е грешка или е нелогичен — махни главата, не я снимай.

### Глави

| Време | Кадър / действие | Глас — пълен текст за началото, точки за останалото |
|---|---|---|
| **0:00–0:30 Hook** | Таб 1, zoom на „27 + 20 + 12 = 59 USD“ → бавен скрол до „What we left out“ | „This is the entire software stack for a three-person agency: client records, invoicing, project tracking — fifty-nine dollars a month. The budget was one hundred and fifty. We left ninety-one dollars unspent, on purpose, and the tools we *didn't* pick are the most useful part of this video. Every price was read from the vendor's own page this week; the date is on screen.“ |
| **0:30–1:00 Какво ще научиш + дисклоузър** | Title card с 4 точки: `The brief · The 3 tools · What we left out · When this is wrong` | „Four parts. The brief. The three tools and the exact math. What we left out and why. And when this stack is the wrong answer for you. All three tools here pay UNSOLERO a commission if you subscribe. That doesn't change the picks — the ranking code can't even read commission data. The link is first in the description.“ |
| **1:00–2:00 Брифът** | Таб 1, „Who this is for“; мишката по редовете | Трима души, клиентска работа, никой не е платен да поддържа софтуер, $150 таван. Ако вече имате CRM или фактуриране — прескочете на последната глава. |
| **2:00–3:15 Bigin** | Таб 2 (Free / Express колоните) → таб 1 точка 1 | $9 месечно, $7 годишно, 3 места = $27. Безплатно за 1. Express: 3 pipelines, 50,000 записа, 30 автоматизации — повече, отколкото трима ще ползват. Защо не Zoho CRM: $60 за три места за дълбочина, която не ви трябва; премини, когато дълбочината е това, което ти липсва. |
| **3:15–4:30 Zoho Books** | Таб 3 → таб 1 точка 2 | $20 фиксирано. Банковият feed — кой реално е платил. Защо не Zoho Invoice (безплатен): не е счетоводство, няма reconciliation. Безплатният план на Books под $50K годишен оборот — провери за твоята държава. |
| **4:30–5:30 Zoho Projects** | Таб 4 → таб 1 точка 3 | $4 на място, само годишно; 3 места = $12. Безплатно до 5 — трима могат да започнат с $0 → общо $47. Трите въпроса: кой какво прави, какво закъснява, какво чака клиента. |
| **5:30–6:45 Какво оставихме** | Таб 1 „What we left out, and why“ — бавен скрол, zoom на всеки ред | monday.com и ClickUp (повече от двойно за същите 3 въпроса; monday Basic няма timeline/calendar); Slack / Teams / Workspace (чатът не спира работа — Workspace, ако ти трябва имейл на домейн); Calendly / Cal.com / Zoho Bookings (когато записването стане тясно място; започни от безплатните). |
| **6:45–7:30 Калкулаторът** | Таб 5, Team size 3 → 5 → 10, кликове на **+** | Какво става с CRM сметката, когато растете. Това е причината да не купуваш за екипа, който още нямаш. |
| **7:30–8:15 Кога това е грешно** | Таб 1 „When this stack is wrong“ | Вече имаш част от него — пази го. Продаваш продукти, не време — друг стек. Сам си — $0 до второто място. Над 10 души — друга страница. |
| **8:15–8:45 Builder (по желание)** | Таб 6: входовете от сценарий A в GROWTH_PLAYBOOK §10, бюджет 150 | „If your brief is different, the builder does this for your constraints.“ Покажи резултата само ако е проверен без запис. |
| **8:45–9:00 CTA** | Финална карта: URL голям + следващото видео | „The full page, with every price, date and rejected tool, is the first link below. Tell me your team size and budget in the comments — I answer every one. Next: what the same stack costs at ten people.“ |

**Описание:**
```
A 3-person agency's whole software stack for $59/month — Bigin, Zoho Books, Zoho Projects — and the $91 we deliberately left unspent. Every price read from the vendor's page on [date], billing basis shown.

▶ The full stack page: https://unsolero.com/stacks/agency-3-people-under-150?utm_source=youtube&utm_medium=video&utm_campaign=2026-10-long-stack
▶ Build your own stack (free): https://unsolero.com/build?utm_source=youtube&utm_medium=video&utm_campaign=2026-10-long-stack

0:00 $59 for the whole stack
0:30 What you'll learn
1:00 The brief
2:00 Client records: Bigin
3:15 Invoicing: Zoho Books
4:30 Projects: Zoho Projects
5:30 What we left out, and why
6:45 What happens when you grow
7:30 When this stack is wrong
8:45 Your own stack

All three tools link to affiliate programs that pay UNSOLERO a commission. Commission is never an input to the ranking — the rule is enforced by a test that fails the build.

#smallbusiness #agency #saas
```
**Закачен коментар:** двата линка отново. **End screen:** Short I (калкулатора) + Subscribe. **Thumbnail (дългите имат thumbnail):** голямо `$59/mo` + `3 PEOPLE · WHOLE STACK`, фон — скрийншот на трите цени, размазан 40%.

---

## 3. График — от днес до първите данни

| Ден | Какво | Време |
|---|---|---|
| **Пт 25.09** | 0.1: цените препрочетени, seed-овете пуснати, офертите `fresh`. **Без това нищо не се публикува.** | 2–3 ч. (или аз) |
| **Сб 26.09** | Проверки на цени за всички 7 (таблиците „Провери“). Глас за I, F, H, G, K, J в Audacity (един сеанс). | 2 ч. |
| **Сб 26.09** | Запис на кадрите: I, K (калкулаторите, един таб), F, J (stack страницата), H, G. | 2 ч. |
| **Нд 27.09** | Глас и запис на Дълго 2. | 3 ч. |
| **Пн 28 – Вт 29.09** | Монтаж: Дълго 2 първо, после I, F. | 5 ч. |
| **Ср 30.09** | Дълго 2 в YouTube (→ Р 5.3). Монтаж H, G, K, J. | 3 ч. |
| **Чт 1.10 → Вт 6.10** | По един Short на ден: I, F, H, G, K, J. Всеки в YouTube Shorts + TikTok + Instagram Reels (→ Р 5.1–5.3), **без воден знак** между платформите. Related video → Дълго 2. | 45 мин./ден |
| Всеки ден | Отговор на **всеки** коментар в първия час. На „team size“ / „monthly sales“ въпросите — реален отговор с числа. | 20 мин. |
| +48 ч. след всеки | Правилото от → Р 6.3. Ако гледанията са под медианата — препубликувай с **Hook B** (само първите 3 сек. се сменят). | 15 мин. |

**Как се мери дали прави пари:** admin analytics → кампаниите `2026-10-*` → affiliate кликове по кампания. Целта за тези две седмици (от 02 §8.5): Shorts медиана ≥ 1,000 гледания; Дълго 2 ≥ 500 гледания предимно от search за 30 дни; първите affiliate кликове с `2026-10-long-stack`. Гледания без кликове = CTA-то е късно или неясно — кажи линка по-рано в Дълго 2, не сменяй темата.
