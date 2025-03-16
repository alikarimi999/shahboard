package event

import "github.com/alikarimi999/shahboard/types"

var (
	TopicDirectChat            = NewTopic(DomainDirectChat, ActionAny)
	TopicDirectChatMsgSent     = NewTopic(DomainDirectChat, ActionMsgSent)
	TopicDirectChatMsgApproved = NewTopic(DomainDirectChat, ActionMsgApproved)
)

type EventDirectChatMsgSent struct {
	ID         types.ObjectId `json:"id"`
	ChatID     types.ObjectId `json:"chat_id"`
	SenderId   types.ObjectId `json:"sender_id"`
	ReceiverId types.ObjectId `json:"receiver_id"`
	Content    string         `json:"content"`
	Timestamp  int64          `json:"timestamp"`
}

func (e EventDirectChatMsgSent) GetResource() string {
	return e.ChatID.String()
}

func (e EventDirectChatMsgSent) GetTopic() Topic {
	return TopicDirectChatMsgSent.SetResource(e.GetResource())
}
