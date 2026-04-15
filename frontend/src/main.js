import './style.css';
import './app.css';
import { WindowMinimise, Quit, EventsOn } from '../wailsjs/runtime/runtime';
import {
    GetPlatformInfo,
    GetConnectedDevice,
    GetDeviceDetail,
    CheckTrustStatus,
    TriggerTrustCheck,
    CheckAppleDevicesInstalled,
    SelectBackupFolder,
    GetDefaultBackupPath,
    GetDiskInfo,
    ManifestExists,
    EstimateBackupSize,
    StartBackup,
    CancelBackup,
    LoadConfig,
    OpenFolder,
    OpenURL,
    InstallAppleDevices,
    InstallHeicCodec,
} from '../wailsjs/go/main/App';
import { t, renderAll, setLang, getLang } from './i18n.js';

// ============================================================
// 狀態機
// ============================================================
const STATES = ['idle', 'device-found', 'trust-guide', 'driver-missing', 'ready', 'backing-up', 'done', 'error'];

let currentState = null;
let currentDevice = null;
let selectedPath = '';
let backupResult = null;
let platformInfo = null;
let appConfig = null;
let trustHintTimer = null;
let trustHardHintTimer = null;

function setState(newState, data = {}) {
    if (!STATES.includes(newState)) {
        console.error('Unknown state:', newState);
        return;
    }

    if (currentState) {
        document.getElementById(`view-${currentState}`)?.classList.remove('active');
    }

    // 離開 trust-guide 時清除提示計時器
    if (currentState === 'trust-guide') {
        if (trustHintTimer) { clearTimeout(trustHintTimer); trustHintTimer = null; }
        if (trustHardHintTimer) { clearTimeout(trustHardHintTimer); trustHardHintTimer = null; }
    }

    // AMDS 提示只在 IDLE 時有效
    if (newState !== 'idle') {
        const amdsEl = document.getElementById('amds-status');
        if (amdsEl) amdsEl.style.display = 'none';
    }

    currentState = newState;
    document.getElementById(`view-${newState}`)?.classList.add('active');

    onEnterState(newState, data);
}

function onEnterState(state, data) {
    switch (state) {
        case 'idle':           onEnterIdle();              break;
        case 'device-found':   onEnterDeviceFound(data);   break;
        case 'trust-guide':    onEnterTrustGuide(data);    break;
        case 'ready':          onEnterReady(data);          break;
        case 'done':           onEnterDone(data);           break;
        case 'error':          onEnterError(data);          break;
    }
}

// ============================================================
// 初始化
// ============================================================
async function init() {
    document.getElementById('btn-minimize')?.addEventListener('click', () => WindowMinimise());
    document.getElementById('btn-close')?.addEventListener('click', () => Quit());

    try {
        [platformInfo, appConfig] = await Promise.all([GetPlatformInfo(), LoadConfig()]);
        document.body.classList.add(`os-${platformInfo.os}`);
        if (platformInfo.darkMode) document.body.classList.add('dark');
    } catch (e) {
        console.error('init:', e);
    }

    renderAll();
    registerEvents();
    bindHandlers();

    if (platformInfo?.os === 'windows') {
        try {
            const installed = await CheckAppleDevicesInstalled();
            if (!installed) {
                setState('driver-missing');
                return;
            }
        } catch (e) {}
    }

    try {
        const dev = await GetConnectedDevice();
        if (dev && dev.udid) {
            currentDevice = dev;
            setState('device-found', dev);
        } else {
            setState('idle');
        }
    } catch (e) {
        setState('idle');
    }
}

