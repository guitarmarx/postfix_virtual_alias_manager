package main

//go get -u github.com/go-sql-driver/mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// log
var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

var config = readConfig("conf.json")
var statusMessage = ""
var statusMessageAddress = &statusMessage
var connectionString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.DbUser, config.DbPassword, config.DbHost, config.DbPort, config.DbName)

func main() {

	Init(os.Stdout, os.Stdout, os.Stderr)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/create", createHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.ListenAndServe(":"+config.ServerPort, nil)
}

//########     struct     ############
type alias struct {
	Alias string `json:"alias"`
	Email string `json:"email"`
}

type response struct {
	AliasList     []alias
	StatusMessage string
}

type Config struct {
	ServerPort     string
	DbHost         string
	DbPort         string
	DbUser         string
	DbPassword     string
	DbName         string
	AliasTableName string
}

//############# server functions    ################
func readConfig(configFile string) Config {
	file, _ := os.Open(configFile)
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		Error.Println("error:", err)
	}
	fmt.Println("read parameter:", configuration)
	return configuration
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	Info.Println("####Aufruf:", r.Host, r.URL.Path)
	//get alias from url
	r.ParseForm()
	alias := alias{Alias: strings.Join(r.Form["alias"], ""), Email: strings.Join(r.Form["email"], "")}
	db, _ := sql.Open("mysql", connectionString)

	Info.Println("Try to delete alias:", alias)
	statusCode := deleteAlias(db, alias)
	Info.Println("ReturnCode:", statusCode)

	var message string
	if statusCode == 0 {
		message = "Alias deleted"
	} else {
		message = "Error occured"
	}
	*statusMessageAddress = message

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	Info.Println("####Aufruf:", r.Host, r.URL.Path)
	// get alias from url
	r.ParseForm()
	alias := alias{Alias: strings.Join(r.Form["alias"], ""), Email: strings.Join(r.Form["email"], "")}

	//open db
	db, _ := sql.Open("mysql", connectionString)

	//check if alias already exist
	checkStatusCode := checkAliasesAlreadyExist(db, alias)
	Info.Println("Check if alias already exist...")
	Info.Println("Returncode:", checkStatusCode)

	var message string
	if checkStatusCode == 0 {
		statusCode := insertAlias(db, alias)
		Info.Println("Insert alias in db...")
		Info.Println("Returncode:", statusCode)
		if statusCode == 0 {
			message = "Alias created"
		} else {
			message = "Error occured"
		}
	} else {
		message = "Alias already exist"
	}

	Info.Println(message)
	*statusMessageAddress = message

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	Info.Println("####Aufruf:", r.Host, r.URL.Path)
	// create database connection
	db, _ := sql.Open("mysql", connectionString)

	// get aliases
	aliasList := getAliases(db)

	// load template
	t, _ := template.ParseFiles("templates/index.html")
	var response = response{AliasList: aliasList, StatusMessage: statusMessage}
	t.Execute(w, response)

}

//#########  database  functions     #########

func getAliases(db *sql.DB) []alias {
	results, _ := db.Query("SELECT * FROM virtual_aliases")

	var aliasList []alias
	for results.Next() {

		var alias alias
		results.Scan(&alias.Alias, &alias.Email)
		aliasList = append(aliasList, alias)
	}
	results.Close()
	return aliasList
}

func checkAliasesAlreadyExist(db *sql.DB, alias alias) int {
	query := fmt.Sprintf("SELECT * FROM virtual_aliases where alias='%s' and email='%s'", alias.Alias, alias.Email)
	result, err := db.Query(query)
	if err != nil {
		Error.Println(err.Error())
		return 0
	}
	for result.Next() {
		result.Close()
		return 1
	}
	return 0
}

func insertAlias(db *sql.DB, alias alias) int {
	query := fmt.Sprintf("INSERT INTO "+config.AliasTableName+" (alias,email) VALUES ('%s','%s')", alias.Alias, alias.Email)
	insert, err := db.Query(query)

	if err != nil {
		Error.Println(err.Error())
		return 1
	}

	insert.Close()
	return 0
}

func deleteAlias(db *sql.DB, alias alias) int {
	query := fmt.Sprintf("DELETE  FROM virtual_aliases WHERE alias='%s' AND email='%s'", alias.Alias, alias.Email)
	delete, err := db.Query(query)

	if err != nil {
		Error.Println(err.Error())
		return 1
	}
	delete.Close()
	return 0

}

func Init(
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime)

	Warning = log.New(warningHandle,
		"WARN: ",
		log.Ldate|log.Ltime)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime)
}
