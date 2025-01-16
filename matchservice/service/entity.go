package match

import "github.com/alikarimi999/shahboard/types"

type Match struct {
	ID        types.ObjectId
	UserA     types.User
	UserB     types.User
	TimeStamp int64
}
