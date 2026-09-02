# Проучване 3 — Аудиторията: какво търси, покриваме ли го, как да я спечелим

**Дата на проучването:** 2 септември 2026
**Въпросите:** кой точно е аудиторията и какво търси най-много; покриваме ли нуждите ѝ — ако да, как и защо, и как да я спечелим; ако не — какво да добавим. Гледна точка: маркетингов стратег и стратег за най-бързо възможно ескалиране в социалните мрежи.
**Свързани:** [01_COMPETITORS_BG.md](./01_COMPETITORS_BG.md) (какво правят конкурентите), [02_SOCIAL_GROWTH_BG.md](./02_SOCIAL_GROWTH_BG.md) (моделът и скриптовете — темите идват от тук).

## Как е направено и какво не можа да се стигне

- Прочетени директно: Capterra / Gartner Digital Markets (2025 и 2026), G2 Buyer Behavior Report 2025 (пълен PDF) и 2026, TrustRadius B2B Buying Disconnect 2024 и 2026, 6sense Buyer Experience Report 2025, Zylo SaaS Management Index 2026, Vertice SaaS Inflation Index 2026, US Chamber 2025, SBA 2025, ICF 2025, Contently (цитирания в LLM, апр. 2026), BrightLocal 2026, Hootsuite Social Trends 2026, ценови страници на вендори и тракери на ценова история, GitHub, AppSumo.
- **Reddit блокира автоматизирано четене** (403/429 на всички пътища и огледала). Заглавия, дати, адреси и откъси са възстановени през търсачка; размерите на общностите — през GummySearch (1 септ. 2026). Броят гласове е наличен само където търсачката го е показала. Нишките **не са четени изцяло** — цитатите са откъси.
- **Обеми на заявки:** Ahrefs/Semrush/SpyFu безплатните инструменти са зад JavaScript или 403. Wordtracker даде **4 заявки** преди лимита. Всичко останало е **Google Autocomplete** (реални предложения, извлечени днес) плюс изрично отбелязана неизточникова оценка. Обемите са редови, не абсолютни.
- Репото: `BUSINESS_MODEL.md` още описва фитнес вертикала. Определението на SaaS аудиторията живее в §6 на стария playbook и в петте типа бизнес на builder-а в seeds: `client_services`, `sell_products_online`, `creator_business`, `solo_consulting`, `software_product`; предпочитания `all_in_one`, `best_of_breed`, `api_first`, `eu_hosted`, `open_source`, `privacy_focused`.

## Резюме в десет реда

1. Аудиторията е точно тази, която builder-ът моделира, и е **по-шумна от преди година**: r/SaaS +106%, r/CRM +131%, r/ProductivityApps +108%, r/automation +86%, r/agency +62% (GummySearch, 1 септ. 2026).
2. **Доминиращата болка е сметката, казана в числа:** „stay under that $25 mark“, „$308 a month to send daily to 7K subs“, „23 subscriptions, $4,100/month“, „death by $29/month“. SaaS инфлацията е 16.4% (Vertice, юни 2026), ~5× CPI.
3. По измерен обем **двете най-печелещи страници вече стоят на двете най-големи заявки**: „mailchimp alternatives“ 4,300/мес. и „calendly alternatives“ 2,700/мес. (Wordtracker). „best crm for small business“ 1,400, „hubspot alternatives“ 1,200.
4. Дупката не е повече категории. Дупката са **„X vs Y“ страници, в които двете страни печелят** (Kit vs MailerLite, Zoho vs Pipedrive, Pipedrive vs monday), **ценовият клъстер** („how much does hubspot cost“, „is hubspot worth it“, „mailchimp free plan limits“), **персона стекове с общо** и **калкулатор по абонати/места**.
5. Как проучват: Google 47%, сайтове с ревюта 38%, **AI търсене 35%** (G2 2025); през 2026 **51% започват с AI чатбот**, но **94% проверяват AI-я** (TrustRadius 2026). Продавачът е **най-малко доверен** източник — 9.3%.
6. **Прозрачната цена е желание №1 четвърта година подред** (TrustRadius). UNSOLERO я има — датирана, нормализирана при 1,000 абонати — **но 25 продукта смесват годишна и месечна база**. Един Reddit отговор, който го посочи, приключва историята за доверие. Поправи преди мащаб.
7. Успешните купувачи предпочитат **ревюта и отраслови съвети пред генеративен AI**, решават до 3 месеца, слагат ≤3 инструмента в късия списък (Capterra 2026; TrustRadius 83%). ~60% от малкия бизнес съжалява за покупка в 18 месеца; 88% от съжаляващите са купили само по вендорска информация (Capterra).
8. **Канали по скорост и доказателства:** Reddit (№1 цитиран в Perplexity, „reddit“ е автоматично довършване към всяка наша заявка) → дълго YouTube (№2 сайт за проучване след Google; №1 цитиран в AI Overviews) → LinkedIn лични постове (14.3% от ChatGPT цитиранията) → SEO (4–6 месеца). Късото видео: без доказателства, че купувачи на бизнес софтуер откриват инструменти там — то е за откриване на канала.
9. Осем от 15 категории не печелят нищо, а най-големите обеми без страница (Zapier, Notion, Slack, Google Analytics алтернативи) са точно там. Ако трафикът дойде първо там, приходът не следва. Насочи всичко ранно към 25-те елемента в раздел 6.
10. **Двадесет теми за говорене** идват дословно от въпросите на аудиторията (раздел 9). Използвай тяхната формулировка като първо изречение.

---

## 1. Кой е аудиторията

| Сегмент (тип в builder-а) | Размер | Бюджет | Спусък за покупка | Къде са (членове, 1 септ. 2026) | Увереност |
|---|---|---|---|---|---|
| **Соло / фрийлансъри / консултанти** (`solo_consulting`) | САЩ: 36.2M малки бизнеса, ~27.1M без служители (SBA, юни 2025). 28% от квалифицираните knowledge workers фрийлансват, $1.5T (Upwork 2025). | $5–$30 на инструмент; соло основател: „$80–90/mo on tools… each tool is $5 or $10 or $20“ (r/Entrepreneur, 19 март 2026). | първи клиент; първа фактура; ударен лимит на free плана (Mailchimp 250, MailerLite 500); имейл за увеличение на цена | r/Entrepreneur 5.3M (+7.9%), r/smallbusiness 2.5M (+14.1%), r/freelance 695K, r/productivity 4.2M | висока за размер |
| **Малки агенции 2–15 души** (`client_services`) | без първичен брой; r/agency 99K, **+61.8%** | 15–30 инструмента: „12 person company… most businesses I talk to are closer to 15–20“ (r/smallbusiness, 31 март 2026); „23 separate subscriptions… $4,100/month“ (r/Entrepreneur, 27 март 2026) | 3–5-ият човек (цена на място); отчети за клиенти; HubSpot Starter→Pro; „$7K onboarding fee?!“ | r/agency 99K, r/marketing 2.0M, r/digital_marketing 373K (+35.6%) | средна |
| **Малки e-commerce** (`sell_products_online`) | ~6.91M живи Shopify магазина (BuiltWith, март 2026) | платформа $39–$105 + приложения; имейл по контакти | Shopify Basic $29→$39 (+34%); имейл сметката расте със списъка | r/ecommerce 672K (+19.7%), r/Emailmarketing 123K (+33.7%) | средна |
| **Създатели на курсове / коучове / newsletter** (`creator_business`) | 122,974 коучове, $5.34B (ICF 2025); Thinkific creators $4.23B | имейлът е най-големият ред: „$308 a month… 7K subs“ (r/MailChimp, март 2026); „20k contacts… $500/month“ (r/Newsletters, февр. 2026); курс платформа $29–$149 | списък минава 1k/5k/10k; първи платен продукт; Mailchimp free cut (февр. 2026) | r/Newsletters 14K (+64%), r/MailChimp 11K, r/ContentCreators | средна |
| **Малки SaaS основатели** (`software_product`) | r/SaaS 794K, **+105.8%**; r/indiehackers 190K, +75.7% | Stripe/Paddle + анализи + help desk + CRM; много free tier и open source | launch; първите 100 клиенти; данъци при Stripe/Paddle | r/SaaS, r/indiehackers, r/selfhosted 831K (+42%), r/n8n 255K (+86.8%) | висока |
| **Локални услуги** (няма тип в builder-а) | част от 36.2M; най-бързо растящите CRM въпроси в Reddit са от услуги („7 salespeople“, „iPad signup, waivers, scheduling, texting“) | CRM под $25/потребител; безплатно записване | първи търговец; излизане от таблици | r/smallbusiness | средна |

