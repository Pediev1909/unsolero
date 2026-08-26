# Кратък и точен план за faceless видеа — UNSOLERO

**Версия:** 26 август 2026 г.  
**Език на видеата:** английски  
**Формат:** 1080×1920, 9:16, 30 fps  
**Дължина:** 25–40 секунди  
**Качване:** един и същ чист MP4 в YouTube Shorts, TikTok и Instagram Reels

Този документ съдържа само практическата работа: какво натискаш, какво записваш,
какво трябва да се вижда и какво казваш.

---

## 1. Най-лесният работещ процес

Използвай този процес за основните видеа:

1. Записваш само сайта с OBS — без лице и без лични данни.
2. Записваш гласа отделно или използваш AI voice.
3. В Kdenlive изрязваш паузите, правиш вертикален кадър и добавяш текста.
4. Експортираш един MP4.
5. Качваш същия файл в трите платформи.

Не търси „анимирани карти“. Картата е обикновен текст върху правоъгълен фон,
който правиш в Kdenlive чрез `Add Title Clip`. Тя не се записва с OBS.

---

## 2. Подготовка преди всеки запис

Направи точно това:

1. Създай папка, например `Videos/2026-08-Video-01`.
2. Отвори Chrome само с нужните страници.
3. Затвори Gmail, банкиране, affiliate dashboards и всички лични tabs.
4. Излез от admin профила на UNSOLERO.
5. Отвори `https://unsolero.com` в Incognito прозорец.
6. Натисни `Ctrl+0`, след това `Ctrl++` веднъж или два пъти, докато основният
   текст стане голям и четим.
7. Направи целия builder веднъж без запис. Записвай само ако резултатът се
   зарежда правилно.
8. Изключи известията на компютъра.
9. Постави курсора до бутона, който ще натиснеш първи.

Не трябва да се виждат:

- имейл адрес;
- име на акаунт;
- IBAN, affiliate приходи или dashboards;
- browser bookmarks с лична информация;
- пароли, API ключове или admin страници.

---

## 3. Настройване на OBS — прави се само веднъж

1. Отвори OBS.
2. Натисни `Settings` → `Video`.
3. На `Base (Canvas) Resolution` избери `1920x1080`.
4. На `Output (Scaled) Resolution` избери `1920x1080`.
5. На `Common FPS Values` избери `30`.
6. Натисни `Apply`.
7. Отвори `Settings` → `Output`.
8. На `Output Mode` избери `Simple`.
9. На `Recording Quality` избери `High Quality, Medium File Size`.
10. На `Recording Format` избери `MKV`.
11. Избери папката за запис и натисни `OK`.
12. В `Scenes` натисни `+` → напиши `UNSOLERO Browser` → `OK`.
13. В `Sources` натисни `+` → `Window Capture` → `OK`.
14. Избери Incognito прозореца с UNSOLERO.
15. Натисни с десен бутон върху preview → `Transform` → `Fit to Screen`.
16. В `Audio Mixer` изключи `Desktop Audio`, за да не запишеш известия.
17. Ако ще говориш по време на записа, остави `Mic/Aux` включен. Ако ще
    използваш AI voice, изключи и `Mic/Aux`.

За запис:

1. Натисни `Start Recording`.
2. Изчакай две секунди без движение.
3. Изпълни действията от сценария бавно.
4. След последния кадър изчакай още две секунди.
5. Натисни `Stop Recording`.
6. В OBS натисни `File` → `Remux Recordings`.
7. Избери MKV файла → `Remux`. Полученият MP4 се използва в Kdenlive.

OBS препоръчва MKV, защото незавършен запис не поврежда целия файл; след това
програмата може да го remux-не до MP4.

---

## 4. Видео №1 — реална демонстрация на UNSOLERO

**Цел:** зрителят вижда, че платформата действително изгражда software stack.  
**Дължина:** 35–39 секунди.  
**Запис:** сайтът с OBS.  
**Глас:** прочети дословно английския текст.

