package env

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

const (
	varNameField = "var-name"
	varValField  = "var-val"
)

func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}

	return value
}

func GetEnvInt(envVar, fallback string) int {
	envVal := GetEnv(envVar, fallback)

	valInt, err := strconv.Atoi(envVal)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			varNameField: envVar,
			varValField:  envVal,
		}).Error("Could not parse int from env")
	}

	return valInt
}
