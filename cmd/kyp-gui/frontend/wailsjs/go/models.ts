export namespace gui {
	
	export class EntryDTO {
	    id?: string;
	    title: string;
	    username?: string;
	    password?: string;
	    url?: string;
	    notes?: string;
	    totpSecret?: string;
	    totpIssuer?: string;
	    totpDigits: number;
	    totpPeriod: number;
	    totpAlgorithm: string;
	
	    static createFrom(source: any = {}) {
	        return new EntryDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.url = source["url"];
	        this.notes = source["notes"];
	        this.totpSecret = source["totpSecret"];
	        this.totpIssuer = source["totpIssuer"];
	        this.totpDigits = source["totpDigits"];
	        this.totpPeriod = source["totpPeriod"];
	        this.totpAlgorithm = source["totpAlgorithm"];
	    }
	}
	export class TOTPState {
	    code: string;
	    remaining: number;
	
	    static createFrom(source: any = {}) {
	        return new TOTPState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.remaining = source["remaining"];
	    }
	}

}