// ============================================================
// Wails Event 監聽
// ============================================================
function registerEvents() {
    EventsOn('device:connected', (info) => {
        currentDevice = info;
        setState('device-found', info);
    });

    EventsOn('device:disconnected', () => {
        currentDevice = null;
        if (currentState !== 'backing-up') {
            setState('idle');
        }
    });

    EventsOn('device:trust-changed', (data) => {
        if (data.trusted && currentState === 'trust-guide') {
            setState('ready', currentDevice || { udid: data.udid });
        }
    });

    EventsOn('trust:timeout', () => {
        if (currentState === 'trust-guide') {
            setState('error', {
                code: 'TRUST_TIMEOUT',
                message: t('error.TRUST_TIMEOUT'),
                recoverable: true,
            });
        }
    });

    EventsOn('backup:progress', (progress) => {
        updateProgressUI(progress);
    });

    EventsOn('backup:complete', (result) => {
        backupResult = result;
        // E-10: 更新前端 appConfig 快取，避免回 READY 時顯示舊歷史
        if (appConfig) {
            appConfig.lastInterrupted = false;
            appConfig.firstBackupDone = true;
            appConfig.history = [{
                date: new Date().toISOString(),
                deviceName: currentDevice?.name || 'iPhone',
                deviceUdid: currentDevice?.udid || '',
                photosCount: result.photosCount ?? 0,
                videosCount: result.videosCount ?? 0,
                newFiles: result.newFiles ?? 0,
                skipped: result.skippedFiles ?? 0,
                failed: result.failedFiles ?? 0,
                totalBytes: result.totalBytes ?? 0,
                duration: result.duration || '',
            }, ...(appConfig.history ?? [])];
        }
        setState('done', result);
    });

    EventsOn('backup:interrupted', (result) => {
        // E-11: 取消後若 iPhone 仍插著 → 直接回 READY，不繞 IDLE
        if (appConfig) {
            appConfig.lastInterrupted = true;
            if (result) {
                appConfig.interruptedDone = result.interruptedDone ?? 0;
                appConfig.interruptedTotal = result.interruptedTotal ?? 0;
            }
        }
        if (currentDevice?.udid) {
            setState('ready', currentDevice);
        } else {
            setState('idle');
        }
    });

    EventsOn('backup:error', (err) => {
        if (appConfig) appConfig.lastInterrupted = true;
        if (currentState === 'backing-up') {
            setState('error', err);
        }
    });

    EventsOn('driver:required', (data) => {
        if (currentState !== 'driver-missing') setState('driver-missing');
        // WMI 偵測到裝置時不顯示 placeholder（P0-6：只在有偵測到時才顯示）
        if (data?.deviceName) {
            // 可在 driver view 附近加設備名稱，目前簡化版不顯示
        }
    });

    EventsOn('driver:installed', () => {
        const pending = document.getElementById('install-pending');
        const initial = document.getElementById('install-initial');
        const success = document.getElementById('install-success');
        if (pending) pending.style.display = 'none';
        if (initial) initial.style.display = 'none';
        if (success) success.style.display = '';
    });

    EventsOn('amds:starting', () => {
        const el = document.getElementById('amds-status');
        if (el) el.style.display = '';
    });

    EventsOn('amds:start_failed', () => {
        if (currentState !== 'error') {
            setState('error', {
                code: 'AMDS_START_FAILED',
                message: t('error.amds_desc'),
                recoverable: true,
            });
        }
    });

    EventsOn('amds:timeout', () => {
        if (currentState !== 'error') {
            setState('error', {
                code: 'AMDS_TIMEOUT',
                message: t('error.AMDS_TIMEOUT'),
                recoverable: true,
            });
        }
    });

    EventsOn('heic:progress', (data) => {
        const section = document.getElementById('heic-convert-section');
        const bar = document.getElementById('heic-progress-bar');
        const label = document.getElementById('heic-convert-label');
        if (section) section.style.display = '';
        if (bar) bar.style.width = (data.percent ?? 0).toFixed(1) + '%';
        if (label) label.textContent = `${t('heic.converting')} ${data.done ?? 0} / ${data.total ?? 0}`;
    });

    EventsOn('heic:complete', (data) => {
        const section = document.getElementById('heic-convert-section');
        const label = document.getElementById('heic-convert-label');
        if (data?.converted > 0) {
            if (label) label.textContent = `${t('heic.done')} ${fmt(data.converted)} ${t('heic.unit')}`;
        } else {
            if (section) section.style.display = 'none';
        }
    });
}

