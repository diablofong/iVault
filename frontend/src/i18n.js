import zhTW from './locales/zh-TW.js';
import en   from './locales/en.js';

const LOCALES = { 'zh-TW': zhTW, en };

function detectLang() {
    const saved = localStorage.getItem('ivault-lang');
    if (saved && LOCALES[saved]) return saved;
    const nav = (navigator.language || 'zh-TW').toLowerCase();
    return nav.startsWith('zh') ? 'zh-TW' : 'en';
}

let currentLang = detectLang();

/** 取得翻譯字串，找不到時回傳 key 本身 */
export function t(key) {
    return LOCALES[currentLang]?.[key] ?? LOCALES['zh-TW']?.[key] ?? key;
}

/** 取得當前語言代碼 */
export function getLang() { return currentLang; }

/** 切換語言並重新渲染所有標記元素 */
export function setLang(lang) {
    if (!LOCALES[lang]) return;
    currentLang = lang;
    localStorage.setItem('ivault-lang', lang);
    renderAll();
}

/**
 * 掃描所有帶 data-i18n 屬性的元素並更新 textContent。
 * 動態填入的文字（如進度數字）不走這裡，由 main.js 各 handler 用 t() 直接設定。
 */
export function renderAll() {
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.dataset.i18n;
        const val = t(key);
        if (val !== key) el.textContent = val;
    });

    // 更新語言切換 toggle active 狀態
    document.getElementById('lang-opt-zh')?.classList.toggle('active', currentLang === 'zh-TW');
    document.getElementById('lang-opt-en')?.classList.toggle('active', currentLang === 'en');
}
