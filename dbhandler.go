package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"strconv"
	"strings"
	"unicode"
)

func GetTables(esquema string) []Tabela {
	tabelas := make([]Tabela, 0)
	db, err := sql.Open("postgres", getStringConexao())
	if err != nil {
		log.Println(err)
		return tabelas
	}
	defer db.Close()
	tablesQuery, err := db.Query("select table_name,table_schema from information_schema.tables where table_schema= $1", esquema)
	if err != nil {
		log.Println(err)
		return tabelas
	}
	defer tablesQuery.Close()

	var tabela Tabela
	for tablesQuery.Next() {
		err = tablesQuery.Scan(&tabela.Nome, &tabela.Schema)
		if err != nil {
			log.Println(err)
			return tabelas
		}
		tabela.Colunas = GetColunas(esquema, tabela.Nome)
		tabelas = append(tabelas, tabela)
	}
	return tabelas
}

func GetColunas(esquema, tabela string) []Coluna {
	Colunas := make([]Coluna, 0)
	db, err := sql.Open("postgres", getStringConexao())
	if err != nil {
		log.Println(err)
		return Colunas
	}
	defer db.Close()
	ColunasQuery, err := db.Query(`select c1.column_name,c1.is_nullable,c1.data_type,
EXISTS((SELECT 
c.column_name
FROM
information_schema.table_constraints tc 
JOIN information_schema.constraint_column_usage AS ccu USING (constraint_schema, constraint_name) 
JOIN information_schema.columns AS c ON c.table_schema = tc.constraint_schema AND tc.table_name = c.table_name AND ccu.column_name = c.column_name
where constraint_type = 'PRIMARY KEY' and tc.table_name = $1 and tc.table_schema = $2 and c.column_name = c1.column_name))
 from information_schema.columns c1 where table_name = $1 and table_schema = $2 `, tabela, esquema)
	if err != nil {
		log.Println(err)
		return Colunas
	}
	defer ColunasQuery.Close()

	var coluna Coluna
	for ColunasQuery.Next() {
		err = ColunasQuery.Scan(&coluna.Nome, &coluna.Nullable, &coluna.Tipo, &coluna.IsPK)
		if err != nil {
			log.Println(err)
			return Colunas
		}
		Colunas = append(Colunas, coluna)
	}
	return Colunas
}

type Tabela struct {
	Nome    string
	Schema  string
	Colunas []Coluna
}

func (t Tabela) HasTime() bool {
	for _, coluna := range t.Colunas {
		if coluna.GetGolangType() == "time.Time" || coluna.GetGolangType() == "*time.Time" {
			return true
		}
	}
	return false
}

func (t Tabela) GetNome() string {
	return t.Nome
}

func (t Tabela) GetSqlString() string {
	var sqlString string
	colunas := make([]string, 0)
	for _, coluna := range t.Colunas {
		colunas = append(colunas, coluna.Nome)
	}
	sqlString = strings.Join(colunas, ",\n")
	return sqlString
}

func (t Tabela) GetScanString() string {
	var scanString string
	colunas := make([]string, 0)
	for _, coluna := range t.Colunas {
		colunas = append(colunas, "&"+t.Nome+"."+GetNomeObjeto(coluna))
	}
	scanString = strings.Join(colunas, ",\n")
	return scanString
}

func (t Tabela) GetUpdateCreateString() string {
	var scanString string
	colunas := make([]string, 0)
	for _, coluna := range t.Colunas {
		colunas = append(colunas, t.Nome+"."+GetNomeObjeto(coluna))
	}
	scanString = strings.Join(colunas, ",\n")
	return scanString
}

func (t Tabela) GetPKParamsString() string {
	var pkParamString string
	colunas := make([]string, 0)
	for _, coluna := range t.Colunas {
		if coluna.IsPK {
			colunas = append(colunas, coluna.Nome+" "+coluna.GetGolangType())
		}
	}
	pkParamString = strings.Join(colunas, ",")
	return pkParamString
}

func (t Tabela) GetPKNomes() string {
	var pkParamString string
	colunas := make([]string, 0)
	for _, coluna := range t.Colunas {
		if coluna.IsPK {
			colunas = append(colunas, coluna.Nome)
		}
	}
	pkParamString = strings.Join(colunas, ",")
	return pkParamString
}

func (t Tabela) GetWhereClause() string {
	var pkParamString string
	colunas := make([]string, 0)
	for i, coluna := range t.Colunas {
		if coluna.IsPK {
			colunas = append(colunas, coluna.Nome+" = $"+strconv.Itoa(i+1))
		}
	}
	pkParamString = strings.Join(colunas, " AND ")
	return pkParamString
}

