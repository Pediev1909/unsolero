# Проучване 2 — Приложение: Производствено ръководство, клик по клик

**Дата на проверка:** 2 септември 2026
**Част от:** [02_SOCIAL_GROWTH_BG.md](./02_SOCIAL_GROWTH_BG.md) (стратегията и скриптовете) и [05_FIRST_VIDEO_MICROTASKS_BG.md](./05_FIRST_VIDEO_MICROTASKS_BG.md) (първото видео стъпка по стъпка).
**Какво е това:** точните менюта, бутони, полета и клавиши за запис, монтаж, глас, музика, качване и измерване, написани за конкретната машина, на която ще се прави. Етикетите на интерфейса са оставени на английски, както се виждат на екрана. Където етикет не е потвърден от официална документация, стои **(провери)**.

---


Написано за: Fedora, GNOME Shell 50.4 на Wayland, един лаптоп екран 1920×1080 @144 Hz, Google Chrome, OBS Studio **32.2.2** (Flatpak `com.obsproject.Studio`), Kdenlive **26.04.3** (Flatpak `org.kde.kdenlive`), RTX 4050 Mobile + Intel iGPU, ffmpeg 8.1.2 на хоста, `python3` без `pip3`, вграден микрофон през PipeWire. Всичко е безплатно. Дата на проверка: 2 септември 2026. Етикетите на бутоните са на английски, точно както ги виждаш на екрана; където етикетът не е потвърден от официален източник, пише **(провери)**.

## Четири ограничения, от които не можеш да излезеш с кликане

1. **Екранът е 1080 px висок.** Реален запис 1080×1920 е физически невъзможен. Най-големият точен 9:16 прозорец, който Chrome може да покаже, е **576×1024**; OBS го увеличава ~1.9×. Компенсира се с page zoom 125–150% (по-големи бутони и текст) и Lanczos филтър. На телефон изглежда добре; не изглежда като запис от телефон 1:1.
2. **Двете приложения са Flatpak.** Захващането на екран минава през диалога на GNOME „Share Screen“; NVENC иска NVIDIA runtime за Flatpak със същата версия като драйвера; Whisper на Kdenlive живее в venv вътре в sandbox-а.
3. **Wayland не дава глобални клавиши на OBS** както X11. Планирай го (виж 1.7).
4. **Affiliate съдържанието е комерсиално.** Това изключва ElevenLabs Free (само некомерсиално) и прави edge-tts правно сиво. Използвай Kokoro (Apache-2.0) или Piper.

---

## 0. Еднократна подготовка (терминал)

```bash
# pip + pipx за TTS; espeak-ng за Kokoro/Piper; Audacity за глас
sudo dnf install python3-pip pipx espeak-ng audacity

# Провери Flatpak-ите и правата (двата имат --filesystem=host, т.е. виждат ~/Videos)
flatpak info com.obsproject.Studio | head
flatpak info org.kde.kdenlive | head

# NVIDIA: версията на драйвера и на Flatpak GL разширението ТРЯБВА да съвпадат
cat /proc/driver/nvidia/version
flatpak list --runtime | grep -i nvidia
# Ако няма съвпадение, например драйвер 580.95.05:
#   flatpak install flathub org.freedesktop.Platform.GL.nvidia-580-95-05
# OBS 32.2.2 използва NVIDIA SDK 13 → драйверът трябва да е >= 570.
sudo dnf install xorg-x11-drv-nvidia-cuda      # CUDA библиотеки за NVENC (RPM Fusion)

mkdir -p ~/Videos/obs ~/Videos/vo ~/Videos/export ~/Videos/cards
```

---

## 1. OBS Studio 32.2.2

Стартиране: `flatpak run com.obsproject.Studio` или от менюто на GNOME. В 32.2 бутонът за добавяне на източник отваря нов диалог с миниатюри; 32.2 също поправи „PipeWire could fail on NVIDIA GPUs“.

### 1.1 Отделен профил и колекция за вертикално (веднъж)

1. Меню **Profile** → **New** → име `Vertical 9-16` → **OK**. (Профилът пази Settings: резолюция, енкодер, клавиши.)
2. Меню **Scene Collection** → **New** → име `Unsolero vertical` → **OK**. (Колекцията пази сцени и източници.)
3. Старият профил остава за 1920×1080 дълги видеа; сменяш от същите две менюта.

### 1.2 Settings → Video

1. **File** → **Settings** (или бутон **Settings** в дока **Controls**) → ляв списък **Video**.
2. **Base (Canvas) Resolution**: кликни в полето, напиши `1080x1920`, `Enter`.
3. **Output (Scaled) Resolution**: `1080x1920`.
4. **Downscale Filter**: **Lanczos (Sharpened scaling, 36 samples)**.
5. **Common FPS Values**: **30**.
6. **Apply**.

За 16:9 профила: `1920x1080` / `1920x1080` / 30.

### 1.3 Settings → Output → Recording

