# Първото видео — микротаскове

**Видеото:** Short A от [02_SOCIAL_GROWTH_BG.md](./02_SOCIAL_GROWTH_BG.md), раздел 10 — **„Mailchimp's free plan is now 250 contacts“**, 38–42 секунди, 1080×1920, запис на екрана + глас + вградени надписи. Публикува се в YouTube Shorts, TikTok и Instagram Reels; води към `https://unsolero.com/guides/mailchimp-alternatives`.
**Защо това видео първо:** страницата има 4 от 5 печелещи продукта (проверено в API-то на 2 септ. 2026); „mailchimp alternatives“ е заявката с най-голям измерен обем в цялата таблица на търсенето (4,300/мес.); спусъкът е реален и датиран (Mailchimp сви free плана на 250 контакта от 17 февр. 2026); hook-ът е доказан модел („Stop paying for X“ — 4.6M и 8.6M при Kevin Stratvert); 40 секунди упражняват цялата производствена линия с минимален риск, преди първото дълго видео.
**Точните менюта и бутони** за всяка стъпка са в [03_PRODUCTION_MANUAL_BG.md](./03_PRODUCTION_MANUAL_BG.md); тук се сочи раздел с „→ Ръководство X.Y“.
**Общо време първия път:** ≈ 8–9 часа, разпределени в 2–3 дни. Второто видео: ≈ 3 часа.

Всеки микротаск има: какво, как (стъпки), колко време, и **„готово, когда“** — критерий, който се проверява, не усещане.

---

## Фаза 0 — Веднъж, преди всичко (≈ 3 ч., може ден по-рано)

### M01 · Акаунтите (правиш ги ти, ръчно)
Едно потребителско име: `unsolero`. Едно лого: тийл `#14605a` буква/марка върху `#f4f5f7`, 800×800 PNG. Едно био: `Build the right software stack. Every price dated, every ranking commission-proof. Free tool ↓`.
1. **YouTube:** studio.youtube.com → канал `UNSOLERO` → **Customization** → **Basic info** (описание с думите „software stack“, „SaaS comparison“, „CRM“, „email marketing“) → **Links**: 1) `https://unsolero.com/build?utm_source=youtube&utm_medium=channel`, 2) `https://unsolero.com/guides/mailchimp-alternatives?utm_source=youtube&utm_medium=channel` → **Branding**: лого + банер 2048×1152. **Settings** → **Channel** → **Feature eligibility** → **Advanced features** → телефонна верификация (нужна за „Related video“ по-късно).
2. **TikTok** (телефон): регистрация → **Profile** → **Edit profile** (лого, био) → **Settings and privacy** → **Account** → **Switch to Business Account** → категория Software/Technology → **Edit profile** → **Website** = `https://unsolero.com/guides/mailchimp-alternatives?utm_source=tiktok&utm_medium=bio`. → Ръководство 5.1.
3. **Instagram** (телефон): регистрация → **Settings** → **Account type and tools** → **Switch to professional account** → **Business** → **Edit profile** → **Links** → същия URL с `utm_source=instagram&utm_medium=bio`. → Ръководство 5.2.
4. **LinkedIn:** личният профил → **Edit** → **Website** → `https://unsolero.com/links?utm_source=linkedin` (или guide URL, докато `/links` не съществува).
**Време:** 60–90 мин. **Готово, когда:** трите профила показват едно лого, едно име, едно био, и полето Website в TikTok е попълнено и видимо от друг телефон.

### M02 · Машината (терминал)
1. Изпълни блока от → Ръководство 0 (`dnf install python3-pip pipx espeak-ng audacity`; проверка на NVIDIA драйвер срещу Flatpak GL runtime; `mkdir -p ~/Videos/{obs,vo,export,cards}`).
2. `flatpak run com.obsproject.Studio` → **Settings** → **Output** → **Recording** → отвори **Video Encoder**: запиши си дали има **NVIDIA NVENC H.264**. Ако няма — решението е **x264, CRF 18, veryfast** (Ръководство 1.3), не губи време да го оправяш днес.
3. `flatpak run org.kde.kdenlive` → **Settings** → **Configure Kdenlive…** → **Plugins** → **Speech to text** → **Whisper** → **Check configuration** → **Install missing dependencies** → модел `small.en` → **Language** English. Изтеглянето е ~2 GB; пусни го и продължи с M03. Ако откаже — ръчният блок в → Ръководство 2.5.
**Време:** 30 мин. активно + изчакване. **Готово, когда:** OBS се отваря, енкодерът е избран; Kdenlive показва `small.en` в **Model**.