| Време      | Какво натискаш                                                   | Какво трябва да се вижда              | Какво казваш                                                                      | Текст върху видеото           |
| ---------- | ---------------------------------------------------------------- | ------------------------------------- | --------------------------------------------------------------------------------- | ----------------------------- |
| 0–3 сек.   | Не движиш мишката                                                | Началната страница и `Build My Setup` | “Most software lists start with products. This one starts with your constraints.” | `STOP STARTING WITH PRODUCTS` |
| 3–6 сек.   | Натискаш `Build My Setup`                                        | Отваря се Question 1                  | “I’ll build a stack for a small client-services business.”                        | `REAL EXAMPLE`                |
| 6–10 сек.  | Избираш `Run a client services business` → `Next`                | Избраният Goal                        | “First, define what the business actually does.”                                  | `1. GOAL`                     |
| 10–14 сек. | Избираш `No dedicated admin` → `Next`                            | Избраният Team вариант                | “Nobody is paid to maintain complicated software.”                                | `2. TEAM CAPACITY`            |
| 14–18 сек. | В `Exact budget in dollars` пишеш `120` → `Next`                 | Бюджет `$120`                         | “The complete stack has a one-hundred-and-twenty-dollar monthly ceiling.”         | `3. TOTAL BUDGET: $120`       |
| 18–22 сек. | Избираш `Team chat` → `Next`                                     | Team chat е маркиран                  | “Team chat already exists, so recommending another one would be waste.”           | `4. KEEP WHAT WORKS`          |
| 22–26 сек. | Избираш `The strongest tool per job` → `Next`                    | Избраната preference карта            | “This team wants the strongest tool for each job.”                                | `5. PREFERENCE`               |
| 26–30 сек. | Избираш `Best value` и `Connects to my stack` → `Build my setup` | Двете тъмни избрани карти             | “Value and integrations matter most. Now the engine can rank the fit.”            | `6. PRIORITIES`               |
| 30–35 сек. | Скролваш бавно през първите резултати                            | Total, match score и първите продукти | “The result stays inside the budget and explains every choice.”                   | `EXPLAINED RESULT`            |
| 35–39 сек. | Спираш движението                                                | Стабилен кадър на първия резултат     | “Build your own brief at unsolero dot com.”                                       | `TRY IT AT UNSOLERO.COM`      |

Правила за този запис:

- Направи действията по-бавно от гласа. После съкрати паузите в Kdenlive.
- Не чакай loading екрана във финалното видео — изрежи го до 0.3–0.8 секунди.
- Ако излезе грешка, не използвай записа.
- Не приближавай affiliate бутона и не отваряй външен merchant линк.

---

## 5. Видео №2 — изцяло без снимане на сайт

**Цел:** образователно видео, което може да бъде направено изцяло от AI или
само с Kdenlive title clips.  
**Дължина:** 25 секунди.

| Време      | Какво има на екрана                        | Какво казваш                                                  |
| ---------- | ------------------------------------------ | ------------------------------------------------------------- |
| 0–3 сек.   | `BEST EMAIL PLATFORM?`                     | “Stop asking which email platform is best overall.”           |
| 3–6 сек.   | `THE ANSWER CHANGES WITH YOUR CONSTRAINTS` | “The answer changes when your constraints change.”            |
| 6–10 сек.  | `1. SUBSCRIBERS`                           | “First: how many subscribers do you have?”                    |
| 10–14 сек. | `2. EMAILS PER MONTH`                      | “Second: how many emails do you send each month?”             |
| 14–18 сек. | `3. AUTOMATION`                            | “Third: how complex must the automation be?”                  |
| 18–22 сек. | `4. REQUIRED INTEGRATIONS`                 | “Fourth: which tools must it connect to?”                     |
| 22–25 сек. | `CONSTRAINTS FIRST. PRODUCT SECOND.`       | “Compare products only after answering those four questions.” |

В Kdenlive създай шест отделни title clips. Не снимай текстовете с OBS.

---

## 6. Резервно видео без глас

**Дължина:** 16 секунди.  
**Звук:** без звук или тиха музика с доказан лиценз.

| Време      | Текст на екрана                      |
| ---------- | ------------------------------------ |
| 0–2 сек.   | `BEST CRM?`                          |
| 2–4 сек.   | `WRONG FIRST QUESTION.`              |
| 4–7 сек.   | `Which workflow is broken?`          |
| 7–10 сек.  | `Who will maintain it?`              |
| 10–13 сек. | `What must it connect to?`           |
| 13–16 сек. | `CONSTRAINTS FIRST. PRODUCT SECOND.` |

Направи шест title clips и постави всеки след предишния. Не добавяй преходи;
използвай обикновени hard cuts.

---

## 7. Монтаж в Kdenlive — точните стъпки

### Създаване на вертикален проект

