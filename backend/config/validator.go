package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

func (v *Validator) Validate(config *Config) error {
	if err := v.validate.Struct(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	if err := v.validateCustomRules(config); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateCustomRules(config *Config) error {
	if config.Database.MaxIdleConns > config.Database.MaxOpenConns {
		return fmt.Errorf("database max_idle_conns cannot exceed max_open_conns")
	}

	if len(config.Security.EncryptionKey) != 32 {
		return fmt.Errorf("security.encryption_key must be exactly 32 characters")
	}

	return nil
}
