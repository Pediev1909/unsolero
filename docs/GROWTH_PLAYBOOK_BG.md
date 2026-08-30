# UNSOLERO — растеж, съдържание и видео

**Актуален към:** 30 август 2026 г.
**Заменя:** `SOCIAL_MEDIA_GROWTH_STRATEGY_BG.md`, `FACELESS_VIDEO_PLAYBOOK_BG.md`
и `FACELESS_VIDEO_QUICK_GUIDE_BG.md`, които бяха слети тук и изтрити. Три
документа за една тема означава, че два от тях тихо остаряват.

**Език на съдържанието:** английски
**Формат на късите видеа:** 1080×1920, 9:16, 30 fps, 25–40 секунди

---

## Съдържание

1. [Какво се промени и защо](#1-какво-се-промени-и-защо)
2. [Първо действие: AI обхождачите](#2-първо-действие-ai-обхождачите)
3. [Ред на каналите](#3-ред-на-каналите)
4. [Кои страници могат да печелят](#4-кои-страници-могат-да-печелят)
5. [Ритъм и 90-дневен план](#5-ритъм-и-90-дневен-план)
6. [Позициониране и аудитория](#6-позициониране-и-аудитория)
7. [Формати и дължини](#7-формати-и-дължини)
8. [Лице или без лице](#8-лице-или-без-лице)
9. [Теми и серии](#9-теми-и-серии)
10. [Сценарии за къси видеа](#10-сценарии-за-къси-видеа)
11. [Сценарий за дълго YouTube видео](#11-сценарий-за-дълго-youtube-видео)
12. [Reddit](#12-reddit)
13. [Подготовка на компютъра](#13-подготовка-на-компютъра)
14. [OBS — настройка и запис](#14-obs--настройка-и-запис)
15. [Kdenlive — монтаж](#15-kdenlive--монтаж)
16. [Може ли AI да го направи вместо теб](#16-може-ли-ai-да-го-направи-вместо-теб)
17. [Визуален стил](#17-визуален-стил)
18. [Качване по платформи](#18-качване-по-платформи)
19. [Измерване и решения](#19-измерване-и-решения)
20. [Контролни списъци](#20-контролни-списъци)
21. [Източници](#21-източници)

---

## 1. Какво се промени и защо

Предишната стратегия беше писана за конверсии и предполагаше, че affiliate
програмите са достъпни. И двете вече не са верни.

**На 30 август 2026 Thinkific, Webflow и Tidio отказаха в един ден**, и трите с
мотив „малък трафик и аудитория". И трите плащат процент от оборота и трите
вървят през PartnerStack, където дотогава бяхме приети пет от пет — Kit,
monday.com, Pipedrive, Teachable, ActiveCampaign.

Значи правилото „процент → приемат те без трафик" вече не важи. Различаващият
фактор не е моделът на плащане, а датата: програмите, които приеха оператор без
аудитория, го направиха в края на август и вратата се затвори.

**Практическото следствие:** спираш да кандидатстваш. Задачата вече не е
конверсия, а **число, което да покажеш на партньорски мениджър**. Всяко
действие в този документ се мери по това дали го увеличава, не по гледания.

Второ следствие: половината публикувано съдържание сочи към страници, където
нито един продукт няма действаща програма. Виж раздел 4.

---

## 2. Първо действие: AI обхождачите

`https://unsolero.com/robots.txt` съдържа блок, озаглавен
`BEGIN Cloudflare Managed content`, който **не идва от този repository** —
приложението сервира само долната част. Cloudflare инжектира горната:

```
User-agent: *
Content-Signal: search=yes,ai-train=no,use=reference

User-agent: GPTBot             Disallow: /
User-agent: ClaudeBot          Disallow: /
User-agent: Google-Extended    Disallow: /
User-agent: CCBot              Disallow: /
User-agent: Bytespider         Disallow: /
User-agent: Amazonbot          Disallow: /
User-agent: Applebot-Extended  Disallow: /
User-agent: meta-externalagent Disallow: /
```

Сайтът публикува `llms.txt`, който обяснява защо данните му са цитируеми, и
едновременно казва на обхождачите да не ги четат.

Точността има значение: това са **тренировъчните и grounding** обхождачи.
`OAI-SearchBot`, `ChatGPT-User`, `Claude-User` и `PerplexityBot` не са в списъка
и минават. `Google-Extended` не влияе на обикновеното търсене — само на AI
Overviews. Не е пълен мрак, но е изключен каналът, който расте най-бързо.

**Това е решение за права, не бъг.** „Позволявам ли модели да се обучават върху
текста ми, срещу това да ме цитират" е легитимен въпрос с два защитими
отговора. Проблемът е, че сега отговорът идва от подразбирането на Cloudflare, а
не от собственика. Реши го съзнателно: Cloudflare → AI Crawl Control.

---

## 3. Ред на каналите

| # | Канал | Защо | Кога плаща |
|---|---|---|---|
| 1 | **Reddit** | №1 цитиран източник във всички answer engines; до 1 от 5 цитирания в Perplexity | седмици |
| 2 | **Дълго YouTube** | търсене, а не алгоритмичен фийд; не иска domain authority | 1–3 месеца |
| 3 | **Къси клипове** | нарязани от дългото, не отделна продукция | месеци |
| 4 | **SEO** | 69% от affiliate трафика, но нов домейн иска 4–6 месеца за първо движение и 6–12 за конкурентни думи | 6–12 месеца |

Reddit е пръв **не защото носи клиенти**. Носи цитирания, а те са това, което
кара ChatGPT и Perplexity да те споменат — и единственото, което расте
достатъчно бързо, за да има число до края на годината.

Късите клипове падат на трето по конкретна причина: наличните данни за B2B SaaS
в TikTok са за **платени** кампании — CPL, retargeting, ROAS. При нулев бюджет
не важат.

SEO остава прав, но бавен, и AI Overviews свиват click-through дори когато
класираш.

**LinkedIn** е четвърти по приоритет, но си струва: изкачи се от ~№11 до №5 сред
цитираните източници в ChatGPT между ноември 2025 и февруари 2026. Публикувай
там дългото видео като текстов анализ, не като линк.

---

## 4. Кои страници могат да печелят

Измерено на 30 август 2026 срещу продукцията. Видео, насочено към страница,
където нищо не печели, не може да направи пари независимо от гледанията.

| Печели | Страница |
|---|---|
| **4/5** | `/guides/mailchimp-alternatives` |
| 2/3 | `/guides/calendly-alternatives` |
| 2/3 | `/compare/zoho-crm-vs-hubspot` |
| 2/3 | `/compare/zoho-invoice-vs-wave` |
| 1/3 | `/compare/teachable-vs-thinkific-vs-gumroad` |
| 1/3 | `/compare/ahrefs-vs-semrush` |
| **0/3** | slack-alternatives · google-analytics-alternatives · zapier-alternatives · canva-vs-figma · webflow-vs-squarespace-vs-framer · freshdesk-vs-help-scout-vs-tidio |

**Първото видео е сценарий B** (раздел 10) — води към страницата с 4/5.

Не прави видео за Canva срещу Figma. И трите марки нямат действаща програма.

**Провери таблицата отново, преди да планираш месец съдържание.** Тя се мени с
всяка приета или отказана програма, а `/api/catalog/products/{slug}/offers`
връща `purchase_path` само когато офертата има жив affiliate линк.

---

## 5. Ритъм и 90-дневен план

### Седмичен ритъм

- **1 дълго YouTube видео** — по сценария в раздел 11
- **2 клипа, нарязани от него** — не отделна продукция
- **3 отговора в Reddit** — в теми, където някой вече пита точно това

Пет оригинални видеа седмично от един човек без бюджет е планът, който умира във
втората седмица. При нула трафик второто спиране струва повече от бавното
начало.

### Дни 1–14 — настройка и първи опити

- реши въпроса с AI обхождачите (раздел 2)
- настрой OBS профила и Kdenlive template-а
- намери и запази 20 подходящи Reddit теми
- запиши първото дълго видео и извади два клипа от него
- настрой bio линковете и UTM параметрите
- направи таблица за резултатите

### Дни 15–45 — последователност

- поддържай ритъма без изключения
- започни постоянните серии (раздел 9)
- отговаряй на всеки смислен коментар
- проверявай funnel-а на телефон
- след 10 къси клипа направи анализ, не след едно

### Дни 46–90 — задълбочаване

- видеа по зрителски казуси от коментарите
- content cluster около темата с най-добри резултати
- обнови сайта със спечелилите теми
- **след като има число**: повторно кандидатстване към Thinkific, Webflow и
  Tidio; отказът не е окончателен, а изявление за текущия обхват

### Реалистична цел за 90 дни

- 12–13 дълги видеа
- 24–26 клипа
- около 120 отговора в Reddit
- поне 3 разпознаваеми серии
- измерим път от съдържание до `affiliate click`
- **число за показване**, което е истинската цел

---

## 6. Позициониране и аудитория

UNSOLERO не трябва да звучи като поредния сайт с класации „10 best tools".
Разликата е:

- избор според конкретния бизнес;
- бюджетът е ограничение, а не цел за изразходване;
- вземат се предвид вече използваните инструменти;
- показват се причините за препоръката;
- показват се **отхвърлените** продукти и защо;
- affiliate комисионата не участва в оценяването.

Основната идея:

> **Stop overbuying software. Build the right stack for your business, budget,
> and existing tools.**

**Аудитория:** собственици на малък бизнес, агенции с 2–10 души, самостоятелни
консултанти, creators, малки SaaS екипи, хора, които плащат за дублиращи се
инструменти.

**Език:** английски. Публичното съдържание на продукта е на английски, SaaS
програмите са международни, а YouTube Search има по-голям дългосрочен
потенциал. Не смесвай английски и български в един профил — български може да
се изгради отделно след валидиране на формата.

### Роля на всяка платформа

| Платформа | Роля | Основен CTA |
|---|---|---|
| Reddit | Цитирания и реални въпроси | Отговор, без линк в първото изречение |
| YouTube (дълго) | Търсене, доверие, библиотека | Виж сравнението |
| YouTube Shorts | Откриване на канала | Гледай свързаното дълго видео |
| TikTok | Тестване на теми | Коментирай своя казус |
| Instagram Reels | Доверие и споделяне | Запази или сподели |
| LinkedIn | B2B разпространение | Прочети анализа |

---

## 7. Формати и дължини

| Формат | Дължина | Цел |
|---|---:|---|
| Кратко решение | 20–40 сек. | Достигане до нови хора |
| Мини сравнение | 35–60 сек. | Покупателно намерение |
| Отговор на коментар | 20–45 сек. | Общност и идеи |
| Пълно сравнение | 7–12 мин. | Търсене и доверие |
| Изграждане на stack | 8–15 мин. | Демонстрация на продукта |
| Методология | 4–8 мин. | Доказване на независимост |

**Основно правило:** видеото трябва да бъде точно толкова дълго, колкото е нужно
за изпълняване на обещанието му. Не удължавай за watch time, не започвай с лого,
не обяснявай минута какво предстои и не повтаряй заключението.

### Структура на кратко видео

**С глас, 25–40 сек.:** hook (0–2) → проблем (2–8) → решение (8–25) → правило
(25–34) → един CTA (34–40).

**Без глас, 12–20 сек.:** въпрос (0–2) → 3–4 визуални стъпки (2–12) →
заключение (12–17) → CTA (17–20). Текстът трябва да се чете спокойно, без
зрителят да спира видеото.

---

## 8. Лице или без лице

И двете работят. Изборът е твой и не е нужно да е окончателен.

**С лице** — за нов профил най-силната структура е:

> лице (0–3 сек., проблемът) → екран (доказателството) → лице (3–5 сек.,
> заключение и CTA)

Лицето не стои през цялото видео. Реалният основател е актив, когато брандът
продава преценка и прозрачност. **AI avatar не се препоръчва като основен
говорител** — точно този бранд губи най-много от него.

**Без лице** — основният формат е:

> собствен глас + запис на екрана + големи субтитри + прости графики

Липсата на лице не е проблем, ако има конкретен проблем в първите 2 секунди,
реален екран или диаграма, ясен човешки глас, визуална промяна на всеки 2–4
секунди, една идея и един CTA. **AI глас върху случайни stock клипове не се
препоръчва** — създава по-слабо доверие и не показва продукта.

### Шест faceless формата

| Формат | Какво е | Подходящ за |
|---|---|---|
| **A. Screen recording + voiceover** | реалният сайт, продукт или builder | сравнения, демонстрации |
| **B. Анимирани карти** | 2–4 карти се подреждат: `Delivery → Invoicing → CRM` | какво да се купи първо, грешки |
| **C. Split-screen** | екранът на две, максимум три критерия | кратки сравнения |
| **D. Kinetic text** | големи думи върху чист фон | митове, правила, видеа без глас |
| **E. Диаграма** | `constraints + facts → recommendation → offers` | обяснение на метода |
| **F. Отговор на коментар** | screenshot на коментара + карти | най-добрият резервен формат |

### Какво се вижда на екрана

Един основен визуален обект наведнъж: два продукта един до друг, три стъпки в
ред, конкретен екран от UNSOLERO, таблица с максимум 3–4 реда, или отхвърлените
продукти с причината.

Вертикалният кадър се дели мислено на три:

```text
┌────────────────────────────┐
│  ГОРЕ: hook, 1–2 реда      │  ~20% от височината
├────────────────────────────┤
│  СРЕДА: сайтът, таблицата  │  ~60% от височината
│  или картите               │
├────────────────────────────┤
│  ДОЛУ: субтитри            │  ~20% — не най-долу
└────────────────────────────┘
```

Бутоните на TikTok, Reels и Shorts закриват десния край и долната част. Не
поставяй важен текст в последните ~120 px отдясно и ~250 px отдолу. Дръж
главния текст в централните 80%.

Ориентировъчно: hook 3–7 думи, текст 58–72 px, субтитри до два реда.

Когато показваш сайт: browser zoom на `125–175%`, показвай само зоната, за
която говориш, движи курсора бавно и го спирай на точния ред, не въртѝ
страницата по време на важно изречение. Ако няма нищо ново 3 секунди — crop,
zoom или смяна на кадъра.

---

## 9. Теми и серии

### Разпределение

- **40%** сравнения с високо покупателно намерение
- **25%** изграждане на цял software stack
- **15%** грешки и ненужни абонаменти
- **10%** методология и affiliate независимост
- **10%** behind the scenes и отговори на аудиторията

### Постоянни серии

`One Question Decides` · `Build the Stack` · `Bought Too Early` ·
`What We Rejected` · `Commission Cannot Rank This` · `Your Business, Your Stack`

### Теми, подредени по това дали могат да печелят

**Първо (страници с оферти):**

1. MailerLite, Brevo or Kit: the five constraints that decide.
2. Calendly alternatives when scheduling is not the bottleneck.
3. Zoho CRM vs HubSpot for a four-person team.
4. Zoho Invoice vs Wave: when free stops being free.
5. Teachable vs Thinkific for a first paid course.

**После (носят доверие, не пари пряко):**

6. Why a four-person agency may not need a paid CRM yet.
7. Build a software stack under $100/month.
8. Three subscriptions small businesses buy too early.
9. What to buy first: invoicing, CRM or project management?
10. Why "best CRM" is usually the wrong question.
11. How affiliate commission is separated from UNSOLERO rankings.
12. When free software becomes more expensive than paid software.
13. Why integrations matter more than feature count.
14. Best stack for a solo consultant.
15. Give me your team size, budget and existing tools — I'll build your stack.

---

## 10. Сценарии за къси видеа

Колоната `Действие` казва какво правиш с мишката. `Екран` казва какво остава
видимо. `Глас` е точният английски текст за четене. **Не изговаряй текста от
колоната `Екран`**, освен ако същото изречение го няма и в `Глас`.

### Сценарий A — демонстрация на builder-а

**Цел:** зрителят да види продукта в действие, не само да чуе обещание.
**Дължина:** 34–40 сек.
**Подготовка:** отвори homepage-а, изключи consent banner-а предварително и се
увери, че builder-ът връща резултат.

| Време | Действие с мишката | Какво се вижда | Точен глас | Текст отгоре |
|---|---|---|---|---|
| 0–2 | Не движи курсора. | Homepage, crop около headline-а и `Build My Setup`. | "Most software lists start with products. This one starts with your constraints." | `STOP STARTING WITH PRODUCTS` |
| 2–5 | Курсор върху `Build My Setup`, задръж половин секунда, кликни. | Бутонът, после Question 1. | "I'll build a stack for a small client-services business." | `REAL EXAMPLE` |
| 5–8 | Кликни `Run a client services business`, после `Next`. | Избраната карта ясно сменя състояние. | "First: what the business actually does." | `1. GOAL` |
| 8–11 | Кликни `No dedicated admin`, после `Next`. | Изборът за поддръжката. | "Nobody is paid to maintain complicated software." | `2. TEAM CAPACITY` |
| 11–15 | Кликни полето за бюджет, `Ctrl+A`, въведи `120`, после `Next`. | Полето показва `$120`. | "The complete stack has a one-hundred-and-twenty-dollar monthly ceiling." | `3. TOTAL BUDGET: $120` |
| 15–19 | Кликни `Team chat`, после `Next`. | Само `Team chat` е избрано. | "Team chat already exists, so recommending another one would be waste." | `4. KEEP WHAT WORKS` |
| 19–23 | Кликни `The strongest tool per job`, после `Next`. | Избраната preference карта. | "This team prefers the strongest tool for each job." | `5. PREFERENCE` |
| 23–27 | Кликни `Best value` и `Connects to my stack`, после `Build my setup`. | Двете priorities избрани; бутонът се натиска. | "Value and integrations matter most. Now the engine can rank the fit." | `6. PRIORITIES` |
| 27–34 | Изчакай резултата. Scroll бавно до сумата и първите карти. | Реалният резултат, score, total. | "The result stays inside the budget, avoids what the team already owns, and explains every choice." | `CONSTRAINTS → EXPLAINED STACK` |
| 34–39 | Спри движението. Курсор върху score или първата reason карта. | Един стабилен close-up. | "Build your own brief at unsolero dot com." | `TRY YOUR OWN BRIEF` |

Ако резултатът покаже error, **не записвай и не монтирай този take.** Повтори
същите inputs без OBS. Едва след успешен резултат прави истинския запис.

### Сценарий B — MailerLite, Brevo или Kit

**Първото видео за снимане.** Води към `/guides/mailchimp-alternatives`, която е
единствената страница с 4 от 5 печелещи продукта.

**Цел:** метод, а не недоказано „winner" твърдение.
**Дължина:** 30–34 сек.
**Подготовка:** пет title cards в Kdenlive. Ако показваш vendor страници, отвори
официалните pricing pages предварително и **не показвай цена, непроверена
същия ден**.

| Време | Действие/монтаж | Какво се вижда | Точен глас | Текст на картата |
|---|---|---|---|---|
| 0–3 | Static title clip, без browser. | Имената, еднакъв размер. | "Stop asking which email platform is best overall." | `MAILERLITE vs BREVO vs KIT?` |
| 3–6 | Имената се свиват; червен X върху `BEST`. | Чист фон, без stock footage. | "The answer changes when your constraints change." | `"BEST" IS THE WRONG FILTER` |
| 6–10 | Появява се card 1. | Голямо число, малка envelope икона. | "First: how many subscribers do you have?" | `1. SUBSCRIBER COUNT` |
| 10–14 | Hard cut към card 2. | Calendar/envelope графика. | "Second: how many emails do you send each month?" | `2. MONTHLY SEND VOLUME` |
| 14–18 | Hard cut към card 3. | Схема `trigger → email → wait → email`. | "Third: how complex must the automation be?" | `3. AUTOMATION COMPLEXITY` |
| 18–22 | Hard cut към card 4. | 1 срещу 4 person icons. | "Fourth: who needs access, and with what permissions?" | `4. TEAM ACCESS` |
| 22–26 | Hard cut към card 5. | Три boxes със свързващи линии. | "Fifth: which tools must it connect to?" | `5. REQUIRED INTEGRATIONS` |
| 26–31 | Трите имена под петте constraints. | **Няма winner badge.** | "Only then compare MailerLite, Brevo and Kit against the same list." | `CONSTRAINTS FIRST. PRODUCT SECOND.` |
| 31–34 | Финален CTA title clip. | UNSOLERO малко долу. | "Comment your list size and must-have automation." | `LIST SIZE + AUTOMATION?` |

Не слагай зелени отметки до vendor, ако не си проверил конкретния план. Изглежда
като факт дори когато е декорация.

### Сценарий C — комисионата и ranking-ът

**Цел:** доверие и affiliate прозрачност.
**Дължина:** 28–34 сек.
**Подготовка:** отвори `/affiliate-disclosure` и `/how-it-works` в два таба.

| Време | Действие с мишката | Какво се вижда | Точен глас | Текст отгоре |
|---|---|---|---|---|
| 0–3 | Diagram title, не browser. | Две boxes: `Ranking` и `Commission data`. | "Can affiliate commission change a software ranking?" | `CAN COMMISSION CHANGE THE RANKING?` |
| 3–8 | Червена прекъсната линия между boxes и X върху нея. | Ясно разделяне. | "On UNSOLERO, commercial data is not an input to recommendation scoring." | `NO COMMERCIAL INPUT` |
| 8–14 | Flow `Goal + budget + current stack → ranking`. | Boxes се появяват една по една. | "The engine reads the goal, budget, current stack, preferences, and validated product facts." | `CONSTRAINTS + VERIFIED FACTS` |
| 14–20 | Резултат box, после отделно `Merchant offers`. | Offers **след** резултата. | "It produces the recommendation first. Merchant offers are attached afterwards." | `RECOMMENDATION → THEN OFFERS` |
| 20–27 | Switch към disclosure таба; zoom върху абзаца за комисионата. Курсорът неподвижен. | Само релевантният абзац. | "That separation is written into the product architecture and tested as a business invariant." | `PROMISE + TESTED SEPARATION` |
| 27–33 | Switch към how-it-works. | Methodology страницата. | "Do not just trust a ranking. Check what its code is allowed to read." | `ASK HOW IT WAS PRODUCED` |

Disclosure се добавя и в caption-а, когато видеото съдържа affiliate линк. Не го
скривай само в profile bio.

### Сценарий D — резервен, без лице и без глас

**Цел:** видео при шумна среда или без желание за voiceover.
**Дължина:** 16–19 сек. **Звук:** тиха лицензирана музика или никакъв.

| Време | Какво правиш в Kdenlive | Точен текст на екрана |
|---|---|---|
| 0–2 | Title clip, тъмен фон, лек zoom 100% → 105%. | `BEST CRM?` |
| 2–4 | Hard cut; думите се свиват, появява се X. | `WRONG FIRST QUESTION.` |
| 4–7 | Card 1 влиза. | `Which workflow is broken?` |
| 7–10 | Card 2 заменя card 1. | `Who will maintain it?` |
| 10–13 | Card 3 заменя card 2. | `What must it connect to?` |
| 13–16 | Трите се подреждат вертикално. | `CONSTRAINTS FIRST.` |
| 16–19 | Финален чист кадър. | `PRODUCT SECOND.  •  UNSOLERO` |

Текстът стои достатъчно за едно спокойно прочитане. Не добавяй пет икони,
particles и преход едновременно — движението води окото, не доказва, че
програмата има ефекти.

### Кратки сценарии за клипове от дългото видео

**Не купувай CRM първо** (30–35 сек., карти + voiceover)

> 0–3: `YOUR SMALL AGENCY MAY NOT NEED A PAID CRM` —
> "A four-person agency often buys its CRM too early."
> 3–10: `Late projects + unpaid invoices ≠ CRM problem` —
> "If projects are late or invoices are not going out, a CRM will not solve the real bottleneck."
> 10–24: карти `1. Delivery / 2. Invoicing / 3. Client records` —
> "Start with delivery tracking, then invoicing, and keep client records simple until manual work or seat limits actually hurt."
> 24–32: `BUY WHEN A WORKFLOW BREAKS — NOT WHEN A LIST TELLS YOU TO`

**Три абонамента, купени прекалено рано** (25–30 сек., kinetic text)

> 0–3: `3 SUBSCRIPTIONS SMALL BUSINESSES BUY TOO EARLY`
> 3–9: `1. Paid CRM before simple records become limiting`
> 9–15: `2. Advanced analytics before meaningful traffic exists`
> 15–21: `3. Help desk while the shared inbox still works`
> 21–30: `GOOD TOOLS. WRONG TIME.` —
> "These can all become valuable. The mistake is buying them before the bottleneck exists."

**Кога безплатният план става скъп** (25–30 сек., сметка на екрана)

> 0–3: `WHEN DOES FREE SOFTWARE BECOME EXPENSIVE?`
> 3–18: `2 hours/week × 4 weeks = 8 hours/month` —
> "A free tool becomes expensive when its missing workflow creates more manual work than the paid plan would cost."
> 18–26: `Compare subscription cost with repeated manual work.` —
> "Compare the subscription price with the repeated work it removes — not with zero."

**Stack за зрител** (30–35 сек., screenshot на коментара + карти)

> Коментар: `"4 people, $120/month, already using Slack."`
> Карти `1. Delivery tracking / 2. Invoicing / 3. Simple client record` —
> "First, delivery tracking. Second, invoicing. Third, a simple client record. Keep Slack because the team already uses it."
> Последна карта: `DON'T SPEND THE WHOLE BUDGET JUST BECAUSE IT EXISTS`

---

## 11. Сценарий за дълго YouTube видео

**Заглавие:** *I Built a Complete Software Stack for a 4-Person Agency Under
$120/Month*

**Thumbnail:** `THE $120 AGENCY STACK` — лицето, `$120` като най-голям елемент,
максимум три продукта, чист фон без дребен текст.

**0:00–0:20 — Cold open**

> "Most software-stack videos begin with products. This one begins with the
> business. We have four people, a monthly budget of 120 dollars, no dedicated
> software administrator, and an existing team-chat tool. By the end, we will
> have a complete stack, a purchase order, and a list of tools we deliberately
> rejected."

**0:20–0:45 — Обещание**

> "I'm not going to spend the whole budget just because it is available. The
> goal is to cover the jobs that stop delivery or payment first, while avoiding
> duplicate tools."

**0:45–1:30 — Brief.** Покажи: Team 4 · Client services · $120/month · No admin ·
Existing: team chat.

> "If your business does not match this brief, your answer should be different.
> That is the point. A software recommendation without constraints is just a
> preference."

**1:30–3:00 — Jobs before products**

> "Before naming a product, we need to cover three jobs: deliver the work,
> invoice for it, and maintain a shared client record. Scheduling, advanced
> analytics and a help desk can wait because none of them currently stops
> delivery or payment."

**3:00–5:30 — Сравнение.** Максимум три проверени кандидата на категория.

> "The deciding factor is not the longest feature list. It is whether the team
> bills for time, needs client access, and has someone available to configure
> the system."

**5:30–7:00 — Отхвърлени продукти**

> "Now the part most comparison sites hide: what we rejected. We rejected a paid
> help desk because the support volume does not justify it. We rejected advanced
> analytics because there is not enough traffic. We rejected a second
> communication tool because the team already has one."

**7:00–8:00 — Резултат и CTA**

> "This stack covers delivery, invoicing and client records without treating the
> budget as a spending target. If your team, budget or existing tools are
> different, use the UNSOLERO builder linked below. It produces a different
> answer from different inputs — and shows why."

End screen води към следващото релевантно сравнение, не към началната страница
на канала.

---

## 12. Reddit

Каналът с най-бърза възвръщаемост и този, който предишната стратегия изобщо не
разглеждаше.

**Как се прави:**

1. Намери теми, в които някой **вече пита** точно това, което сайтът отговаря.
   Подходящи места: `r/smallbusiness`, `r/agency`, `r/Entrepreneur`,
   `r/marketing`, `r/nocode`, `r/SaaS`.
2. Отговори с реалния отговор, изцяло в коментара. Човекът трябва да получи
   стойност дори ако не кликне.
3. Линк само когато добавя нещо, което коментарът не може да събере — цялата
   таблица, датата на четене на цената, методологията. Не в първото изречение.
4. Кажи какъв ти е интересът, когато линкваш. Reddit наказва скритата реклама
   по-сурово от всяка друга платформа, а тук честността е самият продукт.

**Защо работи за този сайт:** отговорите тук са структурирани факти с записан
източник и дата. Това е формата, който answer engines цитират — и Reddit е
най-цитираният домейн сред тях.

**Три седмично.** Не повече: акаунт, който публикува само линкове, се маркира
за спам и загубата е необратима.

---

## 13. Подготовка на компютъра

Прави се преди **всеки** запис.

1. Създай папка `UNSOLERO_VIDEOS`, в нея папка с дата и тема —
   `2026-08-30_email-tools`.
2. Вътре създай:

```text
01_raw/         суровите записи от OBS
02_audio/       отделният voiceover
03_assets/      screenshots, лого, графики
04_project/     Kdenlive проектът
05_exports/     готовите MP4
06_thumbnails/  корици
```

3. Отвори **отделен browser profile** само за записи.
4. Затвори email, чат, password manager, affiliate dashboards и admin панели.
5. Скрий bookmarks bar с `Ctrl+Shift+B`.
6. Включи `Do Not Disturb`.
7. Излез от личните акаунти, които могат да покажат име, email или снимка.
8. Отвори само табовете от сценария, в правилния ред.
9. **Провери всяка цена и функция от официалната страница в деня на записа.**
10. Постави сценария на телефон или втори монитор, който OBS не записва.
11. Прочети текста веднъж на глас без запис.

Никога не показвай парола, API key, recovery code, адрес, клиентски данни,
банкова информация, реална affiliate статистика или browser autocomplete.

---

## 14. OBS — настройка и запис

### Профили (правят се веднъж)

`YT_LONG_1080P` — Base и Output `1920×1080`, 30 fps.
`SHORT_VERTICAL` — Base и Output `1080×1920`, 30 fps.

`Settings → Video` за всеки. Ако се отвори Auto-Configuration Wizard, избери
вариант за **запис**, не за streaming.

### Запис

`Settings → Output → Advanced`:

- Recording format: `MKV` — прекъснат запис обикновено остава възстановим
- Encoder: наличният hardware encoder
- Rate control: `CQP/ICQ`, Quality около `20`
- Keyframe interval: `2`
- Audio sample rate: `48 kHz`

След записа: `File → Remux Recordings` за MP4. Не трий MKV преди финалната
проверка на export-а.

### Звук

`Settings → Audio`: Sample Rate `48 kHz`, Mic на точния микрофон, Desktop Audio
`Disabled`, освен ако сценарият изисква звук от сайта.

Аудио tracks: `1` общ микс, `2` само микрофон, `3` desktop звук.

Филтри на микрофона, **в този ред**:

1. Noise Suppression: RNNoise
2. Expander за лек фонов шум
3. Compressor — ratio `3:1`, threshold около `-18 dB`
4. Limiter — последен, между `-3` и `-6 dB`

Кажи "This is a microphone level test." Зелената лента трябва да стига между
`-18` и `-8 dB`, без червено.

### Сцени

`FACE_FULL` · `SCREEN_FACE` · `SCREEN_FULL` · `COMPARISON` · `CTA_END`

За `SCREEN_FACE`: Window Capture за браузъра + Video Capture Device за камерата
+ кратко текстово поле + малък брандинг елемент.

**Window Capture е за предпочитане пред запис на целия desktop.**

Създаване на source:

1. `Sources` → `+` → `Window Capture`, име `Browser UNSOLERO`
2. Избери прозореца с browser profile-а за записи
3. Включи `Capture Cursor`
4. Десен бутон в preview → `Transform` → `Fit to Screen`
5. При празни ленти: задръж `Alt` и влачи ръбовете, за да crop-неш browser
   рамките; после увеличи и премести source-а в средната зона
6. **Заключи source-а** с катинарчето, за да не го местиш случайно

### Тестов запис — задължителен

1. Отвори `https://unsolero.com`, натисни `Start Recording`, изчакай секунда.
2. Кажи: "Test one. The cursor is pointing at the recommendation button."
3. Премести курсора бавно до бутона, задръж две секунди, scroll веднъж, спри.
4. `File` → `Show Recordings`, отвори последния файл и провери: чете ли се
   текстът на малък прозорец, вижда ли се курсорът, чист ли е гласът, няма ли
   известия или лични данни, гладко ли е движението.

**Монтажът не спасява нечетим запис.**

### Истинският запис

1. Върни първия таб и първата позиция.
2. `Start Recording`, мълчи една секунда — това дава чисто място за изрязване.
3. Изпълни сценария. След грешка спри, мълчи две секунди и повтори **цялото
   изречение**. Не спирай целия запис.
4. След последната дума мълчи една секунда, после `Stop Recording`.

Ако четенето и кликването едновременно е трудно, запиши първо само екрана, после
voiceover отделно, докато гледаш кадрите. Това е нормален и често по-чист
workflow.

---

## 15. Kdenlive — монтаж

### Профил на проекта

`File → New`. Vertical: `1080×1920`, `30/1` fps, Progressive, pixel aspect `1/1`,
display aspect `9/16`. Запази като `UNSOLERO Vertical 30fps`. Поне 3 video и 2
audio tracks.

За дълго видео: `1920×1080`, progressive, 30 fps.

**Не сменяй профила след започване на монтажа** — позициите и keyframes се
разместват.

### Подредба на timeline

```text
V3   Hook, callouts и CTA
V2   Screen recording / screenshots
V1   Фон или втори screen recording
A2   Музика, ако изобщо има
A1   Глас
SUB  Субтитри
```

Ако Kdenlive пита дали да смени profile-а по първия clip — `Keep current
project settings`.

### Ред на работа

1. Импорт на записите; proxy clips, ако preview-ът насича.
2. `S` за Selection tool. Playhead преди първата дума, `Shift+R` за разрязване,
   изтрий празното начало. Същото след последната дума.
3. За всяка грешка: `Shift+R` преди и след нея, изтрий средата, затвори
   празнината.
4. Пускай от 2 секунди преди всеки cut и слушай дали изречението звучи
   естествено.
5. Screen recording и B-roll.
6. Zoom, текст, графики.
7. Звук.
8. Субтитри и **ръчна** проверка.
9. Export.
10. Пълно гледане на компютър **и** телефон.

### Темпо

**Къси:** махай почти всички паузи, визуална промяна на 2–4 секунди, основно
чисти cuts, zoom само за насочване на вниманието, без преход при всяко
изрязване.

**Дълги:** не изрязвай всяко вдишване, запази спокоен ритъм, B-roll само когато
доказва казаното, не сменяй екрана заради движение.

### Zoom върху точен елемент

1. Избери screen clip-а, добави `Position and Zoom` или `Transform`.
2. Увеличи `Size/Scale`, докато думата или бутонът се чете.
3. `X` и `Y` — важният елемент в центъра.
4. Keyframe в началото на изречението; след 8–12 frames втори с по-голям scale.
5. Задръж zoom-а, докато говориш за елемента; после два keyframes обратно.

Едно приближаване има смисъл само когато показва точно това, което изречението
назовава.

### Hook, callouts и CTA

`Add Clip or Folder` → стрелка → `Add Title Clip`. Hook максимум 8–10 думи,
дебел sans-serif, правоъгълник зад текста при нужда, в title-safe зоната.
Плъзни на `V3` от `00:00` до ~`00:03`.

Добър callout: `BUDGET`, `INTEGRATIONS`, `STEP 2`. Лош: цял параграф, който се
конкурира със субтитрите.

### Субтитри

`Settings → Configure Kdenlive → Speech to Text` → Whisper. После иконата
`Edit Subtitle Tool`, `Automatic Subtitling`, правилният език, `Process`.

**Провери ръчно** `UNSOLERO`, `MailerLite`, `Brevo`, `Kit`, `ClickUp` и всяко
число.

Ръчно: playhead на първата дума, `Shift+S`, напиши 3–7 думи, разтегли до края на
фразата.

Правила: максимум два реда, 3–7 думи на subtitle, голям шрифт, бял текст с тъмен
outline или подложка, не върху важния бутон, без „uh" и повторения, появява се
**със** думата, не секунда по-късно.

### Звук

`Normalize (2 Pass)` върху финалния voice track, цел около `-14 LUFS`. Стартирай
анализа, приложи, провери за clipping.

Ако има музика — на `A2`, 18–25 dB под гласа. При съмнение я махни. Само музика
с ясно разрешена commercial употреба; пази документа за лиценза.

Не използвай агресивен „radio voice" ефект. Прекалената noise reduction прави
гласа метален.

### Export

`Ctrl+Enter` → `File → Render`. Preset `MP4-H264/AAC`, `Full Project`, без
`Rescale`, ако профилът вече е верният. Субтитрите трябва да се **burn**-нат във
видеото, не да останат отделен stream.

Провери готовия файл: `1080×1920` (или `1920×1080`), 30 fps, H.264, AAC, без
черни кадри в началото и края. Изгледай го веднъж на цял екран и веднъж в
прозорец с ширина колкото телефон.

---

## 16. Може ли AI да го направи вместо теб

Да, но има две различни ситуации.

### Вариант A — почти всичко от AI

Подходящ за сценарий D и видеата без глас. Най-прекият вариант е InVideo AI:
избери `Script to Video`, формат `9:16`, английски, faceless, AI voice,
subtitles, и постави:

```text
Create a 25-second vertical 9:16 faceless SaaS education video.
Use a clean dark background, large white captions, simple software icons,
hard cuts, no avatar, no talking person, and no fake product interface.
Use exactly the script below without rewriting it. Use a calm professional
English voice. Keep all text inside the mobile safe area.
```

Под prompt-а сложи точния сценарий. После смени всеки кадър, който изглежда
несвързан или показва измислен софтуер, и провери правописа на всички субтитри.
Безплатният export може да има watermark.

### Вариант B — препоръчаният процес

**За реална демонстрация AI не трябва да измисля интерфейса.** Запиши екрана с
OBS без глас, качи MP4 в VEED, Descript или InVideo, постави точния voiceover,
избери AI voice и automatic captions, **премахни автоматично добавените stock
кадри** — сайтът трябва да остане реален — и провери всяка цена и твърдение.

| Инструмент | Подходящ за | Ограничение |
|---|---|---|
| InVideo AI | най-автоматичен script → voice → visuals → captions | избира общи stock кадри; не бива да измисля интерфейса |
| VEED | browser editor, AI voice, captions, вертикално | AI credits и част от export-а са платени |
| Descript | монтаж чрез редактиране на текста | безплатният план е ограничен; 1080p е платен |
| Canva | заглавни карти, икони, кратки b-roll клипове | клиповете са кратки; не е продуктова демонстрация |
| OBS + Kdenlive | най-точната и безплатна демонстрация | ръчен запис и кратък монтаж |

**Няма инструмент, който надеждно да влезе в живия builder, да натисне
правилните бутони и да гарантира, че продуктовите факти са верни.** Тази част е
истински screen recording или се проверява кадър по кадър.

Не плащай годишен план, преди да си направил поне три тестови видеа и да си
проверил качеството на export-а.

---

## 17. Визуален стил

Не се променя между видеата:

- един дебел sans-serif шрифт
- тъмно мастилено, бяло и един зелен/teal акцент
- hook: максимум два реда
- callout: 1–4 думи
- еднакъв финален CTA кадър
- hard cuts или кратки 4–8 frame преходи
- без 3D ефекти, spinning текст и случайни stock клипове
- логото малко и ненатрапчиво, не огромен watermark

---

## 18. Качване по платформи

**Един чист master в трите платформи.** Не сваляй от TikTok, за да качиш в
Reels — качвай оригиналния MP4, за да няма watermark и повторна компресия.
Адаптирай заглавието, cover frame-а и CTA-то, не файла.

**YouTube (дълго):** кратко точно заглавие, custom thumbnail, обещанието от
thumbnail-а се изпълнява в първите секунди, описание с точния URL и UTM,
chapter markers, end screen към следващо релевантно видео, Test & Compare с до
три thumbnails при възможност.

**YouTube Shorts:** свържи със съответното дълго видео. Първия път качи като
`Unlisted`, отвори на телефон, провери crop-а, после публикувай. Гледай
`Engaged views`, не общите views.

**TikTok:** ясна тема в първите 3 секунди, естествено представяне, звук и
субтитри, релевантни hashtags без общи тагове като стратегия, Creator Search
Insights за теми, video replies към качествени коментари. Не добавяй
автоматичен crop, ако preview-то вече е 9:16.

**Instagram:** Reel за достигане, Stories за текущата аудитория, Trial Reels за
експерименти. Провери preview-то в profile grid — важният текст не трябва да е
отрязан в миниатюрата.

**LinkedIn:** дългото видео като текстов анализ, не като линк.

**Affiliate disclosure** при съдържание с affiliate връзки, видимо, не скрито
след дълъг текст:

> "Some links may earn us a commission. Commission never changes UNSOLERO's
> product ranking."

---

## 19. Измерване и решения

**Основната метрика сега е една:** число, което може да се покаже на партньорски
мениджър. Всичко останало е диагностика.

**Диагностични метрики:** задържане в първите секунди, average watch time,
completion rate, shares, saves, смислени коментари, profile visits, clicks към
конкретната страница, започнати и завършени recommendations, `affiliate click`.

**UNSOLERO funnel:** посещение на точната страница → започнат builder →
завършен builder → отворен продукт → започнато сравнение → affiliate click.

Примерен URL — проектът вече поддържа тези параметри:

```text
/build?utm_source=youtube&utm_medium=organic_video&utm_campaign=agency_stack_001
```

| Наблюдение | Вероятен проблем | Следващ тест |
|---|---|---|
| Много swipes | Слаб hook или неясна първа рамка | Нови първи 2–3 секунди |
| Добро гледане, малко clicks | Слаб или неподходящ CTA | CTA към по-точна страница |
| Много clicks, малко completions | Landing page или UX проблем | Проверка на mobile funnel-а |
| Много views, никакви saves/clicks | Нисконамерена аудитория | По-конкретен бизнес казус |
| Малко views, качествени completions | Ценно нишово B2B съдържание | Повече сходни теми |
| Ранен спад в дълго видео | Intro-то не изпълнява обещанието | По-кратък cold open |

**Анализирай след минимум 10 видеа, не след едно.** Не сменяй цялата стратегия
заради едно слабо видео.

---

## 20. Контролни списъци

### Преди запис

- [ ] Продуктовите факти са проверени от официален източник.
- [ ] Цените са проверени **в деня на записа**.
- [ ] Страницата, към която води видеото, има поне един печелещ продукт.
- [ ] Hook-ът е една ясна идея.
- [ ] Видеото има само един CTA.
- [ ] Известията и личните приложения са изключени.
- [ ] Използва се test/demo акаунт.
- [ ] Направен е 30-секунден тест на звук и картина.

### При монтаж

- [ ] Първите три секунди изпълняват обещанието.
- [ ] Няма ненужно лого intro.
- [ ] Екранът е четим на телефон.
- [ ] Субтитрите са проверени ръчно.
- [ ] Не повече от два реда субтитри.
- [ ] Няма лични или чувствителни данни.
- [ ] Affiliate disclosure-ът е ясен, когато е приложим.
- [ ] Звукът е нормализиран и не clipping-ва.

### Преди публикуване

Гледай готовото видео три пъти:

1. **Без звук** — разбира ли се от визуалното и субтитрите?
2. **Само звук** — логично и естествено ли е обяснението?
3. **На телефон** — чете ли се всичко и закриват ли платформените бутони текст?

- [ ] Export-ът е изгледан изцяло.
- [ ] Няма watermark от друга платформа.
- [ ] Заглавието е точно и не подвеждащо.
- [ ] Thumbnail-ът е четим в малък размер.
- [ ] Линкът води към точната страница.
- [ ] Добавени са правилните UTM параметри.
- [ ] Следващото релевантно видео е свързано.

Ако отговорът на някое е „не", върни се в Kdenlive. Не публикувай с надеждата,
че зрителят ще положи повече усилие от автора.

### След публикуване

- [ ] Отговорено е на смислените коментари.
- [ ] Записани са резултатите след 24 часа и след 7 дни.
- [ ] Коментарите са прегледани за нови теми.
- [ ] Не са правени големи изводи от едно видео.

---

## 21. Източници

**Растеж и канали (проверени 30 август 2026):**

- [Top sources LLMs cite](https://contently.com/2026/04/29/top-sources-llms-cite/) — Reddit №1 във всички answer engines
- [Answer Engine Optimization](https://llmrefs.com/answer-engine-optimization) — AI search трафик +527% годишно
- [SaaS affiliate marketing statistics](https://wecantrack.com/insights/saas-affiliate-marketing-statistics/) — 69% от affiliate трафика е SEO
- [How long does SEO take](https://factoryjet.com/blog/how-long-does-seo-take-2026-month-by-month-timeline) — 4–6 месеца до първо движение за нов домейн
- [SaaS TikTok benchmarks](https://theadranker.com/benchmarks/saas-tiktok) — данните са за платени кампании

**Платформи:**

- [YouTube Recommendation System](https://support.google.com/youtube/answer/16533387)
- [YouTube thumbnail and title tips](https://support.google.com/youtube/answer/12340300)
- [YouTube Shorts analytics](https://support.google.com/youtube/answer/12942217)
- [TikTok Creative Best Practices](https://ads.tiktok.com/help/article/creative-best-practices)
- [TikTok Creator Search Insights](https://support.tiktok.com/en/using-tiktok/growing-your-audience/creator-search-insights)
- [Instagram Trial Reels](https://about.fb.com/news/2024/12/trial-reels-try-content-non-followers-first-see-what-perfoms-best/)
- [Meta Sound Collection — commercial use](https://www.facebook.com/help/instagram/402084904469945)

**Изследвания:**

- [Face presence in social media video](https://www.sciencedirect.com/science/article/pii/S0167811625000096)
- [Buffer consistency study](https://buffer.com/resources/consistent-posting-study/)
- [Edelman–LinkedIn B2B Thought Leadership](https://www.edelman.com/insights/thought-leadership-gets-b2b-buyers-back-into-game)

**OBS и Kdenlive (сверени 26 август 2026):**

- [OBS Standard Recording Output Guide](https://obsproject.com/kb/standard-recording-output-guide)
- [OBS Advanced Recording and Multi-Track Audio](https://obsproject.com/kb/advanced-recording-guide-and-multi-track-audio)
- [OBS Filters](https://obsproject.com/kb/filters-guide)
- [Kdenlive Project Settings](https://docs.kdenlive.org/en/project_and_asset_management/project_settings.html)
- [Kdenlive Speech to Text](https://docs.kdenlive.org/en/effects_and_filters/speech_to_text.html)
- [Kdenlive Subtitles](https://docs.kdenlive.org/en/effects_and_filters/subtitles.html)
- [Kdenlive Normalize 2 Pass](https://docs.kdenlive.org/en/effects_and_filters/audio_effects/volume_and_dynamics/normalize_2pass.html)
- [Kdenlive Rendering](https://docs.kdenlive.org/en/exporting/render.html)

Менютата може да са преведени различно според езика на OBS/Kdenlive. Затова тук
е дадено английското име на всяка команда — то е най-надеждният ориентир.

---

## Статус

Раздели 1–5 се менят с всяка приета или отказана програма и с всяка нова
страница. Раздели 6–21 са наръчник за изпълнение и се менят рядко.

Преразгледай раздел 4 преди всеки месец съдържание. Преразгледай раздели 1–3,
когато има число за показване на партньорски мениджър — тогава кандидатстването
се отваря отново и приоритетът на каналите се сменя пак.