1. Отвори Kdenlive.
2. Натисни `File` → `New`.
3. В `Project Profile` избери вертикален профил `1080x1920, 30 fps` от
   категория `Custom`.
4. Ако няма такъв, отвори `Project` → `Project Settings` → `Manage Project
Profiles` → `Create new profile`.
5. Въведи Width `1080`, Height `1920`, Frame rate `30/1`, Progressive.
6. Запази профила като `Vertical 1080x1920 30fps`.
7. Запази проекта в папката на конкретното видео.

### Поставяне на OBS записа

1. В `Project Bin` натисни `Add Clip or Folder`.
2. Избери remux-натия MP4 от OBS.
3. Плъзни файла върху track `V1`.
4. Натисни клипа и отвори `Effects`.
5. Потърси `Transform` и го добави.
6. Увеличи `Size`, докато централната част на сайта запълни вертикалния кадър.
7. Променяй `X`, така че активният бутон и основният текст да останат в средата.
8. Разрежи преди и след всяка ненужна пауза с `Shift+R` или инструмента Razor.
9. Изтрий паузата и събери клиповете без празни места.

Не се опитвай да показваш цялата desktop страница в 9:16. Показвай само
активната централна зона: въпрос, избрана карта, бутон или резултат.

### Добавяне на текстова карта или надпис

1. В `Project Bin` натисни стрелката до `Add Clip or Folder`.
2. Избери `Add Title Clip`.
3. Натисни инструмента за текст или `Alt+T`.
4. Напиши само текста от колоната `Текст върху видеото`.
5. Използвай бял bold текст и тъмен правоъгълник зад него.
6. Постави текста в горната средна част, но не в най-горните 150 px.
7. Натисни `Create Title`.
8. Постави title clip върху track `V2`, над видеото.
9. Скъси го до секундите, посочени в сценария.

Текстът не трябва да покрива бутона или резултата, за който говориш.

### Добавяне на глас

Ако си записал глас отделно:

1. Добави WAV/MP3 файла в `Project Bin`.
2. Плъзни го върху `A1`.
3. Постави първата дума на 0:00.
4. Съкрати паузите във видеото, докато действията съвпаднат с изреченията.
5. Намали евентуалната музика на около 8–12% и остави гласа ясно отпред.

### Експорт

1. Натисни `Ctrl+Enter` или `File` → `Render`.
2. Избери `MP4-H264/AAC`.
3. Име на файла: `unsolero-video-01.mp4`.
4. Провери: `1080x1920`, `30 fps`, цял проект.
5. Натисни `Render to File`.
6. Гледай целия MP4 веднъж преди качване.

---

## 8. Може ли AI да направи видеото вместо теб?

Да, но има две различни ситуации.

### Вариант A — почти всичко се прави от AI

Подходящ за видео №2 и видеото без глас. Най-прекият вариант е **InVideo AI**:

1. Отвори `https://invideo.io`.
2. Избери AI video generator.
3. Избери `Use my script` или workflow `Script to Video`.
4. Избери формат `9:16`, език `English`, faceless video, AI voice и subtitles.
5. Постави следния prompt:

```text
Create a 25-second vertical 9:16 faceless SaaS education video.
Use a clean dark background, large white captions, simple software icons,
hard cuts, no avatar, no talking person, and no fake product interface.
Use exactly the script below without rewriting it. Use a calm professional
English voice. Keep all text inside the mobile safe area.
```

6. Под prompt-а постави точния сценарий от видео №2.
7. Натисни `Generate`.
8. Смени всеки кадър, който изглежда несвързан или показва измислен софтуер.
9. Провери правописа на всички субтитри.
10. Експортирай 1080×1920 без watermark.

InVideo официално предлага script-to-video, избор на voice, subtitles, music и
автоматични визуални сцени. Безплатният export може да има watermark.

### Вариант B — препоръчаният процес за UNSOLERO

За реална демонстрация AI не трябва да измисля интерфейса. Направи един истински
OBS запис и дай на AI да извърши останалото:

1. Запиши действията от видео №1 без глас.
2. Качи MP4 в **VEED**, **Descript** или InVideo.
3. Постави точния voiceover от таблицата.
4. Избери AI voice.
5. Натисни automatic captions.
6. Избери вертикално `9:16`.
7. Премахни автоматично добавените stock кадри — сайтът трябва да остане реален.
8. Провери всяка цена, продуктово име и твърдение.
9. Експортирай и изгледай целия файл.

### Кой инструмент за какво служи

