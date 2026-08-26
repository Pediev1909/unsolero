# Подробен наръчник за видеа без лице — UNSOLERO

**Период:** първите 30 дни

**Платформи:** YouTube Shorts, TikTok, Instagram Reels

**Препоръчителен език:** английски

**Формат:** 1080×1920, 9:16, 30 fps

---

## 1. Най-добрият вариант без лице

Основният формат трябва да бъде:

> **Собствен глас + запис на екрана + големи субтитри + прости графики**

Липсата на лице не е проблем, ако видеото има:

- конкретен проблем още в първите 2 секунди;
- реален екран, таблица или диаграма;
- ясен човешки глас;
- движение или визуална промяна на всеки 2–4 секунди;
- само една основна идея;
- само един CTA.

Не се препоръчва основният формат да бъде AI voice върху случайни stock клипове. Това създава по-слабо доверие и не показва реалната стойност на UNSOLERO.

---

## 2. Шест подходящи faceless формата

### Формат A — Screen recording с voiceover

Показва се:

- страницата на UNSOLERO;
- конкретен продукт;
- comparison таблица;
- работа с builder-а.

Подходящ за:

- сравнения;
- tutorials;
- демонстрация на препоръка;
- обяснение на конкретно решение.

### Формат B — Текстови карти, които създавате в Kdenlive

Тук „карта“ не означава готова анимация, която трябва да намерите или
изтеглите. Това е обикновен правоъгълник с кратък текст, направен от вас в
`Add Title Clip` на Kdenlive. Не го снимате с OBS. OBS се използва само когато
показвате истински сайт или програма; картите се добавят по-късно при монтажа.

Най-лесният вариант е картите изобщо да не се движат: показвате една за 3
секунди, правите hard cut и я заменяте със следващата. След като това стане
лесно, по желание може да добавите плавно влизане отдолу с ефекта `Transform`.

На екрана се показват 2–4 такива текстови карти, например:

```text
Project delivery → Invoicing → CRM
```

Подходящ за:

- purchase order;
- грешки в software stack;
- какво да се купи първо;
- rejected products.

### Формат C — Split-screen сравнение

Екранът се разделя на две:

```text
CLICKUP                   TEAMWORK
Broader workspace         Client-work focus
```

Подходящ за кратки сравнения. Не показвайте повече от три критерия в едно кратко видео.

### Формат D — Kinetic text

Използват се големи думи, икони и кратки изречения върху чист фон.

Подходящ за:

- митове;
- силни правила;
- affiliate transparency;
- видеа без глас.

### Формат E — Диаграма или whiteboard

Пример:

```text
Business constraints + Verified facts
                  ↓
        Objective recommendation
                  ↓
           Merchant offers
```

Подходящ за обяснение на метода на UNSOLERO.

### Формат F — Отговор на коментар

Коментарът стои в горната част, а под него се изгражда отговорът с карти или screen recording.

Това е най-добрият резервен формат, когато няма подготвена нова тема.

---

## 3. Универсална структура

### Видео с voiceover — 25–40 секунди

1. **0–2 сек.:** голям hook.
2. **2–8 сек.:** проблемът.
3. **8–25 сек.:** решение или сравнение.
4. **25–34 сек.:** правило или заключение.
5. **34–40 сек.:** един CTA.

### Видео без лице и без глас — 12–20 секунди

1. **0–2 сек.:** въпрос или конфликт.
2. **2–12 сек.:** 3–4 кратки визуални стъпки.
3. **12–17 сек.:** заключение.
4. **17–20 сек.:** CTA.

При липса на глас текстът трябва да може да бъде прочетен удобно, без зрителят да спира видеото.

---

## 4. Готови сценарии

Следващите четири сценария са най-важната част от документа. Колоната
`Действие` казва какво правите с мишката. Колоната `Екран` казва какво трябва
да остане видимо. Колоната `Глас` е точният английски текст за четене. Не
казвайте текста от колоната `Екран`, освен ако същото изречение не присъства и
в колоната `Глас`.

### Сценарий A — Пълна демонстрация на UNSOLERO builder-а

**Цел:** зрителят да види продукта в действие, не само да чуе обещание.
**Дължина:** 34–40 секунди.
**Подготовка:** отворете homepage-а, изключете consent banner-а предварително и
се уверете, че builder-ът връща резултат.

| Време | Действие с мишката | Какво се вижда | Точен глас | Текст отгоре |
|---|---|---|---|---|
| 0–2 сек. | Не движете курсора. | Homepage, crop около headline-а и `Build My Setup`. | “Most software lists start with products. This one starts with your constraints.” | `STOP STARTING WITH PRODUCTS` |
| 2–5 сек. | Преместете курсора върху `Build My Setup`, задръжте половин секунда и кликнете. | Бутонът и после Question 1. | “I’ll build a stack for a small client-services business.” | `REAL EXAMPLE` |
| 5–8 сек. | Кликнете `Run a client services business`, после `Next`. | Избраната карта трябва ясно да промени състояние. | “First: what the business actually does.” | `1. GOAL` |
| 8–11 сек. | Кликнете `No dedicated admin`, после `Next`. | Вижда се изборът за човека, който поддържа tools. | “Nobody is paid to maintain complicated software.” | `2. TEAM CAPACITY` |
| 11–15 сек. | Кликнете полето `Exact budget in dollars`, натиснете `Ctrl+A`, въведете `120`, после `Next`. | Полето трябва да показва `$120` monthly setup budget. | “The complete stack has a one-hundred-and-twenty-dollar monthly ceiling.” | `3. TOTAL BUDGET: $120` |
| 15–19 сек. | Кликнете `Team chat`, после `Next`. | Само `Team chat` е избрано в Current stack. | “Team chat already exists, so recommending another one would be waste.” | `4. KEEP WHAT WORKS` |
| 19–23 сек. | Кликнете `The strongest tool per job`, после `Next`. | Избраната preference карта. | “This team prefers the strongest tool for each job.” | `5. PREFERENCE` |
| 23–27 сек. | Кликнете `Best value` и `Connects to my stack`, после `Build my setup`. | Двете priorities са тъмни/избрани; бутонът се натиска. | “Value and integrations matter most. Now the engine can rank the fit.” | `6. PRIORITIES` |
| 27–34 сек. | Изчакайте резултата. Scroll-нете бавно до total/monthly amount и първите cards. | Реалният резултат, score, total и продуктови карти. | “The result stays inside the budget, avoids what the team already owns, and explains every choice.” | `CONSTRAINTS → EXPLAINED STACK` |
| 34–39 сек. | Спрете движението. Поставете cursor върху score или първата reason карта. | Един стабилен close-up на резултата. | “Build your own brief at unsolero dot com.” | `TRY YOUR OWN BRIEF` |

