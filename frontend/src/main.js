import './style.css';
import './app.css';
import { WindowMinimise, Quit, EventsOn } from '../wailsjs/runtime/runtime';
import {
    GetPlatformInfo,
    GetConnectedDevice,
    GetDeviceDetail,
    CheckTrustStatus,
    CheckAppleDevicesInstalled,
    SelectBackupFolder,
    GetDefaultBackupPath,
    GetDiskInfo,
    EstimateBackupSize,
    StartBackup,
    CancelBackup,
    LoadConfig,
    OpenFolder,
    OpenURL,
    InstallAppleDevices,
} from '../wailsjs/go/main/App';
import { t, renderAll, setLang, getLang } from './i18n.js';

// ============================================================
// 狀態機
// ============================================================
const STATES = ['idle', 'device-found', 'trust-guide', 'driver-missing', 'ready', 'backing-up', 'done', 'error'];

let currentState = null;
let currentDevice = null;   // device.DeviceInfo
let selectedPath = '';
let backupResult = null;
let platformInfo = null;
let appConfig = null;
let lastError = null;

function setState(newState, data = {}) {
    if (!STATES.includes(newState)) {
        console.error('Unknown state:', newState);
        return;
    }

    if (currentState) {
        document.getElementById(`view-${currentState}`)?.classList.remove('active');
    }
    currentState = newState;
    document.getElementById(`view-${newState}`)?.classList.add('active');

    // AMDS 啟動提示只在 IDLE 時顯示，切換到其他狀態時隱藏
    if (newState !== 'idle') {
        const amdsEl = document.getElementById('amds-status');
        if (amdsEl) amdsEl.style.display = 'none';
    }

    onEnterState(newState, data);
}

function onEnterState(state, data) {
    switch (state) {
        case 'device-found':   onEnterDeviceFound(data);  break;
        case 'trust-guide':    onEnterTrustGuide(data);   break;
        case 'ready':          onEnterReady(data);         break;
        case 'done':           onEnterDone(data);          break;
        case 'error':          onEnterError(data);         break;
    }
}