| Инструмент     | Подходящ за                                                        | Ограничение                                                                    |
| -------------- | ------------------------------------------------------------------ | ------------------------------------------------------------------------------ |
| InVideo AI     | Най-автоматичен script → voice → visuals → captions                | Често избира общи stock кадри; не трябва да измисля интерфейса на UNSOLERO     |
| VEED           | Browser editor, AI voice, captions, вертикално видео и редактиране | AI credits и част от export функциите са платени                               |
| Descript       | Монтаж чрез редактиране на текста, AI voice и captions             | Безплатният план е ограничен; 1080p е в платен план                            |
| Canva          | Заглавни карти, икони, кратки AI b-roll клипове и brand templates  | Генерираните клипове са кратки; не е автоматична точна продуктова демонстрация |
| OBS + Kdenlive | Най-точната и безплатна демонстрация на истинския продукт          | Изисква ръчен запис и кратък монтаж                                            |

**Препоръка:** първо тествай InVideo AI с видео №2. За видео №1 използвай
OBS + AI voice/captions или OBS + Kdenlive. Не плащай годишен план, преди да
направиш поне три тестови видеа и да провериш качеството на export-а.

Няма инструмент, който надеждно да влезе в живия UNSOLERO builder, да натисне
правилните бутони, да покаже актуалния резултат и едновременно да гарантира, че
всички продуктови факти са верни. Тази част трябва да бъде истински screen
recording или да бъде проверена от теб кадър по кадър.

---

## 9. Качване и график

Започни с четири видеа седмично:

- понеделник — демонстрация на builder-а;
- сряда — кратко сравнение или правило;
- петък — отговор на въпрос;
- неделя — резервно видео без глас.

За всяко видео:

1. Качи същия чист MP4 отделно във всяка платформа.
2. Не сваляй видео от TikTok, за да го качиш с TikTok watermark другаде.
3. Използвай едно изречение за caption и един CTA.
4. Ако има affiliate линк или конкретна affiliate оферта, добави:
   `Affiliate disclosure: we may earn a commission at no extra cost to you.`
5. Ако има реалистични AI-generated хора, гласове или събития, включи AI label
   в платформата. AI помощ само за сценарий, обикновени captions или корекция
   на звук обичайно не изисква такъв label в YouTube.

Примерен caption за видео №1:

```text
Software recommendations should start with constraints, not commissions.
Build your own brief at unsolero.com.
```

Примерно заглавие:

```text
Build a Software Stack From Constraints, Not Hype
```

---

## 10. Проверка преди публикуване

Не качвай видеото, докато всеки отговор не е „да“:

- Файлът е 1080×1920 и се гледа вертикално.
- Първите две секунди съдържат ясен проблем или обещание.
- Няма лични tabs, имейли, имена, банкови или admin данни.
- Вижда се точно това, за което говори гласът.
- Няма измислен UI, измислена цена или непроверен winner.
- Субтитрите са прочетени и поправени ръчно.
- Текстът не е покрит от бутоните на TikTok/Reels/Shorts.
- Има само един CTA.
- Affiliate disclosure е добавен, когато е необходим.
- AI label е включен, ако видеото съдържа реалистично генерирано съдържание.

---

## Официални източници

- [OBS Quick Start Guide](https://obsproject.com/kb/quick-start-guide)
- [OBS Recording Output Guide](https://obsproject.com/kb/standard-recording-output-guide)
- [Kdenlive vertical project profiles](https://docs.kdenlive.org/en/project_and_asset_management/project_settings/general_settings.html)
- [Kdenlive Title Clips](https://docs.kdenlive.org/en/titles_and_graphics/titles/titles.html)
- [Kdenlive Exporting](https://docs.kdenlive.org/en/exporting.html)
- [YouTube Shorts upload requirements](https://support.google.com/youtube/answer/12779649)
- [YouTube AI disclosure requirements](https://support.google.com/youtube/answer/14328491)
- [TikTok AI-generated content labels](https://support.tiktok.com/en/using-tiktok/creating-videos/ai-generated-content)
- [InVideo: create a video using your script](https://help.invideo.io/en/articles/9382180-how-can-i-create-a-video-using-my-script)
- [VEED script-to-video](https://www.veed.io/tools/ai-video/script-to-video)
- [Descript AI video maker](https://www.descript.com/video-generator/video-maker)
- [Canva AI video generator](https://www.canva.com/features/ai-video-generator/)
