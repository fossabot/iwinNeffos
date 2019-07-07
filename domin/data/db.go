package data

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

type dbConfig struct {
	Debug    bool   `mapstructure:"debug"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
	Server   string `mapstructure:"server"`
	User     string `mapstructure:"user"`
	Database string `mapstructure:"database"`
}

var db *sqlx.DB
var dbData *sqlx.DB
var dbCore *sqlx.DB

//GetDB ...
func GetDB() {

	var con dbConfig

	if err := viper.UnmarshalKey("db", &con); err != nil {
		panic(err)
	}

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;", con.Server, con.User, con.Password, con.Port, con.Database)

	if con.Debug {
		fmt.Printf(" connString:%s\n", connString)
	}
	conn, err := sqlx.Connect("sqlserver", connString)
	if err != nil {
		//log.Fatal("Open connection failed:", err.Error())
		panic(err)
	}

	conn.SetMaxOpenConns(7)
	conn.SetMaxIdleConns(7)

	fmt.Printf("Connect To Database SmartCCLog\n")

	//	fmt.Printf("Connected!\n")
	//	defer conn.Close()
	db = conn
	//return conn
}

//GetDBData ...
func GetDBData() {

	var con dbConfig

	if err := viper.UnmarshalKey("dbData", &con); err != nil {
		panic(err)
	}

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;", con.Server, con.User, con.Password, con.Port, con.Database)

	if con.Debug {
		fmt.Printf(" connString:%s\n", connString)
	}
	conn, err := sqlx.Connect("sqlserver", connString)
	if err != nil {
		//log.Fatal("Open connection failed:", err.Error())
		panic(err)
	}

	conn.SetMaxOpenConns(5)
	conn.SetMaxIdleConns(5)

	fmt.Printf("Connect To Database SmartCCData\n")

	//	fmt.Printf("Connected!\n")
	//	defer conn.Close()
	dbData = conn
	//return conn
}

//GetDBCore ...
func GetDBCore() {

	var con dbConfig

	if err := viper.UnmarshalKey("dbCore", &con); err != nil {
		panic(err)
	}

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;", con.Server, con.User, con.Password, con.Port, con.Database)

	if con.Debug {
		fmt.Printf(" connString:%s\n", connString)
	}
	conn, err := sqlx.Connect("sqlserver", connString)
	if err != nil {
		//log.Fatal("Open connection failed:", err.Error())
		panic(err)
	}

	conn.SetMaxOpenConns(5)
	conn.SetMaxIdleConns(5)

	fmt.Printf("Connect To Database SmartCCCore\n")

	//	fmt.Printf("Connected!\n")
	//	defer conn.Close()
	dbCore = conn
	//return conn
}
