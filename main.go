package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
	"sync"

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
	err              error
	exePath          string
	server           *neffos.Server
	formList         = make(map[int][]byte)
	mutex            = &sync.Mutex{}
	connections      = make(map[string]*neffos.Conn)
	addConnection    = make(chan *neffos.Conn)
	removeConnection = make(chan *neffos.Conn)
	notify           = make(chan model.Notification)
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
		addConnection <- c
		if c.WasReconnected() {
			log.Printf("[%s] connection is a result of a client-side re-connection, with tries: %d", c.ID(), c.ReconnectTries)
		}
		log.Printf("[%s] connected to the server.", c)
		return nil
	}
	server.OnDisconnect = func(c *neffos.Conn) {
		removeConnection <- c
		log.Printf("[%s] disconnected from the server.", c)
	}

	serveMux := http.NewServeMux()
	serveMux.Handle("/echo", server)
	serveMux.Handle("/setBroadCast", http.HandlerFunc(setBroadcastHandler))
	serveMux.Handle("/broadcast", http.HandlerFunc(broadcastHandler))

	go func() {
		c := cron.New()
		_ = c.AddFunc("@every "+config.NotifTime, func() {
			notificationHandler()
		})
		c.Start()
	}()

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

func setBroadcastHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	number := r.URL.Query().Get("Number")
	num, err := strconv.Atoi(number)
	if err != nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  err.Error(),
			Body:     number,
		})
		http.Error(w, err.Error(), 400)
		return
	}
	mutex.Lock()
	val, ok := formList[num]
	if ok {
		delete(formList, num)
		w.WriteHeader(200)
		w.Write(val)
	} else {
		w.WriteHeader(200)
	}
	mutex.Unlock()
}

func notificationHandler() {

	connectionIDs := []model.IDTvp{}

	for c := range connections {
		var connectionID model.IDTvp
		cID, err := strconv.Atoi(c)
		if err != nil {
			fmt.Println(err)
		}
		connectionID.ID = cID
		connectionIDs = append(connectionIDs, connectionID)
	}

	notifications, err := data.GetNotificationList(connectionIDs)

	if err != nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  err.Error(),
			Body:     "",
		})
		return
	}

	if notifications == nil {
		return
	}

	if n := len(*notifications); n == 0 {
		return
	}

	for _, nf := range *notifications {
		notify <- nf
	}

}

func startConnectionManager(ctx context.Context) {
	if ctx == nil {
		ctx = context.TODO()
	}

	go func() {
		for {
			select {
			case c := <-addConnection:
				connections[c.ID()] = c
			case c := <-removeConnection:
				delete(connections, c.ID())
			case nf := <-notify:
				uID := strconv.Itoa(nf.UserID)
				c, ok := connections[uID]
				if !ok {
					data.UpdateNotification(nf.UserID, 22710)
					continue
				} else {
					ok = c.Write(neffos.Message{
						Namespace: variable.Agent,
						Event:     "notif",
						Body:      neffos.Marshal(nf),
					})
					if ok {
						data.UpdateNotification(nf.UserID, 22712)
					} else {
						data.UpdateNotification(nf.UserID, 22710)
					}
				}

			case <-ctx.Done():
				return
			}
		}
	}()

}
