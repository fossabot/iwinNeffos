package main

import (
	"fmt"
	"log"
	"net/http"
	"path"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/kardianos/minwinsvc"
	"github.com/kardianos/osext"
	"github.com/majidbigdeli/neffosAmi/controller"
	"github.com/majidbigdeli/neffosAmi/domin/config"
	"github.com/majidbigdeli/neffosAmi/domin/data"
	"github.com/majidbigdeli/neffosAmi/job"

	"github.com/rs/cors"
	"github.com/spf13/viper"
)

var (
	err     error
	exePath string
)

func init() {

	// برای ویندوز سرویس کردن اضافه کردم
	exePath, err = osext.ExecutableFolder()

	if err != nil {
		panic(fmt.Errorf("fatal error ExecutableFolder: %s", err.Error()))
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(exePath)
	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err.Error()))
	}

	// باز کردن کانکشن به دیتابیس
	data.GetDB()
	data.GetDBData()
	data.GetDBCore()
	config.GetConfig()
}

func main() {

	controller.NeffosServer()

	serveMux := http.NewServeMux()
	//websocket Handler ...
	// panel be sorate https api echo ra seda mizanad
	serveMux.Handle("/echo", controller.Server)
	//ای پی ای خانم وحید که در زمان پاسخ دادن تماس خانم وحید صدا می کند به صورت http
	serveMux.Handle("/broadcast", http.HandlerFunc(controller.BroadcastHandler))

	serveMux.Handle("/check", http.HandlerFunc(controller.Check))
	serveMux.Handle("/disconnect", http.HandlerFunc(controller.Disconnect))

	handler := cors.Default().Handler(serveMux)

	//run all jobs

	job.Jobs()

	controller.StartConnectionManager()

	go func() {
		log.Printf("Listening on: %s\nPress CTRL/CMD+C to interrupt.", config.HTTPAddr)
		log.Fatal(http.ListenAndServe(config.HTTPAddr, handler))
	}()

	//run server in https
	log.Printf("Listening on: %s\nPress CTRL/CMD+C to interrupt.", config.Addr)

	log.Fatal(http.ListenAndServeTLS(config.Addr, path.Join(exePath, config.CertFile), path.Join(exePath, config.KeyFile), handler))

}
