export namespace backup {
	
	export class CopyResult {
	    remotePath: string;
	    localPath: string;
	    bytesCopied: number;
	
	    static createFrom(source: any = {}) {
	        return new CopyResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.remotePath = source["remotePath"];
	        this.localPath = source["localPath"];
	        this.bytesCopied = source["bytesCopied"];
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
	export class PhotoFile {
	    remotePath: string;
	    fileName: string;
	    size: number;
	    modTime: number;
	
	    static createFrom(source: any = {}) {
	        return new PhotoFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.remotePath = source["remotePath"];
	        this.fileName = source["fileName"];
	        this.size = source["size"];
	        this.modTime = source["modTime"];
	    }
	}

}

