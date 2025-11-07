package configs

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	SERVERPORT string `mapstructure:"SERVERPORT"`
}

func LoadConfig() *Config {
	var config Config
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigFile("../.env")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No .env file found or could not read it: %v", err)
	} else {
		log.Println("Loaded .env file successfully")
	}

	// Explicitly bind expected environment variables
	keys := []string{
		"SERVERPORT",
	}
	for _, key := range keys {
		_ = viper.BindEnv(key)
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into config struct: %v", err)
	}

	return &config
}
