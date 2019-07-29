package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"sync"

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
	serveMux.Handle("/getBroadCast", getBroadcastHandler)

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

	_ = c.AddFunc("@every 10m", func() {
		setExtentionInRedis()
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
	err := userMsg.UnmarshalBinary(body)
	if err != nil {
		data.SetNeffosError1(model.NeffosError{
			SocketID: "",
			Message:  err.Error(),
			Body:     string(body),
		})
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	formList[userMsg.Extension] = body
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func getBroadcastHandler() {
	mutex.Lock()
	defer mutex.Unlock()

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
