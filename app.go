package main

import (
	"context"
	"fmt"
	"strconv"
	"unipilot/internal/app"
	"unipilot/internal/auth"
	"unipilot/internal/client"
	"unipilot/internal/events"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/models/user"
)

// App struct
type App struct {
	ctx    context.Context
	Auth   *auth.Auth
	Events *events.Events
	DB     *app.DatabaseHelper
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		Auth:   auth.NewAuth(),
		Events: events.NewEvents(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize database helper
	dbHelper, err := app.NewDatabaseHelper()
	if err != nil {
		fmt.Printf("Warning: Could not initialize database helper: %v\n", err)
	} else {
		a.DB = dbHelper
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Login handles user authentication
func (a *App) Login(username, password string) error {
	if err := a.Auth.Login(username, password); err != nil {
		return err
	}

	go a.Auth.SSE.Connect(a.Auth.Client)

	a.Events.Start(a.Auth.SSE)

	// Reinitialize database helper after login
	dbHelper, err := app.NewDatabaseHelper()
	if err != nil {
		fmt.Printf("Warning: Could not initialize database helper after login: %v\n", err)
	} else {
		a.DB = dbHelper
	}

	return nil
}

// Logout handles user logout
func (a *App) Logout() error {
	if err := a.Auth.Logout(); err != nil {
		return err
	}

	a.Auth.SSE.StopListener()
	a.Events.Stop()
	a.DB = nil

	return nil
}

// IsAuthenticated checks if the user is currently authenticated
func (a *App) IsAuthenticated() bool {
	return a.Auth.IsAuthenticated()
}

// GetAssignment returns an assignment by ID
func (a *App) GetAssignment(id uint) (*assignment.LocalAssignment, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.DB.GetAssignment(id)
}

// GetCourse returns a course by ID
func (a *App) GetCourse(id uint) (*course.Course, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.DB.GetCourse(id)
}

// GetUser returns a user by ID
func (a *App) GetUser(id uint) (*user.User, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.DB.GetUser(id)
}

// GetAssignments returns all assignments for the current user
func (a *App) GetAssignments() ([]assignment.LocalAssignment, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.DB.GetAssignments()
}

// GetCourses returns all courses for the current user
func (a *App) GetCourses() ([]course.LocalCourse, error) {
	if a.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.DB.GetCourses()
}

// CreateAssignment creates a new assignment
func (a *App) CreateAssignment(assignment *assignment.Assignment) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return a.DB.CreateAssignment(assignment)
}

// UpdateAssignment updates an existing assignment
func (a *App) UpdateAssignment(LocalAssignment *assignment.LocalAssignment, column, value string) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := a.DB.UpdateAssignment(LocalAssignment, column, value); err != nil {
		return err
	}

	assignment_id_int := int(LocalAssignment.RemoteID)

	assignment_id := strconv.Itoa(assignment_id_int)

	if err := client.SendUpdate(assignment_id, column, value); err != nil {
		return err
	}

	return nil
}

// DeleteAssignment deletes an assignment
func (a *App) DeleteAssignment(assignment *assignment.Assignment) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return a.DB.DeleteAssignment(assignment)
}

// CreateCourse creates a new course
func (a *App) CreateCourse(course *course.Course) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return a.DB.CreateCourse(course)
}

// UpdateCourse updates an existing course
func (a *App) UpdateCourse(course *course.Course) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return a.DB.UpdateCourse(course)
}

// DeleteCourse deletes a course
func (a *App) DeleteCourse(course *course.Course) error {
	if a.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return a.DB.DeleteCourse(course)
}