// ============================================================
// 按鈕綁定
// ============================================================
function bindHandlers() {
    document.getElementById('btn-lang-toggle')?.addEventListener('click', () => {
        setLang(getLang() === 'zh-TW' ? 'en' : 'zh-TW');
        if (currentState === 'idle') onEnterIdle();
        if (currentState === 'ready') onEnterReady(currentDevice || {});
    });

    // DRIVER_MISSING
    document.getElementById('btn-install-driver')?.addEventListener('click', () => {
        InstallAppleDevices();
        document.getElementById('install-initial').style.display = 'none';
        const pending = document.getElementById('install-pending');
        if (pending) {
            pending.style.display = '';
            pending.querySelectorAll('[data-i18n]').forEach(el => {
                const key = el.getAttribute('data-i18n');
                const val = t(key);
                if (val) el.textContent = val;
            });
        }
    });

    document.getElementById('btn-recheck-driver')?.addEventListener('click', async () => {
        const btn = document.getElementById('btn-recheck-driver');
        const failMsg = document.getElementById('install-recheck-fail');
        if (btn) btn.disabled = true;
        if (failMsg) failMsg.style.display = 'none';
        try {
            const installed = await CheckAppleDevicesInstalled();
            if (installed) {
                document.getElementById('install-pending').style.display = 'none';
                document.getElementById('install-success').style.display = '';
            } else {
                if (failMsg) failMsg.style.display = '';
            }
        } catch (e) {
            if (failMsg) failMsg.style.display = '';
        } finally {
            if (btn) btn.disabled = false;
        }
    });

    document.getElementById('btn-replug-done')?.addEventListener('click', () => {
        resetDriverMissingView();
        setState('idle');
    });

    document.getElementById('btn-faq-toggle')?.addEventListener('click', () => {
        const content = document.getElementById('driver-faq-content');
        const icon = document.getElementById('faq-toggle-icon');
        if (!content) return;
        const expanded = content.style.display !== 'none';
        content.style.display = expanded ? 'none' : '';
        if (icon) icon.textContent = expanded ? '▶' : '▼';
    });

    // TRUST_GUIDE
    document.getElementById('btn-trust-recheck')?.addEventListener('click', async (e) => {
        const btn = e.currentTarget;
        if (!currentDevice?.udid) return;
        btn.disabled = true;
        try {
            const ok = await TriggerTrustCheck(currentDevice.udid);
            if (!ok) {
                // 仍未偵測到：震一下按鈕 + 立即顯示強力提示
                btn.classList.add('shake');
                setTimeout(() => btn.classList.remove('shake'), 600);
                const h = document.getElementById('trust-hard-hint');
                if (h) h.style.display = '';
            }
            // 若 ok，後端會 emit device:trust-changed，現有 handler 會切 ready
        } catch (err) {
            console.error('TriggerTrustCheck failed', err);
        } finally {
            setTimeout(() => { btn.disabled = false; }, 400);
        }
    });

    // READY
    document.getElementById('btn-select-folder')?.addEventListener('click', onSelectFolder);
    document.getElementById('btn-start-backup')?.addEventListener('click', onStartBackup);

    // BACKING_UP
    document.getElementById('btn-cancel-backup')?.addEventListener('click', () => CancelBackup());

    // DONE
    document.getElementById('btn-open-folder')?.addEventListener('click', () => {
        if (backupResult?.backupPath) OpenFolder(backupResult.backupPath);
    });
document.getElementById('btn-backup-again')?.addEventListener('click', () => {
        // iPhone 仍插著 → 直接回 READY，不需要重新連接
        if (currentDevice?.udid) {
            setState('ready', currentDevice);
        } else {
            setState('idle');
        }
    });

    document.getElementById('btn-toggle-failed')?.addEventListener('click', () => {
        const list = document.getElementById('failed-list');
        const icon = document.getElementById('failed-toggle-icon');
        if (!list) return;
        const expanded = list.style.display !== 'none';
        list.style.display = expanded ? 'none' : '';
        if (icon) icon.textContent = expanded ? '▶' : '▼';
    });

    document.getElementById('btn-install-heic')?.addEventListener('click', () => {
        InstallHeicCodec();
    });

    // ERROR
    document.getElementById('btn-retry')?.addEventListener('click', () => {
        if (currentDevice) setState('ready', currentDevice);
        else setState('idle');
    });
    document.getElementById('btn-back-to-idle')?.addEventListener('click', () => setState('idle'));
    document.getElementById('btn-report-issue')?.addEventListener('click', () => {
        OpenURL('https://github.com/diablofong/ivault/issues/new');
    });
}

