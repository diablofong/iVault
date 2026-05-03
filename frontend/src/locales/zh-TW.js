export default {
    // IDLE — 三種變體
    'idle.first.title':         '連接 iPhone，開始備份',
    'idle.first.trust':         '不需帳號 · 離線備份 · 免費開源',
    'idle.first.cta':           '插上 USB 線，iVault 自動偵測並備份',
    'idle.first.preview':       '備份完成後，照片按月份整理，可直接從資料夾瀏覽',

    'idle.returning.title':     '歡迎回來',
    'idle.returning.safe':      '已安全備份',
    'idle.returning.cta':       '用 USB 線連接 iPhone 繼續',

    'idle.interrupted.title':   '你的上次備份被打斷了',
    'idle.interrupted.progress':'上次備份中斷，連線後將繼續',
    'idle.interrupted.cta':     '重新連接 iPhone 繼續備份',

    'idle.amds_starting':       '正在啟動 Apple Devices 服務，畫面可能短暫出現後自動關閉...',
    'idle.update.prefix':       '新版本',
    'idle.update.suffix':       '已發布（不影響備份資料）→',

    // DRIVER BANNER（IDLE 頁早期警告）
    'banner.driver_missing':    '先安裝 Apple Devices，iPhone 即可連線',
    'banner.install':           '安裝',

    // ONBOARDING（首次啟動三步引導）
    'onboard.archive_hint':   '備份是保險庫，不是鏡像 — 就算在手機上刪掉，照片在電腦上仍然安全。',
    'onboard.step_1_of_3':      '1 / 3',
    'onboard.step_2_of_3':      '2 / 3',
    'onboard.step_3_of_3':      '3 / 3',
    'onboard.next':             '繼續',
    'onboard.s1.title':         '先確認一件事',
    'onboard.s1.desc':          'iVault 需要 Apple Devices（免費）才能讀取 iPhone 的照片',
    'onboard.s1.installed':     'Apple Devices 已安裝',
    'onboard.s1.skip':          '稍後再說',
    'onboard.s2.title':         '備份到哪裡？',
    'onboard.s3.title':         '開機後自動啟動？',
    'onboard.s3.desc':          '插上 iPhone 就自動開始備份，不需要手動開啟 iVault',
    'onboard.s3.yes':           '是，自動啟動',
    'onboard.s3.no':            '不用，我自己開',

    // SETTINGS MODAL
    'settings.title':         '設定',
    'settings.backup_path':   '備份資料夾',
    'settings.backup_size':   '備份大小',
    'settings.change':        '更換',
    'settings.open_folder':   '開啟',
    'settings.autostart':     '開機時自動啟動',
    'settings.close':         '關閉',

    // AUTO BACKUP COUNTDOWN（自動備份規則）
    'ready.auto_now':           '立即',
    'ready.auto_snooze':        '稍後 15 分',
    'ready.auto_skip':          '跳過',

    // DEVICE_FOUND
    'device.reading': '正在驗證裝置...',

    // TRUST_GUIDE
    'trust.title':         '請看你的 iPhone，點「信任」',
    'trust.desc':          '在 iPhone 上點「信任」即可繼續',
    'trust.waiting':       '正在等待 iPhone 信任回應…',
    'trust.hint_slow':     'iPhone 螢幕是否亮著？信任視窗有時需要解鎖手機才會出現',
    'trust.hint_hard':     '如果你已點信任卻仍無反應，請拔掉 USB 線後重新插入一次',
    'trust.recheck':       '我已信任，繼續',
    'trust.timeout_title': '沒收到信任回應',
    'trust.timeout_desc':  '請拔掉重插 iPhone 試試',
    'trust.retry':         '重試',
    'trust.dialog_title':  '信任此電腦？',
    'trust.dialog_body':   '你的設定與資料將\n可透過此電腦存取。',
    'trust.deny':          '不信任',
    'trust.allow':         '信任',

    // DRIVER BANNER（保留安裝按鈕文字）
    'driver.open_store':   '一鍵安裝',

    // 安裝確認 modal（三個入口共用）
    'install.close_confirm.desc':   'iVault 需要先關閉。\n請在 Microsoft Store 安裝 Apple Devices，\n完成後重新開啟 iVault 即可開始備份。',
    'install.close_confirm.ok':     '確定，關閉 app',
    'install.close_confirm.cancel': '取消',

    // READY
    'ready.incremental_hint': 'iVault 只複製新增的照片，已備份的不會重複。',
    'ready.path_change_note': '舊備份保留在原位置，新備份將從這個資料夾開始。',
    'ready.name_hint':        '裝置名稱是預設的「iPhone」，建議在手機設定 > 一般 > 關於本機 中設定你的名字，方便日後辨識。',
    'ready.label_to':    '備份到',
    'ready.choose':      '選擇',
    'ready.heic':        '備份後同時轉存 JPEG 副本',
    'ready.start':       '開始備份',
    'ready.last_backup': '上次備份',
    'ready.no_backup':   '尚未備份過',
    'ready.files_count':    '張',
    'ready.disk_free':      '可用',
    'ready.estimate_label': '預估需要',
    'ready.max_file_label': '最大單檔',
    'ready.icloud_hint':    '如果開啟了 iCloud 最佳化儲存，部分照片可能只有縮圖，建議先在設定中下載完整原始檔。',
    'backup.eta_label':     '剩餘',
    'error.unknown_fallback': '發生未預期的錯誤，請重試。',

    // BACKING_UP（D/AC 改善）
    'backup.minimize_hint':  '可以最小化，備份會繼續在背景執行',
    'backup.title':          '備份中',
    'backup.scanning':       '正在讀取照片清單...',
    'backup.nearly_done':    '快完成了，再等一下...',
    'backup.month':          '正在備份 {year} 年 {month} 月的回憶 · 第 {cur} / {total} 個',
    'backup.nodate':         '正在備份第 {cur} / {total} 個',
    'backup.cancel':         '取消',
    'backup.skipped':        '跳過 {n} 個已備份',
    'backup.comfort_1':      '備份進行中，你可以去做其他事',
    'backup.comfort_2':      '照片一張一張安全地複製中...',
    'backup.comfort_3':      '大量照片需要一些時間，請耐心等待',
    'backup.comfort_4':      '即將完成，請勿拔除 iPhone',

    // DONE — 兩種變體
    'done.first.title':      '你的 {photos} 張照片和 {videos} 段影片安全了',
    'done.returning.title':  '新增 {photos} 張照片和 {videos} 段影片',
    'done.subtitle':         '備份於 {date}',
    'done.first_egg':        '這是你用 iVault 備份的第一次。歡迎。',
    'done.safe_hint':        '照片已安全存在這台電腦，可以拔掉 iPhone 了。',
    'done.archive_hint':     '即使之後在手機上刪除，照片依然保存在備份中。',
    'done.live_photo_note':  '含 Live Photo 原始檔',
    'done.unknown_date_hint':'張照片日期無法讀取，已放入「未知日期」資料夾',
    'done.continue':         '繼續',

    'done.label_photos':     '照片',
    'done.label_videos':     '影片',
    'done.label_size':       '空間',
    'done.label_duration':   '花了',
    'done.label_new':        '新增',
    'done.label_skip':       '略過',
    'done.label_fail':       '失敗',
    'done.saved_to':         '已存到',
    'done.open_folder':      '開啟備份資料夾 →',
    'done.back':             '完成',
    'done.failed_toggle':    '查看',
    'done.failed_suffix':    '個失敗檔案',
    'done.heic_hint':        '你備份的照片含有 .heic 格式，Windows 需安裝擴充才能預覽',
    'done.heic_install':     '一鍵安裝（免費）',

    // ERROR — 各錯誤碼對應人類可讀訊息
    'error.title':               '發生錯誤',
    'error.retry':               '重試',
    'error.back':                '返回首頁',
    'error.report':              '回報問題 →',
    'error.DEVICE_DISCONNECTED': 'iPhone 斷線了（可能是沒電、或 USB 線鬆脫）。照片備份已暫停，重新插上繼續。',
    'error.PERMISSION_DENIED':   '無法寫入備份資料夾。可能是防毒軟體攔截，請暫時關閉後重試，或選擇其他資料夾。',
    'error.AMDS_TIMEOUT':        '無法啟動 Apple 裝置服務。請確認 Apple Devices 已安裝，嘗試重新插拔 iPhone，或重新啟動電腦。',
    'error.DISK_FULL':           '備份硬碟空間不夠。請清出足夠空間後重試。',
    'error.TRUST_TIMEOUT':       'iPhone 沒有回應信任請求。請重新插拔再試。',
    'error.AFC_TIMEOUT':         'iPhone 連線不穩。請換一條 USB 線或重新插拔。',
    'error.AFC_CONNECT_FAILED':  '無法存取 iPhone 照片，請確認 iPhone 已解鎖。',
    'error.BACKUP_PATH_MISSING': '備份資料夾不見了（外接硬碟拔掉了嗎？）',
    'error.path_missing_action': '選擇新資料夾',
    'error.amds_title':          'Apple Devices 未能啟動',
    'error.amds_desc':           'Apple Devices 已安裝，但服務尚未就緒。請點下方按鈕開啟一次，再回到 iVault 重試。',
    'error.amds_retry':          '重試',
    'error.amds_launch_btn':     '開啟 Apple Devices',
    'error.AFC_CONNECT_FAILED':  '無法存取 iPhone 照片。可能原因：① 使用了充電線（不支援資料傳輸）② iPhone 尚未解鎖，請解鎖後重試。',

    // HEIC CONVERT
    'heic.converting': '正在轉換 HEIC...',
    'heic.done':       '已轉換',
    'heic.unit':       '張 JPEG',
};
