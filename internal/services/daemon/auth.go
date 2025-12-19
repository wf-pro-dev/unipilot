package daemon

import (
	"context"
	"fmt"
	"os/exec"

	"unipilot/internal/errors"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// AuthorizationManager handles macOS Authorization Services using osascript
type AuthorizationManager struct {
	ctx        context.Context
	authorized bool
}

// NewAuthorizationManager creates a new authorization manager
func NewAuthorizationManager(ctx context.Context) *AuthorizationManager {
	return &AuthorizationManager{ctx: ctx, authorized: false}
}

// RequestPrivilegesAndExecute requests privileges and executes commands in a single session
func (am *AuthorizationManager) RequestPrivilegesAndExecute(commands []string) error {
	if am.authorized {
		return errors.NewAppError(errors.ValidationInvalid, "Already authorized, use ExecuteWithPrivileges for additional commands", nil)
	}

	// Show a dialog explaining what we're doing
	runtime.MessageDialog(am.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   "System Integration Required",
		Message: "UniPilot needs to install a background service for notifications. This will show a system dialog asking for your password.",
	})

	// Build a single script with all commands
	script := `do shell script "`
	for i, cmd := range commands {
		if i > 0 {
			script += " && "
		}
		script += cmd
	}
	script += `" with administrator privileges`

	cmd := exec.Command("osascript", "-e", script)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errors.SysExecFailed, "Privileged execution failed")
	}

	am.authorized = true
	return nil
}

// ExecuteWithPrivileges executes a command with elevated privileges (for additional commands after authorization)
func (am *AuthorizationManager) ExecuteWithPrivileges(toolPath, arg1, arg2 string) error {
	// Build the shell command
	shellCmd := fmt.Sprintf("%s %s %s", toolPath, arg1, arg2)

	// Use osascript to execute with privileges
	script := fmt.Sprintf(`do shell script "%s" with administrator privileges`, shellCmd)

	cmd := exec.Command("osascript", "-e", script)
	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Run()
}