// ============================================================
// 初始化
// ============================================================
async function init() {
    // 視窗控制（Windows）
    document.getElementById('btn-minimize')?.addEventListener('click', () => WindowMinimise());
    document.getElementById('btn-close')?.addEventListener('click', () => Quit());

    // 取得平台資訊與設定
    try {
        [platformInfo, appConfig] = await Promise.all([GetPlatformInfo(), LoadConfig()]);
        document.body.classList.add(`os-${platformInfo.os}`);
        if (platformInfo.darkMode) document.body.classList.add('dark');
    } catch (e) {
        console.error('init:', e);
    }

    // i18n 初始渲染
    renderAll();

    // 註冊 Wails 事件
    registerEvents();

    // 綁定按鈕
    bindHandlers();

    // Windows：主動確認 Apple Devices 安裝狀態，不等裝置連線觸發
    if (platformInfo?.os === 'windows') {
        try {
            const installed = await CheckAppleDevicesInstalled();
            if (!installed) {
                setState('driver-missing');
                return; // 等裝好後 watchDevices 會 emit driver:installed
            }
        } catch (e) { /* 無法判斷時繼續走正常流程 */ }
    }

    // 檢查是否已有裝置連線（app 重啟場景）
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
        // 備份中若斷線，backup:error 會處理；其他狀態直接回 idle
        if (currentState !== 'backing-up') {
            setState('idle');
        }
    });

    EventsOn('device:trust-changed', (data) => {
        if (data.trusted && currentState === 'trust-guide') {
            setState('ready', currentDevice || { udid: data.udid });
        }
    });

    EventsOn('backup:progress', (progress) => {
        updateProgressUI(progress);
    });

    EventsOn('backup:complete', (result) => {
        backupResult = result;
        setState('done', result);
    });

    EventsOn('backup:error', (err) => {
        lastError = err;
        if (currentState === 'backing-up') {
            setState('error', err);
        }
    });

    // Driver 安裝事件
    // Apple Devices 尚未安裝（watchDevices 偵測到）
    EventsOn('driver:required', (data) => {
        if (currentState !== 'driver-missing') setState('driver-missing');
        // WMI 偵測到裝置名稱
        if (data?.deviceName) {
            const detectedEl = document.getElementById('driver-device-detected');
            const nameEl = document.getElementById('driver-device-name-wmi');
            if (detectedEl) detectedEl.style.display = '';
            if (nameEl) nameEl.textContent = data.deviceName;
        }
    });

    // watchDevices 背景自動偵測到安裝完成（加分項，不依賴此路徑）
    EventsOn('driver:installed', () => {
        const pending = document.getElementById('install-pending');
        const initial = document.getElementById('install-initial');
        const success = document.getElementById('install-success');
        if (pending) pending.style.display = 'none';
        if (initial) initial.style.display = 'none';
        if (success) success.style.display = '';
    });

    // AMDS 正在啟動：Apple Devices UI 即將彈出，在 IDLE 畫面顯示提示
    EventsOn('amds:starting', () => {
        const el = document.getElementById('amds-status');
        if (el) el.style.display = '';
    });

    // AMDS（AppleMobileDeviceProcess）啟動失敗
    // MS Store Apple Devices 的背景服務無法在 8s 內啟動
    EventsOn('amds:start_failed', () => {
        if (currentState !== 'error') {
            setState('error', {
                code: 'AMDS_START_FAILED',
                message: t('error.amds_desc'),
                recoverable: true,
            });
        }
    });

    // HEIC 轉檔事件
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
    // 語言切換
    document.getElementById('btn-lang-toggle')?.addEventListener('click', () => {
        setLang(getLang() === 'zh-TW' ? 'en' : 'zh-TW');
        // 重新渲染動態欄位
        if (currentState === 'ready') onEnterReady(currentDevice || {});
    });

    // DRIVER_MISSING — 開啟 MS Store
    document.getElementById('btn-install-driver')?.addEventListener('click', () => {
        InstallAppleDevices();
        document.getElementById('install-initial').style.display = 'none';
        const pending = document.getElementById('install-pending');
        if (pending) {
            pending.style.display = '';
            // data-i18n 在 display:none 內不會自動更新，手動補
            pending.querySelectorAll('[data-i18n]').forEach(el => {
                const key = el.getAttribute('data-i18n');
                const val = t(key);
                if (val) el.textContent = val;
            });
        }
    });

    // 手動重新偵測
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

    // 安裝完成後重插 iPhone
    document.getElementById('btn-replug-done')?.addEventListener('click', () => {
        resetDriverMissingView();
        setState('idle');
    });

    // FAQ 展開/收合
    document.getElementById('btn-faq-toggle')?.addEventListener('click', () => {
        const content = document.getElementById('driver-faq-content');
        const icon = document.getElementById('faq-toggle-icon');
        if (!content) return;
        const expanded = content.style.display !== 'none';
        content.style.display = expanded ? 'none' : '';
        if (icon) icon.textContent = expanded ? '▶' : '▼';
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
    document.getElementById('btn-sponsor')?.addEventListener('click', () => {
        OpenURL('https://buymeacoffee.com/ivault');
    });
    document.getElementById('btn-backup-again')?.addEventListener('click', () => setState('idle'));

    // 失敗清單展開/收合
    document.getElementById('btn-toggle-failed')?.addEventListener('click', () => {
        const list = document.getElementById('failed-list');
        const icon = document.getElementById('failed-toggle-icon');
        if (!list) return;
        const expanded = list.style.display !== 'none';
        list.style.display = expanded ? 'none' : '';
        if (icon) icon.textContent = expanded ? '▶' : '▼';
    });

    // HEIC 一鍵安裝（Windows HEIF 擴充）
    document.getElementById('btn-install-heic')?.addEventListener('click', () => {
        OpenURL('ms-windows-store://pdp?productId=9PMMSR1CGPWG');
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

async function onEnterDeviceFound(info) {
    setEl('device-name', info.name || 'iPhone');
    setEl('device-ios', `iOS ${info.iosVersion || '-'}`);
    setEl('device-photo-count', t('device.reading'));

    // Windows：先確認 Apple Devices 是否安裝
    if (platformInfo?.os === 'windows') {
        try {
            const installed = await CheckAppleDevicesInstalled();
            if (!installed) {
                setState('driver-missing');
                return;
            }
        } catch (e) { /* 繼續 */ }
    }

    // 檢查信任狀態
    try {
        const trusted = await CheckTrustStatus(info.udid);
        if (!trusted) {
            setState('trust-guide', info);
            return;
        }
    } catch (e) { /* 連線失敗也繼續嘗試 */ }

    // 背景取得照片數（READY 頁顯示）
    fetchPhotoCount(info.udid);

    setState('ready', info);
}

function onEnterTrustGuide(info) {
    // 後端 watchDevices → startTrustPolling 已在背景輪詢
    // 前端只需等待 device:trust-changed event
}

async function onEnterReady(info) {
    setEl('ready-device-name', info.name || 'iPhone');

    // 預設路徑
    if (!selectedPath) {
        try {
            selectedPath = await GetDefaultBackupPath();
        } catch (e) {}
    }
    setEl('backup-path', selectedPath || '-');

    // 磁碟資訊
    if (selectedPath) {
        updateDiskInfo(selectedPath);
    }

    // 估算備份大小
    if (info.udid && selectedPath) {
        estimateSize(info.udid, selectedPath);
    }

    // 還原 HEIC 設定
    const checkbox = document.getElementById('convert-heic');
    if (checkbox && appConfig) checkbox.checked = !!appConfig.convertHeic;

    // 上次備份資訊
    const lastBackupRow = document.getElementById('last-backup-row');
    const lastBackupInfo = document.getElementById('last-backup-info');
    if (appConfig?.history?.length > 0) {
        const rec = appConfig.history[0];
        const dateStr = formatRelativeDate(rec.date);
        const countStr = `${fmt(rec.newFiles)} ${t('ready.files_count')}`;
        if (lastBackupInfo) lastBackupInfo.textContent = `${dateStr}，共 ${countStr}`;
        if (lastBackupRow) lastBackupRow.style.display = '';
    } else {
        if (lastBackupInfo) lastBackupInfo.textContent = t('ready.no_backup');
        if (lastBackupRow) lastBackupRow.style.display = '';
    }
}

function onEnterDone(result) {
    setEl('done-new-count', fmt(result.newFiles ?? 0));
    setEl('done-skip-count', fmt(result.skippedFiles ?? 0));
    setEl('done-fail-count', fmt(result.failedFiles ?? 0));

    const durEl = document.getElementById('done-duration');
    if (durEl) durEl.textContent = result.duration ? `${t('done.duration')} ${result.duration}` : '';

    // 失敗數 > 0 時標紅
    const failEl = document.getElementById('done-fail-count');
    if (failEl) failEl.classList.toggle('danger', (result.failedFiles ?? 0) > 0);

    // 備份大小
    const totalSizeEl = document.getElementById('done-total-size');
    const totalSizeValueEl = document.getElementById('done-total-size-value');
    if (result.totalBytes > 0) {
        if (totalSizeValueEl) totalSizeValueEl.textContent = formatBytes(result.totalBytes);
        if (totalSizeEl) totalSizeEl.style.display = '';
    } else {
        if (totalSizeEl) totalSizeEl.style.display = 'none';
    }

    // 失敗清單（可展開）
    const failedSection = document.getElementById('failed-section');
    const failedList = document.getElementById('failed-list');
    const failedCount = result.failedFiles ?? 0;
    if (failedSection) failedSection.style.display = failedCount > 0 ? '' : 'none';
    setEl('failed-detail-count', fmt(failedCount));
    if (failedList && failedCount > 0 && Array.isArray(result.failedList)) {
        failedList.innerHTML = '';
        failedList.style.display = 'none'; // 預設收合
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
}

function onEnterError(err) {
    const code = err.code || 'UNKNOWN_ERROR';

    // 重設標題為預設值（處理 AMDS 等特殊 code 可能改過標題的情況）
    const titleEl = document.getElementById('error-title');
    if (titleEl) titleEl.textContent = t('error.title');

    // DEVICE_DISCONNECTED → 1.5 秒後自動回 IDLE
    if (code === 'DEVICE_DISCONNECTED') {
        setEl('error-message', err.message || 'iPhone 已斷開連線');
        document.getElementById('btn-retry')?.style && (document.getElementById('btn-retry').style.display = 'none');
        document.getElementById('btn-report-issue')?.style && (document.getElementById('btn-report-issue').style.display = 'none');
        setTimeout(() => setState('idle'), 1500);
        return;
    }

    // AMDS_START_FAILED → 顯示專屬標題和說明
    if (code === 'AMDS_START_FAILED') {
        if (titleEl) titleEl.textContent = t('error.amds_title');
        setEl('error-message', t('error.amds_desc'));
        const retryBtn = document.getElementById('btn-retry');
        if (retryBtn) { retryBtn.style.display = ''; retryBtn.textContent = t('error.amds_retry'); }
        document.getElementById('btn-report-issue').style.display = 'none';
        return;
    }

    setEl('error-message', err.message || '發生未預期的錯誤，請重試。');

    const retryBtn = document.getElementById('btn-retry');
    if (retryBtn) retryBtn.style.display = err.recoverable !== false ? '' : 'none';

    // UNKNOWN_ERROR → 顯示回報問題按鈕
    const issueBtn = document.getElementById('btn-report-issue');
    if (issueBtn) issueBtn.style.display = code === 'UNKNOWN_ERROR' ? '' : 'none';
}

// ============================================================
// 事件處理 — READY 動作
// ============================================================

async function onSelectFolder() {
    try {
        const path = await SelectBackupFolder();
        if (!path) return;
        selectedPath = path;
        setEl('backup-path', path);
        updateDiskInfo(path);
        if (currentDevice?.udid) estimateSize(currentDevice.udid, path);
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
// 進度更新
// ============================================================

function updateProgressUI(p) {
    const pct = (p.percent ?? 0).toFixed(1);
    const bar = document.getElementById('backup-progress-bar');
    if (bar) bar.style.width = pct + '%';

    setEl('backup-percent', pct + '%');
    setEl('backup-speed', p.speedBps > 0 ? formatBytes(p.speedBps) + '/s' : '-');
    setEl('backup-eta', p.eta || '-');
    setEl('backup-current-file', p.currentFile
        ? `${t('backup.current')} ${p.currentFile}`
        : t('backup.scanning'));
    setEl('backup-done-count', `${fmt(p.doneFiles ?? 0)} / ${fmt(p.totalFiles ?? 0)} ${t('backup.sheets')}`);

    const skipEl = document.getElementById('backup-skip-count');
    if (skipEl) {
        const sk = p.skippedFiles ?? 0;
        skipEl.textContent = sk > 0 ? t('backup.skipped').replace('{n}', fmt(sk)) : '';
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
        // 更新 READY 頁照片數
        const readyCount = document.getElementById('ready-photo-count');
        if (readyCount && currentState === 'ready') {
            readyCount.textContent = `${fmt(detail.photoCount)} ${t('ready.files_count')}`;
        }
    } catch (e) {}
}

async function updateDiskInfo(path) {
    try {
        const disk = await GetDiskInfo(path);
        setEl('disk-free', formatBytes(disk.freeSpace) + ' 可用');
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

function resetDriverMissingView() {
    const initial    = document.getElementById('install-initial');
    const pending    = document.getElementById('install-pending');
    const success    = document.getElementById('install-success');
    const detected   = document.getElementById('driver-device-detected');
    const failMsg    = document.getElementById('install-recheck-fail');
    const faqContent = document.getElementById('driver-faq-content');
    const faqIcon    = document.getElementById('faq-toggle-icon');
    if (initial)    initial.style.display = '';
    if (pending)    pending.style.display = 'none';
    if (success)    success.style.display = 'none';
    if (detected)   detected.style.display = 'none';
    if (failMsg)    failMsg.style.display = 'none';
    if (faqContent) faqContent.style.display = 'none';
    if (faqIcon)    faqIcon.textContent = '▶';
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
        if (diffDays < 30) return lang === 'zh-TW' ? `${diffDays} 天前` : `${diffDays} days ago`;
        const diffMonths = Math.floor(diffDays / 30);
        return lang === 'zh-TW' ? `${diffMonths} 個月前` : `${diffMonths} months ago`;
    } catch (e) { return isoStr; }
}

// ============================================================
// 啟動
// ============================================================
window.addEventListener('DOMContentLoaded', init);
