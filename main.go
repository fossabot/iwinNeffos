package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/ivahaev/amigo"
	_ "github.com/kardianos/minwinsvc"
	"github.com/kardianos/osext"
	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gobwas"
	"github.com/majidbigdeli/neffosAmi/domin/data"
	"github.com/spf13/viper"
)

type asteriskConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

var (
	a        *amigo.Amigo
	server   *neffos.Server
	err      error
	exePath  string
	showForm = "showForm"
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
}

func main() {

	//get asterisk config
	var ast asteriskConfig
	if err := viper.UnmarshalKey("asterisk", &ast); err != nil {
		panic(err)
	}

	addr := viper.GetString("server.addr")
	certFile := viper.GetString("server.certFile")
	keyFile := viper.GetString("server.keyFile")

	certPath := path.Join(exePath, certFile)
	keyPath := path.Join(exePath, keyFile)

	settings := &amigo.Settings{Host: ast.Host, Port: ast.Port, Username: ast.Username, Password: ast.Password}

	a = amigo.New(settings)
	a.Connect()

	// Listen for connection events
	a.On("connect", func(message string) {
		fmt.Println("Ami Connected To Astrisk", message)
	})
	a.On("error", func(message string) {
		fmt.Println("Connection error To Astrisk:", message)
	})

	err = a.RegisterHandler("VarSet", varSetHandler)
	if err != nil { // Handle errors reading the config file
		fmt.Printf("Fatal error VarSet: %s", err.Error())
	}

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
	serveMux.Handle("/time", th)

	log.Printf("Listening on: %s\nPress CTRL/CMD+C to interrupt.", addr)
	log.Fatal(http.ListenAndServeTLS(addr, certPath, keyPath, serveMux))

}

func varSetHandler(e map[string]string) {
	switch e["Variable"] {
	case "__ExtCallId":
		{
			// fmt.Println("__ExtCallId :", e["Value"])

			var userMsg userMessage
			s := e["Value"]
			fmt.Println("__ExtCallId :", s)
			s = strings.Replace(s, "Extension", "\"Extension\"", 1)
			s = strings.Replace(s, "CallId", "\"CallId\"", 1)
			s = strings.Replace(s, "Direction", "\"Direction\"", 1)
			s = strings.Replace(s, "CallerNumber", "\"CallerNumber\"", 1)

			err := json.Unmarshal([]byte(s), &userMsg)

			if err != nil {
				log.Print(err)
				return
			}

			if userMsg.Direction != 2910 {
				return
			}

			extensionMessage := strconv.Itoa(userMsg.Extension)

			server.Broadcast(nil, neffos.Message{
				To:        extensionMessage,
				Namespace: namespace,
				Event:     showForm,
				Body:      []byte(s),
			})

			data.InsertLogForForm(userMsg.Extension, userMsg.Direction, 1, userMsg.CallID, userMsg.CallerNumber)

		}
	}
}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	tm := time.Now().Format(time.RFC1123)
	w.Write([]byte("The time is: " + tm))
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
	},
}
