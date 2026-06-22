package main

// "github.com/joho/godotenv"

type Config struct {
	JsonConfigsDir string
	CacheDir       string
}

var Configs Config

func LoadConfigs() {
	Configs = Config{
		JsonConfigsDir: "json_configs",
		CacheDir:       "cache",
	}

	// cfg := Config{
	// 	Token: os.Getenv("DISCORD_TOKEN"),
	// 	AppID: os.Getenv("DISCORD_APPLICATION_ID"),
	// }

	// if cfg.Token == "" {
	// 	log.Fatal("DISCORD_TOKEN required")
	// }

}
