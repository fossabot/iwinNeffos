package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
	"sync"

	"github.com/mediocregopher/radix/v3"
	"github.com/robfig/cron"

	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gobwas"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/kardianos/minwinsvc"
	"github.com/kardianos/osext"
	"github.com/majidbigdeli/neffosAmi/domin/config"
	"github.com/majidbigdeli/neffosAmi/domin/data"
	"github.com/majidbigdeli/neffosAmi/domin/dto"
	"github.com/majidbigdeli/neffosAmi/domin/model"
	"github.com/majidbigdeli/neffosAmi/domin/variable"

	"github.com/spf13/viper"
)

var (
	err      error
	exePath  string
	server   *neffos.Server
	pool     *radix.Pool
	prefex   string
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
	prefex = "Neffos_"
}

var events = neffos.Namespaces{
	variable.Agent: neffos.Events{
		neffos.OnNamespaceConnected: func(c *neffos.NSConn, msg neffos.Message) error {
			log.Printf("[%s] connected to namespace [%s].", c, msg.Namespace)
			return nil
		},

		neffos.OnNamespaceDisconnect: func(c *neffos.NSConn, msg neffos.Message) error {
			log.Printf("[%s] disconnected from namespace [%s].", c, msg.Namespace)
			return nil
		},

		variable.OnUpdateStatusNotification: func(c *neffos.NSConn, msg neffos.Message) error {
			id, _ := strconv.Atoi(string(msg.Body))
			data.UpdateNotification(id)
			return nil
		},
	},
}

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	server = neffos.New(gobwas.DefaultUpgrader, events)
	server.IDGenerator = func(w http.ResponseWriter, r *http.Request) string {
		if userID := r.Header.Get("userID"); userID != "" {
			return userID
		}
		return neffos.DefaultIDGenerator(w, r)
	}
	server.OnUpgradeError = func(err error) {
		log.Printf("ERROR: %v", err)
	}
	server.OnConnect = func(c *neffos.Conn) error {
		if c.WasReconnected() {
			log.Printf("[%s] connection is a result of a client-side re-connection, with tries: %d", c.ID(), c.ReconnectTries)
		}
		log.Printf("[%s] connected to the server.", c)
		return nil
	}
	server.OnDisconnect = func(c *neffos.Conn) {
		log.Printf("[%s] disconnected from the server.", c)
	}

	serveMux := http.NewServeMux()
	serveMux.Handle("/echo", server)
	serveMux.Handle("/getBroadCast", http.HandlerFunc(getBroadcastHandler))
	serveMux.Handle("/broadcast", http.HandlerFunc(broadcastHandler))

	c := cron.New()
	_ = c.AddFunc("@every "+config.NotifTime, func() {
		notificationHandler()
	})
	c.Start()

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
	var t dto.Extension
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  err.Error(),
			Body:     strconv.Itoa(t.Extension),
		})
		http.Error(w, err.Error(), 400)
		return
	}
	mutex.Lock()
	val, ok := formList[t.Extension]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if ok {
		delete(formList, t.Extension)
		w.WriteHeader(200)
		w.Write(val)
	} else {
		w.WriteHeader(200)
	}
	mutex.Unlock()
}

func notificationHandler() {
	notification, err := data.GetNotif()
	if err != nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  err.Error(),
			Body:     "",
		})
		return
	}

	if len(*notification) == 0 {
		return
	}

	server.Broadcast(nil, neffos.Message{
		Namespace: "default",
		Event:     "resiveErja",
		Body:      neffos.Marshal(notification),
	})
}

func setExtentionInRedis() {
	extUser, err := data.GetExtentionUser()
	if err != nil {
		return
	}
	if len(extUser) == 0 {
		return
	}

	// value only
	for _, v := range extUser {
		key := fmt.Sprintf("%s%s", prefex, v.Number)
		pool.Do(radix.Cmd(nil, "SET", key, v.UserID))
	}

}
