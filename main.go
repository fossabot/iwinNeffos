package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"sync"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/kardianos/minwinsvc"
	"github.com/kardianos/osext"
	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gobwas"
	"github.com/majidbigdeli/neffosAmi/domin/config"
	"github.com/majidbigdeli/neffosAmi/domin/data"
	"github.com/majidbigdeli/neffosAmi/domin/dto"
	"github.com/majidbigdeli/neffosAmi/domin/model"
	"github.com/majidbigdeli/neffosAmi/domin/variable"

	"github.com/robfig/cron"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

var (
	server  *neffos.Server
	exePath string
	mutex   = &sync.Mutex{}
)

func init() {

	pathExe, err := osext.ExecutableFolder()

	if err != nil {
		panic(fmt.Errorf("fatal error ExecutableFolder: %s", err.Error()))
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(pathExe)
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

	certPath := path.Join(exePath, config.CertFile)
	keyPath := path.Join(exePath, config.KeyFile)

	server = neffos.New(gobwas.DefaultUpgrader, events)

	server.IDGenerator = func(w http.ResponseWriter, r *http.Request) string {
		if extension := r.Header.Get("Extension"); extension != "" {
			return extension
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

		// if returns non-nil error then it refuses the client to connect to the server.
		return nil
	}

	server.OnDisconnect = func(c *neffos.Conn) {
		log.Printf("[%s] disconnected from the server.", c)
	}

	serveMux := http.NewServeMux()
	serveMux.Handle("/echo", server)

	broadcast := http.HandlerFunc(broadcastHandler)
	serveMux.Handle("/broadcast", broadcast)

	handler := cors.Default().Handler(serveMux)

	go func() {
		log.Printf("Listening on: %s\nPress CTRL/CMD+C to interrupt.", config.HTTPAddr)
		log.Fatal(http.ListenAndServe(config.HTTPAddr, handler))
	}()

	go func() {
		c := cron.New()
		_ = c.AddFunc("@every "+config.NotifTime, func() {
			notificationHandler()
		})

		c.Start()

	}()

	log.Printf("Listening on: %s\nPress CTRL/CMD+C to interrupt.", config.Addr)
	log.Fatal(http.ListenAndServeTLS(config.Addr, certPath, keyPath, handler))

}

func broadcastHandler(w http.ResponseWriter, r *http.Request) {

	var userMsg dto.FormDto

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&userMsg)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	extensionMessage := strconv.Itoa(userMsg.Extension)

	server.Broadcast(nil, neffos.Message{
		To:        extensionMessage,
		Namespace: variable.Agent,
		Event:     variable.OnShowForm,
		Body:      neffos.Marshal(userMsg),
	})

	data.InsertLogForForm(userMsg.Extension, userMsg.Direction, 1, userMsg.CallID, userMsg.CallerNumber)

	w.WriteHeader(http.StatusOK)
}

func notificationHandler() {
	mutex.Lock()

	var gg []model.IDTvp

	connections := server.GetConnections()
	for c := range connections {
		var ff model.IDTvp
		oo, err := strconv.Atoi(c)
		if err != nil {
			fmt.Println(err)
		}
		ff.ID = oo
		gg = append(gg, ff)
	}

	notif, err := data.GetNotificationList(gg)

	if err != nil {
		fmt.Println(err)
		return
	}

	if notif != nil {
		if len(*notif) > 0 {
			for _, element := range *notif {
				extensionMessage := strconv.Itoa(element.Number)
				output, err := json.Marshal(element)
				if err != nil {
					fmt.Println("error in valid json data")
					return
				}

				server.Broadcast(nil, neffos.Message{
					To:        extensionMessage,
					Namespace: variable.Agent,
					Event:     variable.Notification,
					Body:      output,
				})
			}

		}
	}

	mutex.Unlock()

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
		"updateStatusNotification": func(c *neffos.NSConn, msg neffos.Message) error {
			id, _ := strconv.Atoi(string(msg.Body))
			data.UpdateNotification(id)
			return nil
		},
	},
}
