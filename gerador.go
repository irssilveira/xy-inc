package main

import (
	"io"
	"log"
	"os"
	"text/template"
)

func GeraModelo(schema string) {
	tabelas := GetTables(schema)
	funcMap := template.FuncMap{
		"NomeObj": GetNomeObjeto,
	}
	t := template.Must(template.New("model").Funcs(funcMap).Parse(model))

	for _, tabela := range tabelas {
		err := t.Execute(writerToFile("autocreated/"+schema+"/models/"+tabela.GetNome()+".go"), tabela)
		if err != nil {
			log.Println("executing template:", err)
		}
	}
}

const model = `package models
{{if .HasTime}}
import "time" 
{{end}}
type {{NomeObj .}} struct {
	{{range .Colunas}}{{NomeObj .}}				{{.GetGolangType}}			` + "`" + `json:"{{.GetJsonType}}"` + "`" + `
	{{end}}
}
`

func GeraRepositorio(schema string) {
	tabelas := GetTables(schema)
	funcMap := template.FuncMap{
		"NomeObj": GetNomeObjeto,
	}
	t := template.Must(template.New("repository").Funcs(funcMap).Parse(repository))

	for _, tabela := range tabelas {
		err := t.Execute(writerToFile("autocreated/"+schema+"/repository/"+tabela.GetNome()+".go"), tabela)
		if err != nil {
			log.Println("executing template:", err)
		}
	}
	GeraUtil(schema)
}

const repository = `package repository 
import (
	"database/sql"
	"../models"
	"log"
)

func Get{{NomeObj .}}s() ([]models.{{NomeObj .}}, error) {
	var {{.Nome}}s []models.{{NomeObj .}} = make([]models.{{NomeObj .}}, 0)
	db, err := sql.Open("postgres", getStringConexao())
	if err != nil {
		log.Println(err)
		return {{.Nome}}s, err
	}
	defer db.Close()
	rows{{NomeObj .}}s, err := db.Query(` + "`" + `SELECT 
{{.GetSqlString}}
FROM {{.Schema}}.{{.Nome}}` + "`" + `)
	if err != nil {
		log.Println(err)
		return {{.Nome}}s, err
	}
	defer rows{{NomeObj .}}s.Close()
	var {{.Nome}} models.{{NomeObj .}}
	for rows{{NomeObj .}}s.Next() {
		err = rows{{NomeObj .}}s.Scan({{.GetScanString}})
		if err != nil {
			log.Println(err)
			return {{.Nome}}s, err
		}
		{{.Nome}}s = append({{.Nome}}s, {{.Nome}})
	}
	return {{.Nome}}s, err
}


{{if not .IsView}}func Get{{NomeObj .}}ById({{.GetPKParamsString}}) (models.{{NomeObj .}}, error) {
	var {{.Nome}} models.{{NomeObj .}}
	db, err := sql.Open("postgres", getStringConexao())
	if err != nil {
		log.Println(err)
		return {{.Nome}}, err
	}
	defer db.Close()
	query{{NomeObj .}} := db.QueryRow(` + "`" + `SELECT 
{{.GetSqlString}}
FROM {{.Schema}}.{{.Nome}} 
WHERE {{.GetWhereClause}} ` + "`" + `, {{.GetPKNomes}})
	err = query{{NomeObj .}}.Scan({{.GetScanString}})
	if err != nil {
		log.Println(err)
		return {{.Nome}}, err
	}
	return {{.Nome}}, err
}


func Update{{NomeObj .}}({{.Nome}} models.{{NomeObj .}}) error {
	db, err := sql.Open("postgres", getStringConexao())
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	_, err = db.Exec(` + "`" + `UPDATE {{.Schema}}.{{.Nome}} SET 
	   ({{.GetSqlString}}) = ({{.GetNumParamsUpdateCreate}})
       WHERE {{.GetWhereClause}} ` + "`" + `,
		{{.GetUpdateCreateString}})
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}


func Create{{NomeObj .}}({{.Nome}} models.{{NomeObj .}}) error {
	db, err := sql.Open("postgres", getStringConexao())
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	_, err = db.Exec(` + "`" + `INSERT INTO {{.Schema}}.{{.Nome}} ({{.GetSqlString}}) VALUES({{.GetNumParamsUpdateCreate}}) ` + "`" + `,
		{{.GetUpdateCreateString}})
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}

func Delete{{NomeObj .}}({{.GetPKParamsString}}) error {
	db, err := sql.Open("postgres", getStringConexao())
	if err != nil {
		log.Println(err)
		return err
	}
	defer db.Close()
	_, err = db.Exec(` + "`" + `DELETE FROM {{.Schema}}.{{.Nome}} 
       WHERE {{.GetWhereClause}} ` + "`" + `,
		{{.GetPKNomes}} )
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}{{end}}
`

