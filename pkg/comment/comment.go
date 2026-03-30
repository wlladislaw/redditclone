package comment

import (
	"redditclone/pkg/user"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Comment struct {
	ID      bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Author  user.User     `json:"author" bson:"author"`
	Body    string        `json:"body" bson:"body"`
	Created time.Time     `json:"created" bson:"created"`
}
