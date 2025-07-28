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

