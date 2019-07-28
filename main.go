package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/mediocregopher/radix/v3"
	"github.com/robfig/cron"

	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gobwas"
	"github.com/valyala/fasthttp"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/kardianos/minwinsvc"
	"github.com/kardianos/osext"
	"github.com/majidbigdeli/neffosAmi/domin/config"
	"github.com/majidbigdeli/neffosAmi/domin/data"
	"github.com/majidbigdeli/neffosAmi/domin/dto"
	"github.com/majidbigdeli/neffosAmi/domin/model"
	"github.com/majidbigdeli/neffosAmi/domin/variable"

	"github.com/rs/cors"
	"github.com/spf13/viper"
)

var (
	err     error
	exePath string
	server  *neffos.Server
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

		variable.OnUpdateStatusNotification: func(c *neffos.NSConn, msg neffos.Message) error {
			id, _ := strconv.Atoi(string(msg.Body))
			data.UpdateNotification(id)
			return nil
		},
	},
}

func main() {

	customConnFunc := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr,
			radix.DialTimeout(1*time.Minute),
			radix.DialAuthPass("iwinrds"),
		)
	}

	// this pool will use our ConnFunc for all connections it creates
	pool, err := radix.NewPool("tcp", "10.1.10.33:6379", 100, radix.PoolConnFunc(customConnFunc("tcp", "10.1.10.33:6379")))

	if err != nil {
		panic(err)
	}

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
	//serveMux.Handle("/broadcast", http.HandlerFunc(broadcastHandler))
	handler := cors.Default().Handler(serveMux)
	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/broadcast":
			broadcastHandler(ctx)
		default:
			ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}
	c := cron.New()
	_ = c.AddFunc("@every "+config.NotifTime, func() {
		notificationHandler()
	})
	c.Start()

	go func() {
		log.Printf("Listening on: %s\nPress CTRL/CMD+C to interrupt.", config.HTTPAddr)
		log.Fatal(fasthttp.ListenAndServe(config.HTTPAddr, m))
	}()
	//run server in https
	log.Printf("Listening on: %s\nPress CTRL/CMD+C to interrupt.", config.Addr)
	log.Fatal(http.ListenAndServeTLS(config.Addr, path.Join(exePath, config.CertFile), path.Join(exePath, config.KeyFile), handler))
}

func broadcastHandler(ctx *fasthttp.RequestCtx) {
	var userMsg dto.FormDto
	body := ctx.PostBody()
	if body == nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  "Please send a request body",
			Body:     "",
		})
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &userMsg)
	if err != nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  err.Error(),
			Body:     string(body),
		})
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	userID, err := data.GetUserIDByExtentionNumber(userMsg.Extension)
	if err != nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  err.Error(),
			Body:     string(body),
		})
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	if userID == "" {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  "userId is empty",
			Body:     string(body),
		})
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	server.Broadcast(nil, neffos.Message{
		To:        userID,
		Namespace: "default",
		Event:     "showForm",
		Body:      body,
	})
	data.InsertLogForForm(userMsg.Extension, userMsg.Direction, 1, userMsg.CallID, userMsg.CallerNumber)
	ctx.SetStatusCode(fasthttp.StatusOK)
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