1. **Settings** → **Output** → **Output Mode**: **Advanced**.
2. Таб **Recording**:
   - **Type**: **Standard**.
   - **Recording Path**: **Browse** → `~/Videos/obs`. Отметни **Generate File Name without Space**.
   - **Recording Format**: **Hybrid MP4 (.mp4)** — не се чупи при crash като MKV, но се отваря директно в Kdenlive. Ако липсва: **Matroska Video (.mkv)** и remux после (1.7). **Не** ползвай обикновен **MPEG-4 (.mp4)** — crash или принудителен стоп прави файла нечетим.
   - **Video Encoder**: отвори списъка и търси **NVIDIA NVENC H.264**. Ако го има, NVENC работи. Ако виждаш само **x264** и **FFmpeg VAAPI H.264**, GL.nvidia runtime-ът липсва или не съвпада (виж раздел 0); междувременно ползвай резервните настройки долу. В лога ще пише „Not running on NVIDIA GPU, falling back to non-texture encoder“ — това е по-бавен път на копиране, не грешка, защото GNOME рендерира на Intel iGPU.
   - **Audio Encoder**: **FFmpeg AAC**. **Audio Track**: отметни **1** и **2**.
3. Настройки на енкодера (появяват се под избора):

| Енкодер | Rate Control | Качество | Останалото |
|---|---|---|---|
| **NVIDIA NVENC H.264** (предпочитан) | **CQP** | **CQ Level** `18` (20 ако файловете са големи) | **Preset** `P5: Slow (Good Quality)`, **Tuning** `High Quality`, **Multipass** `Two Passes (Quarter Resolution)`, **Profile** `high`, **Psycho Visual Tuning** вкл., **Max B-frames** 2, **Keyframe Interval** `2` s |
| **FFmpeg VAAPI H.264** (Intel iGPU) | **CQP** | **QP** `20` | **Profile** `High`, **Level** `Auto`, keyframe 2 |
| **x264** (CPU, винаги наличен) | **CRF** | **CRF** `18` | **CPU Usage Preset** `veryfast` (`faster` ако падат кадри), **Profile** `high`, **Tune** `(None)`, **Keyframe Interval** `2` |

**Settings** → **Audio**: **Sample Rate** `48 kHz`, **Channels** `Stereo`. **Output → Audio**: битрейт на Track 1/2 `192` или `320`.

### 1.4 9:16 прозорец на unsolero.com в Chrome

1. Отвори `https://unsolero.com`, натисни `F11` (цял екран — цялата площ 1920×1080 става страница).
2. `Ctrl+Shift+I` (DevTools). `Ctrl+Shift+M` (**Toggle device toolbar**).
3. В лентата над страницата: падащо **Dimensions** → **Responsive**. Ширина `576`, височина `1024`. (Ако лентата краде място: `558 × 992`. Винаги ширина = височина × 9/16.) DPR/zoom 100%. Да пишеш 1080×1920 при 50% zoom не печели нищо — пикселите на екрана са пак 540×960.
4. Скрий DevTools от записа: бутон **⋮** (**Customize and control DevTools**) → **Dock side** → **Undock into separate window** (`Ctrl+Shift+D` превключва). Device mode остава активен, а OBS захваща само прозореца на страницата.
5. Направи интерфейса четлив след увеличението: с фокус върху страницата `Ctrl` + `+` до 125–150%.
6. Изключи нотификациите на Chrome и на GNOME (**Settings** → **Notifications** → **Do Not Disturb**) преди запис.

### 1.5 Източници на Wayland и напасване към 9:16

На Wayland източниците се казват точно **Screen Capture (PipeWire)** и **Window Capture (PipeWire)**.

1. Док **Sources** → **+** → **Window Capture (PipeWire)** → име `Chrome unsolero` → **OK**.
2. Отваря се диалогът на GNOME **Share Screen** с табове **Window** и **Display**, отметка **Remember this selection**, бутони **Cancel** / **Share**. Таб **Window** → кликни прозореца на страницата (не DevTools прозореца) → отметни **Remember this selection** → **Share**.
3. В Properties отметни **Show Cursor** ако искаш показалеца.
4. Изрежи до прозореца: в превюто задръж `Alt` и влачи ръбовете на червената рамка навътре, докато остане само областта 576×1024 (ръбовете стават зелени при crop). За точност: **Edit** → **Transform** → **Edit Transform…** (`Ctrl+E`): **Crop** Left/Right/Top/Bottom.
5. Десен клик върху източника → **Transform** → **Fit to screen** (`Ctrl+F`), за да запълни канваса 1080×1920.
6. Десен клик → **Scale Filtering** → **Lanczos**.
7. По желание: **+** → **Color Source** най-отдолу (цвят на бранда `#14605a` или `#f4f5f7`) и **Text (FreeType 2)** за постоянен handle — вътре в safe zone (раздел 5.5).

Ако не се появи нищо: OBS трябва да върви като Wayland клиент — `flatpak run --env=QT_QPA_PLATFORM=wayland com.obsproject.Studio`.

### 1.6 Звук: верига за лаптоп микрофон, без изтичане на системен звук, отделни пътечки

**Settings** → **Audio** → **Global Audio Devices**: **Desktop Audio** → **Disabled** (иначе нотификации и музика влизат в гласа). **Mic/Auxiliary Audio** → вграденият масив (`alsa_input…skl_hda_dsp_generic…`; ако виждаш само „Default“, то е ок).

Филтри: в **Audio Mixer** кликни **⚙** до **Mic/Aux** → **Filters** → **+** (долу вляво), в този ред:

1. **Noise Suppression** → **Method**: **RNNoise (good quality, more CPU usage)**.
2. **Gain** → `+6.0 dB` (нагласи така, че пиковете на речта да са между −12 и −6 dB на метъра, никога червено).
3. **Noise Gate** (по избор; махни го, ако реже тихи думи) → **Close Threshold** `-40 dB`, **Open Threshold** `-34 dB`, **Attack** `10 ms`, **Hold** `200 ms`, **Release** `150 ms`.
4. **Compressor** → **Ratio** `3:1`, **Threshold** `-18 dB`, **Attack** `6 ms`, **Release** `60 ms`, **Output Gain** `+3 dB`.
5. **Limiter** → **Threshold** `-3.0 dB`, **Release** `60 ms`. Винаги последен.