// ============================================================
// 狀態進入處理
// ============================================================

async function onEnterIdle() {
    const cfg = appConfig;
    const history = cfg?.history ?? [];
    const isInterrupted = cfg?.lastInterrupted === true;
    const isReturning = history.length > 0 && !isInterrupted;

    // 隱藏所有 variant
    document.getElementById('idle-variant-first').style.display = 'none';
    document.getElementById('idle-variant-returning').style.display = 'none';
    document.getElementById('idle-variant-interrupted').style.display = 'none';

    if (isInterrupted) {
        document.getElementById('idle-variant-interrupted').style.display = '';
        const done = cfg?.interruptedDone ?? 0;
        const total = cfg?.interruptedTotal ?? 0;
        const infoEl = document.getElementById('idle-interrupted-info');
        if (infoEl) {
            if (done > 0 && total > 0) {
                infoEl.textContent = (getLang() === 'zh-TW'
                    ? `已備份 ${fmt(done)} / ${fmt(total)}`
                    : `Backed up ${fmt(done)} of ${fmt(total)}`)
                    + ` · ${t('idle.interrupted.progress')}`;
            } else {
                infoEl.textContent = t('idle.interrupted.progress');
            }
        }
    } else if (isReturning) {
        // E-13: 確認 manifest 存在，避免資料夾被刪後還顯示 returning 變體
        const rec = history[0];
        const backupPath = cfg?.lastBackupPath || '';
        const backupUdid = rec?.deviceUdid || '';
        let manifestOk = false;
        if (backupPath && backupUdid) {
            try { manifestOk = await ManifestExists(backupPath, backupUdid); } catch (e) {}
        } else {
            manifestOk = true; // 無法驗證時樂觀顯示（避免舊版 config 無 udid 欄位的使用者倒退）
        }
        if (!manifestOk) {
            document.getElementById('idle-variant-first').style.display = '';
            return;
        }
        document.getElementById('idle-variant-returning').style.display = '';
        const dateStr = formatRelativeDate(rec.date);
        const photos = rec.photosCount ?? 0;
        const videos = rec.videosCount ?? 0;
        const infoEl = document.getElementById('idle-last-backup-info');
        if (infoEl) {
            const lang = getLang();
            if (photos === 0 && videos === 0) {
                infoEl.textContent = lang === 'zh-TW'
                    ? `上次備份：${dateStr} · 全部已是最新`
                    : `Last backed up: ${dateStr} · Everything is up to date`;
            } else if (lang === 'zh-TW') {
                infoEl.textContent = `上次備份：${dateStr} · ${fmt(photos)} 張照片 · ${fmt(videos)} 段影片`;
            } else {
                infoEl.textContent = `Last backed up: ${dateStr} · ${fmt(photos)} photos · ${fmt(videos)} videos`;
            }
        }
    } else {
        document.getElementById('idle-variant-first').style.display = '';
    }
}

async function onEnterDeviceFound(info) {
    setEl('device-name', info.name || 'iPhone');
    setEl('device-ios', `iOS ${info.iosVersion || '-'}`);
    setEl('device-photo-count', t('device.reading'));

    if (platformInfo?.os === 'windows') {
        try {
            const installed = await CheckAppleDevicesInstalled();
            if (!installed) {
                setState('driver-missing');
                return;
            }
        } catch (e) {}
    }

    try {
        const trusted = await CheckTrustStatus(info.udid);
        if (!trusted) {
            setState('trust-guide', info);
            return;
        }
    } catch (e) {}

    fetchPhotoCount(info.udid);
    setState('ready', info);
}

