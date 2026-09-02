# Проучване 1 — Конкурентите: сайтове, дизайн, функционалност

**Дата на проучването:** 2 септември 2026
**Обхват:** какво имат конкурентите, какво ползват, как са построени сайтовете им (архитектура, шаблони на страници, дизайн) и какво имат те, което UNSOLERO няма. **Трафикът е изключен** — той е тема на Проучване 2 и 3.
**Резултат:** раздел 8 — всяко възможно подобрение на платформата, разделено на **отлично / добро / ненужно**.

## Как е направено и какви са ограниченията

- Прегледани са **34 сайта** в пет групи (раздел 2). За всеки са извлечени начална страница, продуктова страница, сравнителна/„alternatives“ страница и категория, където съществуват. Списъкът с адресите е в раздел 10.
- Анализът е върху **HTML, CSS, JSON-LD и текст**, не върху пиксели. Chrome разширението не беше свързано в тази сесия, така че няма скрийншоти. Визуалните бележки са изведени от CSS токени, шрифтови декларации и структура. Където това е ограничение, е казано.
- **G2** връща 403 и CAPTCHA на всеки път; описан е по собствената си документация и новини (средно-ниска увереност). **StackShare** и **efficient.app** са зад Vercel защита (ниска увереност). **Slant.co** е паднал (HTTP 526). **Stackfix** — стартъпът „ние тестваме софтуера вместо вас“, вдигнал $3M през декември 2024 — е **затворил продукта**: днес сайтът е лендинг за AI консултации, а `/compare`, `/software`, `/pricing-calculator` връщат 403. Това е находка само по себе си (раздел 5).
- UNSOLERO е сверено с кода в `frontend/src` и с живите отговори на `unsolero.com`, включително API-то за офертите.

## Резюме в десет реда

1. Категорията има **два оцелели бизнес модела**: маркетплейси, които продават на вендорите (Capterra, GetApp, G2, SaaSworthy), и слаби редакционни affiliate сайтове с newsletter и безплатни инструменти (EmailToolTester, Cloudwards, PCMag). Моделите „тестваме вместо вас без пари от вендори“ (Stackfix) не оцеляха.
2. **Нито един** конкурент не показва източник и увереност за всяко твърдение и **нито един** не мисли на ниво „stack“, освен Findstack (ръчно курирани списъци). Това са двете реални различия на UNSOLERO.
3. UNSOLERO губи най-много от **техническа невидимост**: началната страница, `/products` и `/build` подават празен `<div id="root">` — обхождач без JavaScript (и answer engines) вижда само `<title>`. Всички 34 конкурента подават четимо тяло.
4. Най-постоянната конвенция в категорията е **лентата „с един поглед“ в началото на продуктовата страница**: оценка · начална цена · безплатен план/пробен период · „best for“ · един CTA. UNSOLERO започва с проза.
5. **Странична (side-by-side) таблица** на всяка X-vs-Y страница е стандарт (GetApp до 4 продукта, Capterra с под-оценки, TrustRadius по критерии). UNSOLERO-вите сравнения са есета.
6. **Newsletter** има на всеки редакционен affiliate сайт без изключение (TAAFT 2.5M, Futurepedia 350k, SaaSHub 29k). UNSOLERO няма никакъв начин да задържи посетител.
7. **Плюсове/минуси в две колони**, **FAQ блок** и **FAQPage/Product/ItemList/BreadcrumbList схема** са навсякъде; UNSOLERO подава само Article/WebSite/Person.
8. Всеки конкурент върти 2–7 тракера. UNSOLERO върти **нула** (CSP `connect-src 'self'`) — принципно правилно, но означава, че **нищо не се мери**.
9. Дизайнът на UNSOLERO (Inter + редакционен сериф, почти нулеви радиуси, hairline разделители, без баджове) е **различим** сред utility-class каша и WordPress теми. Слабостта му е липсата на всякакво визуално доказателство: нула скрийншоти, нула видео.
10. Списъкът „ненужно“ е дълъг и важен: ревюта, звезди, общност, AI чат, баджове, sponsored слотове, големи vanity числа. Конкурентите ги имат, защото продават на вендори. UNSOLERO не трябва да ги копира.

---

## 1. UNSOLERO днес — базова картина

Проверено в кода и в живия сайт на 2 септември 2026. Това е точката, спрямо която се мери всичко останало.

### 1.1 Какви страници има

| Група | Страници | Индексирани |
|---|---|---|
| Решение | `/build` (6 стъпки), `/compare` (2–4 продукта), `/wishlist`, `/setups` | не (noindex) |
| Каталог | `/products` (53 публикувани), `/categories` (15), `/brands` (52), детайлни страници | да |
| Редакционни | 13 `/compare/x-vs-y`, 9 `/guides/…`, 1 `/articles/…`, автор | да |
| Доверие | `/how-it-works`, `/about`, `/affiliate-disclosure`, `/privacy`, `/terms` | да |
| Акаунт | вход, регистрация, MFA, възстановяване, `/account`, 17 admin екрана | не |

### 1.2 Анатомия на ключовите страници

**Начална:** hero („Build the right software stack.“, бутони „Build My Setup“ / „Explore Categories“, панел „Live from the catalog“) → „How it works“ (3 стъпки) → примерен план („Within budget“, „Deliberately left out“, „Upgrade later“) → 8 категории → демо сравнителна таблица с бадж „Best fit“ → „Why personalized“ → featured tools → принципи („Fit before features“, „Rejections included“, „Commerce stays separate“) → финален CTA.

**Продуктова:** breadcrumb → ценова карта („Entry price“, „Read from the vendor on {дата}“, „Not a live quote“) → описание, „Save product“ → „Suitability at a glance“ (баджове, изрично „not customer ratings or reviews“) → „Decision profile“ (Strengths / Trade-offs / Best use cases като n/100) → **„Evidence record“** (Verified fact / Manufacturer claim / Merchant observation / Editorial assessment; източник; „Observed {дата} · Confidence {n}/100“; „Inspect source“) → „Where to get it“ (само ако има оферта; бутон „View at {merchant}“, бадж „Affiliate link“) → „Alternatives worth considering“ → плаващ бутон „Compare {n}“.

**Редакционна (compare / guide / article):** breadcrumb → тип → H1 → standfirst → автор + „Published“ / „Updated“ → ценова скала „What each one costs, entry paid tier“ → лепкаво съдържание „In this piece“ → тяло (заглавия, абзаци, списъци, цитати, callout, in-body affiliate CTA карта с дисклоузър) → „Continue the research“ → „Products referenced“ → „Related editorial“. Сървърът подава `schema.org/Article`.

**Builder:** Goal (6 карти) → Team (3) → Budget (слайдер $100–$5,000) → Current stack (15 чекбокса) → Preferences (7, вкл. „Open source“, „EU-hosted data“) → Priorities (6 тежести); странична лента „Your brief“ с **„So far we'd suggest“** на живо. Резултат: „Your Personalized Stack“ → 4 метрики → „Save setup“ → класирани карти („Choice {rank}“, „{score}/100 match“, до 4 причини, „Lower-cost / Premium alternative“) → „Considered alternatives“ → **„Products we deliberately rejected“** с причина.

