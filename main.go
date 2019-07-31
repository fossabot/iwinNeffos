package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
	"sync"
	"encoding/json"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/kardianos/minwinsvc"
	"github.com/kardianos/osext"
	"github.com/majidbigdeli/neffosAmi/domin/config"
	"github.com/majidbigdeli/neffosAmi/domin/data"
	"github.com/majidbigdeli/neffosAmi/domin/dto"
	"github.com/majidbigdeli/neffosAmi/domin/model"
	"github.com/spf13/viper"
)

var (
	err      error
	exePath  string
	formList = make(map[int][]byte)
	mutex    = &sync.Mutex{}
)

func init() {
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

	data.GetDB()
	data.GetDBData()
	data.GetDBCore()
	config.GetConfig()
}

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	serveMux := http.NewServeMux()
	serveMux.Handle("/getBroadCast", http.HandlerFunc(getBroadcastHandler))
	serveMux.Handle("/broadcast", http.HandlerFunc(broadcastHandler))
	serveMux.Handle("/GetNotif", http.HandlerFunc(getNotifHandler))

	go func() {
		log.Printf("Listening on: %s\nPress CTRL/CMD+C to interrupt.", config.HTTPAddr)
		log.Fatal(http.ListenAndServe(config.HTTPAddr, serveMux))
	}()
	//run server in https
	log.Printf("Listening on: %s\nPress CTRL/CMD+C to interrupt.", config.Addr)
	log.Fatal(http.ListenAndServeTLS(config.Addr, path.Join(exePath, config.CertFile), path.Join(exePath, config.KeyFile), serveMux))

}

func broadcastHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  err.Error(),
			Body:     "",
		})
		http.Error(w, err.Error(), 400)
		return
	}
	var userMsg dto.FormDto
	err = userMsg.UnmarshalBinary(body)
	if err != nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  err.Error(),
			Body:     string(body),
		})
		http.Error(w, err.Error(), 400)
		return
	}
	mutex.Lock()
	formList[userMsg.Extension] = body
	w.WriteHeader(200)
	mutex.Unlock()

}

func getBroadcastHandler(w http.ResponseWriter, r *http.Request) {

	numberStr := r.URL.Query().Get("Number")

	number, err := strconv.Atoi(numberStr)

	if err != nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  err.Error(),
			Body:     numberStr,
		})
		http.Error(w, err.Error(), 400)
		return
	}
	mutex.Lock()
	val, ok := formList[number]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if ok {
		delete(formList, number)
		w.WriteHeader(200)
		w.Write(val)
	} else {
		w.WriteHeader(200)
	}
	mutex.Unlock()
}

func getNotifHandler(w http.ResponseWriter, r *http.Request) {

	userIDStr := r.URL.Query().Get("UserId")

	userID, err := strconv.Atoi(userIDStr)

	if err != nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  err.Error(),
			Body:     userIDStr,
		})
		http.Error(w, err.Error(), 400)
		return
	}

	notification, err := data.GetNotifByUserID(userID)
	
	if err != nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  err.Error(),
			Body:     "",
		})
		http.Error(w, err.Error(), 400)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	if len(*notification) == 0 {
		w.WriteHeader(200)
		return
	}

	val, err := json.Marshal(notification)

	if err != nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  err.Error(),
			Body:     "",
		})		
		http.Error(w, err.Error(), 400)
		return
	}

	w.WriteHeader(200)
	w.Write(val)

	

}
