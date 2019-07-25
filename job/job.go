package job

import (
	"github.com/majidbigdeli/neffosAmi/controller"
	"github.com/majidbigdeli/neffosAmi/domin/config"
	"github.com/robfig/cron"
)

//Jobs ...
func Jobs() {
	c := cron.New()
	//جاب گرفتن نوتیفیکیشن ها
	_ = c.AddFunc("@every "+config.NotifTime, func() {
		controller.NotificationHandler()
	})
	c.Start()
}