**Каталог:** „Refine results“ (търсене, категория, марка, мин./макс. цена), сортиране (Featured / Name / Price / Quality score / Value score), карти със suitability баджове, „Compare“, „View details“, „View at {merchant}“ + „Commission never changes the ranking.“, пагинация.

### 1.3 Устройства за доверие, които вече съществуват

Affiliate дисклоузър на пет места; всеки изходящ линк е `rel="nofollow noopener sponsored"`; цената носи датата на прочитане от вендора; Evidence record с класификация, източник и увереност; отхвърлените продукти с причина; `/how-it-works` с четири методологични принципа; твърдение, че разделянето комисиона/ранкинг се пази от тест, който чупи билда; `llms.txt`.

### 1.4 Какво липсва (проверено изрично)

Потребителски ревюта и звезди, `aggregateRating`, newsletter, сделки/купони, AI чат, ценов калкулатор, каквото и да е видео или iframe, тъмна тема, FAQ и `FAQPage`, `BreadcrumbList`, бутони за споделяне, коментари, RSS, публичен линк за споделяне на setup, exit-intent, аларми за цена, „best for“ баджове, сравнителна таблица върху продуктовата страница, „editor's choice“/„winner“ баджове. Лепкавият бутон е само „Compare {n}“, никога affiliate CTA. Една и съща `og-default.png` за всички страници.

### 1.5 Дизайн система

Inter за всичко (без webfont), сериф `Iowan Old Style / Charter / Georgia` за цитати и редакционни заглавия, базов размер 17px. Фон `#f4f5f7`, мастило `#14171c`, акцент дълбоко тийл `#14605a` (в кода се нарича „bronze“), moss/amber/ember, „charcoal“ `#1c1f25` за affiliate бутоните. Радиуси 2px, карти без сянка, hairline разделители, решетки с `gap-px`. Контейнери 42/52/90/105rem. Lucide икони. Само светла тема. Дребна несъгласуваност: `<meta theme-color>` е `#f3f0e9` (топло), а фонът е `#f4f5f7` (студено).

Характер: строг, редакционен, „вестникарски“. Това е силата (доверие, различимост) и слабостта (нула визуални доказателства).

### 1.6 Технически факти, които решават видимостта

- **Vite 7 / rolldown + React SPA, Tailwind v4, Go API, Caddy зад Cloudflare, строг CSP.** Сървърът пренаписва `<head>` за всяка страница и **пререндерира тяло за продуктовите, категорийните, марковите, редакционните и авторските страници, но не за началната, `/products`, `/categories`, `/brands`, hub-овете и статичните страници**. Проверено на 2 септември с `curl`: `/guides/mailchimp-alternatives` подава H1 и всички H2; **началната страница подава 3.5 KB shell само със `<title>`**.
- `robots.txt` спира `/build`, `/compare`, `/wishlist`, `/setups` — правилно срещу дублиране, но **най-различаващият продукт е невидим за търсачки**.
- Cloudflare инжектира блок срещу GPTBot, ClaudeBot, CCBot, Google-Extended. G2 твърди, че 51% от купувачите започват проучването си в AI инструмент — блокът трябва да е съзнателно решение, не подразбиране.
- Бюджети за производителност (100 KB gzip JS, 300 KB първи трансфер) — по-строги от всеки конкурент.

---

## 2. Кои са конкурентите и какво ползват

### 2.1 Пет групи, пет бизнес модела

| Група | Сайтове | Модел на приходи | Технология (наблюдавана) |
|---|---|---|---|
| **Маркетплейси за ревюта** | Capterra, GetApp, Software Advice (трите са Gartner Digital Markets), G2, TrustRadius, SaaSworthy, Crozdesk | вендорите плащат: PPC клик („Visit Website“ само за плащащи), lead-gen формуляри, абонаменти за профил, intent data | Next.js + imgix (GetApp; Capterra вероятно същото), Rails + Bootstrap (Crozdesk), PHP-стил utility CSS (SaaSworthy); всички с GTM/Clarity/Mixpanel/Intercom; всички зад Cloudflare |
| **Директории и общност** | Product Hunt, AlternativeTo, StackShare, SaaSHub, Slant (мъртъв) | promoted листинги, „Official Partner“ спонсорства, верификация, sponsored слотове, affiliate (`rel="sponsored"`) | Next.js + Framer Motion + Segment + PostHog + Vercel (PH); класически SSR (AlternativeTo, SaaSHub с bunny.net CDN) |
| **AI-инструменти директории** | Futurepedia, Toolify, There's An AI For That (TAAFT), Toolfinder, aitools.fyi | featured листинги ($347 подаване, $99/мес. highlight, PPC наддаване при TAAFT), newsletter фунел (Futurepedia „350,000+ AI Adopters“ + безплатен курс), deals-first affiliate (Toolfinder) | Next.js/Nuxt + Tailwind + Vercel; Mixpanel/Segment; тъмна тема (Toolify, PH); aitools.fyi е соло проект с Plausible + beehiiv |
| **Редакционни affiliate сайтове** (най-близкият модел до UNSOLERO) | EmailToolTester, Tooltester, Zapier blog, Style Factory, TechRadar, Cloudwards, Email Vendor Selection, Forbes Advisor, NerdWallet, PCMag | affiliate (Impact, PartnerStack, appwiki, собствени `/out/` редиректи), display реклами, ebooks, платени курсове, консултации | WordPress + Yoast/Rank Math + GTM (ETT, Tooltester, Cloudwards, Style Factory); Next.js + Contentful (Zapier); Tailwind + Vue/React islands на Vercel (TechRadar) |
| **Stack builders и quiz-ове** | StackSelector, The SaaS Stack, HelloGrowthCRM quiz (вендорски), efficient.app, Findstack „Curated Tech Stacks“, Stackfix (затворен) | affiliate, вендорска пристрастност (HelloGrowth), lead-gen | Tailwind + Plausible (StackSelector), WordPress (The SaaS Stack), Next.js + shadcn (HelloGrowth), shadcn + PostHog (Findstack) |

### 2.2 Какво следва от технологиите

- Всичко, построено след 2022, е **Tailwind + Next.js/Nuxt със сървърен рендер**. Редакционните affiliate сайтове са **WordPress**. UNSOLERO е единственият чист SPA без сървърно тяло на началната страница.
- Шрифтове: Inter / DM Sans / системен sans доминират. Брандирани display шрифтове са изключение (Zapier Degular, Style Factory Futura PT, TechRadar Libre Franklin). **Серифната двойка на UNSOLERO е различима.**
- Тракери: 2–7 на сайт. Нула при UNSOLERO.
- CDN за изображения навсякъде (imgix, bunny, Cloudflare Images). UNSOLERO няма изображения, които да оптимизира.

---

## 3. Матрица на функциите

Легенда: ✓ има · ~ частично · ✗ няма · ? не се вижда от извлеченото · † средно-ниска увереност (блокиран сайт).
Колони: CAP Capterra · GA GetApp · SA Software Advice · G2† · TR† TrustRadius · SW SaaSworthy · PH Product Hunt · A2 AlternativeTo · SH SaaSHub · FP Futurepedia · TA TAAFT · ETT EmailToolTester · PCM PCMag · CW Cloudwards · FA Forbes Advisor · **UNS UNSOLERO**

