package model

//Notification is model
//Result For sp app.uspGetNotificationList1
type Notification struct {
	NotificationID int    `json:"NotificationId" db:"NotificationId"`
	UserID         int    `json:"UserId" db:"UserId"`
	Type           int    `json:"Type" db:"Type"`
	Status         int    `json:"Status" db:"Status"`
	MessageType    int    `json:"MessageType" db:"MessageType"`
	Data           string `json:"Data" db:"Data"`
	Message        string `json:"Message" db:"Message"`
}

// NeffosError ...
type NeffosError struct {
	SocketID string `json:"socketId" db:"socketId"`
	Message  string `json:"message" db:"message"`
	Body     string `json:"body" db:"body"`
}
