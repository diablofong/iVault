const translations = {
    en: {
        'nav.download':          'Download',
        'trust.free':            'Free',
        'trust.opensource':      'Open Source',
        'hero.title':            'Your iPhone photos,\nbacked up over USB.',
        'hero.subtitle':         'No iCloud. No iTunes sync. Just plug in and go.',
        'hero.cta':              'Free Download',
        'hero.cta.note':         'No account · No subscription · Original quality',
        'how.title':             'How it works',
        'step.1.title':          'Connect your iPhone',
        'step.1.desc':           'Plug in via USB — no Wi-Fi, no account, no setup required.',
        'step.2.title':          'Tap Trust on your iPhone',
        'step.2.desc':           'A one-time prompt. Your computer stays trusted forever.',
        'step.3.title':          'Backup starts',
        'step.3.desc':           'Photos transfer at full USB speed, sorted into monthly folders automatically.',
        'features.title':        'Simple by design',
        'feature.1.title':       'USB direct transfer',
        'feature.1.desc':        'Uses Apple\'s AFC protocol — no sync, no cloud, no account required.',
        'feature.2.title':       'Auto monthly folders',
        'feature.2.desc':        'Photos sorted by EXIF shoot date into monthly folders, automatically.',
        'feature.3.title':       'Resume anytime',
        'feature.3.desc':        'Interrupted backups pick up exactly where they left off.',
        'screenshots.title':     'See it in action',
        'screenshot.1':          'First launch',
        'screenshot.2':          'Device ready',
        'screenshot.3':          'Backing up',
        'screenshot.4':          'All done',
        'download.title':        'Download iVault',
        'download.free':         'Free & open source — always.',
        'download.mac':          'Download for macOS',
        'download.win':          'Download for Windows',
        'download.mac.note':     'If blocked by Gatekeeper: System Settings → Privacy & Security → Open Anyway',
        'download.win.note':     'If blocked by SmartScreen: More info → Run anyway.\nApple Devices (free, Microsoft Store) will be set up on first launch.',
        'download.source':       'View source on GitHub',
        'footer.tagline':        'Built by a developer, for people who just want their photos safe.',
        'footer.license':        'Apache 2.0',
        'footer.github':         'GitHub',
    },
    'zh-TW': {
        'nav.download':          '下載',
        'trust.free':            '免費',
        'trust.opensource':      '開源',
        'hero.title':            '透過 USB 備份\niPhone 照片。',
        'hero.subtitle':         '不需 iCloud，不需 iTunes 同步，插上就能用。',
        'hero.cta':              '免費下載',
        'hero.cta.note':         '不需帳號 · 不需訂閱 · 原始畫質',
        'how.title':             '使用方式',
        'step.1.title':          '連接 iPhone',
        'step.1.desc':           '用 USB 插上，不需 Wi-Fi、不需帳號、不需設定。',
        'step.2.title':          '點選信任',
        'step.2.desc':           'iPhone 上出現一次性提示，點信任即可，之後自動記住。',
        'step.3.title':          '開始備份',
        'step.3.desc':           '以 USB 速度傳輸，照片自動依月份分類整理。',
        'features.title':        '極簡設計',
        'feature.1.title':       'USB 直接傳輸',
        'feature.1.desc':        '使用 Apple AFC 協定，無需同步、無需雲端、無需帳號。',
        'feature.2.title':       '自動月份分類',
        'feature.2.desc':        '依 EXIF 拍攝日期自動按月份建立資料夾。',
        'feature.3.title':       '隨時繼續',
        'feature.3.desc':        '中斷的備份會從中斷處繼續，不重複傳輸。',
        'screenshots.title':     '實際操作畫面',
        'screenshot.1':          '首次啟動',
        'screenshot.2':          '裝置就緒',
        'screenshot.3':          '備份中',
        'screenshot.4':          '備份完成',
        'download.title':        '下載 iVault',
        'download.free':         '永久免費，完全開源。',
        'download.mac':          '下載 macOS 版',
        'download.win':          '下載 Windows 版',
        'download.mac.note':     '若被 Gatekeeper 擋住：系統設定 → 隱私權與安全性 → 仍要打開',
        'download.win.note':     '若被 SmartScreen 擋住：更多資訊 → 仍要執行。\nApple Devices（免費，Microsoft Store）將於首次啟動時自動引導安裝。',
        'download.source':       '在 GitHub 查看原始碼',
        'footer.tagline':        '一個開發者做的工具，只為讓大家的照片安全留存。',
        'footer.license':        'Apache 2.0',
        'footer.github':         'GitHub',
    }
};

function detectLang() {
    const stored = localStorage.getItem('ivault-lang');
    if (stored && translations[stored]) return stored;
    if (navigator.language.startsWith('zh')) return 'zh-TW';
    return 'en';
}

let currentLang = detectLang();

export function t(key) {
    return translations[currentLang]?.[key] ?? translations['en'][key] ?? key;
}

export function setLang(lang) {
    currentLang = lang;
    localStorage.setItem('ivault-lang', lang);
    renderAll();
}

export function getLang() { return currentLang; }

export function renderAll() {
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        const val = t(key);
        if (!val) return;
        if (val.includes('\n')) {
            el.innerHTML = val.replace(/\n/g, '<br>');
        } else {
            el.textContent = val;
        }
    });

    const heroTitle = document.getElementById('hero-title');
    if (heroTitle) heroTitle.innerHTML = t('hero.title').replace(/\n/g, '<br>');

    document.querySelectorAll('[data-lang]').forEach(btn => {
        btn.classList.toggle('active', btn.getAttribute('data-lang') === currentLang);
    });

    document.documentElement.lang = currentLang === 'zh-TW' ? 'zh-TW' : 'en';
}
