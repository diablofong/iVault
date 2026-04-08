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

    // 淡出目前視圖
    if (currentState) {
        const prev = document.getElementById(`view-${currentState}`);
        if (prev) {
            prev.classList.remove('visible');
            setTimeout(() => prev.classList.remove('active'), 200);
        }
    }

    currentState = newState;

    // 淡入新視圖
    const next = document.getElementById(`view-${newState}`);
    if (next) {
        next.classList.add('active');
        requestAnimationFrame(() => {
            requestAnimationFrame(() => next.classList.add('visible'));
        });
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

    // 註冊 Wails 事件
    registerEvents();

    // 綁定按鈕
    bindHandlers();

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
    EventsOn('driver:install-started', (data) => {
        const btn = document.getElementById('btn-install-driver');
        const progress = document.getElementById('install-progress');
        const label = document.getElementById('install-status-label');
        if (btn) btn.style.display = 'none';
        if (progress) progress.style.display = '';
        if (label) {
            label.textContent = data?.method === 'winget' ? '正在自動安裝...' : '請在 Microsoft Store 完成安裝';
        }
    });

    EventsOn('driver:installed', () => {
        if (currentDevice) setState('device-found', currentDevice);
        else setState('idle');
    });

    // HEIC 轉檔事件
    EventsOn('heic:progress', (data) => {
        const section = document.getElementById('heic-convert-section');
        const bar = document.getElementById('heic-progress-bar');
        const label = document.getElementById('heic-convert-label');
        if (section) section.style.display = '';
        if (bar) bar.style.width = (data.percent ?? 0).toFixed(1) + '%';
        if (label) label.textContent = `正在轉換 HEIC... ${data.done ?? 0} / ${data.total ?? 0}`;
    });

    EventsOn('heic:complete', (data) => {
        const section = document.getElementById('heic-convert-section');
        const label = document.getElementById('heic-convert-label');
        if (data?.converted > 0) {
            if (label) label.textContent = `已轉換 ${fmt(data.converted)} 張 JPEG`;
        } else {
            if (section) section.style.display = 'none';
        }
    });

    EventsOn('driver:install-failed', () => {
        // winget 兩步都失敗，已 fallback 到 MS Store，更新提示文字
        const label = document.getElementById('install-status-label');
        if (label) label.textContent = '請在 Microsoft Store 完成安裝';
    });
}

// ============================================================
// 按鈕綁定
// ============================================================
function bindHandlers() {
    // DRIVER_MISSING
    document.getElementById('btn-install-driver')?.addEventListener('click', () => InstallAppleDevices());

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

    // ERROR
    document.getElementById('btn-retry')?.addEventListener('click', () => {
        if (currentDevice) setState('ready', currentDevice);
        else setState('idle');
    });
    document.getElementById('btn-back-to-idle')?.addEventListener('click', () => setState('idle'));
}

// ============================================================
// 狀態進入處理
// ============================================================

async function onEnterDeviceFound(info) {
    setEl('device-name', info.name || 'iPhone');
    setEl('device-ios', `iOS ${info.iosVersion || '-'}`);
    setEl('device-photo-count', '正在驗證裝置...');

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
}

function onEnterDone(result) {
    setEl('done-new-count', fmt(result.newFiles ?? 0));
    setEl('done-skip-count', fmt(result.skippedFiles ?? 0));
    setEl('done-fail-count', fmt(result.failedFiles ?? 0));

    const durEl = document.getElementById('done-duration');
    if (durEl) durEl.textContent = result.duration ? `耗時 ${result.duration}` : '';

    // 失敗數 > 0 時標紅
    const failEl = document.getElementById('done-fail-count');
    if (failEl) failEl.classList.toggle('danger', (result.failedFiles ?? 0) > 0);

    // Windows HEIC 提示
    const heicHint = document.getElementById('heic-hint');
    if (heicHint) {
        heicHint.style.display = (platformInfo?.os === 'windows' && result.hasHeic) ? 'block' : 'none';
    }
}

function onEnterError(err) {
    setEl('error-message', err.message || '發生未預期的錯誤，請重試。');
    const retryBtn = document.getElementById('btn-retry');
    if (retryBtn) retryBtn.style.display = err.recoverable !== false ? '' : 'none';
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
    setEl('backup-current-file', p.currentFile ? `正在備份 ${p.currentFile}` : '掃描照片清單...');
    setEl('backup-done-count', `${fmt(p.doneFiles ?? 0)} / ${fmt(p.totalFiles ?? 0)} 張`);

    const skipEl = document.getElementById('backup-skip-count');
    if (skipEl) {
        const sk = p.skippedFiles ?? 0;
        skipEl.textContent = sk > 0 ? `(跳過 ${fmt(sk)} 張已備份)` : '';
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
            readyCount.textContent = `${fmt(detail.photoCount)} 張照片`;
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

function formatBytes(bytes) {
    if (!bytes || bytes <= 0) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
    return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i];
}

// ============================================================
// 啟動
// ============================================================
window.addEventListener('DOMContentLoaded', init);
