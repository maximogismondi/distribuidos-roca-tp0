package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/common"
)

var log = logging.MustGetLogger("log")

// InitConfig Function that uses viper library to parse configuration parameters.
// Viper is configured to read variables from both environment variables and the
// config file ./config.yaml. Environment variables takes precedence over parameters
// defined in the configuration file. If some of the variables cannot be parsed,
// an error is returned
func InitConfig() (*viper.Viper, error) {
	v := viper.New()

	// Configure viper to read env variables with the CLI_ prefix
	v.AutomaticEnv()
	v.SetEnvPrefix("cli")
	// Use a replacer to replace env variables underscores with points. This let us
	// use nested configurations in the config file and at the same time define
	// env variables for the nested configurations
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Add env variables supported
	v.BindEnv("id")
	v.BindEnv("server", "address")
	v.BindEnv("loop", "period")
	v.BindEnv("loop", "amount")
	v.BindEnv("log", "level")

	// Add env variables for bet configuration
	v.BindEnv("NOMBRE")
	v.BindEnv("APELLIDO")
	v.BindEnv("DOCUMENTO")
	v.BindEnv("NACIMIENTO")
	v.BindEnv("NUMERO")

	// Try to read configuration from config file. If config file
	// does not exists then ReadInConfig will fail but configuration
	// can be loaded from the environment variables so we shouldn't
	// return an error in that case
	v.SetConfigFile("./config.yaml")
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Configuration could not be read from config file. Using env variables instead")
	}

	// Parse time.Time variables and return an error if those variables cannot be parsed

	if _, err := time.Parse("2006-01-02", v.GetString("NACIMIENTO")); err != nil {
		return nil, errors.Wrapf(err, "Could not parse NACIMIENTO env var as time.Time.")
	}

	return v, nil
}

// InitLogger Receives the log level to be set in go-logging as a string. This method
// parses the string and set the level to the logger. If the level string is not
// valid an error is returned
func InitLogger(logLevel string) error {
	baseBackend := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05} %{level:.5s}     %{message}`,
	)
	backendFormatter := logging.NewBackendFormatter(baseBackend, format)

	backendLeveled := logging.AddModuleLevel(backendFormatter)
	logLevelCode, err := logging.LogLevel(logLevel)
	if err != nil {
		return err
	}
	backendLeveled.SetLevel(logLevelCode, "")

	// Set the backends to be used.
	logging.SetBackend(backendLeveled)
	return nil
}

// PrintConfig Print all the configuration parameters of the program.
// For debugging purposes only
func PrintConfig(v *viper.Viper) {
	log.Infof("action: config | result: success | client_id: %s | server_address: %s | dni: %d | numero: %d",
		v.GetString("id"),
		v.GetString("server.address"),
		v.GetInt("DOCUMENTO"),
		v.GetInt("NUMERO"),
	)
}

func main() {
	v, err := InitConfig()
	if err != nil {
		log.Criticalf("%s", err)
	}

	if err := InitLogger(v.GetString("log.level")); err != nil {
		log.Criticalf("%s", err)
	}

	// Print program config with debugging purposes
	PrintConfig(v)

	birthDateStr := v.GetString("NACIMIENTO")
	birthDate, _ := time.Parse("2006-01-02", birthDateStr)

	clientConfig := common.ClientConfig{
		ServerAddress: v.GetString("server.address"),
		ID:            v.GetString("id"),
		Name:          v.GetString("NOMBRE"),
		Surname:       v.GetString("APELLIDO"),
		Document:      v.GetInt("DOCUMENTO"),
		Birthdate:     birthDate,
		Number:        v.GetInt("NUMERO"),
	}

	client := common.NewClient(clientConfig)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	// use an anonymous goroutine to listen for signals and stop the client loop
	go func() {
		signalRecieved := <-signalChannel

		log.Infof("action: signal | result: success | client_id: %s | signal: %v", clientConfig.ID, signalRecieved)
		client.StopClientLoop()
	}()

	client.StartClientLoop()
}