Ако резултатът покаже error, не записвайте и не монтирайте този take. Първо
повторете същите inputs без OBS. Едва след като builder-ът завърши успешно,
направете истинския запис.

### Сценарий B — MailerLite, Brevo или Kit без хаотично отваряне на tabs

**Цел:** образователно видео, което води към метод, а не прави недоказано
“winner” твърдение.
**Дължина:** 30–34 секунди.
**Подготовка:** направете петте constraint карти в Kdenlive по процедурата в
раздел 6.9.1. Направете отделно началния hook, преходната карта, обобщението и
финалния CTA. Това са title clips, създадени от вас, а не свалени анимации.
Ако показвате vendor страници, отворете официалните pricing pages
предварително и не показвайте цена, която не е проверена същия ден.

| Време | Действие/монтаж | Какво се вижда | Точен глас | Текст на картата |
|---|---|---|---|---|
| 0–3 сек. | Static title clip, без browser. | Имената MailerLite, Brevo и Kit, еднакъв размер. | “Stop asking which email platform is best overall.” | `MAILERLITE vs BREVO vs KIT?` |
| 3–6 сек. | Имената се свиват; появява се червен X върху думата `BEST`. | Чист фон, без stock footage. | “The answer changes when your constraints change.” | `“BEST” IS THE WRONG FILTER` |
| 6–10 сек. | Появява се card 1. | Голямо число и малка envelope икона. | “First: how many subscribers do you have?” | `1. SUBSCRIBER COUNT` |
| 10–14 сек. | Hard cut към card 2. | Calendar/envelope графика. | “Second: how many emails do you send each month?” | `2. MONTHLY SEND VOLUME` |
| 14–18 сек. | Hard cut към card 3. | Проста схема `trigger → email → wait → email`. | “Third: how complex must the automation be?” | `3. AUTOMATION COMPLEXITY` |
| 18–22 сек. | Hard cut към card 4. | 1 person срещу 4 person icons. | “Fourth: who needs access, and with what permissions?” | `4. TEAM ACCESS` |
| 22–26 сек. | Hard cut към card 5. | Три boxes със свързващи линии. | “Fifth: which tools must it connect to?” | `5. REQUIRED INTEGRATIONS` |
| 26–31 сек. | Показват се трите имена под петте constraints. | Няма winner badge. | “Only then compare MailerLite, Brevo and Kit against the same list.” | `CONSTRAINTS FIRST. PRODUCT SECOND.` |
| 31–34 сек. | Финален CTA title clip. | UNSOLERO малко долу; въпросът е основен. | “Comment your list size and must-have automation.” | `LIST SIZE + AUTOMATION?` |

Не поставяйте произволни зелени отметки до vendor, ако не сте проверили
конкретния план. Това би изглеждало като факт, дори когато е само decoration.

### Сценарий C — Как комисионата е отделена от ranking-а

**Цел:** доверие и affiliate transparency.
**Дължина:** 28–34 секунди.
**Подготовка:** отворете `https://unsolero.com/affiliate-disclosure` и
`https://unsolero.com/how-it-works` в два tabs.

| Време | Действие с мишката | Какво се вижда | Точен глас | Текст отгоре |
|---|---|---|---|---|
| 0–3 сек. | Показвайте diagram title, не browser. | Две boxes: `Ranking` и `Commission data`. | “Can affiliate commission change a software ranking?” | `CAN COMMISSION CHANGE THE RANKING?` |
| 3–8 сек. | Поставете червена прекъсната линия между boxes и X върху нея. | Ясно визуално разделяне. | “On UNSOLERO, commercial data is not an input to recommendation scoring.” | `NO COMMERCIAL INPUT` |
| 8–14 сек. | Покажете flow `Goal + budget + current stack → ranking`. | Boxes се появяват една по една. | “The engine reads the goal, budget, current stack, preferences, and validated product facts.” | `CONSTRAINTS + VERIFIED FACTS` |
| 14–20 сек. | Добавете резултат box, после отделно `Merchant offers`. | Offers са след result-а, не преди него. | “It produces the recommendation first. Merchant offers are attached afterwards.” | `RECOMMENDATION → THEN OFFERS` |
| 20–27 сек. | Switch към affiliate disclosure tab; zoom върху абзаца, който казва, че commission не влияе. Дръжте cursor неподвижен до текста. | Само релевантният абзац се чете. | “That separation is written into the product architecture and tested as a business invariant.” | `PROMISE + TESTED SEPARATION` |
| 27–33 сек. | Switch към how-it-works; покажете methodology CTA/section. | Methodology page. | “Do not just trust a ranking. Check what its code is allowed to read.” | `ASK HOW IT WAS PRODUCED` |

Affiliate disclosure се добавя и в caption-а, когато конкретното видео съдържа
affiliate линк. Не скривайте disclosure-а само в profile bio.

### Сценарий D — Резервен вариант без лице и без глас

**Цел:** видео, което може да се направи при шумна среда или когато не искате
да записвате voiceover.
**Дължина:** 16–19 секунди.
**Звук:** тиха лицензирана музика или никакъв звук.

| Време | Какво правите в Kdenlive | Точен текст на екрана |
|---|---|---|
| 0–2 сек. | Title clip на тъмен фон; лек zoom от 100% към 105%. | `BEST CRM?` |
| 2–4 сек. | Hard cut; думите се свиват и се появява X. | `WRONG FIRST QUESTION.` |
| 4–7 сек. | Първият Kdenlive title clip просто се появява след hard cut. | `Which workflow is broken?` |
| 7–10 сек. | Вторият title clip заменя първия. | `Who will maintain it?` |
| 10–13 сек. | Третият title clip заменя втория. | `What must it connect to?` |
| 13–16 сек. | Нов title clip показва трите въпроса един под друг. | `CONSTRAINTS FIRST.` |
| 16–19 сек. | Финален чист кадър. | `PRODUCT SECOND.  •  UNSOLERO` |

Текстът се оставя достатъчно дълго за едно спокойно прочитане. Не добавяйте
пет икони, particles и преход едновременно — движението трябва да води окото,
не да доказва, че програмата има effects.

## Видео 1 — Не купувай CRM първо

**Формат:** текстови карти, направени в Kdenlive + voiceover
**Продължителност:** 30–35 секунди

**Екран 0–3 сек.**

> `YOUR SMALL AGENCY MAY NOT NEED A PAID CRM`

**Voiceover:**

> “A four-person agency often buys its CRM too early.”

**Екран 3–10 сек.**

```text
Late projects + unpaid invoices ≠ CRM problem
```

> “If projects are late or invoices are not going out, a CRM will not solve the real bottleneck.”

**Екран 10–24 сек.**

Три последователни Kdenlive title clips:

