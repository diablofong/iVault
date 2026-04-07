import './style.css';
import './app.css';
import { WindowMinimise, Quit } from '../wailsjs/runtime/runtime';

// ============================================================
// 狀態機
// ============================================================
const STATES = ['idle', 'device-found', 'trust-guide', 'driver-missing', 'ready', 'backing-up', 'done', 'error'];

let currentState = null;

function setState(newState) {
    if (!STATES.includes(newState)) {
        console.error('未知狀態:', newState);
        return;
    }

    // 隱藏目前視圖
    if (currentState) {
        const current = document.getElementById(`view-${currentState}`);
        if (current) {
            current.classList.remove('visible');
            setTimeout(() => current.classList.remove('active'), 200);
        }
    }

    currentState = newState;

    // 顯示新視圖
    const next = document.getElementById(`view-${newState}`);
    if (next) {
        next.classList.add('active');
        // 讓 transition 觸發
        requestAnimationFrame(() => {
            requestAnimationFrame(() => next.classList.add('visible'));
        });
    }
}

// ============================================================
// 平台初始化
// ============================================================
async function init() {
    // 預設先顯示 idle
    setState('idle');

    // 標題列按鈕（Windows 用）
    document.getElementById('btn-minimize')?.addEventListener('click', () => WindowMinimise());
    document.getElementById('btn-close')?.addEventListener('click', () => Quit());

    // DRIVER_MISSING 按鈕
    document.getElementById('btn-install-driver')?.addEventListener('click', onInstallDriver);
    document.getElementById('btn-recheck-driver')?.addEventListener('click', onRecheckDriver);

    // READY 按鈕
    document.getElementById('btn-select-folder')?.addEventListener('click', onSelectFolder);
    document.getElementById('btn-start-backup')?.addEventListener('click', onStartBackup);

    // BACKING_UP 按鈕
    document.getElementById('btn-cancel-backup')?.addEventListener('click', onCancelBackup);

    // DONE 按鈕
    document.getElementById('btn-open-folder')?.addEventListener('click', onOpenFolder);
    document.getElementById('btn-backup-again')?.addEventListener('click', () => setState('idle'));

    // ERROR 按鈕
    document.getElementById('btn-retry')?.addEventListener('click', onRetry);
    document.getElementById('btn-back-to-idle')?.addEventListener('click', () => setState('idle'));
}

// ============================================================
// 事件處理（暫時用 placeholder，等 API 完成後接入）
// ============================================================
async function onInstallDriver() {
    // await OpenURL("ms-windows-store://pdp/?productId=9NP83LWLPZ9N");
    console.log('open apple devices store');
}

async function onRecheckDriver() {
    setState('idle');
}

async function onSelectFolder() {
    console.log('select folder');
}

async function onStartBackup() {
    setState('backing-up');
}

async function onCancelBackup() {
    console.log('cancel backup');
    setState('idle');
}

async function onOpenFolder() {
    console.log('open folder');
}

async function onRetry() {
    setState('idle');
}

// ============================================================
// 啟動
// ============================================================
window.addEventListener('DOMContentLoaded', init);
