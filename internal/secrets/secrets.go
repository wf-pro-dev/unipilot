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

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		return "", fmt.Errorf("failed to read config: %w", err)
	}

	var envVar = os.Getenv(envName)
	if envVar == "" {
		envVar = viper.GetString(envName)
	}

	if envVar == "" {
		return "", fmt.Errorf("error getting %s: is empty", envName)
	}

	return envVar, nil

}