Физически: 30–40 см от лаптопа, говори *през* клавиатурата, не *в* нея; на ток, не на батерия (вентилаторът); капакът в ъгъла, в който ще остане.

Мониторинг: **Edit** → **Advanced Audio Properties** → ред **Mic/Aux** → **Audio Monitoring** → **Monitor Only (mute output)** за проба, **Monitor Off** за запис. Пътечки: там отметни **Tracks 1** и **2** за Mic/Aux.

**Препоръка: записвай гласа отделно в Audacity.** Faceless видеото не иска синхрон с кликанията — запиши екрана без глас, после чети скрипта, гледайки клипа в Kdenlive. Audacity: **Audio Setup** → **Recording Device** → вграденият микрофон; запис; **Effect** → **Noise Removal and Repair** → **Noise Reduction** (Get Noise Profile върху 2 сек. тишина, после 12 dB); **Effect** → **Volume and Compression** → **Loudness Normalization** → `-16 LUFS`; **File** → **Export Audio** → WAV 48 kHz. По-чисто от всяка OBS верига и можеш да презапишеш едно изречение.

### 1.7 Клавиши, файлове, remux

1. **Settings** → **Hotkeys** → **Start Recording** / **Stop Recording** → кликни в празното поле → натисни `Ctrl+F9` / `Ctrl+F10`. **Pause Recording** → `Ctrl+F11`.
2. На Wayland тези клавиши работят надеждно само когато OBS прозорецът е на фокус (освен ако се появи известие за GlobalShortcuts portal в **Settings → Hotkeys** — провери). Работещ винаги вариант: **Start Recording** в **Controls**, `Alt+Tab` към Chrome, кадърът, `Alt+Tab` обратно, **Stop Recording**; първите и последните секунди се режат в Kdenlive.
3. **File** → **Show Recordings** отваря папката.
4. Remux (само ако си записал MKV): **File** → **Remux Recordings** → влачи `.mkv` → **Remux**. Или **Settings → Advanced → Recording → Automatically remux to mp4**. В терминал: `ffmpeg -i in.mkv -c copy out.mp4`.

### 1.8 Пробен запис (20 секунди, всяка сесия)

1. Chrome цял екран, device toolbar 576×1024, DevTools в отделен прозорец, zoom 125–150%, нотификации изключени.
2. OBS профил **Vertical 9-16**; превюто е вертикално.
3. Източникът показва само страницата; без сива DevTools рамка.
4. Микрофонът пикира между −12 и −6 dB при говор; тишината е под −50 dB; **Desktop Audio** е Disabled.
5. Запиши 20 сек., спри, отвори файла: остър текст, стабилни 30 fps (**View** → **Stats**: „Frames missed due to rendering lag“ = 0), звук на пътечки 1 и 2.
6. Диск: 1080×1920 при CQ18 ≈ 30–60 MB/минута.

---

## 2. Kdenlive 26.04.3

Kdenlive преподреди менютата в 25.x/26.x: **Project Settings** и **Render** са под **File**, субтитрите под **Sequence** → **Subtitles**, а добавянето на клип е през Project Bin. Старите етикети (**Project → Render**, **Project → Subtitles**, **Project → Add Title Clip**) са дадени в скоби. Стартиране: `flatpak run org.kde.kdenlive`.

### 2.1 Вертикален проект

1. **File** → **New** (`Ctrl+N`) отваря **Project Settings** (за съществуващ проект **File** → **Project Settings**; старо: **Project → Project Settings**).
2. Таб **General Settings** → списък **Video Profile** → категория **Custom** → **Vertical HD 30 fps** (MLT профил `vertical_hd_30`: 1080×1920, 30 fps, progressive). Layout-ът автоматично става вертикален със вертикални safe зони.
3. **Video tracks** `2`, **Audio tracks** `3` (глас, музика, ефекти), **Audio channels** **2 Channels (Stereo)**. **OK**.
4. Ако профилът липсва: иконата гаечен ключ до списъка (или **Settings** → **Manage Project Profiles**) → избери 1080p30 → зелен **+** → **Description** `Vertical HD 1080x1920 30`, **Size** `1080`×`1920`, **Frame rate** `30/1`, **Pixel aspect ratio** `1/1`, **Display aspect ratio** `9/16`, отметка **Progressive**, **Colorspace** `ITU-R 709` → иконата за **Save profile** (Kdenlive не предупреждава, ако забравиш) → **OK**.

### 2.2 Импорт, рязане, ripple, скорост

1. **Project Bin** (горе вляво) → **Add Clip or Folder** (бутон +) → `~/Videos/obs/…mp4` → **Open**. Или влачи от файловия мениджър.
2. Влачи клипа върху **V1**.
3. Инструменти: **Selection** `S`, **Razor** `X`, **Spacer** `M`. Рязане при playhead без смяна на инструмента: `Shift+R`. Зона in/out: `I` / `O`.
4. Ripple delete (маха клипа и затваря дупката): десен клик → **Extract Clip** (`Shift+Delete`). Само дупката: десен клик върху празното → **Remove Space**. Зона през всички пътечки: `Shift+X` (extract), `Z` (lift, оставя дупка).
5. Скорост: десен клик → **Change Speed** → **Speed** `%` (напр. `150`), **Pitch compensation** отметнато за глас → **OK**. Бързо: `Ctrl` + влачене на края на клипа.
6. Раздели звука от OBS записа, за да го заглушиш: десен клик → **Split Audio**; mute иконата на пътечката или изтрий.
7. Timeline на цял екран: `Shift+Z`. Кадър по кадър `←`/`→`, секунда `Shift+←/→`.