| # | Функция / устройство | CAP | GA | SA | G2† | TR† | SW | PH | A2 | SH | FP | TA | ETT | PCM | CW | FA | **UNS** |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | Собствени потребителски ревюта | ✓ | ✓ | ✓ | ✓ | ✓ дълги | ~ агрегирани | ✓ | ~ likes | ~ | ✓ малко | ✓ малко | ~ коментари | ✗ | ~ коментари | ~ CSI | ✗ |
| 2 | Обща оценка | ✓ ★ | ✓ ★ | ✓ ★ | ✓ ★ | ✓ trScore | ✓ /100 | ✓ ★ | ~ likes | ~ гласове | ✓ ★ | ✓ ★ | ✓ редакц. | ✓ редакц. | ✓ редакц. | ✓ редакц. | ~ suitability, не звезди |
| 3 | Под-оценки по критерий | ✓ 4 | ✓ | ✓ 4 | ✓ | ✓ 8+ | ~ | ✓ 3 | ✗ | ✗ | ✓ 10 | ✗ | ✓ | ✗ | ✗ | ✗ | ~ n/100 в Decision profile |
| 4 | Плюсове / минуси блок | ✓ с % | ✓ | ✓ | ✓ | ✓ цитати | ✓ | ✓ | ~ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ~ само проза |
| 5 | Ценови планове на продуктовата | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | ~ | ~ | ✓ | ~ | ✓ по обем | ~ | ✓ | ✓ мес./год. | ✓ с дата |
| 6 | Отделна /pricing страница | ✗ | ✗ | ✗ | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ | ✗ | ✗ | ✗ | ✗ |
| 7 | Скрийншоти | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | ✓ | ~ | ✓ | ✓ | ✓ hands-on | ✓ | ✗ (галерия има, картинки няма) |
| 8 | Видео | ✗ | ✗ | ✗ | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✓ YouTube | ✗ | ✗ | ✗ | ✗ |
| 9 | Матрица с функции | ✓ 94 + „% critical“ | ✓ | ✓ 80+ | ✓ | ✓ | ✓ + речник | ✗ | ~ тагове | ~ | ~ | ~ | ✓ | ✓ specs | ✓ | ~ | ~ в /compare |
| 10 | Списък интеграции | ✓ | ✓ | ~ | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ~ | ✗ | ✓ | ✗ | ~ като вход в builder-а |
| 11 | Алтернативи блок | ✓ | ✓ | ✓ 6 | ✓ | ✓ | ✓ 10 | ✓ | ✓ ядро | ✓ ядро | ✓ | ✓ | ✓ | ✗ | ✗ | ~ | ✓ |
| 12 | X-vs-Y страници | ✓ | ✓ до 4 | ✓ 17 линка | ✓ AI резюме | ✓ по критерии | ✓ (днес 500) | ✗ | ~ | ✓ | ✗ | ✗ | ✓ | ✗ | ✗ | ✗ | ✓ есе |
| 13 | Compare чекбокси / лепкава лента | ✓ | ✓ | ~ | ✓ | ✓ | ✓ „Compare (0)“ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ~ | ✗ | ✗ | ✓ плаващ бутон |
| 14 | Филтри в категория | ✓ | ✓ 6 групи | ~ | ✓ | ✓ | ✓ ценови кошници, AI | ✗ | ✓ | ~ | ~ | ✓ free-режим | ✗ | ✗ | ✗ | ✗ | ~ цена, марка, категория |
| 15 | Сортиране | ✓ | ✓ | ✓ вкл. „Sponsored“ | ✓ | ✓ | ✓ | ~ | ✓ | ✓ | ~ | ✓ | ✗ | ✗ | ✗ | ✗ | ✓ 6 варианта |
| 16 | Sponsored слот, обозначен | ✓ | ✓ | ✓ | ✓ | ✗ | ~ | ✓ | ✓ | ✓ | ~ | ✓ PPC | ~ | ~ | ✗ | ✓ | ✗ |
| 17 | Баджове / награди | ✓ Shortlist | ✓ | ✓ FrontRunners | ✓ Grid | ✓ Top Rated | ✓ | ✓ дневни | ✗ | ✓ | ✓ Editor's Pick | ✓ | ✗ | ✓ Editors' Choice | ✗ | ~ | ✗ |
| 18 | „Best for“ етикети | ✓ | ~ | ✓ | ~ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ~ | ✓ | ✓ | ✓ | ✓ | ~ 3 сценария в текста |
| 19 | FAQ блок | ✓ | ✓ | ✓ | ✓ | ? | ✓ | ✗ | ✗ | ✗ | ~ | ✓ | ✓ | ✗ | ✓ | ✓ | ✗ |
| 20 | Buyer's guide за категория | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ | ✓ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ (2 guides) |
| 21 | Видима дата на обновяване | ✓ | ✓ месец | ~ | ~ | ✓ | ✓ в заглавието | ✗ | ✓ | ✗ | ✓ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ |
| 22 | История на промените | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ~ | ✗ | ✗ | ✗ | ✓ Releases | ✓ „See Updates“ | ✗ | ✓ | ✗ | ✗ (revision ID има, история не) |
| 23 | Автор с биография на страницата | ✓ | ✗ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ | ✓ | ✓ | ✓ | ~ име + линк, без кутия |
| 24 | Редактор / fact-checker | ✓ | ✗ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ~ | ✓ 2 + fact-check | ✓ | ✗ |
| 25 | Методология / „how we test“ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ |
| 26 | Дисклоузър в началото | ~ | ~ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ~ | ✓ | ✓ | ✓ | ~ в края + отделна страница |
| 27 | Newsletter | ✓ | ✗ | ✓ | ✗ | ✗ | ✗ | ✓ | ✗ | ✓ 29k | ✓ ядро | ✓ 2.5M | ✓ | ✓ | ✓ | ✗ | ✗ |
| 28 | Deals / купони страница | ✗ | ✗ | ✗ | ✓ | ✗ | ✗ | ~ | ✗ | ✗ | ✗ | ✓ | ~ | ✓ | ~ | ~ | ~ оферти без страница |
| 29 | Запазване / списъци | ✓ | ✓ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ | ✗ | ✓ | ✓ | ✗ | ✗ | ✗ | ✗ | ✓ wishlist + setups |
| 30 | Аларми / follow | ✗ | ✗ | ✗ | ~ | ✗ | ✗ | ✓ | ~ | ✗ | ✗ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ |
| 31 | Общност / Q&A / коментари | ✗ | ✗ | ✗ | ✓ | ✗ | ✗ | ✓ | ✓ | ✓ | ✗ | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| 32 | AI асистент / чат | ~ | ✗ | ✓ хора | ✓ Monty | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ~ „Maggie“ | ✗ | ✗ | ✗ |
| 33 | Quiz / recommender | ~ | ✗ | ✓ форма | ✗ | ✗ | ~ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ Smart Quiz | ✗ | ✗ | ✗ | **✓ /build (ядро)** |
| 34 | Калкулатор (цена/ROI/TCO) | ✗ | ✗ | ~ | ✗ | ✗ | ✓ Price Estimator | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ по обем | ✗ | ✗ | ✗ | ✗ |
| 35 | Безплатни инструменти | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ статус страници | ✗ | ✓ | ✓ 4+ | ✗ | ✗ | ✗ | ✗ |
| 36 | Browser extension | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ | ✓ | ✗ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ |
| 37 | Публичен API | ✗ | ✗ | ✗ | ✓ платен | ✗ | ✗ | ✓ | ✗ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |
| 38 | Тъмна тема | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✓ | ? | ✗ | ✗ | ? | ✗ | ✗ | ~ | ✗ | ✗ |
| 39 | Product / AggregateRating / FAQ схема | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ всички | ✓ | ✓ | ✓ ItemList | ✓ | ? | ✓ | ✓ | ✓ FAQ | ✓ | ~ Article, WebSite, Person, Product |
| 40 | Лепкав CTA / лента | ? | ? | ? | ✓ | ? | ✓ лента | ✓ | ? | ✓ | ✓ | ✓ | ✓ | ? | ✓ | ? | ~ само Compare |
| 41 | **Куриране на ниво stack** | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ~ | ✗ | ✗ | ~ колекции | ✗ | ✗ | ✗ | ✗ | **✓ ядро** |
| 42 | **Източник за всяко твърдение** | ~ | ✗ | ✗ | ✗ | ~ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | **✓** |
| 43 | **Увереност за всяко твърдение** | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | **✓** |
| 44 | „Комисионата не мести ранкинга“ изрично | ✗ PPC | ✗ | ✗ сорт по наддаване | ✗ | ✓ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ~ | ✓ | ✓ | ✓ | **✓ + „провери сам“** |
| 45 | Тяло, четимо без JavaScript | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | **~ продукти, категории, марки, редакционни; не начална и списъци** |
| 46 | Тракери на трети страни | много | много | много | много | много | 6+ | 5+ | ? | 2 | Mixpanel | ? | GTM, Clarity | много | GTM | много | **нула** |