function onEnterTrustGuide() {
    // 重設兩層提示
    const slow = document.getElementById('trust-slow-hint');
    if (slow) slow.style.display = 'none';
    const hard = document.getElementById('trust-hard-hint');
    if (hard) hard.style.display = 'none';

    // 10 秒後淡入次要提示（手機可能沒解鎖）
    trustHintTimer = setTimeout(() => {
        const h = document.getElementById('trust-slow-hint');
        if (h && currentState === 'trust-guide') h.style.display = '';
    }, 10000);

    // 15 秒後顯示強力提示（請拔插 USB）
    trustHardHintTimer = setTimeout(() => {
        const h = document.getElementById('trust-hard-hint');
        if (h && currentState === 'trust-guide') h.style.display = '';
    }, 15000);
}

async function onEnterReady(info) {
    setEl('ready-device-name', info.name || 'iPhone');

    if (!selectedPath) {
        try { selectedPath = await GetDefaultBackupPath(); } catch (e) {}
    }
    setEl('backup-path', selectedPath || '-');
    if (selectedPath) updateDiskInfo(selectedPath);
    if (info.udid && selectedPath) estimateSize(info.udid, selectedPath);

    const checkbox = document.getElementById('convert-heic');
    if (checkbox && appConfig) checkbox.checked = !!appConfig.convertHeic;

    await updateLastBackupRow(selectedPath, info.udid);
}

// updateLastBackupRow 以正確的優先順序更新 READY 頁的 last-backup-row：
// 1. manifest 不在（資料夾被刪/換路徑）→「尚未備份過」（lastInterrupted 一律忽略）
// 2. manifest 在 + lastInterrupted → 顯示中斷狀態
// 3. manifest 在 + 依 UDID 過濾的 history 紀錄：
//    - photos+videos === 0 且 estimateSize > 0 → 不顯示「全部已是最新」（E-12）
//    - 否則依正常邏輯顯示
// 4. 其他 →「尚未備份過」
async function updateLastBackupRow(path, udid) {
    const lastBackupRow = document.getElementById('last-backup-row');
    const lastBackupInfo = document.getElementById('last-backup-info');
    if (!lastBackupRow || !lastBackupInfo) return;

    // Step 1: manifest check（E-6 / E-7）
    let manifestOk = false;
    if (path && udid) {
        try { manifestOk = await ManifestExists(path, udid); } catch (e) {}
    }

    if (!manifestOk) {
        lastBackupInfo.textContent = t('ready.no_backup');
        lastBackupRow.style.display = '';
        return;
    }

    // Step 2: lastInterrupted（E-6）
    if (appConfig?.lastInterrupted) {
        const done = appConfig.interruptedDone ?? 0;
        const total = appConfig.interruptedTotal ?? 0;
        const lang = getLang();
        let infoText;
        if (done > 0 && total > 0) {
            infoText = lang === 'zh-TW'
                ? `上次備份中斷 · 已完成 ${fmt(done)} / ${fmt(total)}`
                : `Last backup interrupted · ${fmt(done)} of ${fmt(total)} done`;
        } else {
            infoText = lang === 'zh-TW' ? '上次備份被中斷' : 'Last backup was interrupted';
        }
        lastBackupInfo.textContent = infoText;
        lastBackupRow.style.display = '';
        return;
    }

    // Step 3: UDID 過濾的 history（E-8）
    const history = appConfig?.history ?? [];
    const rec = udid ? history.find(h => h.deviceUdid === udid) : history[0];
    if (!rec) {
        lastBackupInfo.textContent = t('ready.no_backup');
        lastBackupRow.style.display = '';
        return;
    }

    const dateStr = formatRelativeDate(rec.date);
    const photos = rec.photosCount ?? 0;
    const videos = rec.videosCount ?? 0;
    const lang = getLang();

    if (photos === 0 && videos === 0) {
        // E-12: estimateSize > 0 時不顯示「全部已是最新」
        let estimateBytes = 0;
        if (path && udid) {
            try { estimateBytes = await EstimateBackupSize(udid, path); } catch (e) {}
        }
        if (estimateBytes > 0) {
            lastBackupInfo.textContent = dateStr; // 有新檔待備份，不加「全部已是最新」
        } else {
            lastBackupInfo.textContent = lang === 'zh-TW'
                ? `${dateStr} · 全部已是最新`
                : `${dateStr} · Everything up to date`;
        }
    } else {
        lastBackupInfo.textContent = lang === 'zh-TW'
            ? `${dateStr} · ${fmt(photos)} 張照片 · ${fmt(videos)} 段影片`
            : `${dateStr} · ${fmt(photos)} photos · ${fmt(videos)} videos`;
    }
    lastBackupRow.style.display = '';
}