### 1.1 Как проучват софтуер (доклади за поведение на купувачите, последни издания)

| Находка | Число | Източник |
|---|---|---|
| Източници при фирми 1–250 служители | Google 47%, сайтове с ревюта 38%, **AI търсене 35%**, анализатори 34%, вендорски сайтове 33%, колеги 28% | G2 Buyer Behavior 2025 (PDF), апр. 2025 |
| Най-доверен източник при решение (Сев. Америка) | GenAI 17.2%, ревюта 13.4%, колеги 13%, вендор 12.6%, анализатори 12.6%, **продавач 9.3%**, форуми 8.4%, инфлуенсъри 7.9% | същият |
| Предпочитат контакт с продажби само късно | 62% (+17 пункта); къс списък 2–3 | същият |
| Започват проучването с AI чатбот по-често от Google | **51%** (29% през апр. 2025); 69% избират друг вендор заради AI; една трета купуват от непознат вендор | G2, 15 апр. 2026 |
| Влияние върху късия списък 2026 | ревюта 38% vs AI 37%; 82% са взели препоръка от AI за 24 мес.; оценяването е най-дългият етап (40%); 47% CFO вето | G2 Buyer Behavior 2026, 22 юли 2026 |
| Ползване на AI и проверка | 63% ползват AI; **94% го проверяват**; 74% ползват ревюта; 83% къс списък ≤3; **прозрачната цена е желание №1 четвърта година** | TrustRadius 2026, 15 юли 2026, n=1,862 |
| Топ сайтове за проучване | Google, **YouTube**, TrustRadius, LinkedIn, вендорски; 73% смятат, че виждат фалшиви ревюта; безплатните пробни периоди са най-влиятелни (74%) | TrustRadius 2024 (по-стар) |
| Самообслужване и LLM | 5.1 вендора оценени, 3.6 в списъка от ден 1; 94% ползват LLM; първият в списъка преди контакт печели ~80% | 6sense 2025, n≈4,000 |
| Съжаление в малкия бизнес | ~60% съжаляват за покупка в 18 мес.; 75% планират да увеличат разхода | Capterra Tech Trends, ян. 2025, n=3,500 |
| Успешните купувачи | само 1 от 3; предпочитат **отраслови съвети и ревюта пред генеративен AI**; решават до 3 мес.; 89% от съжаляващите — проблеми при внедряване | Capterra Software Buying Trends 2026, окт. 2025, n=3,385, 11 страни |
| Само вендорска информация = съжаление | 88% от съжаляващите; съжаляващите ползват социални мрежи 40% vs 18% при успешните; ревютата са вход №1 за успешните | Capterra, ноем. 2024 |
| Натиск върху SaaS разходите | 36% от лицензите неползвани; 78% неочаквани AI такси; 61% отрязани проекти заради увеличения | Zylo 2026 (ентърпрайз тежест) |
| SaaS инфлация | 12.1% (апр. 2026) → **16.4% (юни 2026)** | Vertice |
| AI в малкия бизнес | 58% ползват генеративен AI (40% през 2024) | US Chamber, авг. 2025 |

**Какво следва:** аудиторията се самообслужва, проверява AI-я, вярва на колеги и данни от ревюта, иска цените отпред, слага 2–3 инструмента в късия списък и съжалява, когда е чела само вендора. Сайт с датирани цени, написана методология и къс списък е подравнен с всичко това. Единственото, на което не вярват, е продавачът — и всичко, което звучи като него.

---

## 2. Какво търсят най-много — таблица на търсенето

Ключ за обем: **WT** = Wordtracker (подценява спрямо Ahrefs 2–4×; за подредба). **AC** = Google Autocomplete днес. **оц.** = моя неизточникова оценка, ниска увереност. Статус: ✅ страница има, 🟡 частично, ❌ няма.

