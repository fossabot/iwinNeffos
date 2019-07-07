package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/kardianos/minwinsvc"
	"github.com/kardianos/osext"
	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gobwas"
	"github.com/majidbigdeli/neffosAmi/domin/data"
	"github.com/majidbigdeli/neffosAmi/domin/model"
	"github.com/robfig/cron"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

var (
	server   *neffos.Server
	err      error
	exePath  string
	showForm = "showForm"
	mutex    = &sync.Mutex{}
)

const (
	namespace = "default"
)

type userMessage struct {
	Extension    int   `json:"Extension"`
	Direction    int   `json:"Direction"`
	CallID       int64 `json:"CallId"`
	CallerNumber int64 `json:"CallerNumber"`
}

func init() {

	path, err := osext.ExecutableFolder()

	if err != nil {
		panic(fmt.Errorf("Fatal error ExecutableFolder: %s", err.Error()))
	}

	viper.SetConfigName("config")
	viper.AddConfigPath(path)
	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err.Error()))
	}
	data.GetDB()
	data.GetDBData()
}

func main() {

	trigerTime := viper.GetString("Notification.TrigerTime")
	addr := viper.GetString("server.addr")
	certFile := viper.GetString("server.certFile")
	keyFile := viper.GetString("server.keyFile")

	certPath := path.Join(exePath, certFile)
	keyPath := path.Join(exePath, keyFile)

	server = neffos.New(gobwas.DefaultUpgrader, events)
	server.IDGenerator = func(w http.ResponseWriter, r *http.Request) string {
		if userID := r.Header.Get("Extension"); userID != "" {
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

		// if returns non-nil error then it refuses the client to connect to the server.
		return nil
	}

	server.OnDisconnect = func(c *neffos.Conn) {
		log.Printf("[%s] disconnected from the server.", c)
	}

	serveMux := http.NewServeMux()
	serveMux.Handle("/echo", server)
	th := http.HandlerFunc(timeHandler)
	broadcast := http.HandlerFunc(broadcastHandler)
	getbroadcast := http.HandlerFunc(getbroadcastHandeler)
	serveMux.Handle("/time", th)
	serveMux.Handle("/broadcast", broadcast)
	serveMux.Handle("/getBroadcast", getbroadcast)

	handler := cors.Default().Handler(serveMux)

	go func() {
		log.Printf("Listening on: %s\nPress CTRL/CMD+C to interrupt.", "3812")
		log.Fatal(http.ListenAndServe(":3812", handler))
	}()

	go func() {
		c := cron.New()
		c.AddFunc("@every "+trigerTime, func() {
			notificationHandler()
		})
		c.Start()

	}()

	log.Printf("Listening on: %s\nPress CTRL/CMD+C to interrupt.", addr)
	log.Fatal(http.ListenAndServeTLS(addr, certPath, keyPath, handler))

}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	tm := time.Now().Format(time.RFC1123)
	w.Write([]byte("The time is: " + tm))
}

func broadcastHandler(w http.ResponseWriter, r *http.Request) {

	var userMsg userMessage

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&userMsg)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	output, err := json.Marshal(userMsg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	extensionMessage := strconv.Itoa(userMsg.Extension)

	server.Broadcast(nil, neffos.Message{
		To:        extensionMessage,
		Namespace: namespace,
		Event:     showForm,
		Body:      output,
	})

	data.InsertLogForForm(userMsg.Extension, userMsg.Direction, 1, userMsg.CallID, userMsg.CallerNumber)

	w.WriteHeader(http.StatusOK)
}

func getbroadcastHandeler(w http.ResponseWriter, r *http.Request) {

	queryValues := r.URL.Query()

	extenstionString := queryValues.Get("Extension")
	extenstion, err := strconv.Atoi(extenstionString)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}
	direction, err := strconv.Atoi(queryValues.Get("Direction"))
	if err != nil {
		http.Error(w, err.Error(), 400)
	}

	callID, err := strconv.ParseInt(queryValues.Get("CallId"), 10, 64)

	if err != nil {
		http.Error(w, err.Error(), 400)
	}

	callerNumber, err := strconv.ParseInt(queryValues.Get("CallerNumber"), 10, 64)

	if err != nil {
		http.Error(w, err.Error(), 400)
	}

	userMsg := &userMessage{
		Extension:    extenstion,
		Direction:    direction,
		CallID:       callID,
		CallerNumber: callerNumber,
	}

	output, err := json.Marshal(userMsg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	server.Broadcast(nil, neffos.Message{
		To:        extenstionString,
		Namespace: namespace,
		Event:     showForm,
		Body:      output,
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
					Namespace: namespace,
					Event:     "notification",
					Body:      output,
				})
			}
		}
	}

	mutex.Unlock()

}

var events = neffos.Namespaces{
	namespace: neffos.Events{
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
