export default {
    // IDLE — three variants
    'idle.first.title':         'Back up your iPhone photos and videos to this computer.',
    'idle.first.cta':           'Connect iPhone with a USB cable to get started',
    'idle.first.tagline':       'Offline only · Free & open source · Original quality',

    'idle.returning.title':     'Welcome back',
    'idle.returning.cta':       'Connect iPhone with a USB cable to continue',

    'idle.interrupted.title':   'Your last backup was interrupted',
    'idle.interrupted.progress':'Backup interrupted — will resume on reconnect',
    'idle.interrupted.cta':     'Reconnect iPhone to resume',

    'idle.amds_starting':       'Starting Apple Devices service, the app may briefly appear then close...',
    'idle.update.prefix':       'New version',
    'idle.update.suffix':       'available →',

    // DEVICE_FOUND
    'device.reading': 'Verifying device...',

    // TRUST_GUIDE
    'trust.title':         'Tap "Trust" on your iPhone',
    'trust.desc':          'Tap "Trust" on your iPhone to continue',
    'trust.waiting':       'Waiting for iPhone trust response…',
    'trust.hint_slow':     'Is your iPhone screen on? The trust prompt may require unlocking your iPhone first.',
    'trust.hint_hard':     'If you already tapped Trust but nothing happens, unplug the USB cable and plug it back in.',
    'trust.recheck':       "I've Trusted, Continue",
    'trust.timeout_title': 'No trust response received',
    'trust.timeout_desc':  'Try unplugging and reconnecting your iPhone',
    'trust.retry':         'Retry',
    'trust.dialog_title':  'Trust This Computer?',
    'trust.dialog_body':   'Your settings and data will be\naccessible from this computer.',
    'trust.deny':          'Don\'t Trust',
    'trust.allow':         'Trust',

    // DRIVER_MISSING
    'driver.title':        'Install Apple Devices first (free)',
    'driver.subtitle':     'Official Apple app · One-time setup',
    'driver.open_store':   'Install Now',
    'driver.pending_title':'Microsoft Store opened',
    'driver.hint':         'After installation, click the button below to continue',
    'driver.recheck':      'I\'ve installed it — Re-detect',
    'driver.recheck_fail': 'Not detected yet. Please confirm Apple Devices is fully installed.',
    'driver.success':      'Installation complete!',
    'driver.success_hint': 'Please unplug and reconnect your iPhone to continue',
    'driver.replug_done':  'Continue',
    'driver.faq_toggle':   'Why is this needed?',
    'driver.faq_a':        'Apple Devices is a free official driver from Apple that lets Windows recognize your iPhone. Completely safe — iVault never accesses your Apple account.',

    // READY
    'ready.incremental_hint': 'iVault only copies new photos — previously backed up files are skipped.',
    'ready.label_to':    'Back up to',
    'ready.choose':      'Choose',
    'ready.heic':        'Also save JPEG copies of HEIC photos',
    'ready.start':       'Start Backup',
    'ready.last_backup': 'Last backup',
    'ready.no_backup':   'Never backed up',
    'ready.files_count':    'photos',
    'ready.disk_free':      'available',
    'ready.estimate_label': 'Estimated',
    'backup.eta_label':     'Remaining',
    'error.unknown_fallback': 'An unexpected error occurred. Please try again.',

    // BACKING_UP
    'backup.minimize_hint': 'You can minimize — backup continues in the background',
    'backup.title':    'Backing Up',
    'backup.scanning': 'Scanning photo library...',
    'backup.month':    'Backing up memories from {month}/{year} · {cur} of {total}',
    'backup.nodate':   'Backing up file {cur} of {total}',
    'backup.cancel':   'Cancel',
    'backup.skipped':  '{n} already backed up, skipped',

    // DONE — two variants
    'done.first.title':      'Your {photos} photos and {videos} videos are safe',
    'done.returning.title':  '{photos} new photos and {videos} new videos added',
    'done.subtitle':         'Backed up on {date}',
    'done.first_egg':        'This is your first backup with iVault. Welcome.',
    'done.safe_hint':        'Your photos are safely stored on this computer. You can unplug your iPhone.',
    'done.live_photo_note':  'Includes Live Photo originals',
    'done.unknown_date_hint':'photos could not be dated and were placed in the "unknown-date" folder',
    'done.continue':         'Continue',

    'done.label_photos':     'Photos',
    'done.label_videos':     'Videos',
    'done.label_size':       'Size',
    'done.label_duration':   'Time',
    'done.label_new':        'New',
    'done.label_skip':       'Skipped',
    'done.label_fail':       'Failed',
    'done.open_folder':      'Open Backup Folder →',
    'done.back':             'Done',
    'done.failed_toggle':    'View',
    'done.failed_suffix':    'failed files',
    'done.heic_hint':        'Your backup contains .heic files. Install the free codec to preview them on Windows.',
    'done.heic_install':     'Install Free Codec',

    // ERROR — human-readable messages per code
    'error.title':               'Something Went Wrong',
    'error.retry':               'Try Again',
    'error.back':                'Back to Home',
    'error.report':              'Report Issue →',
    'error.DEVICE_DISCONNECTED': 'iPhone was disconnected. Plug it back in to continue.',
    'error.AMDS_TIMEOUT':        'Could not start Apple Devices service. Try unplugging and reconnecting your iPhone.',
    'error.DISK_FULL':           'Not enough disk space. Please free up space and try again.',
    'error.TRUST_TIMEOUT':       'iPhone did not respond to the trust request. Try unplugging and reconnecting.',
    'error.AFC_TIMEOUT':         'iPhone connection is unstable. Try a different USB cable or reconnect.',
    'error.AFC_CONNECT_FAILED':  'Cannot access iPhone photos. Make sure your iPhone is unlocked.',
    'error.BACKUP_PATH_MISSING': 'Backup folder not found. (Did you unplug an external drive?)',
    'error.path_missing_action': 'Choose a New Folder',
    'error.amds_title':          'Apple Devices Failed to Start',
    'error.amds_desc':           'Please open Apple Devices once, then return to iVault and retry',
    'error.amds_retry':          'Retry',

    // HEIC CONVERT
    'heic.converting': 'Converting HEIC photos...',
    'heic.done':       'Converted',
    'heic.unit':       'JPEG photos',
};
