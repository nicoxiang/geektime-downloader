package config

import "fmt"

// ValidateConfig validates the application configuration.
func ValidateConfig(cfg *AppConfig) error {
	if err := validateCookies(cfg); err != nil {
		return err
	}
	if err := validateComments(cfg); err != nil {
		return err
	}
	if err := validateQuality(cfg); err != nil {
		return err
	}
	if err := validateColumnOutputType(cfg); err != nil {
		return err
	}
	if err := validateLogLevel(cfg); err != nil {
		return err
	}
	return nil
}

func validateCookies(cfg *AppConfig) error {
	if cfg.Gcid == "" || cfg.Gcess == "" {
		return fmt.Errorf("arguments 'gcid' and 'gcess' are required and cannot be empty")
	}
	return nil
}

func validateComments(cfg *AppConfig) error {
	validComments := []int{0, 1, 2}

	isValidCommentFlag := false
	for _, v := range validComments {
		if cfg.DownloadComments == v {
			isValidCommentFlag = true
			break
		}
	}

	if !isValidCommentFlag {
		return fmt.Errorf("argument 'comments' is not valid, must be one of 0, 1, 2")
	}

	return nil
}

func validateQuality(cfg *AppConfig) error {
	validQuality := []string{"ld", "sd", "hd"}

	isValidQualityFlag := false
	for _, v := range validQuality {
		if cfg.Quality == v {
			isValidQualityFlag = true
			break
		}
	}

	if !isValidQualityFlag {
		return fmt.Errorf("argument 'quality' is not valid, must be one of ld, sd, hd")
	}

	return nil
}

func validateLogLevel(cfg *AppConfig) error {
	validLogLevels := []string{"debug", "info", "warn", "error", "none"}

	isValidLogLevel := false
	for _, v := range validLogLevels {
		if cfg.LogLevel == v {
			isValidLogLevel = true
			break
		}
	}

	if !isValidLogLevel {
		return fmt.Errorf("argument 'log-level' is not valid, must be one of debug, info, warn, error, none")
	}

	return nil
}

func validateColumnOutputType(cfg *AppConfig) error {
	if cfg.ColumnOutputType <= 0 || cfg.ColumnOutputType >= 8 {
		return fmt.Errorf("argument 'output' is not valid, must be between 1 and 7")
	}

	return nil
}
