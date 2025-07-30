export namespace assignment {
	
	export class Assignment {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    UserID: number;
	    NotionID: string;
	    Title: string;
	    Todo: string;
	    // Go type: time
	    Deadline: any;
	    Link: string;
	    CourseCode: string;
	    TypeName: string;
	    StatusName: string;
	    Priority: string;
	    Completed: boolean;
	    User: user.User;
	    Course: course.Course;
	    Type: models.AssignmentType;
	    Status: models.AssignmentStatus;
	    Documents: document.Document[];
	
	    static createFrom(source: any = {}) {
	        return new Assignment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.UserID = source["UserID"];
	        this.NotionID = source["NotionID"];
	        this.Title = source["Title"];
	        this.Todo = source["Todo"];
	        this.Deadline = this.convertValues(source["Deadline"], null);
	        this.Link = source["Link"];
	        this.CourseCode = source["CourseCode"];
	        this.TypeName = source["TypeName"];
	        this.StatusName = source["StatusName"];
	        this.Priority = source["Priority"];
	        this.Completed = source["Completed"];
	        this.User = this.convertValues(source["User"], user.User);
	        this.Course = this.convertValues(source["Course"], course.Course);
	        this.Type = this.convertValues(source["Type"], models.AssignmentType);
	        this.Status = this.convertValues(source["Status"], models.AssignmentStatus);
	        this.Documents = this.convertValues(source["Documents"], document.Document);
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
	export class LocalAssignment {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    RemoteID: number;
	    NotionID: string;
	    Title: string;
	    Todo: string;
	    // Go type: time
	    Deadline: any;
	    Link: string;
	    CourseCode: string;
	    TypeName: string;
	    StatusName: string;
	    Priority: string;
	    Completed: boolean;
	    SyncStatus: string;
	    Course: course.LocalCourse;
	    Type: models.LocalAssignmentType;
	    Status: models.LocalAssignmentStatus;
	    Documents: document.LocalDocument[];
	
	    static createFrom(source: any = {}) {
	        return new LocalAssignment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.RemoteID = source["RemoteID"];
	        this.NotionID = source["NotionID"];
	        this.Title = source["Title"];
	        this.Todo = source["Todo"];
	        this.Deadline = this.convertValues(source["Deadline"], null);
	        this.Link = source["Link"];
	        this.CourseCode = source["CourseCode"];
	        this.TypeName = source["TypeName"];
	        this.StatusName = source["StatusName"];
	        this.Priority = source["Priority"];
	        this.Completed = source["Completed"];
	        this.SyncStatus = source["SyncStatus"];
	        this.Course = this.convertValues(source["Course"], course.LocalCourse);
	        this.Type = this.convertValues(source["Type"], models.LocalAssignmentType);
	        this.Status = this.convertValues(source["Status"], models.LocalAssignmentStatus);
	        this.Documents = this.convertValues(source["Documents"], document.LocalDocument);
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

export namespace course {
	
	export class Course {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    UserID: number;
	    User: user.User;
	    NotionID: string;
	    Code: string;
	    Name: string;
	    Color: string;
	    Duration: string;
	    RoomNumber: string;
	    // Go type: time
	    StartDate: any;
	    // Go type: time
	    EndDate: any;
	    Schedule: string;
	    Credits: number;
	    Semester: string;
	    Instructor: string;
	    InstructorEmail: string;
	
	    static createFrom(source: any = {}) {
	        return new Course(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.UserID = source["UserID"];
	        this.User = this.convertValues(source["User"], user.User);
	        this.NotionID = source["NotionID"];
	        this.Code = source["Code"];
	        this.Name = source["Name"];
	        this.Color = source["Color"];
	        this.Duration = source["Duration"];
	        this.RoomNumber = source["RoomNumber"];
	        this.StartDate = this.convertValues(source["StartDate"], null);
	        this.EndDate = this.convertValues(source["EndDate"], null);
	        this.Schedule = source["Schedule"];
	        this.Credits = source["Credits"];
	        this.Semester = source["Semester"];
	        this.Instructor = source["Instructor"];
	        this.InstructorEmail = source["InstructorEmail"];
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
	export class LocalCourse {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    RemoteID: number;
	    Code: string;
	    Name: string;
	    NotionID: string;
	    Duration: string;
	    RoomNumber: string;
	    Color: string;
	    // Go type: time
	    StartDate: any;
	    // Go type: time
	    EndDate: any;
	    Credits: number;
	    Schedule: string;
	    Semester: string;
	    Instructor: string;
	    InstructorEmail: string;
	    SyncStatus: string;
	
	    static createFrom(source: any = {}) {
	        return new LocalCourse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.RemoteID = source["RemoteID"];
	        this.Code = source["Code"];
	        this.Name = source["Name"];
	        this.NotionID = source["NotionID"];
	        this.Duration = source["Duration"];
	        this.RoomNumber = source["RoomNumber"];
	        this.Color = source["Color"];
	        this.StartDate = this.convertValues(source["StartDate"], null);
	        this.EndDate = this.convertValues(source["EndDate"], null);
	        this.Credits = source["Credits"];
	        this.Schedule = source["Schedule"];
	        this.Semester = source["Semester"];
	        this.Instructor = source["Instructor"];
	        this.InstructorEmail = source["InstructorEmail"];
	        this.SyncStatus = source["SyncStatus"];
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

export namespace document {
	
	export class Document {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    AssignmentID: number;
	    UserID: number;
	    LocalID: number;
	    Type: string;
	    FileName: string;
	    FileType: string;
	    FilePath: string;
	    FileSize: number;
	    Version: number;
	    ParentDocID?: number;
	    IsOriginal: boolean;
	    User: user.User;
	    ParentDoc?: Document;
	    Versions: Document[];
	
	    static createFrom(source: any = {}) {
	        return new Document(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.AssignmentID = source["AssignmentID"];
	        this.UserID = source["UserID"];
	        this.LocalID = source["LocalID"];
	        this.Type = source["Type"];
	        this.FileName = source["FileName"];
	        this.FileType = source["FileType"];
	        this.FilePath = source["FilePath"];
	        this.FileSize = source["FileSize"];
	        this.Version = source["Version"];
	        this.ParentDocID = source["ParentDocID"];
	        this.IsOriginal = source["IsOriginal"];
	        this.User = this.convertValues(source["User"], user.User);
	        this.ParentDoc = this.convertValues(source["ParentDoc"], Document);
	        this.Versions = this.convertValues(source["Versions"], Document);
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
	export class LocalDocument {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    AssignmentID: number;
	    UserID: number;
	    Type: string;
	    FileName: string;
	    FileType: string;
	    FilePath: string;
	    FileSize: number;
	    Version: number;
	    ParentDocID?: number;
	    IsOriginal: boolean;
	    HasLocalFile: boolean;
	    // Go type: time
	    LastSyncAt?: any;
	    ParentDoc?: LocalDocument;
	    Versions: LocalDocument[];
	
	    static createFrom(source: any = {}) {
	        return new LocalDocument(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.AssignmentID = source["AssignmentID"];
	        this.UserID = source["UserID"];
	        this.Type = source["Type"];
	        this.FileName = source["FileName"];
	        this.FileType = source["FileType"];
	        this.FilePath = source["FilePath"];
	        this.FileSize = source["FileSize"];
	        this.Version = source["Version"];
	        this.ParentDocID = source["ParentDocID"];
	        this.IsOriginal = source["IsOriginal"];
	        this.HasLocalFile = source["HasLocalFile"];
	        this.LastSyncAt = this.convertValues(source["LastSyncAt"], null);
	        this.ParentDoc = this.convertValues(source["ParentDoc"], LocalDocument);
	        this.Versions = this.convertValues(source["Versions"], LocalDocument);
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
	export class LocalDocumentCache {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    UserID: number;
	    TotalSize: number;
	    DocumentCount: number;
	    // Go type: time
	    LastCalculatedAt: any;
	    // Go type: time
	    LastSyncAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new LocalDocumentCache(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.UserID = source["UserID"];
	        this.TotalSize = source["TotalSize"];
	        this.DocumentCount = source["DocumentCount"];
	        this.LastCalculatedAt = this.convertValues(source["LastCalculatedAt"], null);
	        this.LastSyncAt = this.convertValues(source["LastSyncAt"], null);
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

export namespace models {
	
	export class AssignmentStatus {
	    ID: number;
	    Name: string;
	    Color: string;
	    NotionID: string;
	
	    static createFrom(source: any = {}) {
	        return new AssignmentStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Color = source["Color"];
	        this.NotionID = source["NotionID"];
	    }
	}
	export class AssignmentType {
	    ID: number;
	    Name: string;
	    Color: string;
	    NotionID: string;
	
	    static createFrom(source: any = {}) {
	        return new AssignmentType(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Color = source["Color"];
	        this.NotionID = source["NotionID"];
	    }
	}
	export class LocalAssignmentStatus {
	    ID: number;
	    Name: string;
	    Color: string;
	    NotionID: string;
	
	    static createFrom(source: any = {}) {
	        return new LocalAssignmentStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Color = source["Color"];
	        this.NotionID = source["NotionID"];
	    }
	}
	export class LocalAssignmentType {
	    ID: number;
	    Name: string;
	    Color: string;
	    NotionID: string;
	
	    static createFrom(source: any = {}) {
	        return new LocalAssignmentType(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Color = source["Color"];
	        this.NotionID = source["NotionID"];
	    }
	}

}

export namespace storage {
	
	export class LocalCredentials {
	    is_authenticated: boolean;
	    // Go type: struct { UserID uint "json:\"user_id\""; Username string "json:\"username\"" }
	    user: any;
	
	    static createFrom(source: any = {}) {
	        return new LocalCredentials(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_authenticated = source["is_authenticated"];
	        this.user = this.convertValues(source["user"], Object);
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

export namespace user {
	
	export class User {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    Username: string;
	    Email: string;
	    PasswordHash: string;
	    NotionAPIKey: string;
	    AssignmentsDbId: string;
	    NotionID: string;
	    CoursesDbId: string;
	    // Go type: time
	    LastSync?: any;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.Username = source["Username"];
	        this.Email = source["Email"];
	        this.PasswordHash = source["PasswordHash"];
	        this.NotionAPIKey = source["NotionAPIKey"];
	        this.AssignmentsDbId = source["AssignmentsDbId"];
	        this.NotionID = source["NotionID"];
	        this.CoursesDbId = source["CoursesDbId"];
	        this.LastSync = this.convertValues(source["LastSync"], null);
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

