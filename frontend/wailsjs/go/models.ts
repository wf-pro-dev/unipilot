export namespace client {
	
	export class RemoteUser {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    Username: string;
	    Email: string;
	    Password: string;
	    PasswordHash: string;
	    Avatar: string;
	    University: string;
	    Semester: string;
	    Year: string;
	    IsVerified: boolean;
	    Language: string;
	    CoursesCode: string[];
	    // Go type: time
	    LastSync?: any;
	    Courses: models.Course[];
	    Assignments: models.Assignment[];
	    Notes: models.Note[];
	    OwnerRequests: models.CourseLinkRequest[];
	    ReceiverRequests: models.CourseLinkRequest[];
	    Followers: models.User[];
	    Following: models.User[];
	    CoursesCode: string[];
	
	    static createFrom(source: any = {}) {
	        return new RemoteUser(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.Username = source["Username"];
	        this.Email = source["Email"];
	        this.Password = source["Password"];
	        this.PasswordHash = source["PasswordHash"];
	        this.Avatar = source["Avatar"];
	        this.University = source["University"];
	        this.Semester = source["Semester"];
	        this.Year = source["Year"];
	        this.IsVerified = source["IsVerified"];
	        this.Language = source["Language"];
	        this.CoursesCode = source["CoursesCode"];
	        this.LastSync = this.convertValues(source["LastSync"], null);
	        this.Courses = this.convertValues(source["Courses"], models.Course);
	        this.Assignments = this.convertValues(source["Assignments"], models.Assignment);
	        this.Notes = this.convertValues(source["Notes"], models.Note);
	        this.OwnerRequests = this.convertValues(source["OwnerRequests"], models.CourseLinkRequest);
	        this.ReceiverRequests = this.convertValues(source["ReceiverRequests"], models.CourseLinkRequest);
	        this.Followers = this.convertValues(source["Followers"], models.User);
	        this.Following = this.convertValues(source["Following"], models.User);
	        this.CoursesCode = source["CoursesCode"];
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

export namespace fileops {
	
	export class FileUploadRequest {
	    UploadID: string;
	    AssignmentID: number;
	    RemoteAssignmentID: number;
	    UserID: number;
	    Type: string;
	    FileName: string;
	    FilePath: string;
	    FileSize: number;
	    FileContent: any;
	    StorageKey: string;
	
	    static createFrom(source: any = {}) {
	        return new FileUploadRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.UploadID = source["UploadID"];
	        this.AssignmentID = source["AssignmentID"];
	        this.RemoteAssignmentID = source["RemoteAssignmentID"];
	        this.UserID = source["UserID"];
	        this.Type = source["Type"];
	        this.FileName = source["FileName"];
	        this.FilePath = source["FilePath"];
	        this.FileSize = source["FileSize"];
	        this.FileContent = source["FileContent"];
	        this.StorageKey = source["StorageKey"];
	    }
	}
	export class FileUploadResponse {
	    LocalDocument?: models.LocalDocument;
	    Success: boolean;
	    Message: string;
	
	    static createFrom(source: any = {}) {
	        return new FileUploadResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.LocalDocument = this.convertValues(source["LocalDocument"], models.LocalDocument);
	        this.Success = source["Success"];
	        this.Message = source["Message"];
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

export namespace gorm {
	
	export class DeletedAt {
	    // Go type: time
	    Time: any;
	    Valid: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DeletedAt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Time = this.convertValues(source["Time"], null);
	        this.Valid = source["Valid"];
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

export namespace main {
	
	export class FileInfo {
	    FileName: string;
	    FileSize: number;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.FileName = source["FileName"];
	        this.FileSize = source["FileSize"];
	    }
	}
	export class FollowResponse {
	    users: models.User[];
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new FollowResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.users = this.convertValues(source["users"], models.User);
	        this.count = source["count"];
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
	
	export class Document {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    Type: string;
	    FileName: string;
	    FilePath: string;
	    FileSize: number;
	    StorageKey: string;
	    Version: number;
	    ParentID: number;
	    ParentDocID?: number;
	    IsOriginal: boolean;
	    HasLocalFile: boolean;
	    AssignmentID: number;
	    UploadID: string;
	    UserID: number;
	    User?: User;
	    Assignment?: Assignment;
	    Parent?: Document;
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
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.Type = source["Type"];
	        this.FileName = source["FileName"];
	        this.FilePath = source["FilePath"];
	        this.FileSize = source["FileSize"];
	        this.StorageKey = source["StorageKey"];
	        this.Version = source["Version"];
	        this.ParentID = source["ParentID"];
	        this.ParentDocID = source["ParentDocID"];
	        this.IsOriginal = source["IsOriginal"];
	        this.HasLocalFile = source["HasLocalFile"];
	        this.AssignmentID = source["AssignmentID"];
	        this.UploadID = source["UploadID"];
	        this.UserID = source["UserID"];
	        this.User = this.convertValues(source["User"], User);
	        this.Assignment = this.convertValues(source["Assignment"], Assignment);
	        this.Parent = this.convertValues(source["Parent"], Document);
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
	export class CourseLinkRequest {
	    ID: number;
	    OwnerID: number;
	    ReceiverID: number;
	    CourseID: number;
	    Owner: User;
	    Receiver: User;
	    Course: Course;
	
	    static createFrom(source: any = {}) {
	        return new CourseLinkRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.OwnerID = source["OwnerID"];
	        this.ReceiverID = source["ReceiverID"];
	        this.CourseID = source["CourseID"];
	        this.Owner = this.convertValues(source["Owner"], User);
	        this.Receiver = this.convertValues(source["Receiver"], User);
	        this.Course = this.convertValues(source["Course"], Course);
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
	export class Note {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    Title: string;
	    Subject: string;
	    Content: string;
	    Videos: string;
	    ParentID: number;
	    CourseID: number;
	    CourseCode: string;
	    UserID: number;
	    User: User;
	    Course: Course;
	    Parent?: Note;
	    Children: Note[];
	
	    static createFrom(source: any = {}) {
	        return new Note(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.Title = source["Title"];
	        this.Subject = source["Subject"];
	        this.Content = source["Content"];
	        this.Videos = source["Videos"];
	        this.ParentID = source["ParentID"];
	        this.CourseID = source["CourseID"];
	        this.CourseCode = source["CourseCode"];
	        this.UserID = source["UserID"];
	        this.User = this.convertValues(source["User"], User);
	        this.Course = this.convertValues(source["Course"], Course);
	        this.Parent = this.convertValues(source["Parent"], Note);
	        this.Children = this.convertValues(source["Children"], Note);
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
	export class Course {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    Code: string;
	    Name: string;
	    Color: string;
	    Location: string;
	    // Go type: time
	    StartDate: any;
	    // Go type: time
	    EndDate: any;
	    Schedule: string;
	    Credits: number;
	    Semester: string;
	    Instructor: string;
	    InstructorEmail: string;
	    ParentID: number;
	    UserID: number;
	    Parent?: Course;
	    Children: Course[];
	    User?: User;
	    Assignments: Assignment[];
	    Notes: Note[];
	
	    static createFrom(source: any = {}) {
	        return new Course(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.Code = source["Code"];
	        this.Name = source["Name"];
	        this.Color = source["Color"];
	        this.Location = source["Location"];
	        this.StartDate = this.convertValues(source["StartDate"], null);
	        this.EndDate = this.convertValues(source["EndDate"], null);
	        this.Schedule = source["Schedule"];
	        this.Credits = source["Credits"];
	        this.Semester = source["Semester"];
	        this.Instructor = source["Instructor"];
	        this.InstructorEmail = source["InstructorEmail"];
	        this.ParentID = source["ParentID"];
	        this.UserID = source["UserID"];
	        this.Parent = this.convertValues(source["Parent"], Course);
	        this.Children = this.convertValues(source["Children"], Course);
	        this.User = this.convertValues(source["User"], User);
	        this.Assignments = this.convertValues(source["Assignments"], Assignment);
	        this.Notes = this.convertValues(source["Notes"], Note);
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
	export class User {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    Username: string;
	    Email: string;
	    Password: string;
	    PasswordHash: string;
	    Avatar: string;
	    University: string;
	    Semester: string;
	    Year: string;
	    IsVerified: boolean;
	    Language: string;
	    CoursesCode: string[];
	    // Go type: time
	    LastSync?: any;
	    Courses: Course[];
	    Assignments: Assignment[];
	    Notes: Note[];
	    OwnerRequests: CourseLinkRequest[];
	    ReceiverRequests: CourseLinkRequest[];
	    Followers: User[];
	    Following: User[];
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.Username = source["Username"];
	        this.Email = source["Email"];
	        this.Password = source["Password"];
	        this.PasswordHash = source["PasswordHash"];
	        this.Avatar = source["Avatar"];
	        this.University = source["University"];
	        this.Semester = source["Semester"];
	        this.Year = source["Year"];
	        this.IsVerified = source["IsVerified"];
	        this.Language = source["Language"];
	        this.CoursesCode = source["CoursesCode"];
	        this.LastSync = this.convertValues(source["LastSync"], null);
	        this.Courses = this.convertValues(source["Courses"], Course);
	        this.Assignments = this.convertValues(source["Assignments"], Assignment);
	        this.Notes = this.convertValues(source["Notes"], Note);
	        this.OwnerRequests = this.convertValues(source["OwnerRequests"], CourseLinkRequest);
	        this.ReceiverRequests = this.convertValues(source["ReceiverRequests"], CourseLinkRequest);
	        this.Followers = this.convertValues(source["Followers"], User);
	        this.Following = this.convertValues(source["Following"], User);
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
	export class Assignment {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    Title: string;
	    Type: string;
	    Status: string;
	    Todo: string;
	    // Go type: time
	    Deadline: any;
	    Link: string;
	    CourseID: number;
	    CourseCode: string;
	    Priority: string;
	    ParentID: number;
	    UserID: number;
	    User?: User;
	    Course?: Course;
	    Documents: Document[];
	    Parent?: Assignment;
	    Children: Assignment[];
	
	    static createFrom(source: any = {}) {
	        return new Assignment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.Title = source["Title"];
	        this.Type = source["Type"];
	        this.Status = source["Status"];
	        this.Todo = source["Todo"];
	        this.Deadline = this.convertValues(source["Deadline"], null);
	        this.Link = source["Link"];
	        this.CourseID = source["CourseID"];
	        this.CourseCode = source["CourseCode"];
	        this.Priority = source["Priority"];
	        this.ParentID = source["ParentID"];
	        this.UserID = source["UserID"];
	        this.User = this.convertValues(source["User"], User);
	        this.Course = this.convertValues(source["Course"], Course);
	        this.Documents = this.convertValues(source["Documents"], Document);
	        this.Parent = this.convertValues(source["Parent"], Assignment);
	        this.Children = this.convertValues(source["Children"], Assignment);
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
	
	
	
	export class DocumentStorage {
	    UserID: number;
	    TotalSize: number;
	    DocumentCount: number;
	    // Go type: time
	    LastCalculatedAt: any;
	    User: User;
	
	    static createFrom(source: any = {}) {
	        return new DocumentStorage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.UserID = source["UserID"];
	        this.TotalSize = source["TotalSize"];
	        this.DocumentCount = source["DocumentCount"];
	        this.LastCalculatedAt = this.convertValues(source["LastCalculatedAt"], null);
	        this.User = this.convertValues(source["User"], User);
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
	export class LocalAiMessage {
	    ID: string;
	    AssignmentID: number;
	    Role: string;
	    Parts: number[];
	    Metadata: number[];
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	
	    static createFrom(source: any = {}) {
	        return new LocalAiMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.AssignmentID = source["AssignmentID"];
	        this.Role = source["Role"];
	        this.Parts = source["Parts"];
	        this.Metadata = source["Metadata"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
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
	    DeletedAt: gorm.DeletedAt;
	    Type: string;
	    FileName: string;
	    FilePath: string;
	    FileSize: number;
	    StorageKey: string;
	    Version: number;
	    ParentID: number;
	    ParentDocID?: number;
	    IsOriginal: boolean;
	    HasLocalFile: boolean;
	    AssignmentID: number;
	    UploadID: string;
	    RemoteID: number;
	    RemoteAssignmentID: number;
	    Assignment: LocalAssignment;
	    Parent?: LocalDocument;
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
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.Type = source["Type"];
	        this.FileName = source["FileName"];
	        this.FilePath = source["FilePath"];
	        this.FileSize = source["FileSize"];
	        this.StorageKey = source["StorageKey"];
	        this.Version = source["Version"];
	        this.ParentID = source["ParentID"];
	        this.ParentDocID = source["ParentDocID"];
	        this.IsOriginal = source["IsOriginal"];
	        this.HasLocalFile = source["HasLocalFile"];
	        this.AssignmentID = source["AssignmentID"];
	        this.UploadID = source["UploadID"];
	        this.RemoteID = source["RemoteID"];
	        this.RemoteAssignmentID = source["RemoteAssignmentID"];
	        this.Assignment = this.convertValues(source["Assignment"], LocalAssignment);
	        this.Parent = this.convertValues(source["Parent"], LocalDocument);
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
	export class LocalNote {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    Title: string;
	    Subject: string;
	    Content: string;
	    Videos: string;
	    ParentID: number;
	    CourseID: number;
	    CourseCode: string;
	    RemoteID: number;
	    RemoteCourseID: number;
	    Course: LocalCourse;
	
	    static createFrom(source: any = {}) {
	        return new LocalNote(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.Title = source["Title"];
	        this.Subject = source["Subject"];
	        this.Content = source["Content"];
	        this.Videos = source["Videos"];
	        this.ParentID = source["ParentID"];
	        this.CourseID = source["CourseID"];
	        this.CourseCode = source["CourseCode"];
	        this.RemoteID = source["RemoteID"];
	        this.RemoteCourseID = source["RemoteCourseID"];
	        this.Course = this.convertValues(source["Course"], LocalCourse);
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
	    DeletedAt: gorm.DeletedAt;
	    Code: string;
	    Name: string;
	    Color: string;
	    Location: string;
	    // Go type: time
	    StartDate: any;
	    // Go type: time
	    EndDate: any;
	    Schedule: string;
	    Credits: number;
	    Semester: string;
	    Instructor: string;
	    InstructorEmail: string;
	    ParentID: number;
	    RemoteID: number;
	    Assignments: LocalAssignment[];
	    Notes: LocalNote[];
	
	    static createFrom(source: any = {}) {
	        return new LocalCourse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.Code = source["Code"];
	        this.Name = source["Name"];
	        this.Color = source["Color"];
	        this.Location = source["Location"];
	        this.StartDate = this.convertValues(source["StartDate"], null);
	        this.EndDate = this.convertValues(source["EndDate"], null);
	        this.Schedule = source["Schedule"];
	        this.Credits = source["Credits"];
	        this.Semester = source["Semester"];
	        this.Instructor = source["Instructor"];
	        this.InstructorEmail = source["InstructorEmail"];
	        this.ParentID = source["ParentID"];
	        this.RemoteID = source["RemoteID"];
	        this.Assignments = this.convertValues(source["Assignments"], LocalAssignment);
	        this.Notes = this.convertValues(source["Notes"], LocalNote);
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
	    DeletedAt: gorm.DeletedAt;
	    Title: string;
	    Type: string;
	    Status: string;
	    Todo: string;
	    // Go type: time
	    Deadline: any;
	    Link: string;
	    CourseID: number;
	    CourseCode: string;
	    Priority: string;
	    ParentID: number;
	    RemoteID: number;
	    RemoteCourseID: number;
	    Course?: LocalCourse;
	    Documents: LocalDocument[];
	
	    static createFrom(source: any = {}) {
	        return new LocalAssignment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.Title = source["Title"];
	        this.Type = source["Type"];
	        this.Status = source["Status"];
	        this.Todo = source["Todo"];
	        this.Deadline = this.convertValues(source["Deadline"], null);
	        this.Link = source["Link"];
	        this.CourseID = source["CourseID"];
	        this.CourseCode = source["CourseCode"];
	        this.Priority = source["Priority"];
	        this.ParentID = source["ParentID"];
	        this.RemoteID = source["RemoteID"];
	        this.RemoteCourseID = source["RemoteCourseID"];
	        this.Course = this.convertValues(source["Course"], LocalCourse);
	        this.Documents = this.convertValues(source["Documents"], LocalDocument);
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
	
	
	
	export class LocalNotification {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    sender_id: number;
	    entity: string;
	    entity_id: number;
	    type: string;
	    title: string;
	    action: string;
	    message: string;
	    read: boolean;
	    data: string;
	    sender: User;
	    // Go type: time
	    ExpiresAt?: any;
	
	    static createFrom(source: any = {}) {
	        return new LocalNotification(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.sender_id = source["sender_id"];
	        this.entity = source["entity"];
	        this.entity_id = source["entity_id"];
	        this.type = source["type"];
	        this.title = source["title"];
	        this.action = source["action"];
	        this.message = source["message"];
	        this.read = source["read"];
	        this.data = source["data"];
	        this.sender = this.convertValues(source["sender"], User);
	        this.ExpiresAt = this.convertValues(source["ExpiresAt"], null);
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

export namespace progress {
	
	export class TrackerSnapshot {
	    upload_id: string;
	    total: number;
	    current: number;
	    status: string;
	    // Go type: time
	    start_time: any;
	    error: any;
	    percentage: number;
	
	    static createFrom(source: any = {}) {
	        return new TrackerSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.upload_id = source["upload_id"];
	        this.total = source["total"];
	        this.current = source["current"];
	        this.status = source["status"];
	        this.start_time = this.convertValues(source["start_time"], null);
	        this.error = source["error"];
	        this.percentage = source["percentage"];
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

