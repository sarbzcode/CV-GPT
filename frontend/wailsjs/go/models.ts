export namespace matcher {
	
	export class JDExtract {
	    role_title: string;
	    skills_must: string[];
	    skills_nice: string[];
	    skills_other: string[];
	    years_experience_min: number;
	    education: string[];
	    certifications: string[];
	    titles: string[];
	    responsibilities: string[];
	
	    static createFrom(source: any = {}) {
	        return new JDExtract(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role_title = source["role_title"];
	        this.skills_must = source["skills_must"];
	        this.skills_nice = source["skills_nice"];
	        this.skills_other = source["skills_other"];
	        this.years_experience_min = source["years_experience_min"];
	        this.education = source["education"];
	        this.certifications = source["certifications"];
	        this.titles = source["titles"];
	        this.responsibilities = source["responsibilities"];
	    }
	}
	export class ResumeExtract {
	    skills: string[];
	    years_experience: number;
	    education: string[];
	    certifications: string[];
	    titles: string[];
	
	    static createFrom(source: any = {}) {
	        return new ResumeExtract(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skills = source["skills"];
	        this.years_experience = source["years_experience"];
	        this.education = source["education"];
	        this.certifications = source["certifications"];
	        this.titles = source["titles"];
	    }
	}
	export class Result {
	    rank: number;
	    candidate: string;
	    score: number;
	    strengths: string;
	    weaknesses: string;
	    explanation: string;
	    file: string;
	    extracted?: ResumeExtract;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rank = source["rank"];
	        this.candidate = source["candidate"];
	        this.score = source["score"];
	        this.strengths = source["strengths"];
	        this.weaknesses = source["weaknesses"];
	        this.explanation = source["explanation"];
	        this.file = source["file"];
	        this.extracted = this.convertValues(source["extracted"], ResumeExtract);
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
	export class Output {
	    results: Result[];
	    outPath: string;
	    total: number;
	    jdInfo?: JDExtract;
	
	    static createFrom(source: any = {}) {
	        return new Output(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.results = this.convertValues(source["results"], Result);
	        this.outPath = source["outPath"];
	        this.total = source["total"];
	        this.jdInfo = this.convertValues(source["jdInfo"], JDExtract);
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
	
	export class ResumeAnalysis {
	    strengths: string[];
	    weaknesses: string[];
	    summary: string;
	    skills: string[];
	    years_experience: number;
	    education: string[];
	    certifications: string[];
	    titles: string[];
	
	    static createFrom(source: any = {}) {
	        return new ResumeAnalysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.strengths = source["strengths"];
	        this.weaknesses = source["weaknesses"];
	        this.summary = source["summary"];
	        this.skills = source["skills"];
	        this.years_experience = source["years_experience"];
	        this.education = source["education"];
	        this.certifications = source["certifications"];
	        this.titles = source["titles"];
	    }
	}

}

