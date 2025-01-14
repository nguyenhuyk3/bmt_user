package initializations

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"user_service/global"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func loadConfigs() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	viper.SetConfigType("env")
	viper.AutomaticEnv()

	fillConfigFromENV()

	err = viper.Unmarshal(&global.Config.Server)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	fetchConfigs()
}

func fillConfigFromENV() {
	global.Config.Server.ServiceName = viper.GetString("SERVICE_NAME")
	global.Config.Server.ConfigServiceUrl = viper.GetString("CONFIG_SERVICE_URL")
	global.Config.Server.APIKey = viper.GetString("API_KEY")
}

func fetchConfigs() {
	url := fmt.Sprintf("%s?service_name=%s", global.Config.Server.ConfigServiceUrl, global.Config.Server.ServiceName)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("error making GET request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("error unmarshaling response: %v", err)
	}
	if data, ok := result["data"].(map[string]interface{}); ok {
		dataBytes, _ := json.Marshal(data)
		if err := json.Unmarshal(dataBytes, &global.Config.PostgreSql); err != nil {
			log.Fatalf("error unmarshaling data to postgreSetting: %v", err)
		}
		fmt.Println(global.Config.PostgreSql)
	} else {
		log.Fatalf("no data found in response")
	}
}