```text
1. Delivery
2. Invoicing
3. Client records
```

> “Start with delivery tracking, then invoicing, and keep client records simple until manual work or seat limits actually hurt.”

**Екран 24–32 сек.**

> `BUY WHEN A WORKFLOW BREAKS — NOT WHEN A LIST TELLS YOU TO`

**CTA:**

> `Follow for constraint-based software decisions.`

---

## Видео 2 — ClickUp или Teamwork

**Формат:** split screen + voiceover
**Продължителност:** 25–30 секунди

**Екран 0–3 сек.**

> `CLICKUP OR TEAMWORK?`

> “One question can shorten this comparison.”

**Екран 3–15 сек.**

Ляво:

> `Need billable work and client access?`

Дясно:

> `Put Teamwork on the shortlist.`

**Voiceover:**

> “If the project tool must understand billable work and client access, put Teamwork on your shortlist.”

**Екран 15–24 сек.**

> `Billing handled elsewhere? Compare ClickUp.`

> “If billing lives elsewhere and you want a broader workspace, compare ClickUp first.”

**Екран 24–30 сек.**

> `COMPARE THE WORKFLOW — NOT 100 FEATURES`

Преди публикуване проверете текущите функции от официалните сайтове.

---

## Видео 3 — MailerLite, Brevo или Kit

**Формат:** checklist + screen recording
**Продължителност:** 30–35 секунди

**Екран 0–3 сек.**

> `STOP LOOKING FOR THE “BEST” EMAIL TOOL`

> “Do not choose an email platform from a best-overall list.”

**Екран 3–20 сек.**

Появяват се последователно:

```text
Subscriber count
Monthly email volume
Automation complexity
Team access
Required integrations
```

> “Start with subscriber count, monthly email volume, automation complexity, team access, and the tools it must connect to.”

**Екран 20–29 сек.**

> `THEN compare MailerLite, Brevo and Kit.`

> “Only then compare MailerLite, Brevo and Kit against those constraints.”

**Екран 29–35 сек.**

> `Comment your list size + must-have automation.`

Цените и функциите се проверяват в деня на записа.

---

## Видео 4 — Как UNSOLERO отделя комисионата

**Формат:** диаграма + voiceover
**Продължителност:** 30–35 секунди

**Екран 0–3 сек.**

> `CAN COMMISSION CHANGE THE RANKING?`

> “Every affiliate comparison site claims independence. Here is the real test.”

**Екран 3–12 сек.**

```text
Can ranking code read commission data?
YES → independence is only a promise
```

> “If the ranking system can read the commission rate, independence is only a promise.”

**Екран 12–27 сек.**

```text
Constraints + verified facts → recommendation

Commercial data → shown afterwards
```

> “UNSOLERO ranks from your constraints and validated facts. Merchant offers are added only after the recommendation already exists.”

**Екран 27–35 сек.**

> `DON’T JUST TRUST A RANKING. ASK HOW IT WAS PRODUCED.`

CTA:

> `See the methodology.`

---

## Видео 5 — Три абонамента, купени прекалено рано

**Формат:** kinetic text + икони
**Продължителност:** 25–30 секунди

**Екран 0–3 сек.**

> `3 SUBSCRIPTIONS SMALL BUSINESSES BUY TOO EARLY`

**Екран 3–9 сек.**

> `1. Paid CRM before simple records become limiting`

**Екран 9–15 сек.**

> `2. Advanced analytics before meaningful traffic exists`

**Екран 15–21 сек.**

> `3. Help desk while the shared inbox still works`

**Voiceover за трите части:**

> “A paid CRM before simple records become limiting. Advanced analytics before there is meaningful traffic. And a help desk while the shared inbox still handles support.”

**Екран 21–30 сек.**

> `GOOD TOOLS. WRONG TIME.`

> “These can all become valuable. The mistake is buying them before the bottleneck exists.”

---

## Видео 6 — Stack за зрител

**Формат:** screenshot на коментар + текстови карти от Kdenlive
**Продължителност:** 30–35 секунди

Коментар:

> `“4 people, $120/month, already using Slack.”`

**Voiceover:**

> “Four people, 120 dollars a month, already using Slack. Here is where I would start.”

Title clips се появяват последователно:

```text
1. Delivery tracking
2. Invoicing
3. Simple client record
```

> “First, delivery tracking. Second, invoicing. Third, a simple client record. Keep Slack because the team already uses it.”

Последна карта:

> `DON’T SPEND THE WHOLE BUDGET JUST BECAUSE IT EXISTS`

CTA:

> `Comment your team size, budget and current tools.`

---

## Видео 7 — Features срещу workflow

**Формат:** две колони + cursor animation
**Продължителност:** 20–25 секунди

**Екран 0–3 сек.**

> `100 FEATURES CAN STILL BE THE WRONG TOOL`

**Voiceover:**

> “A longer feature list does not make a tool a better fit.”

**Екран 3–16 сек.**

Лява колона:

> `Features: 100+`

Дясна колона:

```text
Fits the workflow?
Connects to current tools?
Someone can maintain it?
```

> “Check whether it fits the workflow, connects to the current stack, and can actually be maintained by the team.”

**Екран 16–25 сек.**

> `FIT > FEATURE COUNT`

> “The best tool is the one the business can use successfully.”

---

## Видео 8 — Кога безплатният план става скъп

**Формат:** проста сметка на екрана
**Продължителност:** 25–30 секунди

**Екран 0–3 сек.**

> `WHEN DOES FREE SOFTWARE BECOME EXPENSIVE?`

**Екран 3–18 сек.**

```text
2 hours/week of manual work
× 4 weeks
= 8 hours/month
```

> “A free tool becomes expensive when its missing workflow creates more manual work than the paid plan would cost.”

**Екран 18–26 сек.**

> `Compare subscription cost with repeated manual work.`

> “Compare the subscription price with the repeated work it removes—not with zero.”

**CTA:**

> `Save this rule for your next software decision.`

---

## 5. Варианти без собствен глас

Ако не желаете нито лице, нито собствен глас, използвайте:

### Вариант 1 — Само текст и движение

- 12–18 секунди;
- максимум 25–35 думи общо;
- една нова карта на 2–3 секунди;
- тиха музика, разрешена за commercial use;
- последен кадър с ясно заключение.

Пример:

```text
BEST CRM?

Wrong first question.

Ask instead:
Who maintains it?
What must it connect to?
Which workflow is broken?

Constraints first. Product second.
```

### Вариант 2 — Screen recording с текстови callouts

- показва се реално действие;
- cursor-ът сочи важния елемент;
- текстът обяснява защо е важен;
- не се показва цяла страница, ако текстът е дребен.

### Вариант 3 — Slideshow от собствени графики

