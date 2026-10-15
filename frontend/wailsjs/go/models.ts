export namespace main {
	
	export class HistoryItem {
	    title: string;
	    path: string;
	    url: string;
	    finished: string;
	
	    static createFrom(source: any = {}) {
	        return new HistoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.path = source["path"];
	        this.url = source["url"];
	        this.finished = source["finished"];
	    }
	}
	export class Preferences {
	    language: string;
	    downloadDir: string;
	
	    static createFrom(source: any = {}) {
	        return new Preferences(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.downloadDir = source["downloadDir"];
	    }
	}

}

