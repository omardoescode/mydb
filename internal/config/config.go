package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	DATA_DIR string `validate:"required"`
	WAL_DIR  string `validate:"required"`
}

var validate *validator.Validate
var CONFIG *Config = nil

func LoadConfig(filePath string) error {
	if CONFIG != nil {
		return fmt.Errorf("Config has already been initialized")
	}

	validate = validator.New(validator.WithRequiredStructEnabled())

	// 1. Reading the file
	configData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(configData, &CONFIG); err != nil {
		return err
	}

	validate := validator.New()

	// 2. Validate the file
	if err := validate.Struct(CONFIG); err != nil {
		return err
	}

	return nil
}