- 4–6 графики;
- една мисъл на графика;
- без случайни stock снимки;
- подходящо за comparison и checklist видеа.

### Вариант 4 — Subtitle-first voiceover

Ако не искате да използвате естествения запис на гласа си, запишете бавно и спокойно, след което:

- премахнете паузите;
- намалете фоновия шум;
- добавете големи субтитри;
- оставете гласа естествен, без агресивни voice effects.

Това обикновено е по-добър вариант от изцяло синтетичен глас.

---

## 6. Точно как се прави едно видео от нулата

Следващите инструкции са за човек, който никога не е правил такова видео.
Изпълнявайте ги по ред и не прескачайте към монтажа, преди тестовият запис да е
ясен и да се чува добре.

### 6.1. Какво зрителят трябва да вижда

Вертикалният екран е разделен мислено на три зони:

```text
┌────────────────────────────┐
│  ГОРЕ: hook, 1–2 реда      │  приблизително 20% от височината
│  "STOP BUYING TOOLS..."    │
├────────────────────────────┤
│                            │
│  СРЕДА: сайтът, таблицата  │  приблизително 60% от височината
│  или картите, които сочите │
│                            │
├────────────────────────────┤
│  ДОЛУ: субтитри            │  приблизително 20% от височината
│  Не ги слагайте най-долу.  │
└────────────────────────────┘
```

Важно: бутоните на TikTok, Reels и Shorts закриват десния край и долната част.
Затова не поставяйте важен текст в последните приблизително 120 px отдясно и
250 px отдолу. Дръжте главния текст в централните 80% от кадъра.

Когато показвате сайт:

- увеличете browser zoom на `125%`, `150%` или `175%`, докато думите се четат;
- показвайте само зоната, за която говорите, а не цялата страница;
- движете курсора бавно и го спрете върху точния ред;
- не въртете страницата, докато изговаряте важно изречение;
- всяко кликване трябва да има причина в текста;
- ако няма какво ново да се види 3 секунди, направете crop/zoom или сменете кадъра.

### 6.2. Подготовка на компютъра — прави се преди всеки запис

1. Създайте папка `UNSOLERO_VIDEOS`.
2. В нея създайте папка с дата и тема, например
   `2026-08-25_email-tools`.
3. Вътре създайте:
   - `01_raw` — суровите записи от OBS;
   - `02_audio` — отделният voiceover;
   - `03_assets` — screenshots, logo и графики;
   - `04_project` — Kdenlive проектът;
   - `05_exports` — готовите MP4 файлове.
4. Отворете отделен browser profile само за записи.
5. Затворете email, чат, password manager, affiliate dashboards и admin панели.
6. Скрийте bookmarks bar с `Ctrl+Shift+B` в Chrome/Chromium.
7. Включете `Do Not Disturb`/`Не безпокойте` от системните известия.
8. Излезте от личните акаунти, които могат да покажат име, email или снимка.
9. Отворете само табовете, които ще присъстват в сценария, в правилния ред.
10. Проверете всяка цена и функция от официалната страница в деня на записа.
11. Поставете сценария на телефон или втори монитор, който OBS не записва.
12. Изпийте вода и прочетете текста веднъж на глас без запис.

Никога не показвайте парола, API key, recovery code, адрес, клиентски данни,
банкова информация, реална affiliate статистика или browser autocomplete.

### 6.3. Първоначална настройка на OBS — прави се само веднъж

#### A. Създайте отделен profile

1. Отворете OBS.
2. Натиснете `Profile` в горното меню.
3. Натиснете `New`.
4. Въведете `UNSOLERO Vertical 1080x1920`.
5. Потвърдете с `OK`.
6. Ако се отвори Auto-Configuration Wizard, изберете вариант за запис, а не
   за streaming, и завършете съветника.

#### B. Настройте вертикалното видео

1. Натиснете `Settings` долу вдясно.
2. Отляво натиснете `Video`.
3. На `Base (Canvas) Resolution` напишете `1080x1920`.
4. На `Output (Scaled) Resolution` напишете `1080x1920`.
5. На `Common FPS Values` изберете `30`.
6. Натиснете `Apply`, но още не затваряйте прозореца.

#### C. Настройте файла за запис

1. Отляво натиснете `Output`.
2. На `Output Mode` изберете `Simple`.
3. В секцията `Recording` натиснете `Browse` до `Recording Path`.
4. Изберете папката `01_raw` на текущото видео.
5. На `Recording Quality` изберете `High Quality, Medium File Size`.
6. На `Recording Format` изберете `MKV`.
7. На `Encoder` изберете hardware encoder, ако има такъв; ако няма, оставете
   наличния software encoder.
8. Натиснете `Apply`.

OBS препоръчва MKV, защото прекъснат запис обикновено остава възстановим. MP4
е по-удобен за монтаж и качване, затова конвертирането се прави след записа.

#### D. Настройте гласа

1. Отляво натиснете `Audio`.
2. На `Sample Rate` изберете `48 kHz`.
3. На `Mic/Auxiliary Audio` изберете точния микрофон.
4. На `Desktop Audio` изберете `Disabled`, освен ако сценарият изисква звук
   от сайта. За повечето UNSOLERO видеа не е нужен.
5. Натиснете `Apply`, после `OK`.
6. Кажете нормално: “This is a microphone level test.”
7. В `Audio Mixer` зелената лента трябва да стига приблизително между `-18`
   и `-8 dB`, без да влиза в червеното.
8. Ако е твърде тиха или червена, натиснете зъбното колело до `Mic/Aux`, после
   `Properties` или променете плъзгача и повторете теста.

#### E. Създайте scene и source

1. В панела `Scenes` натиснете `+`.
2. Име: `UNSOLERO Screen + Voice`.
3. В панела `Sources` натиснете `+`.
4. Изберете `Window Capture`.
5. Име: `Browser UNSOLERO`.
6. От `Window` изберете прозореца с browser profile-а за запис.
7. Включете `Capture Cursor`.
8. Натиснете `OK`.
9. В preview прозореца натиснете с десен бутон върху source-а.
10. Изберете `Transform` → `Fit to Screen`.
11. Ако има празни ленти, задръжте `Alt` и влачете ръбовете, за да crop-нете
    ненужните browser рамки. После увеличете и преместете source-а в средната
    зона на вертикалния кадър.
12. Заключете source-а с катинарчето в панела `Sources`, за да не го местите
    случайно.

### 6.4. Тестов запис — задължителен преди истинското видео