| # | Заявка | Обем | Намерение | Страница на UNSOLERO | Печелещ на нея |
|---|---|---|---|---|---|
| 1 | mailchimp alternatives | **4,300 WT**; AC: free, cheaper, for small business, for newsletters, open source, reddit | покупка | ✅ `/guides/mailchimp-alternatives` | Kit, MailerLite, ActiveCampaign, Zoho Campaigns (4/5) |
| 2 | calendly alternatives / free | **2,700 WT** + „free calendly alternative“ 570; AC: free, open source, **dsgvo**, reddit | покупка | ✅ `/guides/calendly-alternatives` | Cal.com, Zoho Bookings |
| 3 | best crm for small business | **1,400 WT**; „best free crm for small business“ 130; AC: free, reddit, for agencies, for ecommerce, for real estate | покупка | 🟡 `/categories/crm` + `/guides/best-crm-small-agency` (само агенции) | Zoho CRM, Bigin, Pipedrive |
| 4 | hubspot alternatives | **1,200 WT**; „cheaper alternative to hubspot“ 32 | покупка | ✅ `/guides/hubspot-alternatives` | Zoho, Pipedrive, ActiveCampaign |
| 5 | hubspot vs zoho / zoho crm vs hubspot | AC №2 „hubspot vs“, №1 „zoho crm vs“; оц. 1–3k | сравнение | ✅ `/compare/zoho-crm-vs-hubspot` | Zoho |
| 6 | hubspot vs pipedrive | AC №4 „hubspot vs“, №1 „pipedrive vs“; оц. 1–3k | сравнение | ❌ | Pipedrive |
| 7 | zoho crm vs pipedrive | AC на двете | сравнение | ❌ | **Zoho + Pipedrive (двете)** |
| 8 | zoho crm vs bigin | AC №3 „zoho crm vs“ | сравнение | ❌ | Zoho (двата продукта) |
| 9 | how much does hubspot cost (per month / per user / small business) | AC: 6 варианта; оц. 2–5k | цена | ❌ | Zoho/Pipedrive като изход |
| 10 | is hubspot worth it (…small business, …reddit) | AC №1, №2, №5; оц. 1–2k | оценка | ❌ | Zoho/Pipedrive |
| 11 | cheapest crm (…small business, …reddit) | AC: 10 варианта; оц. 0.5–1.5k | покупка | ❌ | Bigin, Zoho CRM |
| 12 | best free crm (…small business, …reddit) | AC: 10 варианта; оц. 2–5k | покупка | ❌ | Zoho/Bigin free→paid |
| 13 | best crm for agencies | AC №8 „best crm for“ | покупка | ✅ `/guides/best-crm-small-agency` | Zoho, Pipedrive |
| 14 | best email marketing platforms for small businesses / free | AC; оц. 5–15k общо | покупка | 🟡 `/categories/email-marketing` (без персона guide) | Kit, MailerLite, AC, Zoho Campaigns |
| 15 | mailerlite vs mailchimp | AC №1 „mailerlite vs“ | сравнение | ❌ | MailerLite |
| 16 | mailerlite vs kit / convertkit vs mailerlite | AC №2 „mailerlite vs“, №3 „convertkit vs“ | сравнение | ❌ | **Kit + MailerLite (двете)** |
| 17 | convertkit vs beehiiv / kit vs beehiiv | AC №2, №6 „convertkit vs“ | сравнение | ❌ | Kit |
| 18 | convertkit vs substack / mailerlite vs substack | AC на двете | сравнение | ❌ | Kit, MailerLite |
| 19 | mailerlite vs brevo | AC №3 „mailerlite vs“ | сравнение | ❌ | MailerLite |
| 20 | convertkit vs activecampaign | AC №7 „convertkit vs“ | сравнение | ❌ | Kit + ActiveCampaign |
| 21 | mailchimp pricing / mailchimp price increase | оц. 5–10k; много новини | цена | 🟡 (guide споменава, няма ценова страница) | Kit/MailerLite/AC |
| 22 | zapier alternatives (…free, open source, self hosted, lifetime deal, make) | AC: 10 варианта; оц. 3–6k | покупка | ✅ `/guides/zapier-alternatives` | **никой** (0/3) |
| 23 | zapier vs make | — | сравнение | ✅ `/compare/zapier-vs-make` | никой |
| 24 | n8n vs zapier / n8n vs make | r/n8n 255K, r/automation 230K, +86% | сравнение | ❌ | никой |
| 25 | notion vs clickup | AC №1 „clickup vs“ | сравнение | ❌ | никой (monday печели, не е в двойката) |
| 26 | monday.com vs clickup / asana / notion | AC №1–5 „monday.com vs“ | сравнение | ❌ | **monday.com** |
| 27 | hubspot vs monday crm / pipedrive vs monday | AC №6/8 „hubspot vs“, №3 „pipedrive vs“ | сравнение | ❌ | monday.com, Pipedrive |
| 28 | best project management software for small teams / small business | AC №1, №3; оц. 3–8k | покупка | 🟡 `/categories/project-management`, `/compare/clickup-vs-teamwork` | monday.com, Zoho Projects |
| 29 | clickup vs teamwork | — | сравнение | ✅ | никой |
| 30 | free invoicing software (…small business, freelancers, contractors) | AC: 10 варианта; оц. 5–10k | покупка | 🟡 `/compare/zoho-invoice-vs-wave` | Zoho Invoice, Zoho Books |
| 31 | zoho books vs quickbooks / xero / wave | AC №1, №4, №6 „zoho books vs“ | сравнение | ❌ (QuickBooks/Xero не са в каталога) | Zoho Books |
| 32 | zoho books vs zoho invoice | AC №3 | сравнение | ❌ | Zoho (двата) |
| 33 | best scheduling app for small business (free) | AC №1–2 | покупка | 🟡 `/categories/scheduling`, `/compare/calendly-vs-cal-com-vs-zoho-bookings` | Cal.com, Zoho Bookings |
| 34 | teachable vs thinkific / kajabi / skool / podia | AC №1–7 „teachable vs“ | сравнение | 🟡 `/compare/teachable-vs-thinkific-vs-gumroad` (Kajabi, Skool, Podia липсват) | Teachable |
| 35 | best online course platform for coaches | оц. 2–5k | покупка | 🟡 `/categories/course-platform` | Teachable |
| 36 | ai tools for small business (…owners, marketing, automation, accounting, 2026, reddit) | AC: 10 варианта; оц. 5–15k | информация/покупка | ❌ | Zoho (Zia), AC, monday AI като изход |
| 37 | best tools for solopreneurs (…ai, project management) | AC | покупка | ❌ (няма персона страница) | Zoho, Cal.com, MailerLite |
| 38 | semrush vs ahrefs | AC №1 „semrush vs“ | сравнение | ✅ `/compare/ahrefs-vs-semrush` | SE Ranking само като трети (1/3) |
| 39 | semrush vs se ranking | AC №9 „semrush vs“ | сравнение | ❌ | **SE Ranking** |
| 40 | ahrefs alternatives (…cheaper, reddit) / semrush alternatives | AC: 10 варианта | покупка | ❌ | SE Ranking |
| 41 | slack alternatives (…free, open source, self hosted) | AC | покупка | ✅ `/guides/slack-alternatives` | никой |
| 42 | google analytics alternative (…free, self hosted, matomo) | AC | покупка | ✅ `/guides/google-analytics-alternatives` | никой |
| 43 | notion alternatives (…free, open source, self hosted) | AC | покупка | ❌ | никой |
| 44 | notion pricing (…ai, plans) | AC | цена | ❌ | никой |
| 45 | canva vs figma / adobe express | AC №7, №10 | сравнение | ✅ `/compare/canva-vs-figma` | никой |
| 46 | shopify vs woocommerce / wix / squarespace / bigcommerce | AC №2–10 | сравнение | 🟡 `/compare/shopify-vs-bigcommerce` | никой |
| 47–49 | stripe vs paddle; webflow vs squarespace vs framer; freshdesk vs help scout vs tidio | — | сравнение | ✅ | никой |
| 50 | software stack for small agency / tool stack for consulting agency | **Autocomplete не връща нищо** — фразата е Reddit-родна, не търсачка-родна (r/smallbusiness, 1 май 2026) | информация | ✅ `/guides/software-stack-small-agency` | Zoho, monday, Pipedrive, Cal.com |

**Четене на таблицата:** четирите измерени обема нареждат *alternatives* заявките над *best X for small business* (Mailchimp 4,300 > Calendly 2,700 > best CRM 1,400 > HubSpot 1,200). Двете най-печелещи страници на UNSOLERO вече стоят на №1 и №2. Най-големите дупки са **двойките, в които двете страни печелят** (Kit vs MailerLite, Zoho vs Pipedrive, Pipedrive vs monday) и **ценовият/„worth it“ клъстер** около HubSpot и Mailchimp, който няма страница.

---

## 3. Какво се питат помежду си — 30 реални нишки, 2025–2026

Гласовете са показани където търсачката ги е дала; иначе н/д.

| # | Нишка (общност, дата) | Какво питат / казват | Инструменти | Гласове |
|---|---|---|---|---|
| 1 | „Honestly, what is the best CRM for small business in 2026? I need it for a 5 person team“ (r/smallbusiness, 2026) | „contact management… pipeline view… **stay under that $25 mark**?“ | Bigin препоръчан | н/д |
| 2 | „Best and inexpensive CRM for small businesses“ (r/CRM, 2026) | „HubSpot's high costs despite good features“; тества Zoho, Pipedrive, Freshsales | HubSpot, Zoho, Pipedrive | н/д |
| 3 | „What CRM do small business owners actually use (and can afford)?“ (r/Entrepreneur, 2025) | HubSpot free, Google Sheets | HubSpot | н/д |
| 4 | „What CRM do small business owners actually use?“ (r/ITManagers, 2025) | „HubSpot and Zoho are the most common ones I see out in the wild at SMBs“ | HubSpot, Zoho | н/д |
| 5 | „Which CRM would you recommend for our small business?“ (r/CRM, 2026) | интеграция с **Cal.com, Mailchimp, Tawk**; Pipedrive vs Bigin | Cal.com, Pipedrive, Bigin | н/д |
| 6 | „Best affordable CRM + lead/marketing tools“ (r/CRM, 2026) | Attio vs Zoho vs monday CRM vs Pipedrive, „budget constraints“ | Attio, Zoho, monday, Pipedrive | н/д |
| 7 | „What is the best CRM for a small business in 2025“ (r/sales) | HubSpot vs Pipedrive | | н/д |
| 8 | „**Spending $26K/year on HubSpot with no sales team** — feels like massive overkill“ (r/hubspot, 2026) | заглавието е болката | HubSpot | н/д |
| 9 | „**$7K onboarding fee?!**“ (r/hubspot, 2026) | шок от такса | HubSpot | н/д |
| 10 | „Is HubSpot worth it for a small business, or are there better alternatives?“ (r/hubspot, 17 февр. 2025) | топ отговор: „start with requirements and use cases“ | HubSpot, GHL | 10 и 7 |
| 11 | „Is Hubspot worth it?“ (r/marketing, 9 септ. 2024) | Professional $890/мес.; „good jumping-off point“ | HubSpot | ниско |
| 12 | „Alternatives to Mailchimp?“ (r/MailChimp, 10 март 2026) | „**$308 a month to send daily to 7K subs**“; мисли Substack или ActiveCampaign | Mailchimp, Substack, AC | н/д |
| 13 | „What's cheaper than Mailchimp?“ (r/Newsletters, 4 февр. 2026) | 20k контакти, „$500“; Beehiiv $149 | Mailchimp, Beehiiv | н/д |
| 14 | „Mailchimp alternative? It's getting really expensive…“ (r/Emailmarketing, 28 ян. 2025) | мина на **MailerLite** за цена и UI | MailerLite | н/д |
| 15 | „Is it just me or are email marketing platforms getting outrageously priced?“ (r/Emailmarketing, 11 авг. 2025) | иска pay-as-you-go | | н/д |
| 16 | „I am so frustrated. Calendly Alternatives?“ (r/ProductivityApps, 5 май 2026) | „I'm loving **Cal.com**. Calendly, but with more features and a less confusing interface.“ | Cal.com | н/д |
| 17 | „Review of the Best Calendly Alternatives“ (r/ProductivityApps, 3 апр. 2025) | Cal.com, Zcal, TidyCal ($29 еднократно), Acuity | Cal.com, TidyCal | н/д |
| 18 | „Good alternative for calendly and similar“ (r/BuyFromEU, 11 апр. 2025) | иска **EU-based** | | н/д |
| 19 | „Experience with Cal.com — Calendly alternative“ (r/selfhosted, 13 март 2026) | self-hosting | Cal.com | н/д |
| 20 | „If your company is paying for 10+ software subscriptions something is really broken“ (r/smallbusiness, 31 март 2026) | „12 person company… most are closer to 15–20“ | | н/д |
| 21 | „The SaaS model is quietly falling apart for small businesses“ (r/Entrepreneur, 27 март 2026) | „**23 separate subscriptions… $4,100/month**… five years ago $1,200“ | | н/д |
| 22 | „Any other business owners sick of paying per user pricing?“ (r/smallbusiness, 13 септ. 2025) | „wishing for a fixed priced bundle“ | | н/д |
| 23 | „We have 5 subscriptions of the same software because nobody talks to each other“ (r/smallbusiness, 7 ноем. 2025) | пет Notion абонамента, „$900 per month“ | Notion | н/д |
| 24 | „How do you stay on top of all your software subscriptions?“ (r/smallbusiness, 14 юли 2025) | „**death by $29/month**“ | | н/д |
| 25 | „Tired of Zapier“ (r/automation, 9 окт. 2025) | сметка „from $10/month to over $750“ | Zapier, n8n | 3 |
| 26 | „Why is n8n so much more popular than make & zapier?“ (r/n8n, 7 юни 2025) | „Because a bunch of **grifters are selling the dream**“ (**182**); „$24/month for 2500 executions“ (69) | n8n, Make, Zapier | **182 / 69 — най-високият намерен** |
| 27 | „My marketing agency comprehensive 2025 tech stack — 7-fig agency“ (r/agency, 4 юни 2025) | 10 души: QuickBooks, ClickUp, Clockify, Loom, Mailchimp, HubSpot, BrightLocal, Figma, Canva | | н/д |
| 28 | „What's the best tool stack to manage a consulting agency?“ (r/smallbusiness, 1 май 2026) | „the '**one tool to rule them all**' dream rarely survives contact with reality“ | | н/д |
| 29 | „Agencies here – what's your current tool stack and what sucks about it?“ (r/marketingagency, 5 ян. 2026) | Moxie, GChat, Loom, AgencyAnalytics | | н/д |
| 30 | „I switched from Zapier to n8n — here's what actually changed“ (r/nocode, 17 апр. 2026) | „No per-task pricing (huge relief)“ | n8n | н/д |

