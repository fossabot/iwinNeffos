package enum

type enum interface {
	Name() string
	Orginal() int
	Values() *[]string
}

//NotificationTypeEnum ...
type NotificationTypeEnum int

const (
	//BroadcastAll ...
	BroadcastAll NotificationTypeEnum = iota + 23700
	//CustomProject ...
	CustomProject
	//NoficationDisabled ...
	NoficationDisabled
)

var notificationTypeEnumString = []string{
	"BroadcastAll",
	"CustomProject",
	"NoficationDisabled",
}

//Name ...
func (pn NotificationTypeEnum) Name() string {
	return notificationTypeEnumString[pn]
}

//Orginal ...
func (pn NotificationTypeEnum) Orginal() int {
	return int(pn)
}

//Values ...
func (pn NotificationTypeEnum) Values() *[]string {
	return &notificationTypeEnumString
}