function onEnterDone(result) {
    const isFirst = !(appConfig?.firstBackupDone);
    const photos = result.photosCount ?? 0;
    const videos = result.videosCount ?? 0;

    // 主標兩版
    const lang = getLang();
    let mainTitle;
    if (isFirst) {
        if (lang === 'zh-TW') {
            mainTitle = `你的 ${fmt(photos)} 張照片和 ${fmt(videos)} 段影片安全了`;
        } else {
            mainTitle = `Your ${fmt(photos)} photos and ${fmt(videos)} videos are safe`;
        }
    } else {
        const newPhotos = result.photosCount ?? 0;
        const newVideos = result.videosCount ?? 0;
        if (lang === 'zh-TW') {
            mainTitle = `新增 ${fmt(newPhotos)} 張照片和 ${fmt(newVideos)} 段影片`;
        } else {
            mainTitle = `${fmt(newPhotos)} new photos and ${fmt(newVideos)} new videos added`;
        }
    }
    setEl('done-main-title', mainTitle);

    // 副標
    const now = new Date();
    const dateStr = lang === 'zh-TW'
        ? `${now.getMonth() + 1} 月 ${now.getDate()} 日`
        : now.toLocaleDateString('en', { month: 'long', day: 'numeric' });
    setEl('done-subtitle', t('done.subtitle').replace('{date}', dateStr));

    // 首次彩蛋
    const eggEl = document.getElementById('done-first-egg');
    if (eggEl) eggEl.style.display = isFirst ? '' : 'none';

    // Bento 四格
    setEl('done-photos-count', fmt(photos));
    setEl('done-videos-count', fmt(videos));
    const totalSizeEl = document.getElementById('done-total-size-value');
    if (totalSizeEl) totalSizeEl.textContent = result.totalBytes > 0 ? formatBytes(result.totalBytes) : '-';
    setEl('done-duration-value', result.duration || '-');

    // 失敗清單
    const failedSection = document.getElementById('failed-section');
    const failedList = document.getElementById('failed-list');
    const failedCount = result.failedFiles ?? 0;
    if (failedSection) failedSection.style.display = failedCount > 0 ? '' : 'none';
    setEl('failed-detail-count', fmt(failedCount));
    if (failedList && failedCount > 0 && Array.isArray(result.failedList)) {
        failedList.innerHTML = '';
        failedList.style.display = 'none';
        document.getElementById('failed-toggle-icon').textContent = '▶';
        result.failedList.forEach(f => {
            const row = document.createElement('div');
            row.className = 'failed-item';
            row.innerHTML = `<span class="failed-name">${escapeHtml(f.fileName)}</span><span class="failed-reason">${escapeHtml(f.reason)}</span>`;
            failedList.appendChild(row);
        });
    }

    // Windows HEIC 提示
    const heicHint = document.getElementById('heic-hint');
    if (heicHint) {
        heicHint.style.display = (platformInfo?.os === 'windows' && result.hasHeic) ? '' : 'none';
    }

    triggerSuccessParticles();
}

function triggerSuccessParticles() {
    const container = document.getElementById('particle-container');
    if (!container) return;
    const icon = document.querySelector('#view-done .success-icon');
    if (icon && container.parentElement) {
        const iconRect = icon.getBoundingClientRect();
        const parentRect = container.parentElement.getBoundingClientRect();
        container.style.left = (iconRect.left - parentRect.left + iconRect.width / 2) + 'px';
        container.style.top = (iconRect.top - parentRect.top + iconRect.height / 2) + 'px';
    }
    container.innerHTML = '';
    const count = 14;
    for (let i = 0; i < count; i++) {
        const p = document.createElement('div');
        p.className = 'particle';
        const angle = (Math.PI * 2 * i) / count + (Math.random() - 0.5) * 0.3;
        const distance = 60 + Math.random() * 30;
        p.style.setProperty('--tx', Math.cos(angle) * distance + 'px');
        p.style.setProperty('--ty', Math.sin(angle) * distance + 'px');
        p.style.animationDelay = (i * 18) + 'ms';
        container.appendChild(p);
    }
    setTimeout(() => { container.innerHTML = ''; }, 1500);
}