Не стигнати тази сесия: нишки за PM инструменти (ClickUp/Notion/monday), creator нишки (Kit/MailerLite/beehiiv), курс платформи, фактуриране, Indie Hackers.

### 3.1 Повтарящите се болки, с техните думи

| Болка | Тяхната формулировка | Нишки |
|---|---|---|
| Ценово пълзене | „getting really expensive“, „repeated price hikes“, „$10/month to over $750“, „why does every subscription billing software suddenly cost $200/month once you start growing“ | 12, 14, 25, r/smallbusinessesowners 9 юни 2026 |
| Цена на място | „sick of paying per user pricing“, „wishing for a fixed priced bundle“ | 22 |
| Твърде много инструменти | „10+ subscriptions… broken“, „death by $29/month“, „23 separate subscriptions“, „5 subscriptions of the same software“ | 20, 21, 23, 24 |
| Прагове и скрити такси | „$26K/year… no sales team“, „$7K onboarding fee?!“, „$890/mo“ | 8, 9, 11 |
| Бюджетни тавани като числа | „stay under that $25 mark“, „$308 a month… 7K subs“, „20k contacts… $500“ | 1, 12, 13 |
| Интеграцията решава | „integration with Cal.com, Mailchimp and Tawk“, „eliminate duplicate data entry“ | 5 |
| Свиване на free плановете | „Why are so many tools that used to be free now charging crazy prices… free versions shrinking“ | r/productivity 27 септ. 2025 |
| Умора от AI хайп | „a bunch of grifters are selling the dream“ (182) | 26 |
| Недоверие към „едно решение“ | „the 'one tool to rule them all' dream rarely survives contact with reality“ | 28 |
| Страх от миграция | „considering switching… after repeated hikes“ но още на Mailchimp; „100 Zaps running“ | 12, r/n8n дек. 2025 |
| Доверие в ревюта | „honest opinions“, „actually worth it“, „genuine ROI and honest user experiences“ | r/CRMSoftware, r/Entrepreneur |
| Суверенитет / privacy | „EU-based solutions“; autocomplete „…dsgvo“ за Calendly и Zapier | 18 |

**Най-често назовани инструменти:** HubSpot (далеч най-много — като нещото, което напускат или от което се страхуват), Zoho/Bigin, Pipedrive, Mailchimp (напускат), MailerLite, ActiveCampaign, Beehiiv/Substack, Cal.com, Zapier (напускат), n8n, Make, ClickUp, Notion, monday, GoHighLevel, Attio, QuickBooks. **Осем от единадесетте най-назовани са в каталога; пет от тях печелят** (Zoho, Bigin, Pipedrive, MailerLite, ActiveCampaign, Cal.com, monday).

---

## 4. Какво расте през 2026

| Тренд | Доказателство | Връзка с UNSOLERO |
|---|---|---|
| **AI агенти / автоматизация за малък бизнес** | 61% от купувачите ползват или планират AI агенти (G2 2026); 58% от малкия бизнес ползва GenAI (US Chamber); r/n8n 255K (+86.8%), r/automation 230K (+85.9%); n8n 203K GitHub звезди; autocomplete „ai tools for small business automation / accounting / marketing / 2026“ | Automation категорията не печели нищо. Опасност: най-гласуваният коментар по темата нарича маркетинга на AI-автоматизация „grifters selling the dream“ — тонът трябва да е трезв. |
| **Консолидация vs best-of-breed** | 36% лицензи неползвани (Zylo); Reddit „10+ subscriptions… broken“ vs „one tool… rarely survives“; G2: 91% очакват нови ценови структури | **Това е продуктът.** Builder-ът има `all_in_one` vs `best_of_breed` и откриване на дублирания. Най-недоизползваното различие. |
| **Увеличения на цените 2025–2026** | таблицата долу; SaaS инфлация 16.4% | Всяко увеличение е скок на „alternatives“ заявката. Датираните цени са активът — **но дефектът с базата на таксуване трябва да се поправи преди всяко ценово съдържание.** |
| **EU-hosted / privacy-first** | r/BuyFromEU 364K (+50.1%); autocomplete „calendly alternative dsgvo“, „zapier alternative dsgvo“; european-alternatives.eu | Builder-ът има `eu_hosted`. Каталогът: Brevo (FR), MailerLite (LT), Simple Analytics (NL), n8n (DE), Salesflare (BE), Cal.com (EU опция), Zoho (EU DC). „EU-hosted stack“ страница е евтина и никой в affiliate пространството не я прави добре. |
| **Lifetime deals** | AppSumo 1.5M+ предприемачи; autocomplete „zapier alternative lifetime deal“; Reddit препоръчва TidyCal „$29 one-time“ | Извън scoring-а, но добра тема: „кога LTD бие абонамент и кога те оставя без изход“. |
| **Open source / self-hosted** | r/selfhosted 831K (+42%); n8n 203K, Twenty 56K, Cal.com 48K, listmonk 23K звезди; autocomplete „…open source / self hosted“ за Notion, Slack, GA, Zapier, Mailchimp | Builder-ът има `open_source`. Каталогът: n8n, Umami, Cal.com. „Self-host or pay?“ страница подхожда на честната методология и е силен Reddit отговор. |
| **Answer engines като канал** | 51% започват с AI (G2); Reddit е 1 от 5 цитирания в Perplexity; LinkedIn 14.3% от ChatGPT Search отговорите; сайтове с ревюта получават 4.6–6.3 цитирания vs 1.8 (Contently, апр. 2026) | Cloudflare `robots.txt` блокира GPTBot, ClaudeBot, Google-Extended, CCBot — grounding обхождачите, които строят корпуса. Не блокира OAI-SearchBot/PerplexityBot. **Реши го съзнателно.** |

