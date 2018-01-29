package main

import (
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {

	if len(os.Args) != 2 {
		log.Println("schema não especificado, uso: " + os.Args[0] + " schema")
		os.Exit(1)
	}
	schema := os.Args[1]
	os.MkdirAll("autocreated/"+schema+"/models", os.ModePerm)
	os.MkdirAll("autocreated/"+schema+"/repository", os.ModePerm)
	os.MkdirAll("autocreated/"+schema+"/resources", os.ModePerm)
	log.Println("Criando modelos...")
	GeraModelo(schema)
	log.Println("Criando repositórios...")
	GeraRepositorio(schema)
	log.Println("Criando recursos...")
	GeraRecurso(schema)
	log.Println("Criando rotas...")
	GeraRotas(schema)
	GeraMain(schema)
	log.Println("Rest service criado.")
	log.Println("Criaçao finalizada.")
}