---

## 4. Анатомия на страниците — как го правят те и как го правим ние

### 4.1 Продуктова страница

**Стандартът (Capterra, Software Advice, SaaSworthy, Tooltester):**
Лента „с един поглед“ (★ 4.5 от 4,482 · „$20–$75/month · Free version · 3 plans“ · бадж · **„Visit Website“**) → anchor-табове (Description · Use Cases · Alternatives · FAQs · Pros/Cons · Features · Pricing · Integrations · Reviews) → „Compare with a popular alternative“ мини-таблица → плюсове/минуси (Capterra: „93% positive“) → матрица с 80–94 функции („44% of reviewers rated this feature as critical“) → ценови планове → скрийншоти → интеграции с лога → 6–10 алтернативи → 10–17 „Popular comparisons“ линка → FAQ → „Last updated: August 2026“.
SaaSworthy добавя **чекбокс COMPARE до H1** и лепкава лента „Compare (0)“, отделна `/pricing` страница и речник на функциите. Tooltester (най-пълният редакционен шаблон) добавя 17-критерийна таблица с обяснения, 5 видеа, uptime 99.96%, „Techie stuff“ и **история на ревизиите 2012→2026**.

**UNSOLERO:** ценова карта с дата → проза → suitability баджове → Decision profile (n/100) → Evidence record → оферта (само ако има) → алтернативи. Няма лента „с един поглед“, няма плюсове/минуси, няма таблица, няма FAQ, няма скрийншот, няма „popular comparisons“.

**Оценка:** Evidence record е по-силно доверие от всичко в стандарта. Но зрителят го стига след скрол; първият екран не отговаря на трите въпроса, които всеки конкурент отговаря веднага: колко струва, има ли безплатен план, за кого е.

### 4.2 Сравнителна страница (X vs Y)

