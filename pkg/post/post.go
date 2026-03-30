package post

import (
	"redditclone/pkg/comment"
	"redditclone/pkg/user"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Post struct {
	ID               bson.ObjectID      `json:"id" bson:"_id,omitempty"`
	Title            string             `json:"title" bson:"title" valid:"required,length(1|100)"`
	Type             string             `json:"type" bson:"type" valid:"required,in(link|text)"`
	UpvotePercentage float64            `json:"upvotePercentage" bson:"upvotePercentage"`
	Views            uint32             `json:"views" bson:"views"`
	Text             string             `json:"text" bson:"text"`
	URL              string             `json:"url" bson:"url"`
	Score            int                `json:"score" bson:"score"`
	Created          time.Time          `json:"created" bson:"created"`
	Comments         []*comment.Comment `json:"comments" bson:"comments,omitempty"`
	Category         string             `json:"category" bson:"category" valid:"required"`
	Author           *user.User         `json:"author" bson:"author"`
	Votes            []*Vote            `json:"votes" bson:"votes"`
}

type Vote struct {
	UserID int `json:"user" bson:"user"`
	Vote   int `json:"vote" bson:"vote"`
}
