package config

type Config struct{
	DB DBConfig
	Version string
	Port string
}

type DBConfig struct{
	Username string
	Password string
	Host     string
	Port     string
	DBName   string
}

func GetConfig() Config {
	return Config{
		DB:DBConfig{
			Username: "",
			Password: "",
			Host:     "localhost",
			Port:     "27017",
			DBName:   "in-share",
		},
		Version: "1.0",
		Port: "4674",
	}
}