### 4.1 Конкретни увеличения 2025–2026, за които аудиторията се оплаква

| Вендор | Промяна | Източник |
|---|---|---|
| **HubSpot** | seat pricing от март 2024; Starter $15–20/място vs Marketing Hub Professional $890/мес. + $3,000 onboarding (~44× праг); контактите скачат ~$250/мес. при 2,001 | Docket, TinyCommand, Encharge (авг. 2026) |
| **Mailchimp** | ~11% ян. 2026; **free план 500 → 250 контакта / 500 изпращания от 17 февр. 2026** (2,000 през 2022, −87%); автоматизациите махнати от free средата на 2025; legacy планове +11–13% от 13 апр. 2026; такса за отписани контакти | Audienceful (15 юли 2026), ActiveCampaign blog, CampaignHQ, PriceTimeline |
| **Canva** | Teams от $119.99/год. (5 души) на $100/потребител/год. (~+317%) септ. 2024; Pro $12.99→$15→$18 | TechCrunch (3 септ. 2024), SaaSPricePulse |
| **Notion** | Business $15→$20 (юни 2025, +33%); AI add-on $10 махнат, AI вкаран в Business (13 авг. 2025) | PricePulse, UserJot |
| **Zapier** | task overages; Trustpilot 1.5/5 (72% една звезда, юли 2026); Reddit „$10 → $750“ | Siit, StartupOwl, r/automation |
| **Shopify** | Basic $29→$39 (+34%), Shopify $79→$105 (2024); Plus такса нагоре 2026; Flexport минимум $500→$5,000/мес. | PricePulse, Zenventory, Shopify Community |
| **Calendly** | без базово увеличение; два AI add-on-а $8/място от 19 авг. 2026 | Carly |
| **Google Workspace** | Gemini вкаран с увеличение (Business Standard $14), 17 март 2025 | Google Workspace blog (средна увереност за делтата) |
| **MailerLite** (наш affiliate) | free план **250 абонати / 2,500 имейла месечно**, 2 места (прочетено на mailerlite.com/pricing на 2 септ. 2026; вторичен източник твърдеше 500 — грешен) | mailerlite.com/pricing |
| **Klaviyo** | таксуване по активни профили от февр. 2025 | Brevo blog |
| **Semrush** | преопаковка: SEO $139 / Starter $199 / Pro+ $299 / Advanced $549 (годишно $117.33 / $165.17 / $248.17 / $455.67) | Semrush pricing, прочетено 2 септ. 2026 |

---

## 5. Покриваме ли нуждите — анализ на дупките

### 5.1 Покрито добре — и защо е различие

| Нужда | Страница | Защо бие списъка |
|---|---|---|
| Изход от Mailchimp, ценен честно | `/guides/mailchimp-alternatives` | Пет инструмента **при 1,000 абонати ($5.25–$39)**, датирано 21 авг. 2026, изрично „не сме тествали deliverability, всеки който твърди класация по нея гадае“. Нормализацията отговаря точно на Reddit оплакването „$308 за 7K“ така, както никой вендорски блог не прави; 4 от 5 изхода печелят. |
| Изход от Calendly вкл. open source и Zoho | `/guides/calendly-alternatives`, `/compare/calendly-vs-cal-com-vs-zoho-bookings` | Заявка №2 по обем с варианти „free“, „open source“, „dsgvo“; Cal.com и Zoho Bookings печелят; любимецът на Reddit (Cal.com) е печелещият вариант. |
| Zoho vs HubSpot при равна цена | `/compare/zoho-crm-vs-hubspot` | $20/потребител за двата, присъда по персона (1–5 души → HubSpot, вече на Zoho → Zoho, соло → Bigin $9), уговорка „не е присъда от година ежедневна употреба“. Съвпада с това как купувачите правят късия списък (2–3, TrustRadius 83% ≤3). |
| Отговор на ниво stack за „твърде много инструменти“ | builder-ът (`all_in_one`/`best_of_breed`, дублирания, пет типа бизнес) | Единственото, което пазарът няма: болката в нишки 20–24 е *портфолио* проблем, а всеки конкурент отговаря категория по категория. Builder-ът е причината сайтът да съществува; съдържанието още не води с него. |
| Методология и защитна стена за доверие | `/articles/how-unsolero-ranks-software`, `commercial_boundary_test.go` | Продавачът е 9.3% доверие, 94% проверяват AI-я; тест, който чупи билда при комерсиални термини в scoring-а, е твърдение, което никой списък не може да направи. **Невидимо от guide страниците.** |

### 5.2 Покрито частично — какво липсва на съществуващата страница

| Страница | Липсва |
|---|---|
| `/guides/mailchimp-alternatives` | таблица с free лимитите (Mailchimp 250 контакта / MailerLite 250 абонати и 2,500 имейла / Kit 10,000 / Brevo 300 имейла на ден / Zoho Campaigns) — cut-ът от февр. 2026 е *спусъкът*, а страницата го споменава бегло; странична таблица; стъпки за миграция отвъд „бюджетирай две седмици“; скрийншоти или видео; цена при 5k и 10k (Reddit числата) |
| `/compare/zoho-crm-vs-hubspot` | база на таксуване и годишна отстъпка; Pipedrive като трета опция (двете печелят; autocomplete иска „hubspot vs pipedrive“); стълба „колко струва при 5 / 10 места“ — прагът, от който се страхуват; **никакъв CTA** (проверено днес) |
| `/guides/how-to-choose-business-software` | без affiliate дисклоузър, без линк към методологията, без правилото за нормализация на цените. Това е страницата, която answer engine ще цитира за „how to choose“ — трябва да носи устройствата за доверие |
| `/guides/best-crm-small-agency` | само агенции; „best crm for small business“ (1,400), „cheapest crm“, „best free crm“ нямат страница |
| `/compare/teachable-vs-thinkific-vs-gumroad` | autocomplete иска Kajabi, Skool, Podia, Circle; Thinkific и Gumroad не печелят (Thinkific отказа 30 авг. 2026); Teachable рендерирана цена ≠ JSON-LD — страницата да каже коя ползва |
| `/compare/ahrefs-vs-semrush` | Semrush преопакова плановете; печелещият (SE Ranking) е трето колело; „semrush vs se ranking“ няма страница |
| `/categories/*` (15) | без персона филтър, без колона „free plan“, без маркер за база на таксуване (25 продукта смесват); freshness показва „stale“ след 7 дни, докато нищо не препрочита цените |
| `/guides/zapier-alternatives`, `/compare/zapier-vs-make` | n8n е любимецът на общността (255K, 182-гласова нишка) и е в каталога, но няма n8n-vs-Zapier; няма ред „self-host cost“; нищо в двойката не печели |

### 5.3 Непокрито

**Липсващи категории (по доказателства за търсене):** счетоводство (QuickBooks, Xero — „zoho books vs quickbooks/xero“ са топ autocomplete, а Zoho Books печели), форми/анкети, e-signature („$50–60/month just to handle signing documents“, r/SaaS), AI асистенти (ChatGPT/Claude/Gemini са най-експенсваните приложения), social scheduling (Buffer/Planable във всяка агенционна нишка), time tracking (Clockify), password managers, VoIP/SMS, payroll/HR, записване за локални услуги (Fresha, Square Appointments), community/membership (Skool, Circle).

**Липсващи типове страници:** ценови обяснения („how much does hubspot cost“), „is X worth it“, „X free plan limits“, „cheapest X“, „best free X“, „X vs Y vs Z for [персона]“, „best tools for [персона]“, миграционни ръководства (Mailchimp→MailerLite/Kit, Zapier→n8n/Make, HubSpot→Zoho/Pipedrive), stack шаблони по тип бизнес с общо месечно, калкулатор по места/абонати, EU-hosted stack, self-hosted-vs-paid калкулатор, „AI функции по план“ таблица.

---

## 6. Какво да добавим — 25 елемента по приоритет

Оценка: доказателство за търсене × връзка с живи програми × усилие за един човек × годност за късо видео. „Печели“ = само живите днес.

