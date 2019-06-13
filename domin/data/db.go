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

	conn.SetMaxOpenConns(5)
	conn.SetMaxIdleConns(5)

	//	fmt.Printf("Connected!\n")
	//	defer conn.Close()
	db = conn
	//return conn
}
