package secrets

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"unipilot/internal/errors"
)

var (
	CONSTANTS = map[string]string{
		//"BASE_URL": "https://wwwill.dedyn.io",
		//"API_URL":  "https://wwwill.dedyn.io/unipilot/api/v1",
		"BASE_URL": "https://wwwill.xyz",
		"API_URL":  "https://wwwill.xyz/unipilot/api/v1",
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

	} else {
		// Only try to get from viper if ReadInConfig succeeded
		envVar = viper.GetString(envName)
	}

	if envVar == "" {
		return "", errors.NewAppError(errors.ConfigEnvVarNotFound, fmt.Sprintf("error getting %s: is empty", envName), nil)
	}

	return envVar, nil

}
