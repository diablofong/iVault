import './style.css';
import './app.css';
import { WindowMinimise, Quit } from '../wailsjs/runtime/runtime';
import { ListDevices, GetDeviceDetail, ScanDCIM } from '../wailsjs/go/main/App';

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
let currentUDID = null;

async function init() {
    // 預設先顯示 idle
    setState('idle');

    // 標題列按鈕（Windows 用）
    document.getElementById('btn-minimize')?.addEventListener('click', () => WindowMinimise());
    document.getElementById('btn-close')?.addEventListener('click', () => Quit());

    // IDLE：偵測裝置（AFC PoC）
    document.getElementById('btn-detect-device')?.addEventListener('click', onDetectDevice);

    // DEVICE_FOUND：掃描 DCIM（AFC PoC）
    document.getElementById('btn-scan-dcim')?.addEventListener('click', onScanDCIM);

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
// 事件處理
// ============================================================

// AFC PoC：偵測裝置
async function onDetectDevice() {
    const resultEl = document.getElementById('detect-result');
    resultEl.textContent = '偵測中...';
    try {
        const devices = await ListDevices();
        if (!devices || devices.length === 0) {
            resultEl.textContent = '未偵測到裝置，請確認 USB 已連接並已信任此電腦';
            return;
        }
        const d = devices[0];
        currentUDID = d.udid;
        resultEl.textContent = `✅ ${d.name} (${d.model} / iOS ${d.iosVersion})`;

        // 切換到 device-found 狀態
        document.getElementById('device-name').textContent = d.name || 'iPhone';
        document.getElementById('device-info').textContent = `${d.model} · iOS ${d.iosVersion} · UDID: ${d.udid}`;
        setState('device-found');

        // 背景取得詳細資訊
        GetDeviceDetail(d.udid).then(detail => {
            if (!detail) return;
            const gb = (b) => (b / 1024 / 1024 / 1024).toFixed(1) + ' GB';
            document.getElementById('device-info').textContent =
                `${detail.model} · iOS ${detail.iosVersion} · ${detail.photoCount} 張照片 · ${gb(detail.usedSpace)} / ${gb(detail.totalSpace)}`;
        }).catch(console.error);
    } catch (e) {
        resultEl.textContent = `❌ 錯誤：${e}`;
    }
}

// AFC PoC：掃描 DCIM
async function onScanDCIM() {
    if (!currentUDID) return;
    const resultEl = document.getElementById('scan-result');
    resultEl.textContent = '掃描中...';
    try {
        const files = await ScanDCIM(currentUDID);
        if (!files || files.length === 0) {
            resultEl.textContent = '未找到照片';
            return;
        }
        const preview = files.slice(0, 5).map(f => `${f.fileName} (${(f.size/1024).toFixed(0)}KB)`).join('\n');
        resultEl.textContent = `✅ 共 ${files.length} 個檔案\n${preview}${files.length > 5 ? '\n...' : ''}`;
    } catch (e) {
        resultEl.textContent = `❌ 錯誤：${e}`;
    }
}

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