### 2.3 Приближаване (zoom) към бутон с Transform + keyframes

1. Панел **Effects** (долу вляво; **View** → **Effects** ако е скрит) → търси `Transform` → влачи **Transform** върху клипа.
2. В **Effect/Composition Stack** (вдясно) виждаш **X**, **Y**, **W**, **H**, **Size**, **Rotation**, **Compositing**. Keyframing е включен с един keyframe в началото.
3. Постави playhead където започва zoom-ът. Кликни **Add keyframe** (◆+). Стойности 100%.
4. Премести playhead 12–15 кадъра напред (0.4–0.5 сек.). Смени **Size** на `140` и влачи рамката в **Project Monitor** така, че бутонът да е в центъра. Kdenlive сам създава втори keyframe. **(провери)** Падащото меню за тип keyframe дава **Linear** / **Discrete** / **Smooth** и eased варианти като **Ease In/Out (Cubic)** — избери ease за естествен push-in.
5. За връщане: keyframe в края със **Size** `100`. Навигация: **Go to previous/next keyframe**; влачи keyframe наляво/надясно за retiming.

### 2.4 Title clips (hook текст, карти)

1. **Project Bin** → **Add Clip or Folder** ▾ → **Add Title Clip** (или десен клик в празния bin; старо: **Project → Add Title Clip**). Отваря се **Titler** с три червени рамки: външна = кадър, средна = action-safe, вътрешна = title-safe. Текстът стои във вътрешната.
2. Инструмент **Add Text** (T) → клик върху канваса → пиши. В **Properties**: **Font** (тежък sans — Inter Bold или Montserrat Bold), размер `72–96` px, **Color** бяло, **Outline** `6–8` px черно, **Shadow** X/Y `4`, blur `8`, центрирано **(провери етикетите)**.
3. **Duration** долу (напр. `00:00:03:00`). **Create Title**. Влачи от bin-а върху **V2** над кадрите.
4. Повторна употреба: десен клик върху title в bin-а → **Edit** → смени текста → **Update Title**. Шаблон: **Template** ▾ → save.
5. Pop-in анимация: добави **Transform** на title клипа и keyframe **Size** 80→100 за 6 кадъра (както в 2.3); fade — влачи горния ляв ъгъл на клипа (2.6).

### 2.5 Субтитри и Speech-to-Text (Whisper) вътре във Flatpak

Настройка (веднъж):

1. **Settings** → **Configure Kdenlive…** → **Plugins** → секция **Speech to text**.
2. Избери **Whisper** (по-точен от VOSK, слага пунктуация). Kdenlive прави venv в `~/.var/app/org.kde.kdenlive/data/kdenlive/venv`. **Check configuration**; ако пита **Install missing dependencies** — приеми (openai-whisper + torch, ~2 GB, CPU билд).
3. Модел: **Download (1.4GB)** за **turbo**, или **Manage models** → **Install model** → `small.en` (бърз, добър за ясен английски глас) / `base.en` (най-бърз). **Language**: `English` вместо Autodetect. **Device**: **CPU**. Пропусни **Install multilingual translation** (9 GB).
4. Ако вграденият инсталатор откаже (sandbox-ът понякога блокира pip):

```bash
flatpak run --command=/bin/bash org.kde.kdenlive
python3 -m venv $HOME/.var/app/org.kde.kdenlive/data/kdenlive/venv   # само ако папката липсва
$HOME/.var/app/org.kde.kdenlive/data/kdenlive/venv/bin/python -m pip install -U pip openai-whisper torch srt
$HOME/.var/app/org.kde.kdenlive/data/kdenlive/venv/bin/whisper --model small.en ~/Videos/obs/test.mp4   # сваля модела
exit
```
После пак **Configure Kdenlive… → Plugins → Speech to text** и провери, че моделът е в **Model**. След ъпдейт на Flatpak-а може да трябва да повториш 2–4.

Генериране и стил:

1. Лента на timeline → **Edit Subtitle Tool** (добавя subtitle пътечка). Меню: **Sequence** → **Subtitles** (старо: **Project → Subtitles**). `Shift+S` — ръчен субтитър при playhead.
2. В лентата на subtitle пътечката → **Speech recognition** → диалог **Automatic Subtitling**: **Model** (Whisper моделът), **Language** `English`, обхват **Timeline zone (all tracks)** или **Selected clips**, **Maximum character per line** `28` (два кратки реда на телефон), **Translate to English** изключено → **Process**. Поправяш дума с двоен клик върху субтитъра.
3. Стил: **Sequence** → **Subtitles** → **Manage Subtitles** → **Style** редактор с превю: удебелен sans, размер `64–72`, бяло, **Outline** `4–6` черно, малка **Shadow**, **Alignment** долу-център, вертикален **Margin** такъв, че редът да е около y≈1300–1400 от 1920 (над лентата с интерфейса на платформите, раздел 5.5).
4. Експорт SRT за качване като captions: **Sequence** → **Subtitles** → **Export Subtitle File** → SRT. Импорт: **Import Subtitle File** (SRT, ASS, VTT).
5. Вграждане: когато subtitle пътечката е видима, рендерът „изпича“ субтитрите в картината. В **Render → More options** отметката **Embed subtitles instead of burning them in** трябва да е **изключена** за TikTok/Reels/Shorts.

