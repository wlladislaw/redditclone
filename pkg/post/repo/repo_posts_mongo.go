package repo

import (
	"context"
	"fmt"
	"math"
	"redditclone/pkg/comment"
	"redditclone/pkg/post"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type PostsRepo struct {
	Coll IMongoCollection
}

func NewRepo(coll *MongoCollection) *PostsRepo {
	return &PostsRepo{Coll: coll}
}

func (repo *PostsRepo) GetAll() ([]*post.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	posts := make([]*post.Post, 0, 10)
	res, err := repo.Coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("errDb %w in findall method", err)
	}

	err = res.All(ctx, &posts)
	if err != nil {
		return nil, fmt.Errorf("errDecode %w in findall", err)
	}
	return posts, nil
}

func (repo *PostsRepo) FindByCategory(field string) ([]*post.Post, error) {
	posts := make([]*post.Post, 0, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "score", Value: 1}})
	cursor, err := repo.Coll.Find(ctx, bson.D{{Key: "category", Value: field}}, opts)
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &posts)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (repo *PostsRepo) FindByAuthor(login string) ([]*post.Post, error) {
	posts := make([]*post.Post, 0, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cursor, err := repo.Coll.Find(ctx, bson.M{"author.login": login})
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &posts)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (repo *PostsRepo) FindByID(oid bson.ObjectID) (*post.Post, error) {
	post := &post.Post{}
	filter := bson.D{{Key: "_id", Value: oid}}
	errDb := repo.Coll.FindOne(context.Background(), filter).Decode(&post)
	if errDb != nil {
		return nil, errDb
	}
	post.Views++
	updatedViews := bson.M{"$inc": bson.M{"views": 1}}
	_, errUpdate := repo.Coll.UpdateByID(context.Background(), oid, updatedViews)
	if errUpdate != nil {
		return nil, errUpdate
	}

	return post, nil
}

func (repo *PostsRepo) Add(p *post.Post) (*post.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	p.Created = time.Now()
	res, err := repo.Coll.InsertOne(ctx, p)
	if err != nil {
		return nil, err
	}
	oid := res.InsertedId().(bson.ObjectID)
	p.ID = oid
	return p, nil
}

func (repo *PostsRepo) AddComment(c *comment.Comment, oid bson.ObjectID) (*post.Post, error) {
	c.Created = time.Now()
	c.ID = bson.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	updateResult := repo.Coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": oid},
		bson.M{"$push": bson.M{"comments": c}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	updatedPost := &post.Post{}
	errDecode := updateResult.Decode(&updatedPost)
	if errDecode != nil {
		return nil, errDecode
	}
	return updatedPost, nil
}

func (repo *PostsRepo) DestroyComment(commOid, oidPost bson.ObjectID) (*post.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	updateResult := repo.Coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": oidPost},
		bson.M{"$pull": bson.M{"comments": bson.M{"_id": commOid}}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	updatedPost := &post.Post{}
	errDecode := updateResult.Decode(&updatedPost)
	if errDecode != nil {
		return nil, errDecode
	}
	return updatedPost, nil
}

func (repo *PostsRepo) Destroy(oidPost bson.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := repo.Coll.DeleteOne(ctx, bson.M{"_id": oidPost})
	if err != nil {
		return err
	}
	return nil
}

func (repo *PostsRepo) Upvote(oid bson.ObjectID, userId int) (*post.Post, error) {
	return repo.votePost(oid, userId, 1)
}

func (repo *PostsRepo) Downvote(oid bson.ObjectID, userId int) (*post.Post, error) {
	return repo.votePost(oid, userId, -1)
}
func (repo *PostsRepo) Unvote(oid bson.ObjectID, userId int) (*post.Post, error) {
	return repo.votePost(oid, userId, 0)
}

func (repo *PostsRepo) votePost(oid bson.ObjectID, userId int, vote int) (*post.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var currentPost post.Post
	err := repo.Coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&currentPost)
	if err != nil {
		return nil, err
	}

	var foundV *post.Vote
	for _, v := range currentPost.Votes {
		if v.UserID == userId {
			foundV = v
			break
		}
	}

	switch {
	case foundV != nil:
		switch vote {
		case 1:
			if foundV.Vote != 1 {
				currentPost.Score += 2
			}
			foundV.Vote = 1
		case -1:
			if foundV.Vote != -1 {
				currentPost.Score -= 2
			}
			foundV.Vote = -1
		case 0:
			if foundV.Vote == 1 {
				currentPost.Score--
			} else if foundV.Vote == -1 {
				currentPost.Score++
			}
			foundV.Vote = 0
		}
	default:
		if vote == 1 || vote == -1 {
			currentPost.Votes = append(currentPost.Votes, &post.Vote{UserID: userId, Vote: vote})
			currentPost.Score += vote
		}
	}
	currentPost.UpvotePercentage = countUpPercentage(currentPost.Votes)
	update := bson.M{
		"$set": bson.M{
			"score":            currentPost.Score,
			"upvotePercentage": currentPost.UpvotePercentage,
		},
	}

	if vote == 0 {
		update["$pull"] = bson.M{"votes": bson.M{"user": userId}}
	} else {
		update["$set"].(bson.M)["votes"] = currentPost.Votes
	}

	res := repo.Coll.FindOneAndUpdate(
		ctx,
		bson.M{"_id": oid},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	if res.Err() != nil {
		return nil, res.Err()
	}

	var updatedPost post.Post
	err = res.Decode(&updatedPost)
	if err != nil {
		return nil, err
	}

	return &updatedPost, nil
}

func countUpPercentage(votes []*post.Vote) float64 {
	lenVotes := len(votes)
	if lenVotes == 0 {
		return 0
	}
	upvotes := 0
	for _, v := range votes {
		if v.Vote == 1 {
			upvotes++
		}
		if v.Vote == 0 {
			lenVotes--
		}
	}
	if upvotes == 0 {
		return 0
	}

	return math.Floor(float64(upvotes) / float64(lenVotes) * 100)
}