1. В browser-а отворете `https://unsolero.com`.
2. В OBS натиснете `Start Recording`.
3. Изчакайте една секунда.
4. Кажете: “Test one. The cursor is pointing at the recommendation button.”
5. Преместете курсора бавно до бутона и го задръжте там две секунди.
6. Scroll-нете надолу веднъж.
7. Натиснете `Stop Recording`.
8. В OBS натиснете `File` → `Show Recordings`.
9. Отворете последния файл и проверете:
   - текстът чете ли се на малък прозорец;
   - курсорът вижда ли се;
   - гласът чист ли е;
   - няма ли известия или лични данни;
   - движението гладко ли е.
10. Ако нещо не е наред, поправете го и направете втори тест. Не разчитайте
    монтажът да спаси нечетим запис.

### 6.5. Точна процедура за истинския запис

1. Върнете първия tab и първата позиция на страницата.
2. В OBS натиснете `Start Recording`.
3. Мълчете една секунда — това дава чисто място за начално изрязване.
4. Изпълнете сценария. След грешка спрете, мълчете две секунди и повторете
   цялото изречение. Не спирайте непременно целия запис.
5. След последната дума мълчете една секунда.
6. Натиснете `Stop Recording`.
7. Натиснете `File` → `Remux Recordings`.
8. Натиснете бутона с три точки, изберете MKV файла и натиснете `Remux`.
9. Получения MP4 поставете в `01_raw` и не изтривайте MKV, преди финалният
   export да е проверен.

Ако четенето и кликването едновременно е трудно, запишете първо само екрана.
После запишете voiceover отделно, докато гледате кадрите. Това е нормален и
често по-чист workflow.

### 6.6. Първоначална настройка на Kdenlive

1. Отворете Kdenlive.
2. Натиснете `File` → `New`.
3. В прозореца за project profile намерете категория `Custom` и vertical
   profile `1080x1920, 30 fps`. Ако го няма, натиснете бутона за управление на
   profiles и създайте:
   - Width: `1080`;
   - Height: `1920`;
   - Frame rate: `30/1`;
   - Scanning: `Progressive`;
   - Pixel aspect ratio: `1/1`;
   - Display aspect ratio: `9/16`.
4. Запазете profile-а с име `UNSOLERO Vertical 30fps`.
5. Задайте поне `3 video tracks` и `2 audio tracks`.
6. Натиснете `OK`.
7. Натиснете `Ctrl+Shift+S` и запазете проекта като
   `04_project/topic_v01.kdenlive`.

Kdenlive определя ориентацията, резолюцията и fps от project profile-а. Не го
сменяйте след започване на монтажа, защото позициите и keyframes могат да се
разместят.

### 6.7. Импорт и подреждане на timeline

Използвайте тази постоянна подредба:

```text
V3  Hook, callouts и CTA
V2  Screen recording / допълнителни screenshots
V1  Фон или втори screen recording
A2  Музика, ако изобщо има
A1  Глас
SUB Субтитри
```

1. В `Project Bin` натиснете `Add Clip or Folder`.
2. Изберете MP4 screen recording-а и отделния audio файл, ако има такъв.
3. Ако Kdenlive пита дали да смени project profile-а по първия clip, изберете
   `Keep current project settings`.
4. Плъзнете screen recording-а на `V2` в позиция `00:00:00:00`.
5. Ако audio е част от същия файл, той ще отиде на `A1`.
6. Ако voiceover е отделен, плъзнете го на `A1` и подравнете първата дума с
   първото действие.
7. Натиснете `S` за Selection tool.
8. Преместете playhead преди първата дума.
9. Натиснете `Shift+R`, за да разрежете активния clip, и изтрийте празното
   начало. Направете същото след последната дума.
10. За всяка грешка поставете playhead преди и след нея, натиснете `Shift+R`
    на двете места и изтрийте средната част.
11. Затворете празнините, така че да няма черни кадри или тишина.
12. Пускайте от 2 секунди преди всеки cut и слушайте дали изречението звучи
    естествено.

### 6.8. Как да увеличите точния бутон или ред

1. Изберете screen clip-а на timeline.
2. В `Effects` потърсете `Position and Zoom` или `Transform`.
3. Плъзнете ефекта върху clip-а.
4. В `Effect Stack` увеличете `Size/Scale`, докато думата или бутонът се чете.
5. Променете `X` и `Y`, за да поставите важния елемент в центъра.
6. За плавно приближаване активирайте keyframe в началото на изречението.
7. След 8–12 frames добавете втори keyframe с по-голям scale.
8. Задръжте zoom-а, докато говорите за елемента.
9. След края добавете два keyframes, за да се върнете към общия кадър.

Не приближавайте и не отдалечавайте непрекъснато. Едно приближаване има смисъл
само когато показва точно това, което изречението назовава.

### 6.9. Hook, callouts и финален CTA

1. В `Project Bin` натиснете стрелката до `Add Clip or Folder`.
2. Изберете `Add Title Clip`.
3. Добавете текстов обект.
4. Поставете hook-а с максимум 8–10 думи, например
   `STOP BUYING SOFTWARE TOO EARLY`.
5. Използвайте дебел sans-serif шрифт, бяло или почти черно според фона.
6. Добавете правоъгълник зад текста, ако фонът пречи на четенето.
7. Дръжте текста в title-safe зоната, а не до самите ръбове.
8. Натиснете `Create Title`.
9. Плъзнете title clip-а на `V3` от `00:00` до приблизително `00:03`.
10. Същата процедура използвайте за 1–3 кратки callout-а и един CTA.

Добър callout е `BUDGET`, `INTEGRATIONS` или `STEP 2`. Лош callout е цял
параграф, който се конкурира със субтитрите.

### 6.9.1. Как точно да направите „карта“ — нищо не се изтегля

Картата е графика, която изглежда приблизително така:

```text
┌────────────────────────────────┐
│  1. SUBSCRIBER COUNT           │
│  How many people are listed?   │
└────────────────────────────────┘
```

Тя се състои само от фон, правоъгълник и текст. Направете първата карта по
следния ред:

1. Отворете Kdenlive и проекта с профил `1080 × 1920`, `30 fps`.
2. Вляво намерете панела `Project Bin`.
3. Натиснете малката стрелка до бутона `Add Clip or Folder`.
4. Изберете `Add Title Clip`.
5. Ще се отвори прозорецът `Title Clip` с празно платно.
6. Ако искате цял тъмен фон, намерете настройката за background и изберете
   тъмносиво или почти черно. Не използвайте снимка от интернет.
7. Изберете инструмента за правоъгълник от лентата на title editor-а.
8. Начертайте широк правоъгълник в средната част на кадъра. Оставете свободно
   място отляво и отдясно.
9. Задайте на правоъгълника светъл цвят, ако фонът е тъмен, или тъмен цвят,
   ако фонът е светъл. Текстът трябва да има ясен контраст.
10. Изберете инструмента за текст `T`.
11. Кликнете върху правоъгълника и напишете например
    `1. SUBSCRIBER COUNT`.