### 2.6 Звук: нормализация, музика, fade, нива

1. Глас: избери гласовия клип → **Effects** → **Audio** → **Volume and Dynamics** → **Normalize (2 Pass)** → **Target Program Loudness** `-16` LUFS → **Analyse to Apply Effect**. (Еднопасовият **Normalize** е за live; не го ползвай.)
2. Музика: свали трак (раздел 4.2) → **Add Clip or Folder** → влачи на **A2**. **View** → **Audio Mixer** → плъзгач **A2** на `-20 dB`. Правило: музиката 18–24 dB под гласа; master пикове ≤ −1 dBTP.
3. Fade: задръж курсора в горния ъгъл на клипа до появата на дръжката и влачи навътре. Музиката fade-out последната секунда.
4. Ducking: Kdenlive няма sidechain. За 30–60 сек. клип — дръж музиката константно на −22 dB и не я пипай.
5. Заглуши OBS scratch пътечката, ако гласът е от Audacity: mute иконата на пътечката.

### 2.7 Render

1. **File** → **Render** (`Ctrl+Return`) (старо: **Project → Render**).
2. Preset **MP4-H264/AAC**. Резолюцията следва профила (1080×1920); **Rescale** изключено.
3. **Quality** плъзгач ~70–80% надясно (Kdenlive го мапва към CRF; текст на екран иска повече). Ако има битрейт поле: `12000k`. **(провери мапването)**
4. **More options** ▾: **Parallel processing** изключено; **Export audio** включено; **Embed subtitles instead of burning them in** изключено.
5. **Rendering**: **Full project** (или **Selected zone** за проба).
6. **Output file**: `~/Videos/export/unsolero-<тема>-2026-09-02-v1.mp4` (тема-дата-версия; ще ре-експортираш).
7. **Render to File**.

Цел за TikTok / Reels / Shorts: MP4, H.264 High, 1080×1920, 30 fps константно, 8–12 Mbps, AAC 48 kHz стерео 192–320 kbps, ≤ 3 мин. за Shorts, идеално 25–60 сек. Дълго YouTube: 1920×1080, 8–12 Mbps.

---

## 3. CapCut в браузър — алтернатива, с уговорки

Работи в Chrome на Linux (capcut.com), иска вход (Google/TikTok/имейл). 2026 планове: Free / Standard ≈ $9.99 / Pro ≈ $19.99. Free експортира до **1080p без воден знак**, ако не ползваш Pro-маркиран шаблон/ефект/AI инструмент. **Auto captions** са безплатни (≈10 мин. видео на проект); speaker-ID и част от стиловете са Pro.

Път **(провери етикетите — CapCut ги сменя често)**: **Sign in** → **Create new** → **Video** → бутон **Ratio** над плейъра (`16:9`) → **9:16** → ляво **Media** → **Import** → **Upload** → влачи на timeline → ляво **Captions** → **Auto captions** → **Language** `English` → **Generate** → избери caption слоя → **Template/Style** → шрифт, размер, позиция (~70% височина) → **Export** → `1080p`, `30`, `MP4`.

Условията за ползване (промяна от 12 юни 2025) дават на CapCut вечен, световен, безвъзмезден, неотменим лиценз да ползва, редактира и разпространява каченото, включително глас и лик, за реклама; лицензът надживява изтриването на акаунта. CapCut отговори, че секцията „User-generated Content“ не е променена и че не претендира собственост. Практически: нищо, което качваш там, не е само твое в приложим смисъл. Kdenlive дава същия резултат локално. Ползвай CapCut само ако ти трябва стилът на надписите, и никога като единствен архив.

---

## 4. Безплатни ресурси и AI глас

### 4.1 Text-to-speech (2026)

| Опция | Цена | Лиценз за affiliate (комерсиално) съдържание | Качество | Присъда |
|---|---|---|---|---|
| **Kokoro-82M** (локално) | безплатно | Apache-2.0 — комерсиално ОК | високо (близо до ElevenLabs на добрите гласове) | **ползвай това** |
| **Piper** (локално) | безплатно | GPL-3.0 код (ползването е ок), гласовете имат отделни лицензи — проверявай | добро/леко роботизирано | резерва, много бързо |
| **edge-tts** (неофициален достъп до Microsoft Edge) | безплатно | няма даден лиценз; Microsoft казват, че комерсиална употреба без Azure абонамент може да нарушава условията | много добро | само за проби |
| **ElevenLabs Free** | 10k credits/мес. (~10 мин.) | **само некомерсиално**, задължителна атрибуция „elevenlabs.io“ (Starter $6/мес. дава комерсиален лиценз) | най-добро | неизползваемо за affiliate на Free |
| **OpenAI TTS** | платен API | комерсиално ОК | много добро | не е безплатно — пропусни |

Kokoro (препоръка) на тази машина:

