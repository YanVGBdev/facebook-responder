export namespace main {
	
	export class CommentDTO {
	    id: string;
	    autor: string;
	    texto: string;
	    created_time?: string;
	    status: string;
	    sugestao_ia: string;
	    resposta_final: string;
	    respondido_em?: string;
	
	    static createFrom(source: any = {}) {
	        return new CommentDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.autor = source["autor"];
	        this.texto = source["texto"];
	        this.created_time = source["created_time"];
	        this.status = source["status"];
	        this.sugestao_ia = source["sugestao_ia"];
	        this.resposta_final = source["resposta_final"];
	        this.respondido_em = source["respondido_em"];
	    }
	}
	export class ConfigDTO {
	    page_id: string;
	    page_access_token: string;
	    gemini_api_key: string;
	    perfil_empresa: string;
	    usar_mock: boolean;
	    configurado: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page_id = source["page_id"];
	        this.page_access_token = source["page_access_token"];
	        this.gemini_api_key = source["gemini_api_key"];
	        this.perfil_empresa = source["perfil_empresa"];
	        this.usar_mock = source["usar_mock"];
	        this.configurado = source["configurado"];
	    }
	}
	export class PostDTO {
	    id: string;
	    texto_resumo: string;
	    created_time?: string;
	    total: number;
	    pendentes: number;
	    comentarios: CommentDTO[];
	
	    static createFrom(source: any = {}) {
	        return new PostDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.texto_resumo = source["texto_resumo"];
	        this.created_time = source["created_time"];
	        this.total = source["total"];
	        this.pendentes = source["pendentes"];
	        this.comentarios = this.convertValues(source["comentarios"], CommentDTO);
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
	export class PostsListDTO {
	    posts: PostDTO[];
	
	    static createFrom(source: any = {}) {
	        return new PostsListDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.posts = this.convertValues(source["posts"], PostDTO);
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

