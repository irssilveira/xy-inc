package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func getStringConexao() string {
	cfg, err := ReadConfig()
	if err != nil {
		log.Fatalln("Erro no arquivo de configuração")
	}
	stringConexao := fmt.Sprintf("user=%s dbname=%s host=%s password=%s sslmode=disable", cfg.User, cfg.DbName, cfg.Host, cfg.Password)
	return stringConexao
}

// Info from config file
type DbConfig struct {
	Host     string
	User     string
	Password string
	DbName   string
}

// Reads info from config file
func ReadConfig() (DbConfig, error) {
	config := DbConfig{}
	file, err := os.Open("config.json")
	if err != nil {
		fmt.Println("error:", err)
		return config, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
		return config, err
	}
	return config, err
}