| # | Страница / функция | Заявка | Персона | Печели | Hook за видео |
|---|---|---|---|---|---|
| 1 | **Обяснение на Mailchimp free cut + 3-стъпкова миграция към MailerLite/Kit** (разшири guide-а) | „mailchimp free plan limits“, „mailchimp alternatives free“, „mailchimp price increase“ | creator, solo | MailerLite, Kit, AC | „Mailchimp's free plan went from 2,000 contacts to 250. Here's what 1,000 subscribers costs on five tools today.“ |
| 2 | **Kit vs MailerLite** | „mailerlite vs kit“, „convertkit vs mailerlite“ | creator | **Kit + MailerLite** | „Both pay me the same commission, so here is the honest split at 1k, 5k, 10k.“ |
| 3 | **Pipedrive vs Zoho CRM (vs Bigin)** | „pipedrive vs zoho“, „zoho crm vs bigin“ | client_services, solo | **Zoho + Pipedrive** | „Two CRMs under $25 a seat. The 5-person team from Reddit asked; here's the answer.“ |
| 4 | **HubSpot pricing explained 2026 — Starter→Pro прагът** | „how much does hubspot cost“, „is hubspot worth it“ | client_services, software | Zoho, Pipedrive, AC като изход | „$20 a seat to $890 a month plus $3,000 onboarding. Where the cliff is.“ |
| 5 | **Best CRM for small business (under $25/seat)** | „best crm for small business“, „cheapest crm“, „best free crm“ | всички | Zoho CRM, Bigin, Pipedrive | „Every CRM Reddit actually recommends, priced for a 5-person team.“ |
| 6 | **Stack шаблони по тип бизнес с общо месечно** (изведи builder-а) | „software stack for consulting agency“ — Reddit-родна | петте типа | Zoho suite, monday, Pipedrive, Cal.com, MailerLite | „A 3-person agency's whole stack for under $150 a month — every price read this week.“ |
| 7 | **Калкулатор по абонати/места** в email и CRM страниците | „email marketing pricing 10,000 subscribers“, „crm cost 10 users“ | creator, agency | Kit, MailerLite, AC, Zoho, Pipedrive, monday | „Type your list size, see the real bill on five tools. No ‚starting at'.“ |
| 8 | **Best email marketing platform for small business / for newsletters** | „best email marketing platforms for small businesses“, „…free“ | solo, creator, ecommerce | Kit, MailerLite, AC, Zoho Campaigns | „Five email tools, one rule: what does 1,000 subscribers cost, and what happens at 5,000.“ |
| 9 | **Kit vs beehiiv vs Substack** (добави beehiiv, Substack като непечелещи) | „convertkit vs beehiiv“, „convertkit vs substack“ | creator | Kit | „Substack takes 10%. Beehiiv is $149 at 25k. Kit is per-subscriber. Break-even month by month.“ |
| 10 | **monday.com vs ClickUp vs Notion за малки екипи** | „monday.com vs clickup“, „clickup vs notion“, „best project management software for small teams“ | client_services, software | monday.com | „Per-seat maths for a 5-person team on three PM tools — one charges for guests.“ |
| 11 | **Pipedrive vs monday CRM vs HubSpot** | „pipedrive vs monday“, „hubspot vs monday crm“ | client_services | Pipedrive, monday | „monday sells a CRM now. Is it a CRM?“ |
| 12 | **Zoho Books vs Zoho Invoice vs Wave (vs FreshBooks)** + „free invoicing software for freelancers“ | „free invoicing software for freelancers“, „zoho books vs zoho invoice“ | solo, client_services | Zoho Books, Zoho Invoice | „Zoho Invoice is free. Zoho Books is not. The exact feature where you switch.“ |
| 13 | **Zoho Books vs QuickBooks vs Xero** (добави QB/Xero като непечелещи) | „zoho books vs quickbooks“, „…xero“ | solo, client_services | Zoho Books | „The two accounting tools everyone defaults to, against the one that costs a third.“ |
| 14 | **EU-hosted малък бизнес stack** (`eu_hosted` → страница) | „calendly alternative dsgvo“, „zapier alternative dsgvo“, „european alternative to mailchimp/hubspot“ | всички, EU | Zoho (EU DC), MailerLite, Cal.com | „A stack where every vendor keeps your data in the EU — and what it costs vs the US default.“ |
| 15 | **Self-host or pay? n8n / Cal.com / Umami / listmonk vs SaaS** | „…self hosted“ | software, solo | Cal.com cloud; предимно доверие | „A €5 VPS runs n8n, Cal.com and Umami. Here's the real hourly cost.“ |
| 16 | **Zapier vs Make vs n8n — цена при 1k/10k/50k tasks** (разшири compare) | „n8n vs zapier“, „zapier alternatives“ | software, agency | никой — игра за цитирания (182-гласова тема) | „Reddit's most-upvoted comment says the n8n hype is grifters. Let's price it anyway.“ |
| 17 | **Teachable vs Thinkific vs Kajabi vs Skool vs Podia** | „teachable vs kajabi / skool“ | creator | Teachable | „Kajabi is $149. Skool is $99. Teachable Starter is $39 monthly, $29 annual — and its own page shows two prices.“ |
| 18 | **Best tools for solopreneurs (stack под $50/мес.)** | „best tools for solopreneurs“ | solo | Zoho Invoice/Bigin, Cal.com, MailerLite | „The solo founder on Reddit pays $80–90/month without noticing. Here is a $45 stack.“ |
| 19 | **Best scheduling app for small business (free)** — локални услуги | „best scheduling app for small business free“ | local, solo | Cal.com, Zoho Bookings | „Free booking page, payments, reminders — which still costs $0 at 100 bookings a month.“ |
| 20 | **SE Ranking vs Semrush vs Ahrefs** (пресечи около печелещия) | „semrush vs se ranking“, „ahrefs cheaper alternative“ | agency | SE Ranking | „Semrush just repackaged to $139–$549. What a 3-client agency needs, at $103.20 billed yearly.“ |
| 21 | **„Is X worth it“ серия** (HubSpot, Zapier, Notion Business, Canva Pro) | „is hubspot worth it for small business“ | всички | Zoho, Pipedrive, AC като изход | „Worth it for whom? A persona-by-persona verdict, not a star rating.“ |
| 22 | **AI в стека ти: какво струва и прави AI планът на всеки вендор** | „ai tools for small business 2026“, „notion pricing ai“ | всички | Zoho (Zia), AC, monday | „Every vendor added an AI tier in 2025. What each charges and what it removed from your plan.“ |
| 23 | **Миграционни ръководства**: Mailchimp→MailerLite/Kit; Zapier→Make/n8n; HubSpot→Zoho/Pipedrive | „how to migrate from mailchimp to mailerlite“ | creator, agency | MailerLite, Kit, Zoho, Pipedrive | „Export, import, re-consent: the 40 minutes it actually takes to leave Mailchimp.“ |
| 24 | **Индекс на free плановете** (една таблица, всички продукти, датирана) | „X free plan limits“, „free crm“ | всички | Zoho, MailerLite, Kit, Cal.com | „Every free plan in one table, re-read monthly. Mailchimp cut theirs twice; so did MailerLite.“ |
| 25 | **Поправка на базата на таксуване + „price read on“ навсякъде** (инфраструктура) | подкрепя всяка ценова заявка | всички | всички | Позволява честния hook „every price read on [date], monthly billing“ и предотвратява зрител да хване годишна цена, показана като месечна. |

Елементи 1–8 могат да станат през септември без нови програми; 9, 13, 17 изискват добавяне на непечелещи продукти в каталога, за да се сравнява печелещият с това, което хората реално пишат.

---

## 7. Как да ги спечелим

### 7.1 Какво кара тази аудитория да вярва на източник

