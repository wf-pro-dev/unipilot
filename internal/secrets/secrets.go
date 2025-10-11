package secrets

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func GetEnvVar(envName string) (string, error) {

	var envVar string

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		envVar = os.Getenv(envName)

	} else {
		envVar = viper.GetString(envName)
	}
	if envVar == "" {
		return "", fmt.Errorf("error getting %s: is empty", envName)
	}

	return envVar, nil

}
