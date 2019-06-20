package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/majidbigdeli/neffosAmi/domin/data"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/ivahaev/amigo"
	_ "github.com/kardianos/minwinsvc"
	"github.com/kardianos/osext"
	"github.com/kataras/neffos"
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
	extenStr := e["Exten"]
	statusStr := e["Status"]
	statusText := e["StatusText"]
	extenNum, err := strconv.Atoi(extenStr)
	if err != nil {
		fmt.Printf(err.Error())
	}
	StatusNum, err := strconv.Atoi(statusStr)
	if err != nil {
		fmt.Printf(err.Error())
	}
	log.Printf("[%s] =====> [%s]", extenStr, statusText)
	data.InsertExtensionStatus(extenNum, StatusNum)
}

func queueMemberAddedHandler(e map[string]string) {
	split1 := strings.Split(e["Interface"], "/")[1]
	extensionStr := strings.Split(split1, "@")[0]
	extensionNum, err := strconv.Atoi(extensionStr)

	if err != nil {
		fmt.Printf(err.Error())
	}
	queueNum, err := strconv.Atoi(e["Queue"])

	if err != nil {
		fmt.Printf(err.Error())
	}

	data.InsertAgentQueueStatus(extensionNum, queueNum, 1550)

}

func queueMemberRemovedHandler(e map[string]string) {
	split1 := strings.Split(e["Interface"], "/")[1]
	extensionStr := strings.Split(split1, "@")[0]
	extensionNum, err := strconv.Atoi(extensionStr)

	if err != nil {
		fmt.Printf(err.Error())
	}
	queueNum, err := strconv.Atoi(e["Queue"])

	if err != nil {
		fmt.Printf(err.Error())
	}

	data.InsertAgentQueueStatus(extensionNum, queueNum, 1552)

}

func queueMemberPauseHandler(e map[string]string) {
	split1 := strings.Split(e["Interface"], "/")[1]
	extensionStr := strings.Split(split1, "@")[0]
	extensionNum, err := strconv.Atoi(extensionStr)

	if err != nil {
		fmt.Printf(err.Error())
	}
	queueNum, err := strconv.Atoi(e["Queue"])

	if err != nil {
		fmt.Printf(err.Error())
	}

	status, err := strconv.Atoi(e["Paused"])
	if err != nil {
		fmt.Printf(err.Error())
	}
	statusNum := 1551

	if status == 0 {
		statusNum = 1550
	}

	data.InsertAgentQueueStatus(extensionNum, queueNum, statusNum)

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

	// var (
	// 	dialer = gobwas.Dialer(gobwas.Options{Header: gobwas.Header{"Extension": []string{"AMI"}}})
	// )
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

	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error ExtensionStatus: %s", err.Error()))
	}

	err = a.RegisterHandler("QueueMemberAdded", queueMemberAddedHandler)

	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error QueueMemberAdded: %s", err.Error()))
	}

	err = a.RegisterHandler("QueueMemberRemoved", queueMemberRemovedHandler)

	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error QueueMemberRemoved: %s", err.Error()))
	}

	err = a.RegisterHandler("QueueMemberPause", queueMemberPauseHandler)

	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error QueueMemberPause: %s", err.Error()))
	}

	//client(dialer)

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