| Устройство за доверие | Доказателство | Можем ли да го твърдим сега |
|---|---|---|
| Прозрачна, датирана цена | желание №1 четвърта година (TrustRadius 2026) | **Да, с уговорка:** датирани и нормализирани, но 25 продукта смесват база и нищо не препрочита — поправи преди да се облегнеш |
| Гласът на колеги / ревюта | ревютата са топ вход за късия списък (38%, G2 2026); успешните ги предпочитат пред GenAI (Capterra 2026); 74% ги ползват | **Още не** — няма ревюта и не могат да се измислят. Заместител: цитирай и линквай Reddit нишките на всяка страница, с дати |
| Написана методология, без pay-to-play | продавачът 9.3%; 88% от съжаляващите — само вендорска информация | **Да, уникално** — но не е линкнато от guide-овете; how-to-choose няма дисклоузър |
| Показани отхвърляния и компромиси | късият списък е 2–3; недоверие към „one tool“; първият в списъка печели 80% | **Да** — engine-ът дава отхвърлени с причини; редакционните страници не ги показват |
| Актуалност | 74% искат ревюта от последните 3 месеца (BrightLocal 2026) | **Да** — всяка страница има дата; направи „price read on“ изрично и под 30 дни |
| Пробен период / демо / „ползвам го“ доказателство | пробните периоди са най-влиятелни за 74% (TrustRadius 2024); 96% са гледали explainer видео (Wyzowl, малка извадка) | **Частично** — екранни записи на free плановете са най-евтиното достоверно доказателство за соло оператор; днес няма нито един скрийншот в guide |
| Човек, не бранд | служителите са по-доверени от „безлични брандове“ (Edelman 2025); 59% от LinkedIn цитиранията в ChatGPT са лични постове | **Да** — оценките вече са с автор в seeds; сложи името на страницата и във видеото |

**На какво не вярват:** фалшиви ревюта — 73% смятат, че ги виждат (TrustRadius), 97% искат наказание (BrightLocal); AI-направен маркетинг — почти една трета по-малко вероятно да изберат бранд с AI реклами (eMarketer 2025 през Hootsuite); непроверени AI отговори — 94% проверяват; само вендорска информация — корелира със съжаление; хайп — най-гласуваният намерен коментар (182) нарича маркетьорите на автоматизация „grifters selling the dream“. Първично проучване специално за „sponsored“ баджове или списъци с 20 инструмента не се намери (ниска увереност, че съществува на B2B ниво); най-близкото е 7.9% доверие в „industry influencers“ (G2).

### 7.2 Къде да ги стигнем най-бързо без бюджет

| Ранг | Канал | Доказателство, че купувачите на бизнес софтуер откриват инструменти там | Скорост | Присъда |
|---|---|---|---|---|
| 1 | **Reddit** (r/smallbusiness 2.5M, r/Entrepreneur 5.3M, r/SaaS 794K +106%, r/CRM 57K +131%, r/Emailmarketing 123K, r/agency 99K, r/automation 230K, r/ProductivityApps 192K +108%) | купувачите буквално пишат „stay under $25“; Reddit е №1 в Perplexity (1 от 5 цитирания), топ 3 в ChatGPT, Gemini, AI Overviews (Contently); „…reddit“ е autocomplete към CRM, Mailchimp, Calendly, Zapier, Ahrefs заявките | дни–седмици | **Първи.** Отговаряй на нишки 1–30 с таблицата в коментара и линка на второ място. Три на седмица. |
| 2 | **YouTube дълго** | №2 сайт за проучване след Google (TrustRadius 2024); №1 цитиран домейн в AI Overviews, №2 в Google AI Mode и Gemini; най-силна корелация с AI видимост (0.737, Ahrefs 75K бранда) | 1–3 месеца | **Втори.** 7–12 мин. сравнения по таблицата на търсенето, екранни записи на free плановете. |
| 3 | **Answer engines** (ChatGPT, Perplexity, Google AI Mode) | 51% започват там (G2 2026); 45% от потребителите ползват AI за ревюта, от 6% (BrightLocal 2026) | седмици, през цитирания на 1, 2, 4 | Не канал, в който постваш — **последствие** от 1, 2 и 4. Реши политиката за обхождачите първо. |
| 4 | **LinkedIn (лични текстове)** | 14.3% от ChatGPT Search и 13.5% от Google AI Mode цитиранията; от ~№11 на №5 в ChatGPT (ноем. 2025–февр. 2026); 59% лични постове; №4 сайт за проучване (TrustRadius) | седмици | **Четвърти.** Дългото видео като текстов анализ с ценовата таблица; агенционните собственици са тук. |
| 5 | **SEO (собствени страници)** | Google №1 за фирми 1–250 (47%); affiliate сайтовете носят ~85% от органичния трафик на SaaS вендорите (wecantrack); AI Overviews се припокриват само 20–26% с органичния топ 10 | 4–6 месеца на нов домейн | Публикувай 25-те страници — те са това, към което Reddit коментарите и видеата сочат. |
| 6 | **YouTube Shorts / TikTok / Reels** | **няма доказателства**, че купувачи на бизнес софтуер откриват инструменти през късо видео; B2B SaaS TikTok бенчмарковете са за платени кампании | месеци | Нарязани от дългото; откриване на канала, не конверсия. |
| 7 | Indie Hackers / Product Hunt | не стигнати; r/indiehackers +76% подсказва, че общността расте в Reddit; PH е еднодневен скок за *продукт* | един ден | Само за builder-а като launch, след поправка на цените. |
| 8 | Facebook групи, Slack/Discord, newsletters | без първични данни; Facebook е №2 за четене на ревюта при потребители (BrightLocal) — локални услуги | седмици, нецитируемо | По-късно, за локалната персона. |
| 9 | X | в нито един доклад за купувачи | — | Пропусни. |

### 7.3 Най-бързият реалистичен път за ескалация от нула

1. **Седмици 1–4:** поправи базата на таксуване и добави „price read on“; публикувай елементи 1–5; отговори на 12 Reddit нишки (1–16 са живи и без отговор от UNSOLERO) с *таблицата в коментара*; две дълги YouTube видеа (Mailchimp free cut; HubSpot праг). Метрика: referral сесии от reddit.com и youtube.com; първо споменаване в Perplexity/ChatGPT отговор за „mailchimp alternatives at 1,000 subscribers“.
2. **Месеци 2–3:** едно дълго видео седмично, всяко със страница (елементи 6–12); LinkedIn текстова версия; Shorts от всяко; месечен пост „price changes this month“ — повтарящата се болка и естествен повод за връщане (Vertice 16.4%).
3. **Месеци 3–6:** числото за партньорски мениджъри е „N referral сесии/мес. от Reddit + YouTube и M цитирания“; кандидатствай отново при Thinkific, Webflow, Tidio (отказали 30 авг. 2026 за „малък трафик и аудитория“) и при HubSpot/Shopify през Impact.

**Трите най-рискови предположения в този път:**

1. *Че Reddit отговорите се превръщат в цитирания и трафик, а не в shadow-ban.* Акаунтът е нов, филтрите наказват link-heavy акаунти, данните за цитирания са за платформата, не за акаунт. Смекчаване: без линк в първото изречение, декларирай affiliate интереса, таблицата в коментара да е стойността.
2. *Че ценово нормализираната честност е различие, което аудиторията вижда.* Дефектът с базата на таксуване означава, че сайтът извършва точно греха, който критикува, на 25 продукта; един Reddit отговор, който го посочи, приключва историята. Поправи преди мащаб.
3. *Че печелещите и търсените страници са същите страници.* Осем от 15 категории не печелят; най-големите обеми (Zapier, Notion, Slack, GA) са в непечелещи категории. Ако трафикът дойде първо там, приходът не следва. Смекчаване: цялото ранно усилие в Reddit/YouTube към елементи 1–12, които стоят на живи програми.

---

## 8. Двадесет теми за говорене — от думите на аудиторията

> **Поправка, 2 септември 2026.** Два числа в тази таблица бяха грешни при
> писането: „Teachable Starter $39 annual“ (39 е месечната цена, 29 е
> годишната) и „$65“ за SE Ranking, което не отговаряше на нито една цена на
> вендора. Одитът на базата на таксуване същия ден провери всяка цена на
> страницата на вендора: SE Ranking Core е $129 месечно и $103.20 при годишно
> плащане. Всяка цена, която казваш на камера, се проверява в деня на записа —
> точно защото такива грешки минават незабелязано.


