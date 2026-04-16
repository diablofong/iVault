import { t, setLang as setLangCore, getLang, renderAll } from './i18n.js';

function setLang(lang) {
    setLangCore(lang);
    updateScreenshots(lang);
}

// 以 addEventListener 取代 inline onclick，讓 CSP 可以維持嚴格的 script-src 'self'（無 'unsafe-inline'）。
document.querySelectorAll('[data-lang]').forEach(btn => {
    btn.addEventListener('click', () => setLang(btn.getAttribute('data-lang')));
});

const SS_MAP = {
    en: {
        img: {
            'hero-ss':  'screenshots/ready-en.png',
            'ss-1-img': 'screenshots/idle-first-en.png',
            'ss-2-img': 'screenshots/ready-en.png',
            'ss-3-img': 'screenshots/backing-up-en.png',
            'ss-4-img': 'screenshots/idle-returning-en.png',
        },
        webp: {
            'hero-ss-src':  'screenshots/ready-en.webp',
            'ss-1-src': 'screenshots/idle-first-en.webp',
            'ss-2-src': 'screenshots/ready-en.webp',
            'ss-3-src': 'screenshots/backing-up-en.webp',
            'ss-4-src': 'screenshots/idle-returning-en.webp',
        },
    },
    'zh-TW': {
        img: {
            'hero-ss':  'screenshots/ready-zh.png',
            'ss-1-img': 'screenshots/idle-first-zh.png',
            'ss-2-img': 'screenshots/ready-zh.png',
            'ss-3-img': 'screenshots/backing-up-zh.png',
            'ss-4-img': 'screenshots/idle-returning-zh.png',
        },
        webp: {
            'hero-ss-src':  'screenshots/ready-zh.webp',
            'ss-1-src': 'screenshots/idle-first-zh.webp',
            'ss-2-src': 'screenshots/ready-zh.webp',
            'ss-3-src': 'screenshots/backing-up-zh.webp',
            'ss-4-src': 'screenshots/idle-returning-zh.webp',
        },
    }
};

function updateScreenshots(lang) {
    const map = SS_MAP[lang] || SS_MAP['en'];
    for (const [id, src] of Object.entries(map.img)) {
        const el = document.getElementById(id);
        if (el) el.src = src;
    }
    for (const [id, srcset] of Object.entries(map.webp)) {
        const el = document.getElementById(id);
        if (el) el.srcset = srcset;
    }
}

// Scroll reveal (whole sections)
const observer = new IntersectionObserver((entries) => {
    entries.forEach(e => {
        if (e.isIntersecting) {
            e.target.classList.add('visible');
            observer.unobserve(e.target);
        }
    });
}, { threshold: 0.08 });

document.querySelectorAll('.reveal').forEach(el => observer.observe(el));

// Stagger reveal (grids)
const staggerObs = new IntersectionObserver((entries) => {
    entries.forEach(e => {
        if (e.isIntersecting) {
            e.target.classList.add('visible');
            staggerObs.unobserve(e.target);
        }
    });
}, { threshold: 0.1 });

document.querySelectorAll('.steps-grid, .features-grid').forEach(el => staggerObs.observe(el));

renderAll();
updateScreenshots(getLang());
