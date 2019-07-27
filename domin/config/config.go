package config

import (
	"github.com/spf13/viper"
)

var (
	//NotifTime برای اینترول نوتیفیکیشن
	NotifTime string
	//Addr برای ادرس https
	Addr string
	//CertFile ادرس سرتیفیکیت فایل
	CertFile string
	//KeyFile آدرس کی فایل
	KeyFile string
	//HTTPAddr آدرس برای http
	HTTPAddr string
)

//GetConfig ....
// گرفتن کانفیگ ها از فایل کانفیگ
func GetConfig() {

	NotifTime = viper.GetString("Notification.Time")
	Addr = viper.GetString("https.addr")
	CertFile = viper.GetString("https.certFile")
	KeyFile = viper.GetString("https.keyFile")
	HTTPAddr = viper.GetString("http.addr")
}