func GeraRecurso(schema string) {
	tabelas := GetTables(schema)
	funcMap := template.FuncMap{
		"NomeObj": GetNomeObjeto,
	}
	t := template.Must(template.New("resource").Funcs(funcMap).Parse(resource))

	for _, tabela := range tabelas {
		err := t.Execute(writerToFile("autocreated/"+schema+"/resources/"+tabela.GetNome()+".go"), tabela)
		if err != nil {
			log.Println("executing template:", err)
		}
	}
}

const resource = `package resources

import (
	"../repository"{{if not .IsView}}
	"../models"{{end}}
	"github.com/labstack/echo"
	"log"
	"net/http"{{.GetImportsResources}}
)


func Get{{NomeObj .}}s(c echo.Context) error {
	{{.Nome}}s, err := repository.Get{{NomeObj .}}s()
	if err != nil {
		log.Println(err)
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, {{.Nome}}s)
}
{{if not .IsView}}
func Get{{NomeObj .}}ById(c echo.Context) error {
	{{.GetNonStringParamsRota}}{{.Nome}}, err := repository.Get{{NomeObj .}}ById({{.GetStrParamsRota}})
	if err != nil {
		log.Println(err)
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, {{.Nome}})
}

func Update{{NomeObj .}}(c echo.Context) error {
	{{.Nome}} := new(models.{{NomeObj .}})
	if err := c.Bind({{.Nome}}); err != nil {
		log.Println(err)
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	err := repository.Update{{NomeObj .}}(*{{.Nome}})
	if err != nil {
		log.Println(err)
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusNoContent, "")
}

func Create{{NomeObj .}}(c echo.Context) error {
	{{.Nome}} := new(models.{{NomeObj .}})
	if err := c.Bind({{.Nome}}); err != nil {
		log.Println(err)
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	err := repository.Create{{NomeObj .}}(*{{.Nome}})
	if err != nil {
		log.Println(err)
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusNoContent, "")
}

func Delete{{NomeObj .}}(c echo.Context) error {
	{{.GetNonStringParamsRota}}err := repository.Delete{{NomeObj .}}({{.GetStrParamsRota}})
	if err != nil {
		log.Println(err)
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusNoContent, "")
}{{end}}

`

func GeraRotas(schema string) {
	tabelas := GetTables(schema)
	funcMap := template.FuncMap{
		"NomeObj": GetNomeObjeto,
	}
	t := template.Must(template.New("routes").Funcs(funcMap).Parse(routes))
	err := t.Execute(writerToFile("autocreated/"+schema+"/routes.go"), tabelas)
	if err != nil {
		log.Println("executing template:", err)

	}
}

const routes = `package main 

import (
	"./resources"
	"github.com/labstack/echo"
	)


func routes(e *echo.Group) {
	{{range .}}
	e.GET("{{.Nome}}", resources.Get{{NomeObj .}}s )
	{{if not .IsView}}e.GET("{{.Nome}}/{{.GetStrPKRotas}}",resources.Get{{NomeObj .}}ById)
	e.PUT("{{.Nome}}", resources.Update{{NomeObj .}})
	e.POST("{{.Nome}}", resources.Create{{NomeObj .}})
	e.DELETE("{{.Nome}}/{{.GetStrPKRotas}}",resources.Delete{{NomeObj .}}){{end}}
	{{end}}
}

`

func GeraUtil(schema string) {
	t := template.Must(template.New("util").Parse(util))
	err := t.Execute(writerToFile("autocreated/"+schema+"/repository/util.go"), nil)
	if err != nil {
		log.Println("executing template:", err)

	}
}

const util = `package repository

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
`

func GeraMain(schema string) {
	t := template.Must(template.New("principal").Parse(principal))
	err := t.Execute(writerToFile("autocreated/"+schema+"/main.go"), nil)
	if err != nil {
		log.Println("executing template:", err)

	}
}

const principal = `package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	log.SetFlags(log.Llongfile + log.Ltime + log.Ldate)
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	r := e.Group("/api/")
	routes(r)

	e.Logger.Fatal(e.Start(":8080"))
}

`

func writerToFile(path string) io.Writer {
	f, err := os.Create(path)
	if err != nil {
		log.Panicln(err)
	}
	return f
}