12. Използвайте дебел sans-serif шрифт. Започнете с приблизително `70–90 px` и
    намалете само ако текстът не се събира.
13. Центрирайте текста хоризонтално и вертикално в правоъгълника.
14. Уверете се, че текстът не е близо до десния край или най-долните 250 px,
    защото там интерфейсът на TikTok/Reels/Shorts може да го закрие.
15. Натиснете `Create Title`.
16. Новият title clip ще се появи в `Project Bin`. Преименувайте го ясно,
    например `card-01-subscriber-count`.
17. Плъзнете clip-а от `Project Bin` върху видео пътеката `V3`.
18. Поставете началото му точно на `00:00:06:00`, ако картата трябва да започне
    на шестата секунда.
19. Хванете десния край на clip-а и го разтеглете до `00:00:10:00`. Така
    картата ще остане на екрана четири секунди.
20. Натиснете `Space` и гледайте preview-то. Ако четете текста трудно,
    увеличете го преди да добавяте каквото и да е движение.

За втората карта повторете същите стъпки, но напишете
`2. MONTHLY SEND VOLUME` и я поставете от `00:00:10:00` до `00:00:14:00`.
После направете третата, четвъртата и петата. Не ви трябват Canva, stock
animation, template pack или друг сайт.

#### Най-лесният и препоръчителен вариант: без анимация

Поставете картите една след друга на `V3`, без празно място между тях. Когато
playhead премине от единия clip към следващия, картата ще се смени моментално.
Това е hard cut и е напълно достатъчно за Shorts, Reels и TikTok.

```text
V3: [ CARD 1 ][ CARD 2 ][ CARD 3 ][ CARD 4 ][ CARD 5 ]
     06–10 s   10–14 s   14–18 s   18–22 s   22–26 s
```

#### Ако искате картата да влезе плавно отдолу

Правете това едва след като статичната версия е готова:

1. Кликнете върху конкретния title clip на `V3`.
2. Отворете панела `Effects`.
3. В полето за търсене напишете `Transform`.
4. Плъзнете ефекта `Transform` върху избрания title clip.
5. Поставете playhead на първия frame на clip-а.
6. В `Effect Stack` включете keyframe за позицията.
7. На първия keyframe преместете картата надолу, докато е извън видимия кадър.
   Не е нужно да въвеждате еднакво число във всеки проект; гледайте preview-то
   и спрете, когато картата вече не се вижда.
8. Преместете playhead с около `8–10 frames` напред.
9. Добавете втори keyframe.
10. На втория keyframe върнете картата в централната ѝ позиция.
11. Пуснете clip-а. Картата трябва да влезе веднъж и после да остане неподвижна.
12. Ако движението е прекалено рязко, отдалечете втория keyframe до около
    `12 frames`. Ако е бавно, приближете го до първия.

Не анимирайте всеки отделен текст, икона и правоъгълник. Движете цялата карта
като един title clip. Ако не намирате keyframes или резултатът изглежда лошо,
махнете `Transform` и оставете hard cut — посланието е по-важно от ефекта.

#### Какво записвате с OBS и какво правите в Kdenlive

| Елемент | OBS запис | Kdenlive монтаж |
|---|---:|---:|
| UNSOLERO builder, кликове и scroll | Да | Изрязване и zoom |
| Pricing страница на продукт | Да | Изрязване, zoom и callout |
| Правоъгълник с кратък текст | Не | Създавате го с `Add Title Clip` |
| Hook в началото | Не | Създавате го с `Add Title Clip` |
| Субтитри | Не | Добавяте ги след гласа |
| Плавно влизане на карта | Не | Добавяте `Transform` по желание |

### 6.9.2. Къде стои картата, за какво служи и как се синхронизира с гласа

Всяка карта има само една задача: да превърне едно изговорено изречение в
кратка визуална мисъл. Тя не е украса и не трябва да показва информация, за
която не говорите в същия момент.

Следвайте това правило:

```text
ЕДНА КАРТА = ЕДНА ИДЕЯ = ЕДНО ИЗРЕЧЕНИЕ ОТ VOICEOVER-А
```

Картата се появява точно когато започнете съответното изречение. Тя остава на
екрана до края на изречението и след това се заменя от следващата. Не говорете
за `subscriber count`, докато на екрана още пише `monthly send volume`.

#### Точни безопасни позиции за вертикален проект 1080 × 1920

Координатите започват от горния ляв ъгъл: `X` се движи надясно, а `Y` надолу.
Следните размери оставят място за интерфейса на социалната мрежа:

| Елемент | X | Y | Ширина | Височина | Предназначение |
|---|---:|---:|---:|---:|---|
| Горна hook карта върху screen recording | 90 | 120 | 800 | 260 | Изречението, което спира scroll-а |
| Основна карта при видео само с текст | 100 | 520 | 780 | 460 | Главната идея за текущите 3–4 секунди |
| Малък номер `1`, `2`, `3` | 130 | 560 | 100 | 100 | Показва реда на стъпките |
| Основен текст вътре в картата | 250 | 590 | 570 | 180 | Максимум два кратки реда |
| Допълнителен поясняващ ред | 250 | 770 | 570 | 100 | Само ако е нужен за разбирането |
| Субтитри | 90 | 1320 | 760 | 260 | Точните думи от voiceover-а |
| Малко лого/име `UNSOLERO` | 90 | 1600 | 500 | 80 | Само във финалния CTA |

Не поставяйте важен текст след `X = 900` и след `Y = 1650`. Дясната част се
закрива от бутоните за like/comment/share, а долната — от caption-а и
навигацията на приложението.

Схемата на текстово видео изглежда така:

```text
1080 × 1920
┌──────────────────────────────┐
│  свободно място              │
│                              │
│  ┌──────────────────────┐    │  X 100, Y 520
│  │ 1                    │    │
│  │ SUBSCRIBER COUNT     │    │  основна карта 780 × 460
│  │ How many people?     │    │
│  └──────────────────────┘    │
│                              │
│  точните субтитри тук        │  X 90, Y 1320
│                              │
│  не слагайте важен текст     │  последните 250 px
└──────────────────────────────┘
```

#### Когато под картата има запис на сайт

Не слагайте голямата карта в центъра, защото ще закрие бутона или резултата,
за който говорите. Използвайте следната подредба:

```text
V3  малка hook/callout карта горе: X 90, Y 120, 800 × 260
V2  записът на сайта:             центърът на кадъра
V1  фон, ако записът не запълва целия вертикален кадър
SUB субтитри:                     X 90, Y 1320, 760 × 260
```

Докато казвате “The complete stack has a one-hundred-and-twenty-dollar monthly
ceiling”, на `V2` се вижда budget полето, курсорът е спрян до `$120`, а на `V3`
стои малка карта `TOTAL BUDGET: $120`. Не показвайте в този момент карта за
интеграции или друг следващ въпрос.

