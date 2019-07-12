package config

import (
	"github.com/majidbigdeli/neffosAmi/domin/data"
	"github.com/majidbigdeli/neffosAmi/domin/enum"
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
	//CallNotificationType برای این که بفهمیم فرم باز بشود یا خیر
	CallNotificationType enum.NotificationTypeEnum
)

//GetConfig ....
// گرفتن کانفیگ ها از فایل کانفیگ
func GetConfig() {

	NotifTime = viper.GetString("Notification.Time")
	Addr = viper.GetString("https.addr")
	CertFile = viper.GetString("https.certFile")
	KeyFile = viper.GetString("https.keyFile")
	HTTPAddr = viper.GetString("http.addr")
	CallNotificationType, _ = data.GetCallNotificationType()
}