| # | Hook (фразата на аудиторията, за дословна употреба) | Платформа | Води към | Печели |
|---|---|---|---|---|
| 1 | „**$308 a month to send daily to 7K subs**“ — какво таксуват пет инструмента за точно този списък | YouTube дълго + Reddit | `/guides/mailchimp-alternatives` (+ калкулатор #7) | Kit, MailerLite, AC |
| 2 | „Mailchimp's free plan is now **250 contacts**. It was 2,000 in 2022.“ | Shorts/Reels, LinkedIn | #1 | MailerLite, Kit |
| 3 | „**Stay under that $25 mark** — CRM for a 5-person team“ | Reddit + YouTube | #5 / #3 | Zoho CRM, Bigin, Pipedrive |
| 4 | „**$26K/year on HubSpot with no sales team**“ | YouTube дълго, LinkedIn | #4 | Zoho, Pipedrive |
| 5 | „**$7K onboarding fee?!**“ — таксите, които вендорите не слагат на ценовата страница | Shorts, LinkedIn | #4 | Zoho, Pipedrive |
| 6 | „Is HubSpot worth it… **for small business**“ — присъда по персона, не по звезди | YouTube дълго | #21 | Zoho, Pipedrive, AC |
| 7 | „**Death by $29/month**“ — събрах стека на 12-членна агенция | YouTube дълго (stack-build) | `/guides/software-stack-small-agency` + #6 | Zoho suite, monday, Pipedrive, Cal.com |
| 8 | „**23 subscriptions, $4,100 a month**… five years ago $1,200“ — SaaS инфлацията е 16.4% | LinkedIn, YouTube | #6 + месечен пост | всички |
| 9 | „**Sick of paying per user pricing**“ — инструментите, които още таксуват плоско | Reddit, Shorts | категории с колона „pricing model“ (#25) | Zoho (Campaigns/Bookings), Kit |
| 10 | „The '**one tool to rule them all**' dream rarely survives contact with reality“ — all-in-one vs best-of-breed, ценено | YouTube дълго | builder (`all_in_one` vs `best_of_breed`) | Zoho One vs Pipedrive+MailerLite+Cal.com |
| 11 | „Zapier went **from $10 a month to over $750**“ | YouTube дълго, Reddit | #16 | никой (цитирания; n8n е в каталога) |
| 12 | „**A bunch of grifters are selling the dream**“ — 182 гласа за n8n хайпа; колко струва self-hosting реално? | YouTube дълго | #15 | доверие; Cal.com cloud |
| 13 | „**Calendly, but with more features and a less confusing interface**“ — Cal.com vs Calendly vs Zoho Bookings при 100 записвания/мес. | Shorts + YouTube | `/compare/calendly-vs-cal-com-vs-zoho-bookings` | Cal.com, Zoho Bookings |
| 14 | „calendly alternative **dsgvo**“ — EU-hosted stack за 3-членен бизнес | LinkedIn, Reddit (r/BuyFromEU 364K) | #14 | Zoho (EU DC), MailerLite, Cal.com |
| 15 | „Needs to integrate with **Cal.com, Mailchimp and Tawk**“ — CRM-ът, решен от интеграциите | Reddit, YouTube | #3 + compatibility view в builder-а | Zoho, Pipedrive, Cal.com |
| 16 | „**Which platform for a newsletter media company** — Substack, Beehiiv or Ghost?“ — break-even по месеци | YouTube дълго | #9 | Kit |
| 17 | „Teachable's own page shows **two different prices**“ — курс платформи при 100 студенти | YouTube, Shorts | #17 | Teachable |
| 18 | „**Free versions shrinking**“ — всеки free план в каталога в една датирана таблица | LinkedIn, Reddit, месечно | #24 | Zoho, MailerLite, Kit, Cal.com |
| 19 | „**Why does every subscription billing software suddenly cost $200/month once you start growing**“ — праговете на 10 инструмента | YouTube дълго | #7 + #4 | Zoho Books/Invoice, Kit, MailerLite |
| 20 | „Semrush just repackaged: **$139, $199, $299, $549**“ — какво трябва на агенция с 3 клиента при $129 месечно / $103.20 годишно (SE Ranking Core) | YouTube дълго, r/SEO | #20 | SE Ranking |

---

## 9. Източници

**Доклади за купувачи:** G2 Buyer Behavior Report 2025 (PDF, апр. 2025) · G2 „Half of B2B software buyers now start with AI chatbots“ (PR Newswire, 15 апр. 2026) · G2 Buyer Behavior 2026 (company.g2.com/news, 22 юли 2026) · TrustRadius 2026 B2B Buying Disconnect (PR Newswire, 15 юли 2026) · TrustRadius 2024 „The year of the brand crisis“ (10 юни 2024) · 6sense 2025 Buyer Experience Report (+ BusinessWire, 12 ноем. 2025) · Capterra Tech Trends (24 ян. 2025) · Capterra 2026 Software Buying Trends (7 окт. 2025; BusinessWire) · Capterra retail trends (18 ноем. 2024) · Zylo 2026 SaaS Management Index (29 ян. 2026) · Vertice SaaS Inflation Index · US Chamber „Empowering small business“ (18 авг. 2025) · Contently „Top sources LLMs cite“ (29 апр. 2026) · BrightLocal Local Consumer Review Survey 2026 · Hootsuite Social Trends 2026 · Wyzowl Video Marketing Statistics 2026 · wecantrack SaaS affiliate statistics · Semrush AI Overviews study (септ. 2024).
**Размери:** SBA Advocacy (30 юни 2025; 2025 US profile PDF) · Upwork freelance study 2025 · Backlinko Shopify stores (6 март 2026) · ICF Global Coaching Study 2025 · Thinkific about · Goldman Sachs creator economy (вторично) · GummySearch (r/smallbusiness, r/agency, r/SaaS, r/n8n, r/automation, r/BuyFromEU, r/hubspot, r/Newsletters; 1 септ. 2026) · GitHub (n8n, twenty, cal.com, listmonk) · AppSumo about · CostLoop SaaS waste report (юни 2026, вторично).
**Reddit нишки (адресите са в таблицата на раздел 3):** r/smallbusiness 1r71vm8, 1s8qtor, 1nfyeh9, 1or3gas, 1lzyu4b, 1t105u4, 1uvw2te · r/CRM 1qzuphb, 1t6jj1s, 1r257q6, 1kmp1f9 · r/Entrepreneur 1okknos, 1s5i577, 1rxvmve, 1r7vo0s · r/ITManagers 1owwt8j · r/sales 1iv34pb · r/hubspot 1iri8ud · r/marketing 1fck3qx · r/MailChimp 1rplv6f · r/Newsletters 1qvspsj · r/Emailmarketing 1htxd1f, 1mns7pw · r/ProductivityApps 1t3z8h9, 1jqhmug · r/BuyFromEU 1jwwzpl · r/selfhosted 1rt1sze · r/automation 1o1rl8h · r/n8n 1l5tl9a, 1pff2au · r/agency 1l2u472 · r/marketingagency 1q4km4l · r/nocode 1so47ef · r/productivity 1nrm7y7 · r/smallbusinessesowners 1u0zo9s · r/SaaS 1pttgli · r/CRMSoftware 1sv06lp.
**Цени и увеличения:** Wordtracker (4 заявки) · Google Autocomplete (suggestqueries.google.com, 2 септ. 2026) · Docket, TinyCommand, Encharge (HubSpot) · Audienceful (15 юли 2026), ActiveCampaign blog, CampaignHQ, PriceTimeline (Mailchimp) · TechCrunch (3 септ. 2024), SaaSPricePulse, Subkept (Canva) · PricePulse, UserJot (Notion) · Siit, StartupOwl (Zapier) · PricePulse, Zenventory, Shopify Community (Shopify) · Carly (Calendly) · Google Workspace blog (16 ян. 2025) · EmailOctopus (MailerLite) · Brevo blog (Klaviyo) · semrush.com/prices (2 септ. 2026) · ACTGSYS Reddit CRM survey (17 ноем. 2025) · european-alternatives.eu · factoryjet SEO timeline 2026 · theadranker SaaS TikTok benchmarks.
**Репо:** `backend/seeds/*.sql` (типове бизнес, предпочитания, категории), `docs/GROWTH_PLAYBOOK_BG.md` §2, §6, `docs/AFFILIATE_PROGRAMS.md`, `backend/.../commercial_boundary_test.go`; живи страници `/guides/mailchimp-alternatives`, `/compare/zoho-crm-vs-hubspot`, `/guides/how-to-choose-business-software` (2 септ. 2026); `GET /api/catalog/products/{slug}/offers` за всичките 53 продукта.
