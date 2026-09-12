export namespace binmanager {
	
	export class BinaryStatus {
	    YtDlpPath: string;
	    YtDlpFound: boolean;
	    FFmpegPath: string;
	    FFmpegFound: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BinaryStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.YtDlpPath = source["YtDlpPath"];
	        this.YtDlpFound = source["YtDlpFound"];
	        this.FFmpegPath = source["FFmpegPath"];
	        this.FFmpegFound = source["FFmpegFound"];
	    }
	}

}

export namespace config {
	
	export class Settings {
	    DownloadDir: string;
	    OutputTemplate: string;
	    Concurrency: number;
	    Proxy: string;
	    YtDlpPath: string;
	    FFmpegPath: string;
	    Cookies: ytdlp.CookieSource;
	    Theme: string;
	    Language: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.DownloadDir = source["DownloadDir"];
	        this.OutputTemplate = source["OutputTemplate"];
	        this.Concurrency = source["Concurrency"];
	        this.Proxy = source["Proxy"];
	        this.YtDlpPath = source["YtDlpPath"];
	        this.FFmpegPath = source["FFmpegPath"];
	        this.Cookies = this.convertValues(source["Cookies"], ytdlp.CookieSource);
	        this.Theme = source["Theme"];
	        this.Language = source["Language"];
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

export namespace ytdlp {
	
	export class CookieSource {
	    Kind: string;
	    Browser: string;
	    FilePath: string;
	
	    static createFrom(source: any = {}) {
	        return new CookieSource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Kind = source["Kind"];
	        this.Browser = source["Browser"];
	        this.FilePath = source["FilePath"];
	    }
	}
	export class DownloadOptions {
	    URL: string;
	    FormatID: string;
	    OutputDir: string;
	    OutputTemplate: string;
	    AudioOnly: boolean;
	    AudioFormat: string;
	    Subtitles: boolean;
	    SubLangs: string;
	    EmbedSubs: boolean;
	    Playlist: boolean;
	    PlaylistItems: string;
	    Proxy: string;
	    Cookies: CookieSource;
	    ExtraArgs: string[];
	
	    static createFrom(source: any = {}) {
	        return new DownloadOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.URL = source["URL"];
	        this.FormatID = source["FormatID"];
	        this.OutputDir = source["OutputDir"];
	        this.OutputTemplate = source["OutputTemplate"];
	        this.AudioOnly = source["AudioOnly"];
	        this.AudioFormat = source["AudioFormat"];
	        this.Subtitles = source["Subtitles"];
	        this.SubLangs = source["SubLangs"];
	        this.EmbedSubs = source["EmbedSubs"];
	        this.Playlist = source["Playlist"];
	        this.PlaylistItems = source["PlaylistItems"];
	        this.Proxy = source["Proxy"];
	        this.Cookies = this.convertValues(source["Cookies"], CookieSource);
	        this.ExtraArgs = source["ExtraArgs"];
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
	export class Format {
	    format_id: string;
	    ext: string;
	    resolution: string;
	    format_note: string;
	    vcodec: string;
	    acodec: string;
	    fps?: number;
	    filesize?: number;
	    filesize_approx?: number;
	    tbr?: number;
	
	    static createFrom(source: any = {}) {
	        return new Format(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format_id = source["format_id"];
	        this.ext = source["ext"];
	        this.resolution = source["resolution"];
	        this.format_note = source["format_note"];
	        this.vcodec = source["vcodec"];
	        this.acodec = source["acodec"];
	        this.fps = source["fps"];
	        this.filesize = source["filesize"];
	        this.filesize_approx = source["filesize_approx"];
	        this.tbr = source["tbr"];
	    }
	}
	export class PlaylistEntry {
	    id: string;
	    title: string;
	    url: string;
	    duration: number;
	
	    static createFrom(source: any = {}) {
	        return new PlaylistEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.url = source["url"];
	        this.duration = source["duration"];
	    }
	}
	export class Metadata {
	    id: string;
	    title: string;
	    thumbnail: string;
	    duration: number;
	    formats: Format[];
	    _type: string;
	    entries: PlaylistEntry[];
	
	    static createFrom(source: any = {}) {
	        return new Metadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.thumbnail = source["thumbnail"];
	        this.duration = source["duration"];
	        this.formats = this.convertValues(source["formats"], Format);
	        this._type = source["_type"];
	        this.entries = this.convertValues(source["entries"], PlaylistEntry);
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