function onEnterError(err) {
    const code = err.code || 'UNKNOWN_ERROR';
    const titleEl = document.getElementById('error-title');
    const retryBtn = document.getElementById('btn-retry');
    const issueBtn = document.getElementById('btn-report-issue');

    // 重設按鈕狀態
    if (retryBtn) { retryBtn.style.display = ''; retryBtn.textContent = t('error.retry'); }
    if (issueBtn) issueBtn.style.display = 'none';

    // DEVICE_DISCONNECTED → 1.5 秒後自動回 IDLE
    if (code === 'DEVICE_DISCONNECTED') {
        if (titleEl) titleEl.textContent = t('error.title');
        setEl('error-message', t('error.DEVICE_DISCONNECTED'));
        if (retryBtn) retryBtn.style.display = 'none';
        setTimeout(() => setState('idle'), 1500);
        return;
    }

    // AMDS_START_FAILED → 專屬標題
    if (code === 'AMDS_START_FAILED') {
        if (titleEl) titleEl.textContent = t('error.amds_title');
        setEl('error-message', t('error.amds_desc'));
        if (retryBtn) retryBtn.textContent = t('error.amds_retry');
        return;
    }

    // 其他錯誤碼：查 i18n 字典
    if (titleEl) titleEl.textContent = t('error.title');
    const localizedMsg = t(`error.${code}`);
    setEl('error-message', localizedMsg !== `error.${code}` ? localizedMsg : (err.message || t('error.unknown_fallback')));

    if (retryBtn) retryBtn.style.display = err.recoverable !== false ? '' : 'none';

    // 只有真正未知的錯誤才顯示「回報問題」
    if (issueBtn) issueBtn.style.display = code === 'UNKNOWN_ERROR' ? '' : 'none';
}

// ============================================================
// 進度更新（P1-2 故事感）
// ============================================================

function updateProgressUI(p) {
    const pct = (p.percent ?? 0).toFixed(1);
    const bar = document.getElementById('backup-progress-bar');
    if (bar) bar.style.width = pct + '%';

    setEl('backup-percent', pct + '%');
    setEl('backup-speed', p.speedBps > 0 ? formatBytes(p.speedBps) + '/s' : '-');
    setEl('backup-eta', p.eta || '-');

    // 故事感進度文字
    const storyEl = document.getElementById('backup-current-story');
    if (storyEl) {
        if (p.phase === 'scanning') {
            storyEl.textContent = t('backup.scanning');
        } else if (p.currentMonth && p.totalFiles > 0) {
            const [year, month] = p.currentMonth.split('-');
            const lang = getLang();
            if (lang === 'zh-TW') {
                storyEl.textContent = `正在備份 ${year} 年 ${parseInt(month)} 月的回憶 · 第 ${fmt(p.doneFiles)} / ${fmt(p.totalFiles)} 個`;
            } else {
                const monthName = new Date(+year, +month - 1).toLocaleString('en', { month: 'long' });
                storyEl.textContent = `Backing up memories from ${monthName} ${year} · ${fmt(p.doneFiles)} of ${fmt(p.totalFiles)}`;
            }
        } else if (p.totalFiles > 0) {
            storyEl.textContent = getLang() === 'zh-TW'
                ? `正在備份第 ${fmt(p.doneFiles)} / ${fmt(p.totalFiles)} 個`
                : `Backing up file ${fmt(p.doneFiles)} of ${fmt(p.totalFiles)}`;
        }
    }

    const skipEl = document.getElementById('backup-skip-count');
    if (skipEl) {
        const sk = p.skippedFiles ?? 0;
        skipEl.textContent = sk > 0
            ? t('backup.skipped').replace('{n}', fmt(sk))
            : '';
    }
}

// ============================================================
// READY 動作
// ============================================================