func (t Tabela) GetNumParamsUpdateCreate() string {
	var pkParamString string
	colunas := make([]string, 0)
	for i, _ := range t.Colunas {
		colunas = append(colunas, "$"+strconv.Itoa(i+1))
	}
	pkParamString = strings.Join(colunas, ", ")
	return pkParamString
}

func (t Tabela) GetStrPKRotas() string {
	var pkParamString string
	colunas := make([]string, 0)
	for _, coluna := range t.Colunas {
		if coluna.IsPK {
			colunas = append(colunas, ":"+coluna.Nome)
		}
	}
	pkParamString = strings.Join(colunas, "/")
	return pkParamString
}

func (t Tabela) GetStrParamsRota() string {
	var pkParamString string
	colunas := make([]string, 0)
	for _, coluna := range t.Colunas {
		if coluna.IsPK {
			if coluna.GetGolangType() == "int" || coluna.GetGolangType() == "float64" {
				colunas = append(colunas, coluna.Nome)
			} else if coluna.GetGolangType() == "time.Time" {
				colunas = append(colunas, coluna.Nome)
			} else {
				colunas = append(colunas, "c.Param(\""+coluna.Nome+"\")")
			}
		}
	}
	pkParamString = strings.Join(colunas, ", ")
	return pkParamString
}

func (t Tabela) GetNonStringParamsRota() string {
	var params string
	for _, coluna := range t.Colunas {
		if coluna.IsPK {
			if coluna.GetGolangType() == "int" {
				params += coluna.Nome + `, errCv := strconv.Atoi(c.Param("` + coluna.Nome + `"))
	if errCv != nil {
		log.Println(errCv)
		return c.HTML(http.StatusInternalServerError, errCv.Error())
	}
	`
			} else if coluna.GetGolangType() == "time.Time" {
				params += coluna.Nome + `, errCv := time.Parse(\"2006-01-02T15:04:05.000Z\", c.Param("` + coluna.Nome + `"))
	if errCv != nil {
		log.Println(errCv)
		return c.HTML(http.StatusInternalServerError, errCv.Error())
	}
	`
			} else if coluna.GetGolangType() == "float64" {
				params += coluna.Nome + `, errCv := strconv.ParseFloat(c.Param("` + coluna.Nome + `"), 64)
	if errCv != nil {
		log.Println(errCv)
		return c.HTML(http.StatusInternalServerError, errCv.Error())
	}
	`
			}
		}
	}
	return params
}

func (t Tabela) GetImportsResources() string {
	var imports string
	impInt, impTime := false, false
	for _, coluna := range t.Colunas {
		if coluna.IsPK {
			if (coluna.GetGolangType() == "int" || coluna.GetGolangType() == "float64") && impInt == false {
				imports += "\n\t\"strconv\""
				impInt = true
			}
			if coluna.GetGolangType() == "time.Time" && impTime == false {
				imports += "\n\t\"time\""
				impTime = true
			}
		}
	}
	return imports
}

func (t Tabela) IsView() bool {
	for _, coluna := range t.Colunas {
		if coluna.IsPK {
			return false
		}
	}
	return true
}

type Coluna struct {
	Nome     string
	Nullable string
	Tipo     string
	IsPK     bool
}

func (c Coluna) GetNome() string {
	return c.Nome
}

func (c Coluna) GetGolangType() string {
	var ponteiro string
	if c.Nullable == "YES" {
		ponteiro = "*"
	}
	if c.Tipo == "integer" || c.Tipo == "smallint" {
		return ponteiro + "int"
	} else if c.Tipo == "float" || c.Tipo == "numeric" {
		return ponteiro + "float64"
	} else if c.Tipo == "timestamp without time zone" || c.Tipo == "timestamp with time zone" {
		return ponteiro + "time.Time"
	}
	return ponteiro + "string"
}

func (c Coluna) GetJsonType() string {
	if c.Nullable == "YES" {
		return c.Nome + ",omitempty"
	}
	return c.Nome
}

type DbObj interface {
	GetNome() string
}

func GetNomeObjeto(obj DbObj) string {
	partes := strings.Split(obj.GetNome(), "_")
	var nomeObj string
	for _, parte := range partes {
		nomeObj += UpcaseInitial(parte)
	}
	return nomeObj
}

func UpcaseInitial(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]

	}
	return ""
}
