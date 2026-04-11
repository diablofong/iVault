export default {
    // IDLE — 三種變體
    'idle.first.title':         '把 iPhone 的照片與影片，備份到這台電腦。',
    'idle.first.cta':           '用 USB 線連接 iPhone 開始',
    'idle.first.tagline':       '完全離線 · 開源免費 · 原始畫質',

    'idle.returning.title':     '歡迎回來',
    'idle.returning.cta':       '用 USB 線連接 iPhone 繼續',

    'idle.interrupted.title':   '你的上次備份被打斷了',
    'idle.interrupted.progress':'進度安全保留',
    'idle.interrupted.cta':     '重新連接 iPhone 繼續備份',

    'idle.amds_starting':       '正在啟動 Apple Devices 服務，畫面可能短暫出現後自動關閉...',

    // DEVICE_FOUND
    'device.reading': '正在驗證裝置...',

    // TRUST_GUIDE
    'trust.title':         '請看你的 iPhone，點「信任」',
    'trust.desc':          '在 iPhone 上點「信任」即可繼續',
    'trust.waiting':       '正在等待 iPhone 信任回應…',
    'trust.hint_slow':     'iPhone 螢幕是否亮著？信任視窗有時需要解鎖手機才會出現',
    'trust.timeout_title': '沒收到信任回應',
    'trust.timeout_desc':  '請拔掉重插 iPhone 試試',
    'trust.retry':         '重試',
    'trust.dialog_title':  '信任此電腦？',
    'trust.dialog_body':   '你的設定與資料將\n可透過此電腦存取。',
    'trust.deny':          '不信任',
    'trust.allow':         '信任',

    // DRIVER_MISSING
    'driver.title':        '需要先安裝 Apple Devices（免費）',
    'driver.subtitle':     '由 Apple 官方提供，只需安裝一次',
    'driver.open_store':   '一鍵安裝',
    'driver.pending_title':'已開啟 Microsoft Store',
    'driver.hint':         '安裝完成後，點下方按鈕繼續',
    'driver.recheck':      '我已安裝完成，重新偵測',
    'driver.recheck_fail': '尚未偵測到，請確認 Apple Devices 已安裝完成',
    'driver.success':      '安裝完成！',
    'driver.success_hint': '請重新插拔 iPhone 以繼續',
    'driver.replug_done':  '繼續',
    'driver.faq_toggle':   '為什麼需要？',
    'driver.faq_a':        'Apple Devices 是 Apple 官方提供的免費驅動程式，讓 Windows 能辨識 iPhone。完全安全，iVault 不會存取你的 Apple 帳號。',

    // READY
    'ready.label_to':    '備份到',
    'ready.choose':      '選擇',
    'ready.heic':        '備份後同時轉存 JPEG 副本',
    'ready.start':       '開始備份',
    'ready.last_backup': '上次備份',
    'ready.no_backup':   '尚未備份過',
    'ready.files_count': '張',

    // BACKING_UP
    'backup.title':    '備份中',
    'backup.scanning': '掃描照片清單...',
    'backup.month':    '正在備份 {year} 年 {month} 月的回憶 · 第 {cur} / {total} 個',
    'backup.nodate':   '正在備份第 {cur} / {total} 個',
    'backup.cancel':   '取消',
    'backup.skipped':  '跳過 {n} 個已備份',

    // DONE — 兩種變體
    'done.first.title':      '你的 {photos} 張照片和 {videos} 段影片安全了',
    'done.returning.title':  '新增 {photos} 張照片和 {videos} 段影片',
    'done.subtitle':         '備份於 {date}',
    'done.first_egg':        '這是你用 iVault 備份的第一次。歡迎。',

    'done.label_photos':     '照片',
    'done.label_videos':     '影片',
    'done.label_size':       '空間',
    'done.label_duration':   '花了',
    'done.label_new':        '新增',
    'done.label_skip':       '略過',
    'done.label_fail':       '失敗',
    'done.open_folder':      '開啟備份資料夾 →',
    'done.back':             '完成',
    'done.failed_toggle':    '查看',
    'done.failed_suffix':    '個失敗檔案',
    'done.heic_hint':        '你備份的照片含有 .heic 格式，Windows 需安裝擴充才能預覽',
    'done.heic_install':     '一鍵安裝（免費）',

    // SPONSOR
    'sponsor.text': 'iVault 永遠免費開源。如果它幫到你，請我喝杯咖啡 →',
    'sponsor.btn':  '$5 支持開發',

    // ERROR — 各錯誤碼對應人類可讀訊息
    'error.title':               '發生錯誤',
    'error.retry':               '重試',
    'error.back':                '返回首頁',
    'error.report':              '回報問題 →',
    'error.DEVICE_DISCONNECTED': 'iPhone 被拔掉了。插回來就可以繼續。',
    'error.AMDS_TIMEOUT':        '無法啟動 Apple 裝置服務。請重新插拔 iPhone 再試。',
    'error.DISK_FULL':           '備份硬碟空間不夠。請清出足夠空間後重試。',
    'error.TRUST_TIMEOUT':       'iPhone 沒有回應信任請求。請重新插拔再試。',
    'error.AFC_TIMEOUT':         'iPhone 連線不穩。請換一條 USB 線或重新插拔。',
    'error.AFC_CONNECT_FAILED':  '無法存取 iPhone 照片，請確認 iPhone 已解鎖。',
    'error.BACKUP_PATH_MISSING': '備份資料夾不見了（外接硬碟拔掉了嗎？）',
    'error.amds_title':          'Apple Devices 未能啟動',
    'error.amds_desc':           '請手動打開 Apple Devices 一次後，回到 iVault 重試',
    'error.amds_retry':          '重試',

    // HEIC CONVERT
    'heic.converting': '正在轉換 HEIC...',
    'heic.done':       '已轉換',
    'heic.unit':       '張 JPEG',
};
