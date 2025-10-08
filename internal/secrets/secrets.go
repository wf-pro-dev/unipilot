package secrets

import (
	"os"
	"log"
	"fmt"	
	"github.com/spf13/viper"
)

func GetEnvVar(envName string) (string, error) {
	
	var envVar string
	
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println("Note: .env file not found, using environment variables")
		envVar = os.Getenv(envName)
		 
	} else {
		
		envVar = viper.GetString(envName)
	}
	if envVar == "" {
		return "", fmt.Errorf("error getting %s: %w", envName ,err)
	}

	return envVar, nil


}
