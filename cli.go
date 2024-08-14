package main

import (
	"bufio"
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	LocalHost  string
	LocalPort  uint16
	Server     string
	ServerPort uint16
	ClientID   string
	SecretKey  string
}

func main() {
	displayWelcomeMessage()

	var config Config
	var configFile string

	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	runApp(&config, configFile)
}

func displayWelcomeMessage() {
	art := figure.NewColorFigure("Jerusalem", "slant", "green", true)
	art.Print()
	fmt.Println("\n\n👋 Welcome to the Jerusalem Client Application!")
}

func runApp(config *Config, configFile string) {
	if configFile != "" {
		viper.SetConfigType("yaml")
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("❌ Failed to read config file: %v", err)
		}
	}
	readConfigFromViper(config)

	if configFile == "" || config.Server == "" || config.ClientID == "" || config.SecretKey == "" {
		promptForMissingConfig(config)
	}

	client, err := NewClient(config.ServerPort, config.LocalHost, config.LocalPort, config.Server, config.ClientID, config.SecretKey)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	if err := client.Listen(); err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}
}

func readConfigFromViper(config *Config) {
	config.LocalHost = viper.GetString("local-host")
	config.Server = viper.GetString("server")
	config.ClientID = viper.GetString("client-id")
	config.SecretKey = viper.GetString("secret-key")
	config.LocalPort = uint16(viper.GetInt("local-port"))
	config.ServerPort = uint16(viper.GetInt("server-port"))
}

func promptForMissingConfig(config *Config) {
	if config.Server == "" {
		config.Server = getEnvOrPrompt("SERVER", "Server address 🛠️")
	}
	if config.ClientID == "" {
		config.ClientID = getEnvOrPrompt("CLIENT_ID", "Client ID 🆔")
	}
	if config.SecretKey == "" {
		config.SecretKey = getEnvOrPrompt("SECRET_KEY", "Secret key 🔑 (64 chars)")
	}
	if config.LocalHost == "" {
		config.LocalHost = getEnvOrPrompt("LOCAL_HOST", "Local host 💻 (default is 127.0.0.1)", config.LocalHost)
	}
	if config.LocalPort == 0 {
		config.LocalPort = getEnvOrPromptUint16("LOCAL_PORT", "Local port 🔌")
	}
	if config.ServerPort == 0 {
		config.ServerPort = getEnvOrPromptUint16("SERVER_PORT", "Server port 🌐")
	}
}

func getEnvOrPrompt(envVar, prompt string, def ...string) string {
	value := viper.GetString(envVar)
	if value == "" {
		return promptUserInput(prompt, def...)
	}
	return value
}

func getEnvOrPromptUint16(envVar, prompt string) uint16 {
	value := viper.GetString(envVar)
	if value == "" {
		return parseUint16(promptUserInput(prompt))
	}
	return parseUint16(value)
}

func parseUint16(s string) uint16 {
	val, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		log.Fatalf("❌ Invalid port: %v", err)
	}
	return uint16(val)
}

func promptUserInput(fieldName string, def ...string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("➡️ Enter %s: ", fieldName)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("❌ Failed to read %s: %v", fieldName, err)
	}
	v := strings.TrimSpace(input)
	if v == "" && len(def) > 0 {
		return def[0]
	}
	return v
}
