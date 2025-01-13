package event

type Event interface {
	GetTopic() Topic
	GetAction() Action
	TimeStamp() int64
	Encode() []byte
}
