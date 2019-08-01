package data

import (
	"database/sql"
	"fmt"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/majidbigdeli/neffosAmi/domin/model"
)

//InsertLogForForm ...
// لاگ اول برای زمانی که خانم وحید ای پی آی را صدا میزند
func InsertLogForForm(extension, direction, typeSend int, callID, callerNumber int64) {
	_, err := db.Exec("logging.uspInsertLogForForm1",
		sql.Named("PI_Extension", extension),
		sql.Named("PI_Direction", direction),
		sql.Named("PI_CallID", callID),
		sql.Named("PI_CallerNumber", callerNumber),
		sql.Named("PI_Type", typeSend),
	)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

//GetNotificationList ...
//گرفتن لیست نوتیفیکیشن ها
// لیست اکستنشن ها را پاس می دهیم مدل IDTvp
// در نتیجه مدل Notification میگیریم
func GetNotificationList(listUserID []model.IDTvp) (*[]model.Notification, error) {

	if len(listUserID) > 0 {
		tvpType := mssql.TVP{
			TypeName: "typ.BigIntIDType",
			Value:    listUserID,
		}

		var d []model.Notification
		err := dbData.Select(&d, "[app].[uspGetNotificationListByUserID1]",
			sql.Named("PI_UserId", tvpType),
		)
		return &d, err
	}
	return nil, nil
}

//GetNotifByUserID ...
func GetNotifByUserID(userID int) (*[]model.Notification, error) {

	d := []model.Notification{}
	err := dbData.Select(&d, "[message].[UspGetNotificationbyUserId]",
		sql.Named("PI_UserId", userID),
	)
	return &d, err
}

//GetExtentionUser ...
func GetExtentionUser() ([]model.ExtUser, error) {
	var extUser []model.ExtUser
	err := dbCore.Select(&extUser, "[app].[UspGetExtentionUser1]")
	return extUser, err

}

//UpdateNotification ...
func UpdateNotification(id, status int) {
	_, err := dbData.Exec("[app].[uspUpdateNotification1]",
		sql.Named("PI_NotificationId", id),
		sql.Named("PI_Status", status),
	)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

//SetNeffosError1 ....
func SetNeffosError1(neffosError model.NeffosError) {

	_, err := db.Exec("[logging].[UspSetNeffosError1]",
		sql.Named("PI_SocketId", neffosError.SocketID),
		sql.Named("PI_Message", neffosError.Message),
		sql.Named("PI_Body", neffosError.Body),
	)
	if err != nil {
		fmt.Printf(err.Error())
	}

}