### M03 · Целевата страница
1. Отвори `https://unsolero.com/guides/mailchimp-alternatives` в Chrome. Провери, че секциите „The options, all at 1,000 subscribers“ и „Which one, in one line each“ се виждат — те са това, което ще снимаш.
2. Направи UTM линковете с **Campaign URL Builder** (→ Ръководство 6.2) и ги запиши в `~/Videos/short-a-notes.md`:
   - YouTube: `?utm_source=youtube&utm_medium=shorts&utm_campaign=2026-09-mailchimp-250`
   - TikTok: `?utm_source=tiktok&utm_medium=bio&utm_campaign=2026-09-mailchimp-250`
   - Instagram: `?utm_source=instagram&utm_medium=bio&utm_campaign=2026-09-mailchimp-250`
3. **По желание, но силно препоръчано (Проучване 1, #6):** страницата има само един affiliate CTA при четири печелещи продукта. Добави `cta` блок след всеки от четирите (Zoho Campaigns, ActiveCampaign, MailerLite, Kit) в съдържанието на guide-а, за да има къде да кликне зрителят, дошъл от видеото. Ако не го направиш преди видеото, направи го преди второто.
4. **Измерване (Проучване 1, #8):** без Umami или Cloudflare Web Analytics кликовете от UTM няма как да се прочетат. За първото видео може да се пропусне; за второто — не.
**Време:** 20 мин. (+2–3 ч. за CTA блоковете). **Готово, когда:** трите UTM адреса са в бележките и отварят страницата.

---

## Фаза 1 — Съдържанието (денят на записа, ≈ 1.5 ч.)

### M04 · Цените, прочетени днес
Правилото на сайта и на видеото: **никаква цена, непроверена същия ден.** Отвори шестте страници и попълни таблицата в `~/Videos/short-a-notes.md`:

| Инструмент | Страница | План | Цена при 1,000 абонати/контакти | База (месечно / годишно) | Час на прочитане |
|---|---|---|---|---|---|
| Mailchimp | mailchimp.com/pricing → плъзгач на 1,000 контакта | Essentials или Standard (запиши кой) | $ | | |
| Mailchimp Free | същата → колона Free | Free | 250 контакта / 500 изпращания (потвърди) | — | |
| Zoho Campaigns | zoho.com/campaigns/pricing → 1,000 контакта | Standard | $5.25 очаквано | годишно (очаквано) | |
| Brevo | brevo.com/pricing | Starter | $9 очаквано | месечно; **по изпращания** | |
| ActiveCampaign | activecampaign.com/pricing → 1,000 контакта | Starter | $15 очаквано | годишно (очаквано) | |
| MailerLite | mailerlite.com/pricing → 1,000 абонати | (планът, който каталогът нарича Comfort) | $19 очаквано | месечно (потвърди) | |
| Kit | kit.com/pricing → 1,000 абонати | Creator | $39 очаквано | месечно (потвърди) | |

Ако число се различава от страницата на UNSOLERO → запиши го в отделен ред „за поправка в guide-а“ (M29) и **ползвай прочетеното днес във видеото**. Не кликай affiliate бутоните на UNSOLERO, докато проверяваш — отвори вендорските страници директно.
**Време:** 25 мин. **Готово, когда:** седем реда с цена, база и час; нула „очаквано“ остава.

### M05 · Скриптът, финализиран
1. Копирай таблицата на Short A (→ 02_SOCIAL_GROWTH_BG.md, раздел 10) в `~/Videos/vo/short-a.txt` само с колоната „Точен глас“, ред по ред.
2. Замени `[цена, прочетена днес]` с числото от M04 и `[дата]` с днешната дата в формат „2 September 2026“.
3. Прочети на глас с хронометър на телефона, нормално темпо, без бързане. Целта е **36–41 секунди**. Ако е над 42: махни „— pricey, but built for publishers who sell“ и остави „Kit: thirty-nine.“; после махни „and what we're not telling you“ от CTA-то. Ако е под 34: не добавяй нищо — бавно темпо е по-добро от пълнеж.
4. Отбележи с `|` местата, където ще има рязка смяна на екрана (края на всеки ред от таблицата).
**Време:** 20 мин. **Готово, когда:** текстът е записан, три прочитания на глас са в интервала 36–41 сек.

### M06 · Скрийншотите за картите (по желание)
`Print Screen` → GNOME запазва в `~/Pictures/Screenshots`. Нужни: (1) Mailchimp Free колоната с „250 contacts“; (2) стара цифра 2,000 — от статия за ценовата история (Audienceful / CampaignHQ), със източника видим. Премести в `~/Videos/cards`.
**Време:** 10 мин. **Готово, когда:** два PNG файла.

---

## Фаза 2 — Гласът (≈ 30 мин.)

### M07 · Запис на гласа отделно (препоръчано) → Ръководство 1.6, параграф Audacity
1. Затвори прозорци, изключи вентилатора, ако може (лаптопът на ток, но не под товар); телефон на тихо; вратата затворена.
2. **Audacity** → **Audio Setup** → **Recording Device** → вграденият микрофон → натисни **Record** (червеното копче или `R`) → 2 секунди тишина → прочети целия скрипт → 2 секунди тишина → **Stop**. Направи **три пълни дубъла** в един запис, с 3 сек. пауза между тях.
3. Избери най-равния дубъл (без запъване, без „ъъ“); изтрий останалите (`Ctrl+I` за разделяне на клип, `Delete`).
4. Маркирай първите 2 сек. тишина → **Effect** → **Noise Removal and Repair** → **Noise Reduction** → **Get Noise Profile** → `Ctrl+A` → същият ефект → 12 dB → **Apply**.
5. `Ctrl+A` → **Effect** → **Volume and Compression** → **Loudness Normalization** → `-16 LUFS` → **Apply**.
6. **File** → **Export Audio** → WAV, 48000 Hz, 16-bit → `~/Videos/vo/short-a.wav`.
**Алтернатива с AI глас (Kokoro):** блокът в → Ръководство 4.1 с `~/tts/script.txt` = скриптът; глас `af_heart` или `am_michael`; после TikTok/YouTube AI декларацията се включва при качване (M22, M21).
**Време:** 25 мин. **Готово, когда:** `short-a.wav` е 36–41 сек., пиковете под −1 dB, без шум в паузите.

---

## Фаза 3 — Екранът (≈ 45 мин.)

### M08 · Chrome в 9:16 → Ръководство 1.4
1. GNOME: горен панел → **Do Not Disturb** включено. Chrome: затвори всички табове освен тези, които ще снимаш. Изключи разширения с иконки (или нов профил на Chrome без разширения).
2. Отвори табове в този ред: (1) `unsolero.com/guides/mailchimp-alternatives`, (2) `mailchimp.com/pricing`, (3) `unsolero.com`.
3. `F11` → `Ctrl+Shift+I` → `Ctrl+Shift+M` → **Dimensions: Responsive** → `576` × `1024` → **⋮** → **Dock side** → **Undock into separate window** → премести DevTools прозореца настрани.
4. Върни фокуса на страницата, `Ctrl` + `+` до **125%** (ако текстът на цените е дребен — 150%).
**Готово, когда:** страницата стои вертикално, без DevTools в кадъра, цените се четат на око от метър.

### M09 · OBS профил, сцена, пробен запис → Ръководство 1.1, 1.2, 1.3, 1.5, 1.8
1. **Profile** → **New** → `Vertical 9-16`; **Scene Collection** → **New** → `Unsolero vertical`.
2. **Settings** → **Video**: `1080x1920` / `1080x1920` / Lanczos / 30. **Output** → Advanced → **Recording**: `~/Videos/obs`, **Hybrid MP4**, енкодерът от M02, CQ/CRF 18, AAC, Track 1 и 2.
3. **Sources** → **+** → **Window Capture (PipeWire)** → `Chrome unsolero` → в диалога **Share Screen** таб **Window** → прозорецът на страницата → **Remember this selection** → **Share**.
4. `Alt` + влачене на ръбовете до областта 576×1024 → десен клик → **Transform** → **Fit to screen** (`Ctrl+F`) → **Scale Filtering** → **Lanczos**.
5. **Settings** → **Audio** → **Desktop Audio: Disabled**. Микрофонът може да остане (scratch пътечка) — гласът е от M07.
6. **Пробен запис 20 сек.** → отвори файла → остър текст, 30 fps, без DevTools рамка. **View** → **Stats**: 0 пропуснати кадри.
**Готово, когда:** пробният файл е чист. Не продължавай, ако не е.

### M10 · Снимане на кадрите — по един файл на кадър
Записвай всеки кадър като отделен файл (**Start Recording** → кадър → **Stop Recording**, `Alt+Tab` между OBS и Chrome; → Ръководство 1.7). Правила за мишката: **влиза бавно от долния десен ъгъл, спира върху целта, стои 2 секунди, не трепери; никакви кликове върху „View at …“ или „Affiliate link“ бутони** — записан клик върху собствен affiliate линк е реален клик и условията на програмите забраняват изкуствени кликове. Само скрол и hover. Пауза 1 сек. преди и след всяко движение — дава въздух за рязане.

| Кадър | Таб | Какво прави мишката | Дължина на записа | Служи за секунди |
|---|---|---|---|---|
| S1 | Mailchimp pricing | скрол до колоните с планове; мишката спира върху **Free** и числото контакти; стои | 8 сек. | 0–2 |
| S2 | Mailchimp pricing | плъзгачът за контакти на **1,000** (ако има; иначе падащото меню); мишката спира върху цената на плана | 8 сек. | 2–6 и 22–27 |
| S3 | UNSOLERO guide | скрол от началото до заглавието „The options, all at 1,000 subscribers“; спри | 10 сек. | 6–10 |
| S4 | UNSOLERO guide | бавен скрол през петте реда (Zoho → Brevo → ActiveCampaign → MailerLite → Kit); мишката следва реда, спира на цената | 20 сек. | 6–22 |
| S5 | UNSOLERO guide | скрол до „Which one, in one line each“; бавен скрол през петте реда | 12 сек. | 27–34 |
| S6 | UNSOLERO начална | hero-то с „Build the right software stack.“; мишката спира върху бутона „Build My Setup“ **без да кликва** | 8 сек. | 34–40 фон под CTA картата |

Ако ценовата страница на Mailchimp е геолокирана (EUR): снимай я така, но кажи в описанието, че цените са в EUR, или ползвай скрийншота от M06 за S1.
**Време:** 30 мин. **Готово, когда:** шест файла в `~/Videos/obs`, всеки прегледан веднъж, без нотификации в кадъра.

---

## Фаза 4 — Монтажът (≈ 1.5 ч. първия път)

### M11 · Проект и подредба → Ръководство 2.1, 2.2
1. **File** → **New** → **Vertical HD 30 fps** → Video tracks 2, Audio tracks 3 → **OK**. Запази веднага: `~/Videos/kdenlive/short-a.kdenlive`.
2. **Project Bin** → **Add Clip or Folder** → шестте MP4 + `short-a.wav`.
3. Влачи `short-a.wav` на **A1**, в началото. Това е гръбнакът — картината се подрежда по гласа.
4. Слушай гласа и с `Shift+S`-подобни маркери (или просто **Guides**: десен клик на линийката → **Add Guide** при всяка смяна на изречение). Влачи S1 на **V1** от 0; режи (`Shift+R`) при края на първото изречение; следва S2, S3, S4 (раздели го на четири части — по една за всеки ценов ред), S2 отново, S5, S6. Всеки клип се разтяга/реже до дължината на съответния ред от таблицата на Short A.
5. Заглуши OBS звука: десен клик върху видео клип → **Split Audio** → mute на пътечката (или изтрий аудио клиповете).
**Готово, когда:** timeline-ът е без дупки от 0 до края на гласа, всяка смяна на кадър пада на края на изречение.

### M12 · Zoom-ове → Ръководство 2.3
Добави **Transform** и keyframe **Size** 100 → 140 за 12–15 кадъра върху: „250 contacts“ (S1), цената на Mailchimp (S2 — два пъти), всеки от петте ценови реда (S4), „Which one“ реда (S5). Центрирай рамката върху числото. Не зумирай CTA картата.
**Готово, когда:** 8–9 zoom-а, всеки започва до 0.3 сек. след началото на изречението, което го обяснява.

### M13 · Надписите на екрана (title clips) → Ръководство 2.4
Inter Bold (или Montserrat Bold), 84 px, бяло, черен outline 6 px, сянка 4/4/8, центрирано, **вътре в вътрешната safe рамка**, горна трета (y ≈ 350–600). Един title clip за всеки ред от колоната „Надпис на екрана“ на Short A — 10 карти. Постави ги на **V2** с точно същите времена като редовете. Финалната CTA карта (34–40): фон `#14605a`, URL в бяло 64 px, `LINK IN BIO` 84 px, `Comment your list size ↓` 56 px, UNSOLERO марката малка долу (y ≈ 1450, вътре в safe зоната).
**Готово, когда:** всяка карта се появява със своето изречение и нито една буква не е в долните 420 px или десните 200 px.

### M14 · Субтитри → Ръководство 2.5
1. **Edit Subtitle Tool** → **Speech recognition** → **Automatic Subtitling** → Model `small.en`, Language English, **Maximum character per line** 28 → **Process**.
2. Прочети всеки субтитър срещу `short-a.txt`; поправи числата (Whisper често пише „5.25“ като „525“) и имената (Brevo, MailerLite, Kit, Zoho).
3. **Manage Subtitles** → **Style**: bold sans 66 px, бяло, outline 5 черно, долу-център, вертикален margin така, че редът да е около **y ≈ 1350**. Провери, че субтитрите **не се пресичат** с title картите (те са в горната трета).
4. **Export Subtitle File** → SRT → `~/Videos/export/short-a.srt` (за YouTube captions).
**Готово, когда:** всяка дума е вярна, два реда максимум, нищо в долните 420 px.

### M15 · Звук → Ръководство 2.6
Гласът: **Normalize (2 Pass)** → −16 LUFS → **Analyse to Apply Effect**. Музика (по избор): Pixabay Music → инструментал, спокоен, без вокал → **Download** → запиши URL на трака в бележките → **A2** на −22 dB, fade out последната секунда. Master пик ≤ −1 dBTP в **Audio Mixer**.
**Готово, когда:** гласът е ясен над музиката при слушане на телефонен високоговорител.

### M16 · Рендер → Ръководство 2.7
**File** → **Render** → **MP4-H264/AAC** → Quality ~75% → **Embed subtitles instead of burning them in** изключено → **Output file** `~/Videos/export/unsolero-mailchimp-250-2026-09-XX-v1.mp4` → **Render to File**.
**Готово, когда:** файлът е 1080×1920, 30 fps, 36–42 сек., под 60 MB.

### M17 · Контрол на качеството — на телефон
Прати файла на телефона (KDE Connect / Telegram „Saved Messages“ / USB). Гледай го **три пъти**:
- [ ] Първите 2 секунди: „250 CONTACTS.“ се чете, без да спираш видеото.
- [ ] Всяка цена има база (yearly / per send / monthly) на екрана или в гласа.
- [ ] „Prices read [дата]“ се вижда.
- [ ] Дисклоузърът е казан („Some links pay us a commission; it changes nothing about the order“).
- [ ] Нито един текст в долните 420 px и десните 200 px (постави пръст върху мястото на TikTok иконите).
- [ ] Няма кадър, в който мишката кликва affiliate бутон.
- [ ] Няма нотификация, DevTools рамка, друг таб.
- [ ] Звукът е равен, без изкривяване, музиката не покрива глас.
- [ ] Дължината ≤ 42 сек.
Ако нещо пада — поправи и рендерирай `v2`. Не публикувай `v1` с известен дефект „за да тръгне“.

---

## Фаза 5 — Публикуване (≈ 45 мин.)

### M18 · Cover
Кадърът от 0–2 сек. с надписа „250 CONTACTS.“ Ако платформата не дава да избереш точния кадър — Canva → **Custom size** 1080×1920 → скрийншотът от M06 + същия надпис → **Download PNG** (→ Ръководство 4.3).

### M19 · Текстовете (в бележките, готови за copy-paste)
- **Заглавие (YouTube, ≤100 знака):** `Mailchimp's free plan is now 250 contacts. What 1,000 subscribers costs on 5 tools (Sept 2026)`
- **Описание** — шаблонът от Short A в 02_SOCIAL_GROWTH_BG.md, с UTM линка за съответната платформа (M03) и датата.
- **Хаштагове:** YouTube `#emailmarketing #mailchimp #smallbusiness #saas`; TikTok `#emailmarketing #mailchimp #smallbusinesstips #saastools #creatortools`; Instagram 3–5 (`#emailmarketing #mailchimp #smallbusiness #newsletter`).
- **Закачен коментар (YouTube):** линкът към guide-а + „Prices read on [дата]. Ask me your list size.“

### M20 · YouTube Shorts → Ръководство 5.3
studio.youtube.com → **Create** → **Upload videos** → файлът → **Title**, **Description** → **Thumbnail**: избери кадъра 0–2 сек. (или custom, ако е налично) → **Audience: No, it's not made for kids** → **Show more** → ако гласът е Kokoro: **AI use / Altered content → Yes** → **Next** → **Video elements** → **Add subtitles** → `short-a.srt` → **Next** → **Checks** → **Next** → **Visibility: Public** → **Publish**. (Първото видео се публикува веднага — данните са по-важни от часа.) След публикуване: закачи коментара с линка.
**Готово, когда:** видеото е публично, показва се като Short (вертикално, ≤3 мин.), субтитрите са качени.

### M21 · TikTok → Ръководство 5.1
tiktok.com/tiktokstudio/upload → **Select video** → **Description** (описанието + хаштаговете; ключовата фраза „Mailchimp alternatives“ първа) → **Cover** → **Edit cover** → кадърът 0–2 сек. → **Who can watch: Everyone** → **Allow: Comment, Duet, Stitch** → **Content disclosure and ads** → **Disclose commercial content** → **Your brand** ✔ (affiliate = промоционално съдържание; това е задължително) → **AI-generated content** ✔ само ако гласът е Kokoro → **Run a copyright check** → **Post**.
**Готово, когда:** клипът е публикуван с етикет „Promotional content“ и линкът в био е жив.

### M22 · Instagram Reel → Ръководство 5.2
**От телефона**, за да имаш Trial: приложението → **+** → **Reel** → файлът → **Next** → **Edit cover** → кадърът 0–2 сек. → caption (описанието с 3–5 хаштага; UTM линкът не е кликаем в caption — стои в био) → ако има превключвател **Trial** и акаунтът отговаря на условията — включи го (показва се първо на не-последователи; след 24 ч. **Share with everyone**); иначе → **Share**. **Не** добавяй Instagram аудио отгоре (музиката е вградена и лицензирана).
**Готово, когда:** Reel-ът е публикуван; в **Edit profile** → **Links** стои UTM адресът.

### M23 · LinkedIn (по избор за първото, задължително от второто)
**Start a post** → иконата за медия → същият MP4 (вертикалното работи в LinkedIn) → текст: hook ред + петредова таблица (инструмент — цена — база) + въпрос „What does your list cost you right now?“ → линкът в първия коментар с `utm_source=linkedin` → **Post**.

### M24 · Един Reddit отговор (не е част от видеото, но е същата седмица)
r/MailChimp „Alternatives to Mailchimp?“ (10 март 2026) или r/Newsletters „What's cheaper than Mailchimp?“ — отговор по шаблона от 02_SOCIAL_GROWTH_BG.md, раздел 10: числата в първите два реда, таблица, линкът на трето място с дисклоузър. Само ако акаунтът има поне две седмици история без линкове; иначе — отговор без линк.

---

## Фаза 6 — 48 часа след публикуването

### M25 · Коментарите
Нотификациите на трите приложения включени на телефона. Всеки коментар — отговор в първия час (Buffer: +21% IG, +30% LinkedIn; van der Blom: +35% видимост). На „what about X?“ — отговори с цена и база, ако ги знаеш проверени; иначе „I'll read it today and reply“.

### M26 · Метриките на 24 ч. и на 48 ч. → Ръководство 6.1
Запиши в таблицата (Google Sheets / LibreOffice) реда за това видео: дата · платформа · hook текст · дължина · **TikTok: Retention rate при 3 сек., Watched full video %** · **Instagram: Skip rate, Average watch time** · **YouTube: Viewed vs. swiped away, Average percentage viewed** · shares · saves · profile visits · кликове (ако има Umami/GA4).

### M27 · Решението по правилото → Ръководство 6.3
- TikTok retention при 3 сек. под 60% / IG skip rate над 45% / YouTube viewed под 40% → **проблемът е първата 1.5 секунда**: пренапиши hook-а (варианти: „Mailchimp just cut its free plan by 87%.“ / „Stop paying Mailchimp for a list you barely email.“ / „$308 a month to email 7,000 people. Here's the fix.“), пресни само S1 и първата карта, рендерирай `v2`, публикувай като **ново** видео (не заменяй — Sabrina публикува вариациите отделно).
- Hook-ът минава, довършването е под 50% → съкрати с 20% (махни Kit детайла и „what we are not telling you“).
- Shares/saves добри, кликове нула → провери линка в био, UTM-а и дали страницата се отваря на телефон.
- Всичко над праговете → **не пипай нищо**; премини към Short B (CRM под $25) по същия процес; второто видео трябва да отнеме ≈ 3 часа.

### M28 · Поправката на страницата
Ако M04 е показал разлика между вендорската цена и страницата на UNSOLERO — поправи seed-а/съдържанието и датата „Read from the vendor on“ същата седмица. Видеото сочи към страницата; страницата не може да противоречи на видеото.

---

## Проверка накрая — всичките „готово, когда“

| # | Микротаск | Готово |
|---|---|---|
| M01 | три профила, едно лого, Website в TikTok | ☐ |
| M02 | OBS енкодер избран; Kdenlive Whisper `small.en` | ☐ |
| M03 | три UTM адреса в бележките | ☐ |
| M04 | седем цени с база и час, прочетени днес | ☐ |
| M05 | скрипт с числа и дата; 36–41 сек. на глас | ☐ |
| M06 | два PNG за карти | ☐ |
| M07 | `short-a.wav`, −16 LUFS, чисти паузи | ☐ |
| M08 | Chrome 576×1024, DevTools отделно, 125% | ☐ |
| M09 | пробен запис чист, 0 пропуснати кадри | ☐ |
| M10 | шест кадъра, без кликове на affiliate бутони | ☐ |
| M11 | timeline без дупки, смени на края на изречения | ☐ |
| M12 | 8–9 zoom-а | ☐ |
| M13 | 10 карти + CTA, всичко в safe зоната | ☐ |
| M14 | субтитри верни, SRT експортиран | ☐ |
| M15 | глас −16 LUFS, музика −22 dB, пик ≤ −1 | ☐ |
| M16 | MP4 1080×1920, 36–42 сек. | ☐ |
| M17 | 9 точки QC на телефон | ☐ |
| M18 | cover с „250 CONTACTS.“ | ☐ |
| M19 | заглавие, 3 описания, хаштагове, закачен коментар | ☐ |
| M20 | YouTube Short публичен със субтитри | ☐ |
| M21 | TikTok с „Promotional content“ | ☐ |
| M22 | Instagram Reel (Trial, ако е налично) | ☐ |
| M23 | LinkedIn (по избор) | ☐ |
| M24 | един Reddit отговор (ако акаунтът има история) | ☐ |
| M25 | всеки коментар отговорен в първия час | ☐ |
| M26 | редът в таблицата на 24 и 48 ч. | ☐ |
| M27 | решение по правилото, записано | ☐ |
| M28 | страницата поправена, ако е трябвало | ☐ |
