export namespace backup {
	
	export class BackupConfig {
	    deviceUdid: string;
	    deviceName: string;
	    backupPath: string;
	    convertHeic: boolean;
	    organizeByDate: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BackupConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deviceUdid = source["deviceUdid"];
	        this.deviceName = source["deviceName"];
	        this.backupPath = source["backupPath"];
	        this.convertHeic = source["convertHeic"];
	        this.organizeByDate = source["organizeByDate"];
	    }
	}

}

export namespace config {
	
	export class BackupRecord {
	    date: string;
	    deviceName: string;
	    deviceUdid: string;
	    newFiles: number;
	    photosCount: number;
	    videosCount: number;
	    skipped: number;
	    failed: number;
	    totalBytes: number;
	    duration: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.deviceName = source["deviceName"];
	        this.deviceUdid = source["deviceUdid"];
	        this.newFiles = source["newFiles"];
	        this.photosCount = source["photosCount"];
	        this.videosCount = source["videosCount"];
	        this.skipped = source["skipped"];
	        this.failed = source["failed"];
	        this.totalBytes = source["totalBytes"];
	        this.duration = source["duration"];
	    }
	}
	export class AppConfig {
	    lastBackupPath: string;
	    convertHeic: boolean;
	    organizeByDate: boolean;
	    history: BackupRecord[];
	    lastInterrupted: boolean;
	    interruptedDone: number;
	    interruptedTotal: number;
	    firstBackupDone: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lastBackupPath = source["lastBackupPath"];
	        this.convertHeic = source["convertHeic"];
	        this.organizeByDate = source["organizeByDate"];
	        this.history = this.convertValues(source["history"], BackupRecord);
	        this.lastInterrupted = source["lastInterrupted"];
	        this.interruptedDone = source["interruptedDone"];
	        this.interruptedTotal = source["interruptedTotal"];
	        this.firstBackupDone = source["firstBackupDone"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace device {
	
	export class DeviceDetail {
	    udid: string;
	    name: string;
	    model: string;
	    iosVersion: string;
	    trusted: boolean;
	    photoCount: number;
	    usedSpace: number;
	    totalSpace: number;
	
	    static createFrom(source: any = {}) {
	        return new DeviceDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.udid = source["udid"];
	        this.name = source["name"];
	        this.model = source["model"];
	        this.iosVersion = source["iosVersion"];
	        this.trusted = source["trusted"];
	        this.photoCount = source["photoCount"];
	        this.usedSpace = source["usedSpace"];
	        this.totalSpace = source["totalSpace"];
	    }
	}
	export class DeviceInfo {
	    udid: string;
	    name: string;
	    model: string;
	    iosVersion: string;
	    trusted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.udid = source["udid"];
	        this.name = source["name"];
	        this.model = source["model"];
	        this.iosVersion = source["iosVersion"];
	        this.trusted = source["trusted"];
	    }
	}

}

export namespace platform {
	
	export class DiskInfo {
	    path: string;
	    totalSpace: number;
	    freeSpace: number;
	    isSystem: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DiskInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.totalSpace = source["totalSpace"];
	        this.freeSpace = source["freeSpace"];
	        this.isSystem = source["isSystem"];
	    }
	}
	export class Info {
	    os: string;
	    arch: string;
	    appleDevicesInstalled: boolean;
	    heicSupported: boolean;
	    darkMode: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.appleDevicesInstalled = source["appleDevicesInstalled"];
	        this.heicSupported = source["heicSupported"];
	        this.darkMode = source["darkMode"];
	    }
	}

}