```bash
sudo dnf install espeak-ng python3-pip
python3 -m venv ~/tts && source ~/tts/bin/activate
pip install -U pip "kokoro>=0.9.4" soundfile numpy
cat > ~/tts/say.py <<'EOF'
import sys, numpy as np, soundfile as sf
from kokoro import KPipeline
text = open(sys.argv[1]).read()
voice = sys.argv[3] if len(sys.argv) > 3 else 'af_heart'
p = KPipeline(lang_code='a')          # 'a' американски английски, 'b' британски
audio = np.concatenate([a for _, _, a in p(text, voice=voice, speed=1.05)])
sf.write(sys.argv[2], audio, 24000)
EOF
echo "Stop paying Mailchimp for a list you barely email." > ~/tts/script.txt
python ~/tts/say.py ~/tts/script.txt ~/Videos/vo/voice.wav af_heart
```
Първото пускане сваля модела (~330 MB) от Hugging Face; после е офлайн. Оценки на гласовете от официалния VOICES.md: **af_heart** (A), **af_bella** (A-), **af_nicole** (B-), **bf_emma** (B-); мъжките са по-слаби — **am_michael**, **am_fenrir**, **am_puck** (C+). Импортирай WAV-а в Kdenlive на A1 и нормализирай (2.6).

Piper:

```bash
source ~/tts/bin/activate
pip install piper-tts
python3 -m piper.download_voices en_US-lessac-medium en_US-ryan-high en_GB-alan-medium
python3 -m piper -m en_US-ryan-high -f ~/Videos/vo/voice.wav --input-file ~/tts/script.txt --sentence-silence 0.3
```

edge-tts (само проби; дава и SRT, удобен за чернова на субтитри):

```bash
pipx install edge-tts
edge-tts --list-voices | grep -E '^en-(US|GB)'
edge-tts --voice en-US-AndrewMultilingualNeural --rate=+5% \
  --text "$(cat ~/tts/script.txt)" --write-media ~/Videos/vo/voice.mp3 --write-subtitles ~/Videos/vo/voice.srt
```

### 4.2 Музика и ефекти с ясен лиценз

- **YouTube Audio Library**: studio.youtube.com → ляво меню **Audio library** → таб **Music** / **Sound effects** → филтър **Attribution not required** → **Download**. Чисто за YouTube, вкл. монетизирани видеа. Официалната страница не казва нищо за други платформи — третирай я като **само за YouTube**.
- **Pixabay Music**: pixabay.com → **Music** → филтър по настроение/дължина → **Download**. Pixabay Content License: комерсиално, без атрибуция, може да се променя. Пази URL-а на трака в бележките си.
- **Free Music Archive**: филтрирай само **CC BY** или **CC BY-SA** (всичко с **NC** е некомерсиално и неизползваемо); кредитирай точно както иска лицензът.
- **Uppbeat Free**: 3 сваляния/месец, всяко с еднократен **Uppbeat credit**, който трябва да е в описанието на точно това видео.
- **TikTok Commercial Music Library**: бизнес акаунтите виждат само нея под **Add sound**. Ако музиката е вградена в MP4-а, пикерът се заобикаля изцяло.
- **Instagram**: бизнес/професионални акаунти са ограничени до Meta Sound Collection. Вграждай лицензираната музика във файла и не добавяй Instagram аудио отгоре.
- **LinkedIn**: няма библиотека; вграждай.

### 4.3 Canva Free за title карти и thumbnail

1. canva.com → **Create a design** (горе вдясно) → **Custom size** → **px** → **Width** `1080`, **Height** `1920` → **Create new design**. Още един дизайн `1280` × `720` за thumbnail на дълги видеа.
2. Ляво **Design** → търси „story“ или „reel cover“; **Text** → заглавие; **Elements** → форми; **Uploads** → твои скрийншоти.
3. **Share** (горе вдясно) → **Download** → **File type** **PNG** → **Download**. Прозрачен фон и **Resize** са Pro — прави отделни дизайни за всеки размер.
4. Brand Kit е Pro. Безплатни заместители: **Figma Starter** (3 файла) или **Penpot** (безплатно, MPL-2.0).

---

## 5. Качване — точни пътища (2026)

### 5.1 TikTok (десктоп, TikTok Studio)

Помощните страници на TikTok са JavaScript и не можаха да се извлекат — етикетите долу са от актуални ръководства; **провери на екрана**.

1. `https://www.tiktok.com/tiktokstudio/upload` → **Select video** (или влачи MP4). Уеб приема MP4/MOV/WebM, до **60 минути** и **10 GB**.
2. **Description**: пиши; `#` вади предложения за хаштагове, `@` за споменавания. Ключовата фраза първа — TikTok search индексира описанията.
3. **Cover** → **Edit cover** → плъзгача до чист, четим кадър (или качи PNG) → **Save (провери)**.
4. **Who can watch this video** → **Everyone**.
5. **Allow users to:** отметни **Comment**, **Duet**, **Stitch** (всичко, за обхват).
6. **Комерсиална декларация — задължителна за affiliate:** включи **Content disclosure and ads** → **Disclose commercial content** (старо: **Disclose post content**) → отметни **Your brand** (показва „Promotional content“) и/или **Branded content** („Paid partnership“, ако таг-ваш партньор). Липса на декларация при промоционално съдържание води до преглед по правилата от 2025–2026.
7. **AI-generated content**: включи, ако гласът е синтетичен и звучи като реален човек.
8. **Run a copyright check** (по избор).
9. **Schedule**: **Schedule video** → дата/час. Само десктоп, Creator/Business акаунт, до **10 дни** напред.
10. **Post**.