#### Напълно разглобен пример: MailerLite/Brevo/Kit

Това видео не изисква OBS. Всичко на екрана се създава като title clips в
Kdenlive. Запишете първо voiceover-а, поставете го на `A1`, а после подравнете
картите по waveform-а на гласа.

| Време | Какво има на timeline | Точно място и вид на екрана | Какво казвате през това време | Какво трябва да се случи |
|---|---|---|---|---|
| 0–3 сек. | `hook-01` на `V3`, глас на `A1` | Тъмен фон. В центъра, `X 100 / Y 520 / 780 × 460`, текст `MAILERLITE vs BREVO vs KIT?` | “Stop asking which email platform is best overall.” | Картата се появява веднага на 0:00 и не се движи. |
| 3–6 сек. | `hook-02` заменя първата карта | Същото място. Голям текст `“BEST” IS THE WRONG FILTER`; червен X върху `BEST` | “The answer changes when your constraints change.” | Hard cut на 3:00. Не оставяйте празен или черен frame. |
| 6–10 сек. | `card-01-subscribers` | Номер `1` вляво; вдясно `SUBSCRIBER COUNT`. Иконата с плик е по желание. | “First: how many subscribers do you have?” | Картата стои през цялото изречение. Нищо друго не се движи. |
| 10–14 сек. | `card-02-volume` | Номер `2`; текст `MONTHLY SEND VOLUME` на същото място | “Second: how many emails do you send each month?” | На думата “Second” правите hard cut към карта 2. |
| 14–18 сек. | `card-03-automation` | Номер `3`; текст `AUTOMATION COMPLEXITY`; малък ред `trigger → email → wait → email` | “Third: how complex must the automation be?” | Схемата се вижда, но не се анимира отделно. |
| 18–22 сек. | `card-04-team` | Номер `4`; текст `TEAM ACCESS`; малък ред `1 USER OR A TEAM?` | “Fourth: who needs access, and with what permissions?” | Не търсите снимки на хора; достатъчни са текстът и числата. |
| 22–26 сек. | `card-05-integrations` | Номер `5`; текст `REQUIRED INTEGRATIONS`; малък ред `CRM → EMAIL → CHECKOUT` | “Fifth: which tools must it connect to?” | Картата се сменя точно на “Fifth”. |
| 26–31 сек. | `summary-01` | Трите имена са малки горе; в центъра `CONSTRAINTS FIRST. PRODUCT SECOND.` | “Only then compare MailerLite, Brevo and Kit against the same list.” | Няма winner, отметка или измислена цена. |
| 31–34 сек. | `cta-01` | В центъра `LIST SIZE + AUTOMATION?`; долу в безопасната зона `UNSOLERO` | “Comment your list size and must-have automation.” | Последната карта стои до края. После видеото приключва директно. |

На `SUB` добавете същите изговорени думи, разделени на кратки фрази от 3–7
думи. Субтитрите не заменят основната карта: картата показва идеята, а
субтитрите показват точните думи.

#### Как да подравните картата към гласа

1. Запишете целия voiceover и го поставете на `A1`.
2. Увеличете timeline-а, докато виждате началото на звуковата вълна за всяко
   изречение.
3. Поставете началото на съответния title clip точно под първата дума.
4. Поставете края му точно след последната дума.
5. Следващият title clip трябва да започне на първата дума на следващото
   изречение.
6. Пуснете видеото без звук. Ако разбирате логиката само от картите, редът е
   ясен.
7. Пуснете го само със звук, без да гледате. Ако изреченията са разбираеми,
   voiceover-ът е ясен.
8. Пуснете го нормално. Ако екранът показва същата идея, която чувате, монтажът
   е синхронизиран.

### 6.10. Субтитри — ръчно или с Whisper

#### Ако автоматичните субтитри са настроени

1. Натиснете иконата `Edit Subtitle Tool` в timeline toolbar-а, за да се появи
   subtitle track.
2. Отворете инструмента `Speech Recognition/Automatic Subtitling`.
3. Изберете английски Whisper model, когато voiceover-ът е на английски.
4. Изберете `Selected clips` или цялата timeline zone.
5. Натиснете `Process`.
6. Прочетете всеки subtitle и поправете имената `UNSOLERO`, `MailerLite`,
   `Brevo`, `Kit`, `ClickUp` и всяко число.

#### Ако автоматиката не работи

1. Поставете playhead на първата дума.
2. Натиснете `Shift+S` или `Project` → `Subtitle` → `Add Subtitle`.
3. Напишете 3–7 думи.
4. Разтеглете subtitle clip-а до края на изговорената фраза.
5. Повторете за следващата фраза.

Правила за четимост:

- максимум два реда;
- 3–7 думи в един subtitle;
- голям шрифт, четим на телефон;
- бял текст с тъмен outline или тъмна полупрозрачна подложка;
- не поставяйте subtitle върху важния бутон;
- не изписвайте всяко “uh” или повторение;
- subtitle трябва да се появи със съответната дума, не секунда по-късно.

### 6.11. Звук

1. Пуснете целия клип и наблюдавайте audio meter-а.
2. Гласът не трябва да влиза в червеното.
3. Ако е неравномерен, добавете лек compressor и limiter; не използвайте
   агресивен “radio voice” effect.
4. Ако има шум, използвайте noise reduction умерено. Прекаляването прави гласа
   метален.
5. Ако има музика, сложете я на `A2` и я намалете приблизително 18–25 dB под
   гласа. При съмнение я махнете.
6. Използвайте само музика с ясно разрешена commercial употреба и пазете
   документа за лиценза.

### 6.12. Export от Kdenlive

1. Натиснете `Ctrl+Enter` или `File` → `Render`.
2. На `Output file` изберете `05_exports/topic_master.mp4`.
3. Изберете preset `MP4-H264/AAC`, който е най-съвместимият общ вариант.
4. Изберете `Full Project`.
5. Не включвайте `Rescale`, ако project profile-ът вече е `1080x1920`.
6. Уверете се, че subtitles се burn-ват във видеото, а не остават само като
   отделен subtitle stream.
7. Натиснете `Render to File`.
8. След края отворете MP4 файла и го гледайте веднъж на цял екран и веднъж в
   малък прозорец с ширина приблизително колкото телефон.
9. Проверете чрез file properties или `ffprobe`, ако го използвате:
   - `1080×1920`;
   - `30 fps`;
   - H.264 video;
   - AAC audio;
   - без черни кадри в началото и края.

### 6.13. Качване — един чист master на трите платформи