**Стандартът:** GetApp — „Add up to 4 apps“ с табове Overview / Screenshots / Pricing / User reviews / Key features / Integrations / Alternatives. Capterra — рейтинг + 4 под-оценки едно до друго, оценка по функция с обяснение, „Best for (according to reviews)“ демография („Small businesses account for 77% of reviews“), интеграции 197 vs 70, цитати, „Reviewer verdict“ параграф. TrustRadius — без таблица: секции по критерий (Likelihood to Recommend, Pros, Cons, Usability, Support, Implementation, Contract Terms) с колона за всеки продукт и цитати. G2 — AI резюме на разликите + запазване в „My Lists“. EmailToolTester — редакционно есе със сценарийни заглавия („You're on a tight budget“) и таблица Pros/Cons/Price + „Full review“ + „Try for free“.

**UNSOLERO:** H1 → автор + дата → „The prices“ → „What actually separates them“ → „Where each one is genuinely better“ → „What we are not telling you“ → три сценарийни препоръки. Един inline affiliate линк. Няма таблица, няма оценки, няма FAQ.

**Оценка:** „What we are not telling you“ няма аналог никъде освен Forbes „Why some companies didn't make the cut“ и е ценно. Но читателят на X-vs-Y страница сканира за таблица; без нея UNSOLERO губи първите три секунди.

### 4.3 Категория и списък

**Стандартът:** „Best CRM Software“ + „Last updated August 28, 2026“; филтри в 6 групи (функции, интеграции, ценови модел, устройства, тип организация, оценка); SaaSworthy — ценови кошници (Up to $10, $10–20 … Above $40) и „AI-Powered“; карти с рейтинг + разпределение, плюсове/минуси, ценови snippet, най-високо/най-ниско оценени функции; **„Add to compare“ чекбокс на всяка карта**; топ-5 сравнителна таблица; buyer's guide; FAQ; имейл форма „we'll send a list of the top-rated software“. NerdWallet слага филтри и сортиране и върху редакционна статия.

**UNSOLERO:** търсене, категория, марка, мин./макс. цена; 6 сортирания; карти със suitability баджове; Compare + View details + „View at {merchant}“; пагинация. Няма ценови кошници, няма „free plan“ филтър, няма „компании до N души“, няма таблица, няма FAQ, няма дата в заглавието.

### 4.4 Редакционен affiliate шаблон (Cloudwards, PCMag, Forbes, Style Factory)

**Стандартът:** байлайн с 2 автори + 2 редактори + fact-checker + дата + дисклоузър с линк към editorial-integrity и методология → **Key Takeaways** кутия → бърза таблица (free plan, месечна цена, 3 функции) → за всеки продукт: ценов банер с **„Visit“ + „Review“** (Cloudwards ×17 всеки), плюсове/минуси, „Who is it for“, **„Hands-On Testing“ със скрийншот**, ценова таблица с годишна отстъпка → **„Why some companies didn't make the cut“** → методология таблица → FAQ → биографии → **история на обновяванията** → „Download PDF“. Style Factory: verdict първо, 26 критерийни секции, **„Start Shopify trial >“ ×7**. PCMag: „Jump To“ котви по „best for“, **„Why We Picked It“ / „Who It's For“** с bold персони („Newbies:“, „Growing teams:“), „Bottom Line“, атрибуция на клика по модул в URL-а.

**UNSOLERO:** автор + дата → ценова скала → есе с един inline CTA → продукти → related. Няма Key Takeaways, няма таблица, няма плюсове/минуси, няма повторени CTA, няма FAQ, няма история.

### 4.5 Quiz и stack builders срещу `/build`

- **StackSelector:** 4 въпроса (роля, бюджет <$20 / 20–50 / 50–100 / 100+, екип Solo / 2–5 / 6+, предпочитание All-in-one / Best-of-breed / Mixed), прогрес %, „Why These Questions?“, CTA „Get My Stack Recommendation“, newsletter. Индексиран, публичен.
- **The SaaS Stack:** 15 въпроса / 10 мин. → архетип (Lean Startup, Growing Team, Scale-Up Suite…) + инструменти с цена и сложност на настройка + **90-дневен roadmap + ROI проекция**.
- **HelloGrowthCRM quiz:** вендорски, „Question 1 of 5“ като големи бутони, блок About (What it does / Why it matters / Assumptions / How to interpret), свързани безплатни инструменти (CRM ROI Calculator).
- **Findstack „Curated Tech Stacks“:** Coaching, Creator, Ecommerce, Indie Hacker, Marketing, Personal Productivity — единственият маркетплейс, който мисли над единичния продукт, но ръчно и без причини.
- **Software Advice** (форма → съветник за 15 мин.) и **G2 Monty** (AI чат) са версиите на големите. Никой не връща многопродуктов stack с причини и отхвърляния.

**UNSOLERO `/build`:** по-дълбок от всеки (6 стъпки, живо превю, отхвърлени продукти с причина, Lower-cost / Premium алтернативи). **Но е noindex и JavaScript-only.** StackSelector с 4 въпроса е видим за Google; нашият builder не е. Никой конкурент няма публична страница с резултат за споделяне — UNSOLERO също няма.

---

## 5. Дизайн конвенции през 2026 и анти-шаблони

### 5.1 Конвенции (с по 3+ примера)

1. **Анатомия на картата:** лого → име → един ред tagline → ценово хапче (Free / Freemium / Paid / Free Trial) → едно число за социално доказателство (saves, upvotes, брой ревюта) → 2–3 тага → един изходящ бутон. (Futurepedia, TAAFT, Toolify, aitools.fyi, Product Hunt, GetApp „visit website / Compare / save“.)
2. **Лента „с един поглед“ преди прозата:** оценка + брой, начална цена, безплатен план/пробен период, основен CTA. (Capterra, Software Advice, Tooltester, Forbes, PCMag.)
3. **Anchor-табове под заглавието на дълги страници.** (Capterra, GetApp, TAAFT Overview/Releases/Alternatives/Pricing, Tooltester TOC, PCMag „Jump To“.)
4. **„Best for“ като главен организиращ етикет.** (Forbes, PCMag, Zapier, TechRadar, Software Advice, Cloudwards „Who is it for“, Capterra „Highly rated for“.)
5. **Лента за доверие в началото:** автор + редактор + дата + линк към методология + едно изречение дисклоузър. (Cloudwards, PCMag, Forbes, NerdWallet, Software Advice.)
6. **Плюсове/минуси в две колони с ✓/✗ икони**, понякога с числа („93% positive“, „View 36 more pros“).
7. **Цени като таблица по планове с бележка за базата на таксуване** и все по-често **по обем** (ETT по абонати 250→50,000, Cloudwards годишна отстъпка, Forbes месечно/годишно, Vendr диапазони на договори).
8. **Спонсорството казано на прост език:** Toolfinder „A paid placement… always marked as sponsored“; Software Advice обяснява сортирането по наддаване.
9. **Лепкави елементи:** CTA в хедъра (Tooltester „Try for free“), compare лента (SaaSworthy), фиксирани долни ленти (TechRadar, Findstack).
10. **Сигнали за актуалност навсякъде:** „(September 2026)“ в заглавия, „Last updated“ в хедъра на категорията, история на промените.
11. **Програматични семейства страници за всеки вендор:** review / pricing / vs / alternatives (ETT мега меню, SaaSworthy, Capterra).
12. **Големи числа като доверие** („2.5M+ reviews“, „241,259 products“). UNSOLERO не може и не трябва да се състезава тук; дълбочината и датите са контра-сигналът.
13. **Тъмна тема** е стандарт в AI/tech директориите (PH, Toolify, TechRadar), отсъства в B2B маркетплейсите и редакционните affiliate сайтове.

### 5.2 Анти-шаблони (да не се копират)

- **Pay-to-play по подразбиране:** „Sponsored“ като сортиране по подразбиране (Software Advice); „Visit Website“ само за плащащи, „Learn More“ за останалите.
- **Вендорски шейминг за такси:** банер „It's been two months since this profile received a new review“ и реклами на конкуренти върху неплатени профили (G2); баджове зад абонамент (G2, TrustRadius).
- **Изпиране на ревюта:** агрегирани чужди ревюта показани като собствени („2,383,000+ reviews“ — Findstack, SaaSworthy).
- **Шаблонни страници без съдържание:** Toolify (трафик статистики + спонсорски карти); Toolfinder „No deal available — claim page“.
- **Инжектиране на нерелевантни сделки:** NordPass купони върху CRM класация (TechRadar).
- **Скрити affiliate редиректи без дисклоузър на страницата** (Zapier през `vendors.selectsoftwarereviews.com`).
- **Lead-gated съдържание като основен CTA** (Crozdesk „Get The Report — DOWNLOAD“).
- **Предварително отметнато маркетингово съгласие** (TechRadar newsletter).
- **AI-генериран FAQ пълнеж** (TAAFT, Toolify) — тънък, повторяем, все по-наказван.
- **JS-only shell** (Findstack дублирани решетки; собствените ни начална/products/build).
- **Крехкост на модела:** Stackfix („AI агенти тестват софтуера“) обърна към консултации ~20 месеца след $3M; Venture Harbour и Authority Hacker излязоха от „best tools“; Slant е паднал. Оцеляват маркетплейсите, които монетизират вендори, и слабите редакционни affiliate сайтове с newsletter и безплатни инструменти.

---

## 6. Какво имат те, което ние нямаме — подредено по значение за соло оператор без бюджет

Усилие: S = до един ден, M = 2–5 дни, L = седмици. Стекът е React + Go, което е взето предвид.

| # | Липса | Кой го прави най-добре | Защо има значение тук | Усилие |
|---|---|---|---|---|
| 1 | **Сървърно рендерирано тяло за началната, `/products`, `/categories`, `/brands`, трите hub-а и статичните страници** | всички; GetApp/Futurepedia (Next.js), SaaSHub/ETT (класически SSR) | обхождачи, link preview и answer engines виждат само `<title>` на тези маршрути; продуктовите, категорийните, марковите и редакционните страници вече имат сървърно тяло (проверено в `pagemeta.go`), така че липсват точно входните точки | M |
| 2 | **Лента „с един поглед“** в началото на продуктова и сравнителна страница | Software Advice („$20–$75/month · Free version · 3 plans“), Capterra, Tooltester | най-постоянната конвенция; UNSOLERO започва с проза | S |
| 3 | **Странична сравнителна таблица** с редове по критерий и ред „нашата преценка“ + „добави трети продукт“ | GetApp (до 4), Capterra (под-оценки + verdict), G2 (AI резюме) | X-vs-Y страниците са есета; таблица над есето пази гласа и дава това, което читателят сканира | M |
| 4 | **Плюсове/минуси в две колони** | Tooltester, Capterra (с %), Cloudwards | евтино за рендер от съществуващия текст „where each is genuinely better“ | S |
| 5 | **FAQ блок + `FAQPage`, `ItemList`, `BreadcrumbList` схема** | SaaSworthy (всички), Capterra, ETT, Cloudwards | UNSOLERO подава Article/WebSite/Person/Product; списъците нямат структурирани данни | S |
| 6 | **Newsletter** | Futurepedia (ядро на фунела), TAAFT (2.5M), SaaSHub (29k), ETT, Cloudwards, PCMag | единственото устройство за задържане, което всеки affiliate сайт има; домейн имейлът работи; Go endpoint + Postgres таблица + double opt-in стигат | S–M |
| 7 | **Страница с оферти/сделки** за 15-те живи оферти | Toolfinder (deal-first), TAAFT Deals, Findstack „63 deals“, PCMag | превръща съществуващите проследени оферти в дестинация с ясни „Try free“ етикети | S |
| 8 | **Филтри на /products** (безплатен план, ценови кошници, размер на екипа) | SaaSworthy, GetApp, NerdWallet | сигнализира каталог, не блог; builder-ът вече моделира бюджет и екип | M |
| 9 | **Авторска кутия + линк към редакционните стандарти** на всяка статия | Cloudwards (2 автори + 2 редактори + fact-checker), PCMag, Tooltester | страницата за автор съществува; статията показва само име и дата | S |
| 10 | **История на цените/ревизиите** („See Updates“) | Tooltester (2012→2026), ETT, TAAFT „Releases“ | различието на UNSOLERO са датирани цени; видима история го доказва, и **никой маркетплейс не го прави** | S–M |
| 11 | **Калкулатор „колко струва при N потребители/абонати“** | ETT по абонати, SaaSworthy Price Estimator, Vendr | builder-ът взима бюджет; публичен, индексиран калкулатор по категория е публичното му лице (`/build` е noindex) | M |
| 12 | **Compare чекбокс + запазване на картите + лепкава лента** | SaaSworthy „Compare (0)“, GetApp, G2 „My Lists“ | wishlist и /compare съществуват; не са изведени на каталожните карти | S |
| 13 | **Последователни, повторени изходящи CTA** („Visit website“, „Try for free“, „View pricing“) | Style Factory (×7), Cloudwards (×17), SaaSHub `rel="sponsored"` | guides носят един inline линк; 15 живи оферти са недоекспонирани | S |
| 14 | **Алтернативи + „Popular comparisons“ на продуктовата** | Capterra (10), Software Advice (17), SaaSworthy | кръстосано свързване на съществуващите compare и alternatives страници | S |
| 15 | **OG изображение за всяка страница** | всички редакционни | една `og-default.png` за всичко; продукти и сравнения получават generic preview в социалните мрежи | S |
| 16 | **Публични stack шаблони** („Curated Tech Stacks“) | Findstack (6), TAAFT колекции | setup-ите са частни и noindex; публикувани курирани setup-и правят stack тезата видима | S–M |
| 17 | **First-party анализи** | aitools.fyi, StackSelector (Plausible), Style Factory (Fathom) | CSP забранява всичко; нищо не се мери; self-hosted Umami на същия VPS пази privacy позицията | S |
| 18 | **Актуалност в заглавията** („Best X (September 2026)“) | SaaSworthy, NerdWallet, GetApp | евтино и универсално | S |
| 19 | **Тъмна тема** | Product Hunt, Toolify, TechRadar | често в AI директориите, рядко в B2B; токените са налични | S–M |
| 20 | **„Why X didn't make the cut“** | Forbes Advisor | върви с „What we are not telling you“ | S |
| 21 | Потребителски ревюта, общност, Q&A | G2, Capterra, TAAFT, SaaSHub | игра на обем; непобедима соло и товар за модерация | L (не) |
| 22 | AI асистент | G2 Monty, PCMag „Maggie“ | детерминистичният builder е честната алтернатива | L |
| 23 | Browser extension / публичен API | SaaSHub, AlternativeTo, TAAFT | без смисъл при 53 продукта | L (не) |

---

## 7. Какво имаме ние, което те нямат

- **Детерминистичен, обясним recommendation engine** с източник и увереност за всяко твърдение. Нито един конкурент не показва това; TrustRadius е най-близо с линк от цитат към пълно ревю.
- **Изход на ниво stack**, който отчита вече притежаваните инструменти. Само Findstack (ръчни списъци) и StackShare (developer стекове) мислят над единичния продукт; никой не изчислява съвместимост.
- **„Комисионата не може да мести ранкинг“, казано и направено проверимо.** Software Advice сортира по наддаване по подразбиране; Capterra/GetApp продават „Visit Website“; само TrustRadius и големите редакционни сайтове имат сравнимо твърдение.
- **Датирани, прочетени от вендора цени** като единица истина, с база на таксуване. Маркетплейсите показват подадени от вендора или докладвани от ревюъри цени. (Смесената годишна/месечна база в каталога е известен дефект за поправяне.)
- **Секции с откровеност:** „What we are not telling you“ и „When to stay“ нямат аналог, освен Forbes.
- **Без pay-to-play, без sponsored слотове, без тракери** — наложено от CSP, не само заявено.
- **Един отговорен автор** с Person entity и dateModified на всяка статия.
- **Истинска дизайн система:** Tailwind v4 токени, редакционен сериф + Inter, сдържани семантични акценти. Повечето конкуренти са utility-class каша или WordPress теми.

Честните минуси до това: builder-ът е noindex и JS-only; няма оценки, скрийншоти или плюсове/минуси на първия екран; няма измерване; едно OG изображение; `robots.txt` (Cloudflare) блокира AI обхождачите, с които G2 казва, че започва 51% от проучването на купувачите.

---

## 8. Класификация: ОТЛИЧНО / ДОБРО / НЕНУЖНО

Критерият е един: **дали подобрението увеличава доверието и кликовете към 15-те живи оферти при един човек без бюджет**, без да предава принципите на бранда. „Отлично“ = направи го първо, най-голям ефект за най-малко усилие. „Добро“ = втора вълна, струва си, но не е блокиращо. „Ненужно“ = конкурентите го имат, защото продават на вендори; за UNSOLERO е грешно или невъзможно.

### 8.1 ОТЛИЧНО — направи ги, в този ред

| # | Подобрение | Какво точно | Къде в кода | Усилие |
|---|---|---|---|---|
| 1 | **Тяло за обхождачи на началната, `/products`, `/categories`, `/brands`, hub-овете и статичните страници** | Разшири пререндера, който вече работи за продукти, категории, марки, редакционни и авторски страници, към статичните маршрути: H1, водещ текст, списъци с линкове към продукти/категории/статии. Не е нужен пълен SSR — статичен HTML в `#root` стига. | `backend/internal/transport/httpapi/pagemeta.go`, `public_routes.go` (`resolvePublicRoute` връща тяло само за динамичните маршрути) | M |
| 2 | **Лента „с един поглед“** на продуктова и сравнителна страница | Един ред под H1: начална цена + база на таксуване · безплатен план / пробен период (да/не) · „Best for: …“ (от Best use cases) · бутон „View at {merchant}“ (ако има оферта) · „Evidence: n facts, confidence m/100“. | `ProductDetailPage.tsx`, `ContentDetailPage.tsx` над `EditorialHero` | S |
| 3 | **Странична таблица на всяка `/compare/x-vs-y`** | Рендерирай съществуващия `ComparisonTable` (Money / Suitability / Judgement) за продуктите на статията над есето, със същия ред „Go to vendor“ и същата бележка „We deliberately do not name a winner“. | `ComparisonTable.tsx` + `ContentDetailPage.tsx` | M |
| 4 | **Плюсове/минуси в две колони** | Нов блок тип `pros_cons` в редакционното тяло + автоматичен на продуктовата от Strengths / Trade-offs. | `ContentBody.tsx`, `ProductDetailPage.tsx` | S |
| 5 | **FAQ блок + `FAQPage`, `BreadcrumbList`, `ItemList` схема** | 3–5 реални въпроса на статия и продукт (взети от Reddit, виж Проучване 3). Breadcrumb схемата вече има видими breadcrumbs — само липсва JSON-LD. ItemList за категории и guides. | `pagemeta.go` (сървърният JSON-LD), `ContentBody.tsx` | S |
| 6 | **Повторени изходящи CTA в guides** | Всеки продукт, споменат в guide, получава собствена карта „View at {merchant} · $X/mo · Affiliate link“ на мястото, където е обсъден, не един inline линк. Днес `/guides/mailchimp-alternatives` има 4 печелещи продукта и **един** CTA. | блок `cta` вече съществува в `ContentBody.tsx`; проблемът е редакционен — добави го в съдържанието на всеки guide | S |
| 7 | **Страница `/offers` (или `/deals`) + страница `/links`** | `/offers`: 15-те живи оферти като карти „{Product} · ${price} · Try free/View pricing“, датата на проверка, дисклоузър. `/links`: страница за био в социалните мрежи с по един бутон за всяко текущо видео, всеки с UTM. Двете са задължителни за Проучване 2. | нов маршрут в `router.tsx`, данни от `/api/catalog/products/{slug}/offers` | S |
| 8 | **First-party анализи** | Self-hosted Umami на същия VPS (Node + Postgres, ~512 MB) или Cloudflare Web Analytics (нула инсталация, но по-малко данни). Разреши домейна в CSP `connect-src`. Без това Проучване 2 няма как да се измери. | `compose.production.yaml`, CSP в Caddy/`index.html` | S |
| 9 | **История на цените на продуктовата** | Evidence record вече пази „Fact revision · Score revision“. Покажи хронология „$19 (26 Aug 2026) ← $15 (12 Jun 2026)“ под ценовата карта. Никой маркетплейс не го прави и е точно нашето различие. | `ProductGallery.tsx` (ценовата карта) + endpoint за ревизии | S–M |
| 10 | **Публични курирани stack страници** | `/stacks/{slug}`: „Stack for a 3-person agency under $150/month“ — резултат от builder-а, публикуван като редакционна страница с причини и отхвърляния. Индексируем заместител на noindex `/build`. Всяко видео от Проучване 2 води към такава страница. | нов тип съдържание; резултатът на `RecommendationResults.tsx` вече има всичко нужно | S–M |
| 11 | **Newsletter с double opt-in** | Форма във футъра и в края на всяка статия („Get the stack checklist“). Go endpoint, таблица, потвърждаващ имейл през съществуващия SMTP. Не външен доставчик — противоречи на CSP и на бранда. | нов handler + миграция; SMTP boundary съществува | S–M |
| 12 | **Авторска кутия + „Как е проверено“** | Под заглавието: снимка/инициал, роля, „Prices read from vendor pages on {дата}“, линк към `/articles/how-unsolero-ranks-software`. | `ContentDetailPage.tsx` L95–109 | S |

### 8.2 ДОБРО — втора вълна, след като горните са живи

| # | Подобрение | Бележка | Усилие |
|---|---|---|---|
| 13 | Филтри „Free plan“, ценови кошници, „Team size“ на `/products` | Каталогът е 53 продукта; ефектът е сигнал за каталог, не за намиране. | M |
| 14 | Compare чекбокс + „Save“ на каталожните карти + лепкава лента „Compare (n) · Saved (m)“ | Плаващият бутон вече е там; изведи и wishlist. | S |
| 15 | „Popular comparisons“ + алтернативи с линкове към съществуващите compare/guides на продуктовата | Кръстосано свързване; добро за SEO и за престой. | S |
| 16 | Калкулатор „цена при N абонати / N места“ по категория | Email marketing (по абонати) и CRM (по места) първо. Публичен, индексиран, с CTA към офертите. | M |
| 17 | Собствени скрийншоти на 15-те печелещи продукта | Галерията съществува, картинки няма. Cloudwards „Hands-On Testing“ е най-силното доверие в редакционния сегмент. Само печелещите, само с дата. | M (време, не код) |
| 18 | Вграждане на собствените YouTube видеа в съответната страница | След първите видеа от Проучване 2. Единствено ETT и Tooltester го правят в сегмента. | S |
| 19 | „Why X didn't make the cut“ секция в guides | Builder-ът вече показва отхвърляния; пренеси го в редакционния формат. | S |
| 20 | OG изображение за всяка страница | Генерирано на сървъра (Go шаблон → PNG) или поне по тип страница. | S–M |
| 21 | Актуалност в заглавията („at 1,000 subscribers, September 2026“) | Само за страници, които реално се проверяват месечно. Фалшива актуалност е по-лоша от липса. | S |
| 22 | Anchor-табове / „Jump to“ на продуктовата страница | TOC съществува на статиите; пренеси го. | S |
| 23 | Публичен линк за споделяне на setup (`/setups/{id}/public`) | Прави резултата от builder-а споделяем в Reddit/LinkedIn — виж Проучване 2. | M |
| 24 | Поправи `theme-color` (`#f3f0e9` → `#f4f5f7`) и добави webfont за Inter или приеми системния fallback съзнателно | Дребно, но е дизайн-система. | S |
| 25 | Тъмна тема | Токените са готови; ефектът за B2B аудитория е малък. Последно в списъка. | S–M |

### 8.3 НЕНУЖНО — конкурентите го имат, ние не трябва

| Какво | Кой го има | Защо не |
|---|---|---|
| **Потребителски ревюта, звезди, `aggregateRating`** | G2, Capterra, GetApp, PH, Futurepedia | Игра на обем (Capterra: 4,482 ревюта за един продукт). Соло оператор няма как да събере и модерира. Освен това UNSOLERO изрично казва „not customer ratings or reviews“ — звездите ще противоречат на позицията. Фалшиви/„seeded“ ревюта са правен и репутационен риск. |
| **Общност, форуми, Q&A, коментари** | G2, PH, SaaSHub, TAAFT | Модерация и спам без екип. Reddit вече е общността (Проучване 2). |
| **AI чат асистент** | G2 Monty, PCMag Maggie | Детерминистичният builder е по-честният отговор. Естествен език отгоре — евентуално по-късно, не сега. |
| **Баджове, награди, „Leader“ решетки, „Editor's Choice“** | G2 Grid, Capterra Shortlist, TrustRadius Top Rated, PCMag | Съществуват, за да се продават на вендори. „Best fit“ баджът на началната страница е достатъчен и е контекстуален. |
| **Sponsored / promoted слотове, PPC сортиране** | всички маркетплейси, PH, AlternativeTo | Директно противоречи на „Commerce stays separate“ и на теста, който чупи билда. |
| **Vanity числа** („2.5M+ reviews“, „241,259 products“) | Capterra, SaaSHub, AlternativeTo | UNSOLERO губи всяко такова сравнение. Контра-сигналът е „53 tools, every price dated“. |
| **Трафик статистики на продукта** („Monthly Visitors: 11.9M“) | Toolify | Пълнеж без стойност за решение. |
| **Lead-gated PDF доклади, „Book a call“ съветници** | Crozdesk, Software Advice | Бариера пред информацията. |
| **AI-генерирани FAQ и „What is X“ секции** | TAAFT, Toolify | Тънко съдържание; FAQ да, но с реални въпроси на реални хора. |
| **Browser extension, публичен API** | SaaSHub, AlternativeTo, TAAFT | Без смисъл под няколкостотин продукта и без потребители. |
| **Многоезичност (7–10 езика)** | Tooltester, Toolify | Един език, докато английският не работи. |
| **Injected deals от други категории** | TechRadar | Убива доверието, което е единственият актив. |
| **Претоварени рекламни header/footer ленти, exit-intent попъпи** | TechRadar, Forbes | Дизайнът на UNSOLERO печели точно от липсата им. |
| **Vendor „Claim this page“ форми** | SaaSworthy, TAAFT, Toolfinder | Данните идват от вендорските страници по дата, не от вендорите. |
| **Дълги матрици с 80–94 функции** | Capterra, Software Advice | Шум. Три реда, които решават, са по-ценни. `/compare` вече има „Show only where they differ“ — това е правилният подход. |

### 8.4 Какво да се пази непроменено (силите, които конкурентите нямат)

Evidence record с източник и увереност; отхвърлените продукти с причина; „What we are not telling you“; датата на всяка цена; нула тракери на трети страни (освен собствен self-hosted); нула sponsored слотове; серифът + Inter; hairline дизайнът без сенки и баджове. Всяко „подобрение“, което ги размива, е влошаване.

---

## 9. Ред на изпълнение — 30 дни

| Седмица | Какво | Защо в този ред |
|---|---|---|
| 1 | #7 `/offers` + `/links`, #8 Umami, #6 CTA карти във всеки guide | Проучване 2 стартира видеата тази седмица и трябва да има къде да сочат и как да се мери. |
| 2 | #2 лента „с един поглед“, #4 плюсове/минуси, #12 авторска кутия, #5 FAQ + схеми | Най-евтините промени с най-голям ефект върху първия екран. |
| 3 | #1 тяло за обхождачи на началната и продуктовите | Най-голямото техническо усилие; след него всяка страница е видима. |
| 4 | #3 таблица на compare страниците, #10 първи две публични stack страници, #9 история на цените, #11 newsletter | Правят различията видими и дават дестинации за дългите видеа. |

Всичко от 8.2 — след 30-ия ден и само ако измерването от #8 показва, че хората стигат до страниците.

---

## 10. Източници (всички извлечени на 2 септември 2026)

**UNSOLERO:** unsolero.com/ · /build · /compare/zoho-crm-vs-hubspot · /guides/mailchimp-alternatives · /products · /about · /affiliate-disclosure · /how-it-works · /articles/how-unsolero-ranks-software · /author/andon-pediev · /brands/hubspot · /robots.txt · /sitemap.xml · /api/catalog/products (2 страници) · /api/catalog/products/{slug}/offers за всички 53 продукта · `frontend/src` (router, pages, components, tokens.css) · `backend/internal/transport/httpapi/pagemeta.go`.

**Capterra:** capterra.com/p/152373/HubSpot-CRM/ · /customer-relationship-management-software/ · /customer-relationship-management-software/s/small-businesses/ · /compare/152373-155928/HubSpot-CRM-vs-Zoho-CRM
**GetApp:** getapp.com/ · /customer-management-software/crm/ · /customer-management-software/a/hubspot-crm/ · …/compare/zoho-crm/
**Software Advice:** softwareadvice.com/ · /crm/ · /crm/hubspot-profile/
**G2 (блокиран; вторични):** company.g2.com/news/g2s-refreshed-compare-pages… · documentation.g2.com/docs/release-notes-2025 · documentation.g2.com/docs/reference-pages · documentation.g2.com/docs/product-information · sell.g2.com/profiles · apify.com/zen-studio/g2-product-profile-scraper · blastra.io/blog/g2-capterra-vendor-pricing-compared/ · learn.g2.com/your-buyers-are-using-ai-to-find-software…
**TrustRadius:** trustradius.com/compare-products/hubspot-crm-vs-zoho-crm · blastra.io/guides/how-to-earn-trustradius-badges/ · spotlight-io.com/blog/trustradius-reviews-b2b-buyer-decisions-guide/
**SaaSworthy:** saasworthy.com/product/hubspot-crm · /list/crm-software · /sw-score-methodology · /price-estimator
**Crozdesk:** crozdesk.com/ · /sales/crm-software
**Product Hunt:** producthunt.com/ · /products/hubspot · /products/stackfix
**AlternativeTo:** alternativeto.net/ · /software/mailchimp/
**SaaSHub:** saashub.com/ · /mailchimp-alternatives
**StackShare (429):** stackshare.io/ · /stackups/hubspot-vs-mailchimp
**Slant (526):** slant.co/topics/1305/~best-crm-software
**Futurepedia:** futurepedia.io/ · /tool/notion-ai · /ai-tools/productivity
**Toolify:** toolify.ai/ · /tool/notion-ai
**TAAFT:** theresanaiforthat.com/ai/notion-ai/ · /get-featured/
**Toolfinder:** toolfinder.com/ · /tools/notion
**aitools.fyi:** aitools.fyi/
**EmailToolTester:** emailtooltester.com/en/ · /en/reviews/mailchimp/ · /en/email-marketing-services/ · /en/blog/mailchimp-alternatives/
**Tooltester:** tooltester.com/en/ · /en/reviews/squarespace-review/
**Zapier:** zapier.com/blog/best-crm-app/
**Crazy Egg:** crazyegg.com/blog/best-website-builders/
**Style Factory:** stylefactoryproductions.com/blog/shopify-vs-squarespace
**TechRadar:** techradar.com/best/the-best-crm-software
**Cloudwards:** cloudwards.net/best-project-management-software/
**Email Vendor Selection:** emailvendorselection.com/email-marketing-software/
**Forbes Advisor:** forbes.com/advisor/business/software/best-crm-software/
**NerdWallet:** nerdwallet.com/business/software/best/accounting-software
**PCMag:** pcmag.com/picks/the-best-crm-software
**Vendr:** vendr.com/marketplace/hubspot
**Ramp:** ramp.com/vendor-management (pricing-explorer 404)
**Stackfix:** stackfix.com/ (+ /compare, /software, /categories, /pricing-calculator → 403; blog → 403); techcrunch.com за $3M (дек. 2024)
**Findstack:** findstack.com/
**Venture Harbour:** ventureharbour.com/ (best-email статията 404)
**Authority Hacker:** authorityhacker.com/best-keyword-research-tools/ (пренасочва към началната)
**Stack quiz-ове:** stackselector.com/quiz · thesaasstack.com/saas-stack-selection-quiz-save-money/ · hellogrowthcrm.com/tools/crm-comparison-quiz · efficient.app/quiz, /best/crm, /compare/attio-vs-folk (429)

Не са достигнати: Wirecutter (само потребителски софтуер), Venture Harbour и Authority Hacker „best“ статиите (404 / пренасочени).