Линк в био: **Profile** → **Edit profile** → **Website** (само в мобилното приложение). Business акаунт може при всякакъв брой последователи; личен — по историческото правило ≥1,000, но няколко ръководства от 2026 казват, че прагът е махнат през 2024 — провери дали полето **Website** го има при теб. Смяна: **Settings and privacy** → **Account** → **Switch to Business Account** (ограничава музиката до Commercial Music Library, което за affiliate съдържание е коректната класификация).

Анализи: `tiktok.com/tiktokstudio` → ляво **Analytics** → **Overview** / **Content** / **Viewers** / **Followers**. **Content** → клик върху видео → **Retention rate** графика (по секунди), **Average watch time**, **Watched full video** %, **New followers**, **Traffic source**, **Search queries**. Експорт: **Download data** (горе вдясно, CSV) **(провери)**.

### 5.2 Instagram Reels (качване от десктоп)

1. `https://www.instagram.com` → лява лента **Create** (**+**) → **Post**.
2. **Select from computer** → MP4 → иконата **crop** (долу вляво в превюто) → **9:16** → **Next**.
3. (Филтри) → **Next**.
4. Дясно: **Caption** (до 2,200 знака; CTA и 3–5 хаштага в края), **Add location**, **Add collaborators**, **Accessibility** → **Write alt text**, **Advanced settings** → **Hide like and view counts** (изкл.), **Turn off commenting** (изкл.). **Share to Facebook** се показва само ако има свързана Page. **Cover photo** пикер в превюто **(провери — при някои акаунти е само в мобилното Edit cover)**.
5. **Share**.
6. Не е налично на десктоп: Instagram аудио, стикери, **Add paid partnership label** (мобилно: share екран → **Advanced settings**), **Trial** превключвателят.

Дължина: до 15 мин. качено (20 мин. се разпространява), но Reels над ~3 мин. не се препоръчват на не-последователи — affiliate клиповете стоят на 30–90 сек.

Trial reels (само мобилно): **+** → **Reel** → видео → **Next** → на share екрана включи **Trial** → **Share**. Показва се първо само на не-последователи; изисква публичен професионален акаунт, обикновено ≥1,000 последователи; статистики след 24 ч; **Share with everyone** го пуска на всички.

Insights (мобилно): отвори Reel → **View insights** → **Views** (followers vs non-followers), **Watch time**, **Average watch time**, карта **Retention** с кривата и **Skip rate** (добавен август 2025: % отметнати в първите 3 сек.; под ~30–40% е здравословно, над ~50% — hook-ът е паднал), **Interactions** (Likes, Comments, Shares, Saves), **Profile activity** (Profile visits, External link taps). Сравнение между Reels: **Professional dashboard** → **Insights** → **Content you shared** → филтър **Reels** → сортирай по метрика. CSV: business.facebook.com → **Insights** → **Content** → **Export data**.

### 5.3 YouTube Shorts и дълги видеа (YouTube Studio)

1. `https://studio.youtube.com` → **Create** (горе вдясно) → **Upload videos** → **Select files**.
2. **Details**: **Title** (≤100 знака; при Shorts това е видимият текст), **Description** (първи ред = hook, после линкове с UTM, после атрибуция за музика), **Thumbnail** — за Shorts десктопът дава три кадъра; custom Shorts thumbnails тръгнаха към YPP канали от 24 юли 2026 и може да не са налични при теб (заобикаляне: дизайнерски title кадър на 0:00 и го избираш). Дълги: **Upload thumbnail** (1280×720 PNG). **Playlists**. **Audience** → **No, it's not made for kids**.
3. **Show more**: **Paid promotion** (само ако трета страна ти е платила; affiliate линковете сами по себе си не го изискват), AI декларацията — в **Attributes** като **AI use** с **Yes/No** (старо: **Altered content**; редизайнът на Studio от юли 2026 премести неща — гледай на двете места). За TTS глас върху реален екран много creators отговарят Yes. **Tags**, **Language**, **Category**, **Comments and ratings**.
4. **Next** → **Video elements** (**Add subtitles** → качи SRT-а от Kdenlive) → **Next** → **Checks** → **Next** → **Visibility**: **Public** / **Unlisted** / **Private** или **Schedule** → **Schedule** / **Publish**.
5. Shorts се разпознават автоматично: вертикално или квадратно и **≤ 3 минути**. `#Shorts` не е задължителен.
6. Свързване на Short с дълго видео: **Content** → кликни Short-а → **Related video** → избери публично/unlisted видео от канала → **SAVE**. Иска **advanced features** (телефонна верификация). Зрителят вижда кликаем линк под името ти.

Анализи: **Analytics** → **Content** → чип **Shorts** → **Shown in feed** и **Viewed vs. swiped away** (≥70% viewed е силно, <30% умира във feed-а). По видео: **Content** → видео → **Analytics** → **Overview** / **Reach** / **Engagement**; **Key moments for audience retention** (**Intro**, **Top moments**, **Spikes**, **Dips**) — само за видеа ≥60 сек. с ≥100 гледания, т.е. за дългите; за Shorts чети **Average percentage viewed** и кривата. Експорт: **ADVANCED MODE** → **Export current view** → **.csv** / **Google Sheets**.

### 5.4 LinkedIn

Спецификации (официална страница за Pages): макс **10 минути**, **5 GB**, 256×144 до 4096×2304, съотношение **1:2.4 до 2.4:1** (9:16 влиза), 10–60 fps, MP4/WebM/MKV; MOV и AVI вече не се приемат. Лични постове от десктоп — до 15 мин. по повечето ръководства **(провери)**.