async function onSelectFolder() {
    try {
        const path = await SelectBackupFolder();
        if (!path) return;
        selectedPath = path;
        setEl('backup-path', path);
        updateDiskInfo(path);
        if (currentDevice?.udid) {
            estimateSize(currentDevice.udid, path);
            // E-9: 換路徑後重算 last-backup-row
            await updateLastBackupRow(path, currentDevice.udid);
        }
    } catch (e) {}
}

async function onStartBackup() {
    if (!currentDevice?.udid || !selectedPath) return;
    const btn = document.getElementById('btn-start-backup');
    if (btn) btn.disabled = true;
    try {
        const convertHeic = document.getElementById('convert-heic')?.checked ?? false;
        await StartBackup({
            deviceUdid:     currentDevice.udid,
            deviceName:     currentDevice.name || 'iPhone',
            backupPath:     selectedPath,
            convertHeic,
            organizeByDate: true,
        });
        setState('backing-up');
    } catch (e) {
        setState('error', { code: 'UNKNOWN_ERROR', message: String(e), recoverable: true });
    } finally {
        if (btn) btn.disabled = false;
    }
}

// ============================================================
// 輔助函式
// ============================================================

async function fetchPhotoCount(udid) {
    try {
        const detail = await GetDeviceDetail(udid);
        if (!detail) return;
        if (currentDevice) currentDevice.photoCount = detail.photoCount;
        const readyCount = document.getElementById('ready-photo-count');
        if (readyCount && currentState === 'ready') {
            readyCount.textContent = `${fmt(detail.photoCount)} ${t('ready.files_count')}`;
        }
    } catch (e) {}
}

async function updateDiskInfo(path) {
    try {
        const disk = await GetDiskInfo(path);
        setEl('disk-free', formatBytes(disk.freeSpace) + ' ' + t('ready.disk_free'));
    } catch (e) {
        setEl('disk-free', '-');
    }
}

async function estimateSize(udid, path) {
    try {
        const bytes = await EstimateBackupSize(udid, path);
        setEl('estimate-size', formatBytes(bytes));
    } catch (e) {
        setEl('estimate-size', '-');
    }
}

function resetDriverMissingView() {
    const initial    = document.getElementById('install-initial');
    const pending    = document.getElementById('install-pending');
    const success    = document.getElementById('install-success');
    const failMsg    = document.getElementById('install-recheck-fail');
    const faqContent = document.getElementById('driver-faq-content');
    const faqIcon    = document.getElementById('faq-toggle-icon');
    if (initial)    initial.style.display = '';
    if (pending)    pending.style.display = 'none';
    if (success)    success.style.display = 'none';
    if (failMsg)    failMsg.style.display = 'none';
    if (faqContent) faqContent.style.display = 'none';
    if (faqIcon)    faqIcon.textContent = '▶';
}

function setEl(id, text) {
    const el = document.getElementById(id);
    if (el) el.textContent = text;
}

function fmt(n) {
    return Number(n).toLocaleString('zh-Hant');
}

function escapeHtml(str) {
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;');
}

function formatBytes(bytes) {
    if (!bytes || bytes <= 0) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
    return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i];
}

function formatRelativeDate(isoStr) {
    if (!isoStr) return '';
    try {
        const d = new Date(isoStr);
        const diffMs = Date.now() - d.getTime();
        const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
        const lang = getLang();
        if (diffDays === 0) return lang === 'zh-TW' ? '今天' : 'Today';
        if (diffDays === 1) return lang === 'zh-TW' ? '昨天' : 'Yesterday';
        if (diffDays < 7)  return lang === 'zh-TW' ? `${diffDays} 天前` : `${diffDays} days ago`;
        const diffWeeks = Math.floor(diffDays / 7);
        if (diffWeeks < 5) return lang === 'zh-TW' ? `${diffWeeks} 週前` : `${diffWeeks} weeks ago`;
        const diffMonths = Math.floor(diffDays / 30);
        return lang === 'zh-TW' ? `${diffMonths} 個月前` : `${diffMonths} months ago`;
    } catch (e) { return isoStr; }
}

// ============================================================
// 啟動
// ============================================================
window.addEventListener('DOMContentLoaded', init);
