package main

import (
	"fmt"
	"log"

	"github.com/majidbigdeli/neffosAmi/domin/data"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/ivahaev/amigo"
	_ "github.com/kardianos/minwinsvc"
	"github.com/kardianos/osext"
	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gobwas"
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
	c        *neffos.NSConn
	err      error
	endpoint string
)

const (
	namespace = "default"
)

func varSetHandler(e map[string]string) {
	switch e["Variable"] {
	case "__ExtCallId":
		{
			fmt.Println("__ExtCallId :", e["Value"])
			_ = c.Emit("send", []byte(e["Value"]))
		}
	}
}

func extensionStatusHandler(e map[string]string) {

}

var handler = neffos.Namespaces{
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

	var (
		dialer = gobwas.Dialer(gobwas.Options{Header: gobwas.Header{"Extension": []string{"AMI"}}})
	)
	endpoint = viper.GetString("socket.endpoint")

	//get asterisk config
	var ast asteriskConfig
	if err := viper.UnmarshalKey("asterisk", &ast); err != nil {
		panic(err)
	}
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
		panic(fmt.Errorf("Fatal error VarSet: %s", err.Error()))
	}

	err = a.RegisterHandler("ExtensionStatus", extensionStatusHandler)

	client(dialer)

	ch := make(chan bool)
	<-ch
}

func client(dialer neffos.Dialer) {

	client, err := neffos.Dial(nil, dialer, endpoint, handler)
	if err != nil {
		log.Print(err)
	}
	//defer client.Close()
	c, err = client.Connect(nil, namespace)
	if err != nil {
		panic(err)
	}
}