1. linkedin.com → **Start a post** → иконата за медия → MP4 → **Next** → текст (hook първи ред; субтитрите вградени — повечето LinkedIn видео се гледа без звук) → **Post**.
2. LinkedIn пусна вертикален видео feed на мобилно през 2025 и го разшири към десктоп; нативното 9:16 получава мястото.
3. Линкове: изследването на Richard van der Blom за 2026 (1.3M поста) мери −18.8% медианен обхват за един външен линк в тялото и намалена видимост и на „линк в първия коментар“; други данни от 2026 не са съгласни. Практически: когда целта е трафик, слагай линка в тялото с UTM и приемай удара; иначе — **Custom button** / **Website** в профила.

### 5.5 Safe зони и файлов спец (всички 9:16 платформи)

Работни полета върху кадър 1080×1920 (измерени, не официални — платформите местят интерфейса):

| Платформа | Горе | Долу | Ляво | Дясно |
|---|---|---|---|---|
| TikTok | ~130–250 px | ~250–400 px (описание + звук + прогрес) | ~60 px | ~180–200 px (колоната с икони; бутонът „Add to Playlist“ от януари 2026 я разшири) |
| Instagram Reels | ~108–140 px | ~320–500 px | ~60 px | ~120–180 px |
| YouTube Shorts | ~120 px | ~320 px | ~60 px | ~140 px |

**Универсално правило за всичките три:** всяка дума в x = 60…880 и y = 250…1500. Вградените субтитри центрирани хоризонтално с център около **y ≈ 1300–1400** (68–73% надолу). Hook текстът на y ≈ 350–600. Никога не разчитай на долните 420 px и десните 200 px. Вътрешната safe рамка на Kdenlive е добро приближение.

Файл: MP4, H.264 High, 1080×1920, 30 fps константно, 8–12 Mbps, AAC 48 kHz стерео 192–320 kbps.

---

## 6. Измерване и итерация

### 6.1 Карта на метриките

| Въпрос | TikTok Studio → Analytics → Content → видео | Instagram Reel → View insights | YouTube Studio → Analytics |
|---|---|---|---|
| Hook (първите 3 сек.) | **Retention rate** графика при 0–3 сек. | **Skip rate** (карта Retention) | **Viewed vs. swiped away** (Content → Shorts); **Intro** в Key moments за ≥60 сек. |
| Довършване | **Watched full video** %, **Average watch time** спрямо дължината | **Average watch time** спрямо дължината; кривата | **Average percentage viewed**, кривата |
| Повторни гледания | сегменти от кривата >100%, **Total play time** / views | **Watch time** включва replays | **Spikes**; average percentage viewed >100% |
| Споделяния/записвания | **Shares**, **Saves** | **Interactions** → **Shares**, **Saves** | **Engagement** → **Shares**, **Likes** |
| Профил → клик | **Overview** → **Profile views**; кликовете по линк не са надеждни → UTM | **Profile activity** → **Profile visits**, **External link taps** | няма метрика за клик → UTM |

Една таблица (Google Sheets или LibreOffice): дата, платформа, hook текст, дължина, 3-сек. retention, довършване, shares, saves, кликове по линк (от 6.2).

### 6.2 Безплатно следене на линковете

UTM: **Campaign URL Builder** `https://ga-dev-tools.web.app/ga4/campaign-url-builder/`. Полета: **website URL**, **campaign source** (`tiktok`, `instagram`, `youtube`, `linkedin`), **campaign medium** (`bio` или `caption`, за да разделиш местата), **campaign name** (slug на видеото, напр. `2026-09-05-mailchimp-alt`). Пример: `https://unsolero.com/guides/mailchimp-alternatives?utm_source=tiktok&utm_medium=bio&utm_campaign=2026-09-05-mailchimp-alt`.

Линк в био: вместо Linktree Free (бадж „Powered by Linktree“) — страница `https://unsolero.com/links` с по един бутон за всяко текущо видео, всеки със свой UTM, плюс къси пренасочвания `unsolero.com/tt` → 302 към UTM адреса. Кликът каца в твоите анализи веднага и нищо не стои между зрителя и сайта. (Това е и точка от списъка с подобрения в Проучване 1.)

Четене:
- **GA4**: analytics.google.com → **Reports** → **Acquisition** → **Traffic acquisition** → първа дименсия **Session source / medium**, после **Session campaign**.
- **Umami** (self-hosted, влиза в €5 VPS: Node + Postgres, ~512 MB): `git clone https://github.com/umami-software/umami.git && cd umami && docker compose up -d` → `http://<vps>:3000` → **admin** / **umami** (смени го) → **Settings** → **Websites** → **Add website** → **Tracking code** → `<script>` в `<head>` на сайта. UTM: **Reports** → **Create report** → **UTM** **(провери името)**. MIT.
- **Plausible CE** иска ClickHouse и ≥2 GB RAM — не за този VPS.

### 6.3 Правило за решение (в 48 ч. след всяко видео)

- 3-сек. retention <60% (TikTok) / skip rate >45% (IG) / viewed <40% (YT) → проблемът е първата 1.5 сек. Пренапиши hook текста и пресни само началото.
- Hook-ът е добре, довършването е слабо → съкрати с 20%.
- Shares/saves високи, кликове ниски → проблемът е CTA/био пътят, не видеото.
