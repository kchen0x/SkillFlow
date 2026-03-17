export namespace backup {
	
	export class RemoteFile {
	    path: string;
	    size: number;
	    isDir: boolean;
	    action?: string;
	
	    static createFrom(source: any = {}) {
	        return new RemoteFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.size = source["size"];
	        this.isDir = source["isDir"];
	        this.action = source["action"];
	    }
	}

}

export namespace config {
	
	export class AgentConfig {
	    name: string;
	    scanDirs: string[];
	    pushDir: string;
	    enabled: boolean;
	    custom: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AgentConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.scanDirs = source["scanDirs"];
	        this.pushDir = source["pushDir"];
	        this.enabled = source["enabled"];
	        this.custom = source["custom"];
	    }
	}
	export class ProxyConfig {
	    mode: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.url = source["url"];
	    }
	}
	export class CloudProviderConfig {
	    bucketName: string;
	    remotePath: string;
	    credentials: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new CloudProviderConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bucketName = source["bucketName"];
	        this.remotePath = source["remotePath"];
	        this.credentials = source["credentials"];
	    }
	}
	export class CloudConfig {
	    provider: string;
	    enabled: boolean;
	    bucketName: string;
	    remotePath: string;
	    credentials: Record<string, string>;
	    syncIntervalMinutes: number;
	
	    static createFrom(source: any = {}) {
	        return new CloudConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.enabled = source["enabled"];
	        this.bucketName = source["bucketName"];
	        this.remotePath = source["remotePath"];
	        this.credentials = source["credentials"];
	        this.syncIntervalMinutes = source["syncIntervalMinutes"];
	    }
	}
	export class SkillStatusVisibilityConfig {
	    mySkills: string[];
	    myAgents: string[];
	    pushToAgent: string[];
	    pullFromAgent: string[];
	    starredRepos: string[];
	    githubInstall: string[];
	
	    static createFrom(source: any = {}) {
	        return new SkillStatusVisibilityConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mySkills = source["mySkills"];
	        this.myAgents = source["myAgents"];
	        this.pushToAgent = source["pushToAgent"];
	        this.pullFromAgent = source["pullFromAgent"];
	        this.starredRepos = source["starredRepos"];
	        this.githubInstall = source["githubInstall"];
	    }
	}
	export class AppConfig {
	    skillsStorageDir: string;
	    autoPushAgents: string[];
	    launchAtLogin: boolean;
	    defaultCategory: string;
	    logLevel: string;
	    repoScanMaxDepth: number;
	    skillStatusVisibility: SkillStatusVisibilityConfig;
	    agents: AgentConfig[];
	    cloud: CloudConfig;
	    cloudProfiles?: Record<string, CloudProviderConfig>;
	    proxy: ProxyConfig;
	    skippedUpdateVersion?: string;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillsStorageDir = source["skillsStorageDir"];
	        this.autoPushAgents = source["autoPushAgents"];
	        this.launchAtLogin = source["launchAtLogin"];
	        this.defaultCategory = source["defaultCategory"];
	        this.logLevel = source["logLevel"];
	        this.repoScanMaxDepth = source["repoScanMaxDepth"];
	        this.skillStatusVisibility = this.convertValues(source["skillStatusVisibility"], SkillStatusVisibilityConfig);
	        this.agents = this.convertValues(source["agents"], AgentConfig);
	        this.cloud = this.convertValues(source["cloud"], CloudConfig);
	        this.cloudProfiles = this.convertValues(source["cloudProfiles"], CloudProviderConfig, true);
	        this.proxy = this.convertValues(source["proxy"], ProxyConfig);
	        this.skippedUpdateVersion = source["skippedUpdateVersion"];
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

export namespace git {
	
	export class StarSkill {
	    name: string;
	    path: string;
	    subPath: string;
	    repoUrl: string;
	    repoName: string;
	    source: string;
	    logicalKey: string;
	    installed: boolean;
	    imported: boolean;
	    updatable: boolean;
	    pushed: boolean;
	    pushedAgents: string[];
	
	    static createFrom(source: any = {}) {
	        return new StarSkill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.subPath = source["subPath"];
	        this.repoUrl = source["repoUrl"];
	        this.repoName = source["repoName"];
	        this.source = source["source"];
	        this.logicalKey = source["logicalKey"];
	        this.installed = source["installed"];
	        this.imported = source["imported"];
	        this.updatable = source["updatable"];
	        this.pushed = source["pushed"];
	        this.pushedAgents = source["pushedAgents"];
	    }
	}
	export class StarredRepo {
	    url: string;
	    name: string;
	    source: string;
	    localDir: string;
	    // Go type: time
	    lastSync: any;
	    syncError?: string;
	
	    static createFrom(source: any = {}) {
	        return new StarredRepo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.name = source["name"];
	        this.source = source["source"];
	        this.localDir = source["localDir"];
	        this.lastSync = this.convertValues(source["lastSync"], null);
	        this.syncError = source["syncError"];
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

export namespace install {
	
	export class SkillCandidate {
	    Name: string;
	    Path: string;
	    LogicalKey: string;
	    Installed: boolean;
	    Updatable: boolean;
	    Pushed: boolean;
	    PushedAgents: string[];
	
	    static createFrom(source: any = {}) {
	        return new SkillCandidate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Path = source["Path"];
	        this.LogicalKey = source["LogicalKey"];
	        this.Installed = source["Installed"];
	        this.Updatable = source["Updatable"];
	        this.Pushed = source["Pushed"];
	        this.PushedAgents = source["PushedAgents"];
	    }
	}

}

export namespace main {
	
	export class AgentSkillCandidate {
	    name: string;
	    path: string;
	    source: string;
	    logicalKey: string;
	    installed: boolean;
	    imported: boolean;
	    updatable: boolean;
	    pushed: boolean;
	    pushedAgents: string[];
	
	    static createFrom(source: any = {}) {
	        return new AgentSkillCandidate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.source = source["source"];
	        this.logicalKey = source["logicalKey"];
	        this.installed = source["installed"];
	        this.imported = source["imported"];
	        this.updatable = source["updatable"];
	        this.pushed = source["pushed"];
	        this.pushedAgents = source["pushedAgents"];
	    }
	}
	export class AgentSkillEntry {
	    name: string;
	    path: string;
	    source: string;
	    logicalKey: string;
	    installed: boolean;
	    imported: boolean;
	    updatable: boolean;
	    pushed: boolean;
	    pushedAgents: string[];
	    seenInAgentScan: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AgentSkillEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.source = source["source"];
	        this.logicalKey = source["logicalKey"];
	        this.installed = source["installed"];
	        this.imported = source["imported"];
	        this.updatable = source["updatable"];
	        this.pushed = source["pushed"];
	        this.pushedAgents = source["pushedAgents"];
	        this.seenInAgentScan = source["seenInAgentScan"];
	    }
	}
	export class AppUpdateInfo {
	    hasUpdate: boolean;
	    currentVersion: string;
	    latestVersion: string;
	    releaseUrl: string;
	    downloadUrl: string;
	    releaseNotes: string;
	    canAutoUpdate: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppUpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasUpdate = source["hasUpdate"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.releaseUrl = source["releaseUrl"];
	        this.downloadUrl = source["downloadUrl"];
	        this.releaseNotes = source["releaseNotes"];
	        this.canAutoUpdate = source["canAutoUpdate"];
	    }
	}
	export class InstalledSkillEntry {
	    id: string;
	    name: string;
	    path: string;
	    category: string;
	    source: string;
	    sourceSha: string;
	    latestSha: string;
	    updatable: boolean;
	    pushed: boolean;
	    pushedAgents: string[];
	
	    static createFrom(source: any = {}) {
	        return new InstalledSkillEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.category = source["category"];
	        this.source = source["source"];
	        this.sourceSha = source["sourceSha"];
	        this.latestSha = source["latestSha"];
	        this.updatable = source["updatable"];
	        this.pushed = source["pushed"];
	        this.pushedAgents = source["pushedAgents"];
	    }
	}
	export class PromptImportPrepareResult {
	    sessionId: string;
	    creates: prompt.ImportPrompt[];
	    conflicts: prompt.ImportPrompt[];
	
	    static createFrom(source: any = {}) {
	        return new PromptImportPrepareResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.creates = this.convertValues(source["creates"], prompt.ImportPrompt);
	        this.conflicts = this.convertValues(source["conflicts"], prompt.ImportPrompt);
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
	export class PushConflict {
	    skillId?: string;
	    skillName: string;
	    skillPath?: string;
	    agentName: string;
	    targetPath: string;
	
	    static createFrom(source: any = {}) {
	        return new PushConflict(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.skillId = source["skillId"];
	        this.skillName = source["skillName"];
	        this.skillPath = source["skillPath"];
	        this.agentName = source["agentName"];
	        this.targetPath = source["targetPath"];
	    }
	}

}

export namespace prompt {
	
	export class PromptLink {
	    label: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new PromptLink(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.url = source["url"];
	    }
	}
	export class ImportPrompt {
	    name: string;
	    description?: string;
	    category: string;
	    content: string;
	    imageURLs?: string[];
	    webLinks?: PromptLink[];
	
	    static createFrom(source: any = {}) {
	        return new ImportPrompt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.content = source["content"];
	        this.imageURLs = source["imageURLs"];
	        this.webLinks = this.convertValues(source["webLinks"], PromptLink);
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
	export class Prompt {
	    name: string;
	    description?: string;
	    category: string;
	    path: string;
	    filePath: string;
	    content: string;
	    imageURLs?: string[];
	    webLinks?: PromptLink[];
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Prompt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.path = source["path"];
	        this.filePath = source["filePath"];
	        this.content = source["content"];
	        this.imageURLs = source["imageURLs"];
	        this.webLinks = this.convertValues(source["webLinks"], PromptLink);
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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

export namespace skill {
	
	export class Skill {
	    ID: string;
	    Name: string;
	    Path: string;
	    Category: string;
	    Source: string;
	    SourceURL: string;
	    SourceSubPath: string;
	    SourceSHA: string;
	    LatestSHA: string;
	    // Go type: time
	    InstalledAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: time
	    LastCheckedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Skill(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Path = source["Path"];
	        this.Category = source["Category"];
	        this.Source = source["Source"];
	        this.SourceURL = source["SourceURL"];
	        this.SourceSubPath = source["SourceSubPath"];
	        this.SourceSHA = source["SourceSHA"];
	        this.LatestSHA = source["LatestSHA"];
	        this.InstalledAt = this.convertValues(source["InstalledAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.LastCheckedAt = this.convertValues(source["LastCheckedAt"], null);
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
	export class SkillMeta {
	    Name: string;
	    Description: string;
	    ArgumentHint: string;
	    AllowedTools: string;
	    Context: string;
	    DisableModelInvocation: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SkillMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.ArgumentHint = source["ArgumentHint"];
	        this.AllowedTools = source["AllowedTools"];
	        this.Context = source["Context"];
	        this.DisableModelInvocation = source["DisableModelInvocation"];
	    }
	}

}

