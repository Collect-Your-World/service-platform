package config

type GoogleConfig struct {
	SpreadsheetID       string `mapstructure:"spreadsheet_id"`
	CredentialsFilePath string `mapstructure:"credentials_file_path"`
	CredentialsFileJSON string `mapstructure:"credentials_file_json"`
}