Не сваляйте видеото от TikTok, за да го качите в Reels или Shorts. Качвайте
оригиналния `topic_master.mp4`, за да няма watermark и повторна компресия.

#### YouTube Shorts от компютър

1. Влезте в YouTube Studio.
2. Натиснете `Create` → `Upload videos`.
3. Изберете master MP4 файла.
4. Въведете конкретно заглавие, например
   `The 5 constraints to check before choosing an email tool`.
5. Добавете 1–2 изречения описание и точния UNSOLERO URL с UTM параметри.
6. Посочете правилната audience настройка; business software съдържанието не е
   насочено специално към деца.
7. Първия път изберете `Unlisted`, отворете видеото на телефон и проверете
   crop-а. После го направете `Public` или насрочете.

#### TikTok

1. Влезте в TikTok с business профила.
2. Натиснете `Upload`/`+`.
3. Изберете master MP4 файла.
4. Не добавяйте автоматичен crop, ако preview-то вече е 9:16.
5. Напишете кратък caption: една теза + един въпрос.
6. Изберете cover frame, на който hook-ът се чете.
7. Проверете subtitle зоната в preview-то и публикувайте или насрочете.

#### Instagram Reels

1. Влезте в Instagram business профила.
2. Натиснете `Create` → `Reel` или `+` → `Reel`.
3. Изберете master MP4 файла.
4. Изберете cover от кадър с четим hook или качете собствен cover.
5. Напишете caption с една полезна теза, disclosure при нужда и един CTA.
6. Проверете preview-то в profile grid; важният текст не трябва да бъде
   отрязан в квадратната/портретната миниатюра.
7. Публикувайте или насрочете през наличния Meta business инструмент.

Имената на бутоните може леко да се променят. Логиката остава: оригинален
файл → preview → privacy/audience → cover → caption → финална проверка → publish.

### 6.14. Визуален стил, който не се променя между видеата

- един дебел sans-serif шрифт;
- тъмно мастилено, бяло и един зелен/teal акцент от UNSOLERO;
- hook: максимум два реда;
- callout: максимум 1–4 думи;
- еднакъв финален CTA кадър;
- hard cuts или кратки 4–8 frame transitions;
- без 3D ефекти, spinning текст и случайни stock клипове;
- UNSOLERO logo само малко и ненатрапчиво, не като огромен watermark.

### 6.15. Контрол на качеството преди публикуване

Гледайте готовото видео три пъти:

1. **Без звук:** разбира ли се от визуалното и субтитрите?
2. **Само звук, без да гледате:** логично и естествено ли е обяснението?
3. **На телефон:** чете ли се всичко и закриват ли платформените бутони текст?

Ако отговорът на някое е „не“, върнете се в Kdenlive. Не публикувайте с
надеждата, че зрителят ще положи повече усилие от автора.

---

## 7. Кратък график

| Ден | Видео |
|---|---|
| Понеделник | Грешка или мит |
| Вторник | Кратко сравнение |
| Четвъртък | Практическо правило |
| Петък | Viewer stack или отговор на коментар |

Един чист master се публикува в:

- YouTube Shorts;
- TikTok;
- Instagram Reels.

Файлът не трябва да съдържа watermark от друга платформа.

---

## 8. Какво да се измерва

- задържане в първите секунди;
- average watch time;
- completion rate;
- shares и saves;
- смислени коментари;
- profile visits;
- clicks към конкретната страница;
- започнати и завършени UNSOLERO recommendations.

Правила:

- много swipes → сменете hook-а;
- добро гледане, но малко clicks → сменете CTA;
- много clicks, но малко completions → проверете landing page-а;
- много views без business actions → темата вероятно привлича нисконамерена аудитория;
- анализирайте след минимум 10 видеа, не след едно.

---

## 9. Контролен списък

### Преди запис

- [ ] Фактите и цените са проверени.
- [ ] Има една ясна идея.
- [ ] Hook-ът е готов предварително.
- [ ] Има само един CTA.
- [ ] Няма лични данни на екрана.

### Преди публикуване

- [ ] Видеото е 1080×1920.
- [ ] Текстът се чете на телефон.
- [ ] Субтитрите са проверени.
- [ ] Няма watermark.
- [ ] Affiliate disclosure-ът е добавен, когато е необходим.
- [ ] Линкът води към точната страница.
- [ ] Добавени са UTM параметри.

---

## 10. Препоръчителен старт

Започнете с тези четири видеа:

1. `Your small agency may not need a paid CRM.`
2. `ClickUp or Teamwork? One question decides.`
3. `MailerLite, Brevo or Kit: start with these five constraints.`
4. `Can affiliate commission change the UNSOLERO ranking?`

След това използвайте коментарите, за да създадете viewer-stack видеа. Това осигурява постоянен източник на идеи, без да се налага всяка тема да бъде измисляна от нулата.

---

## 11. Официални източници и проверка на настройките

Точните менюта и техническите препоръки в този документ са сверени на
26 август 2026 г. с официалната документация на използваните приложения и
платформи:

- [OBS Studio — Standard Recording Output Guide](https://obsproject.com/kb/standard-recording-output-guide)
- [OBS Studio — Audio/Video Formats Guide](https://obsproject.com/kb/audio-video-formats-guide)
- [OBS Studio — Advanced Recording and Multi-Track Audio](https://obsproject.com/kb/advanced-recording-guide-and-multi-track-audio)
- [OBS Studio — Recording Encoder Presets](https://obsproject.com/kb/recording-encoder-presets-guide)
- [Kdenlive — Project Settings](https://docs.kdenlive.org/en/project_and_asset_management/project_settings.html)
- [Kdenlive — Keyboard Shortcuts](https://docs.kdenlive.org/en/user_interface/shortcuts.html)
- [Kdenlive — Transform, Distort and Perspective](https://docs.kdenlive.org/en/effects_and_filters/video_effects/transform_distort_perspective.html)
- [Kdenlive — Subtitles](https://docs.kdenlive.org/en/effects_and_filters/subtitles.html)
- [Kdenlive — Speech-to-Text](https://docs.kdenlive.org/en/effects_and_filters/speech_to_text.html)
- [Kdenlive — Rendering](https://docs.kdenlive.org/en/exporting/render.html)
- [YouTube Help — Upload YouTube Shorts](https://support.google.com/youtube/answer/12779649)
- [TikTok Help Center — Making a post](https://support.tiktok.com/en/using-tiktok/creating-videos/making-a-post)
- [Instagram Help Center — Music access and Meta Sound Collection for commercial use](https://www.facebook.com/help/instagram/402084904469945)

Менютата могат да бъдат преведени различно според езика на OBS/Kdenlive.
Затова в инструкциите е дадено английското име на всяка команда — то е
най-надеждният ориентир, ако интерфейсът на вашата версия изглежда различно.
