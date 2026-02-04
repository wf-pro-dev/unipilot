export namespace client {
	
	export class RemoteUser {
	    ID: string;
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
	    OwnerRequests: models.CourseInvitation[];
	    ReceiverRequests: models.CourseInvitation[];
	    SentFriendRequests: models.Friendship[];
	    ReceivedFriendRequests: models.Friendship[];
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
	        this.OwnerRequests = this.convertValues(source["OwnerRequests"], models.CourseInvitation);
	        this.ReceiverRequests = this.convertValues(source["ReceiverRequests"], models.CourseInvitation);
	        this.SentFriendRequests = this.convertValues(source["SentFriendRequests"], models.Friendship);
	        this.ReceivedFriendRequests = this.convertValues(source["ReceivedFriendRequests"], models.Friendship);
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
	    DocumentID: string;
	    AssignmentID: string;
	    UserID: string;
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
	        this.DocumentID = source["DocumentID"];
	        this.AssignmentID = source["AssignmentID"];
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

}

export namespace models {
	
	export class Document {
	    ID: string;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    Type: string;
	    FileName: string;
	    FilePath: string;
	    FileSize: number;
	    StorageKey?: string;
	    Version: number;
	    ParentDocID?: string;
	    IsOriginal: boolean;
	    HasLocalFile: boolean;
	    AssignmentID: string;
	    UserID: string;
	    User?: User;
	    Assignment?: Assignment;
	    Parent?: Document;
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
	        this.ParentDocID = source["ParentDocID"];
	        this.IsOriginal = source["IsOriginal"];
	        this.HasLocalFile = source["HasLocalFile"];
	        this.AssignmentID = source["AssignmentID"];
	        this.UserID = source["UserID"];
	        this.User = this.convertValues(source["User"], User);
	        this.Assignment = this.convertValues(source["Assignment"], Assignment);
	        this.Parent = this.convertValues(source["Parent"], Document);
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
	export class Friendship {
	    ID: string;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    RequesterID: string;
	    AddresseeID: string;
	    Status: string;
	    Requester: User;
	    Addressee: User;
	
	    static createFrom(source: any = {}) {
	        return new Friendship(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.RequesterID = source["RequesterID"];
	        this.AddresseeID = source["AddresseeID"];
	        this.Status = source["Status"];
	        this.Requester = this.convertValues(source["Requester"], User);
	        this.Addressee = this.convertValues(source["Addressee"], User);
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
	export class CourseInvitation {
	    ID: string;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    OwnerID: string;
	    ReceiverID: string;
	    SenderID: string;
	    CourseID: string;
	    CourseCode: string;
	    Status: string;
	    Owner?: User;
	    Receiver?: User;
	    Sender?: User;
	    Course?: Course;
	
	    static createFrom(source: any = {}) {
	        return new CourseInvitation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], gorm.DeletedAt);
	        this.OwnerID = source["OwnerID"];
	        this.ReceiverID = source["ReceiverID"];
	        this.SenderID = source["SenderID"];
	        this.CourseID = source["CourseID"];
	        this.CourseCode = source["CourseCode"];
	        this.Status = source["Status"];
	        this.Owner = this.convertValues(source["Owner"], User);
	        this.Receiver = this.convertValues(source["Receiver"], User);
	        this.Sender = this.convertValues(source["Sender"], User);
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
	    ID: string;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    Title: string;
	    Subject: string;
	    Content: string;
	    Videos: string;
	    ParentID?: string;
	    CourseID: string;
	    UserID: string;
	    User?: User;
	    Course?: Course;
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
	    ID: string;
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
	    ClusterID?: string;
	    UserID: string;
	    Cluster?: Course;
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
	        this.ClusterID = source["ClusterID"];
	        this.UserID = source["UserID"];
	        this.Cluster = this.convertValues(source["Cluster"], Course);
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
	    ID: string;
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
	    OwnerRequests: CourseInvitation[];
	    ReceiverRequests: CourseInvitation[];
	    SentFriendRequests: Friendship[];
	    ReceivedFriendRequests: Friendship[];
	
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
	        this.OwnerRequests = this.convertValues(source["OwnerRequests"], CourseInvitation);
	        this.ReceiverRequests = this.convertValues(source["ReceiverRequests"], CourseInvitation);
	        this.SentFriendRequests = this.convertValues(source["SentFriendRequests"], Friendship);
	        this.ReceivedFriendRequests = this.convertValues(source["ReceivedFriendRequests"], Friendship);
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
	    ID: string;
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
	    CourseID: string;
	    Priority: string;
	    ParentID?: string;
	    UserID: string;
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
	    UserID: string;
	    TotalSize: number;
	    DocumentCount: number;
	    // Go type: time
	    LastCalculatedAt: any;
	    User?: User;
	
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
	    AssignmentID: string;
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
	    ID: string;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    Type: string;
	    FileName: string;
	    FilePath: string;
	    FileSize: number;
	    StorageKey?: string;
	    Version: number;
	    ParentDocID?: string;
	    IsOriginal: boolean;
	    HasLocalFile: boolean;
	    AssignmentID: string;
	    // Go type: time
	    SyncedAt?: any;
	    Assignment: LocalAssignment;
	    Parent?: LocalDocument;
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
	        this.ParentDocID = source["ParentDocID"];
	        this.IsOriginal = source["IsOriginal"];
	        this.HasLocalFile = source["HasLocalFile"];
	        this.AssignmentID = source["AssignmentID"];
	        this.SyncedAt = this.convertValues(source["SyncedAt"], null);
	        this.Assignment = this.convertValues(source["Assignment"], LocalAssignment);
	        this.Parent = this.convertValues(source["Parent"], LocalDocument);
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
	    ID: string;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    DeletedAt: gorm.DeletedAt;
	    Title: string;
	    Subject: string;
	    Content: string;
	    Videos: string;
	    ParentID?: string;
	    CourseID: string;
	    // Go type: time
	    SyncedAt?: any;
	    Course?: LocalCourse;
	
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
	        this.SyncedAt = this.convertValues(source["SyncedAt"], null);
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
	    ID: string;
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
	    ClusterID?: string;
	    // Go type: time
	    SyncedAt?: any;
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
	        this.ClusterID = source["ClusterID"];
	        this.SyncedAt = this.convertValues(source["SyncedAt"], null);
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
	    ID: string;
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
	    CourseID: string;
	    Priority: string;
	    ParentID?: string;
	    // Go type: time
	    SyncedAt?: any;
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
	        this.Priority = source["Priority"];
	        this.ParentID = source["ParentID"];
	        this.SyncedAt = this.convertValues(source["SyncedAt"], null);
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
	export class LocalAssignmentStorage {
	    AssignmentID: string;
	    TotalCount: number;
	    DocumentCount: number;
	    TotalSize: number;
	    Size: number;
	    // Go type: time
	    LastCalculatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new LocalAssignmentStorage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.AssignmentID = source["AssignmentID"];
	        this.TotalCount = source["TotalCount"];
	        this.DocumentCount = source["DocumentCount"];
	        this.TotalSize = source["TotalSize"];
	        this.Size = source["Size"];
	        this.LastCalculatedAt = this.convertValues(source["LastCalculatedAt"], null);
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

