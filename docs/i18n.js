const translations = {
    en: {
        'nav.download':          'Download',
        'trust.platform':        'Windows 11 / 10',
        'trust.free':            'Free',
        'trust.opensource':      'Open Source',
        'hero.title':            'Free iPhone photo backup\nfor Windows.',
        'hero.subtitle':         'Built because Microsoft Photos keeps failing\nand iCloud runs out of space.',
        'hero.tagline':          'Free, open source. No iCloud. No iTunes sync.',
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
        'compare.title':         'How iVault compares',
        'compare.feature':       'Feature',
        'compare.ivault':        'iVault',
        'compare.msphotos':      'Microsoft Photos',
        'compare.icloud':        'iCloud',
        'compare.price':         'Price',
        'compare.price.ivault':  'Free',
        'compare.price.msphotos':'Free',
        'compare.price.icloud':  'Free (5 GB) / Subscription',
        'compare.transfer':      'Transfer method',
        'compare.transfer.ivault':'USB (AFC direct)',
        'compare.transfer.msphotos':'USB (PTP/MTP — often fails)',
        'compare.transfer.icloud':'iCloud sync',
        'compare.storage':       'Storage location',
        'compare.storage.ivault':'Your PC',
        'compare.storage.msphotos':'Your PC',
        'compare.storage.icloud':'iCloud (5 GB free)',
        'compare.opensource':    'Open source',
        'compare.yes':           '✓',
        'compare.no':            '—',
        'download.title':        'Download iVault',
        'download.free':         'Free & open source — always.',
        'download.win':          'Download for Windows',
        'download.win.note':     'If blocked by SmartScreen: More info → Run anyway.\nApple Devices (free, Microsoft Store) will be set up on first launch.',
        'download.source':       'View source on GitHub',
        'footer.tagline':        'Free iPhone photo backup for Windows. Open source.',
        'footer.privacy':        'Privacy',
        'footer.license':        'Apache 2.0',
        'footer.github':         'GitHub',
        'req.title':             'System Requirements',
        'req.os':                'OS',
        'req.driver':            'Driver',
        'req.driver.note':       '(free, Microsoft Store)',
        'req.runtime':           'Runtime',
        'req.runtime.note':      '(built into Windows 11; Windows 10 may need separate install)',
        'faq.title':             'FAQ',
        'faq.q1':                'Do I need iTunes?',
        'faq.a1':                'No. iVault uses Apple Devices — a newer, lighter app from Apple. iTunes can conflict with iVault, so close it if it\'s running.',
        'faq.q2':                'Where are my photos saved?',
        'faq.a2':                'To any folder you choose — default is the largest available drive. Photos are organized into YYYY-MM/ subfolders by shoot date.',
        'faq.q3':                'What if iCloud Optimize Storage is on?',
        'faq.a3':                'iCloud Optimize Storage keeps only thumbnails on the iPhone. iVault will back those up. For full-resolution originals, disable Optimize Storage in iPhone Settings → Photos first.',
        'faq.q4':                'What if my cable comes loose mid-backup?',
        'faq.a4':                'iVault saves progress and resumes exactly where it left off — no duplicates, no starting over.',
        'faq.q5':                'Is my data private?',
        'faq.a5':                'Yes. iVault transfers photos locally over USB and collects no data whatsoever. No account, no analytics, no telemetry.',
        'faq.privacy_link':      'privacy policy',
        'faq.q6':                'Why does Windows show a security warning?',
        'faq.a6':                'iVault doesn\'t have paid code signing. Click "More info" → "Run anyway". The source is fully open on GitHub.',
        'kofi.note':             'iVault is free to use. If it saves you time, consider buying me a coffee.',
        'kofi.btn':              '☕ Support on Ko-fi',
    },
    'zh-TW': {
        'nav.download':          '下載',
        'trust.platform':        'Windows 11 / 10',
        'trust.free':            '免費',
        'trust.opensource':      '開源',
        'hero.title':            'Windows 的免費\niPhone 照片備份工具。',
        'hero.subtitle':         '因為 Microsoft Photos 老是失敗，\niCloud 5GB 又不夠用。',
        'hero.tagline':          '免費、開源。不碰 iCloud，不需 iTunes 同步。',
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
        'compare.title':         '與其他工具比較',
        'compare.feature':       '功能',
        'compare.ivault':        'iVault',
        'compare.msphotos':      'Microsoft Photos',
        'compare.icloud':        'iCloud',
        'compare.price':         '價格',
        'compare.price.ivault':  '免費',
        'compare.price.msphotos':'免費',
        'compare.price.icloud':  '免費（5 GB）/ 訂閱',
        'compare.transfer':      '傳輸方式',
        'compare.transfer.ivault':'USB（AFC 直傳）',
        'compare.transfer.msphotos':'USB（PTP/MTP — 常出錯）',
        'compare.transfer.icloud':'iCloud 同步',
        'compare.storage':       '儲存位置',
        'compare.storage.ivault':'本機',
        'compare.storage.msphotos':'本機',
        'compare.storage.icloud':'iCloud（5 GB 免費）',
        'compare.opensource':    '開源',
        'compare.yes':           '✓',
        'compare.no':            '—',
        'download.title':        '下載 iVault',
        'download.free':         '永久免費，完全開源。',
        'download.win':          '下載 Windows 版',
        'download.win.note':     '若被 SmartScreen 擋住：更多資訊 → 仍要執行。\nApple Devices（免費，Microsoft Store）將於首次啟動時自動引導安裝。',
        'download.source':       '在 GitHub 查看原始碼',
        'footer.tagline':        'Windows 的 iPhone 照片備份工具。免費、開源。',
        'footer.privacy':        '隱私政策',
        'footer.license':        'Apache 2.0',
        'footer.github':         'GitHub',
        'req.title':             '系統需求',
        'req.os':                '作業系統',
        'req.driver':            '驅動程式',
        'req.driver.note':       '（免費，Microsoft Store）',
        'req.runtime':           '執行環境',
        'req.runtime.note':      '（Windows 11 內建；Windows 10 可能需要另行安裝）',
        'faq.title':             '常見問題',
        'faq.q1':                '需要安裝 iTunes 嗎？',
        'faq.a1':                '不需要。iVault 使用 Apple Devices（Apple 的新版驅動 App）。iTunes 反而可能衝突，若有開啟請先關閉。',
        'faq.q2':                '照片會存到哪裡？',
        'faq.a2':                '存到你選擇的資料夾（預設為最大的可用磁碟），並自動按拍攝月份整理成 YYYY-MM/ 子資料夾。',
        'faq.q3':                '如果開了 iCloud 最佳化儲存怎麼辦？',
        'faq.a3':                '最佳化儲存會讓 iPhone 只保留縮圖。iVault 會備份這些縮圖。若要完整原始檔，請先到「設定 → 照片」關閉最佳化儲存。',
        'faq.q4':                '備份中途 USB 斷掉會怎樣？',
        'faq.a4':                'iVault 記錄進度，下次插上後從中斷點繼續，不重複備份。',
        'faq.q5':                '我的資料安全嗎？',
        'faq.a5':                '是的。iVault 完全在本機 USB 傳輸，不收集任何資料，沒有帳號、沒有分析、沒有遙測。',
        'faq.privacy_link':      '隱私政策',
        'faq.q6':                '為什麼 Windows 顯示安全警告？',
        'faq.a6':                'iVault 沒有付費程式碼簽章。點「更多資訊」→「仍要執行」即可。原始碼完全開放可在 GitHub 查閱。',
        'kofi.note':             'iVault 永遠免費使用。若覺得好用，可以請我喝杯咖啡支持開發。',
        'kofi.btn':              '☕ 在 Ko-fi 贊助',
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

// setTextWithBreaks 以 DOM API 將多行字串寫入元素，保留換行為 <br>。
// 避免用 innerHTML — 縱深防禦：若未來翻譯值來自遠端或動態來源，不會被當成 HTML 解析。
function setTextWithBreaks(el, val) {
    el.replaceChildren();
    const parts = val.split('\n');
    parts.forEach((part, i) => {
        if (part) el.appendChild(document.createTextNode(part));
        if (i < parts.length - 1) el.appendChild(document.createElement('br'));
    });
}

export function renderAll() {
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        const val = t(key);
        if (!val) return;
        if (val.includes('\n')) {
            setTextWithBreaks(el, val);
        } else {
            el.textContent = val;
        }
    });

    const heroTitle = document.getElementById('hero-title');
    if (heroTitle) setTextWithBreaks(heroTitle, t('hero.title'));

    document.querySelectorAll('[data-lang]').forEach(btn => {
        btn.classList.toggle('active', btn.getAttribute('data-lang') === currentLang);
    });

    document.documentElement.lang = currentLang === 'zh-TW' ? 'zh-TW' : 'en';
}
