package apiserver

type Config struct {
	BindAddr    string `toml:"bind_addr"`
	LogLevel    string `toml:"log_level"`
	DatabaseURL string `toml:"database_url`
	SessionKey  string `toml:session_url`
}

//New Config
func NewConfig() *Config {
	return &Config{
		BindAddr:    ":8080",
		LogLevel:    "debug",
		DatabaseURL: "root:password123@tcp(127.0.0.1:3306)/vkr",
	}
}
