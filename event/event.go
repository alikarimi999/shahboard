package event

type Event interface {
	GetTopic() Topic
	TimeStamp() int64
	Encode() []byte
}
