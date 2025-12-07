package secrets

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var (
	CONSTANTS = map[string]string{
		"BASE_URL": "https://wwwill.dedyn.io",
		"API_URL":  "https://wwwill.dedyn.io/unipilot/api/v1",
	}
)

func GetEnvVar(envName string) (string, error) {

	var envVar = os.Getenv(envName)
	if envVar != "" {
		return envVar, nil
	}

	// Only try to read file if not found in environment
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		// If file is missing, that's fine, we just can't look there.
		// Only return error if it's a file permission/format error, NOT a NotExist error.
		// For simplicity, we can just log/ignore or assume the var is missing.
		// Better: Don't return error here. Just proceed to check viper.GetString
	} else {
		// Only try to get from viper if ReadInConfig succeeded
		envVar = viper.GetString(envName)
	}

	if envVar == "" {
		return "", fmt.Errorf("error getting %s: is empty", envName)
	}

	return envVar, nil

}
