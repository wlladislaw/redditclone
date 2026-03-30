package repo

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"redditclone/pkg/comment"
	"redditclone/pkg/post"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/mock/gomock"
)

func TestPostsRepoFindByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	mockCollection := NewMockIMongoCollection(ctrl)
	mockSingleResult := NewMockIMongoSingleResult(ctrl)
	mockUpdateResult := NewMockIMongoUpdateResult(ctrl)
	repo := &PostsRepo{
		Coll: mockCollection,
	}

	objID := bson.NewObjectID()
	expectedPost := &post.Post{
		ID: objID, Title: "tested", Text: "test text",
		Type: "text", Category: "music", Views: 0,
	}
	mockCollection.EXPECT().
		FindOne(ctx, gomock.Any(), gomock.Any()).
		Return(mockSingleResult)
	mockSingleResult.EXPECT().
		Decode(gomock.Any()).SetArg(0, expectedPost).
		Return(nil)
	mockCollection.EXPECT().
		UpdateByID(ctx, objID, bson.M{"$inc": bson.M{"views": 1}}).
		Return(mockUpdateResult, nil)

	res, err := repo.FindByID(objID)
	if err != nil {
		t.Errorf("unexpected error, got %v", err)
		return
	}
	if res.Views != 1 {
		t.Errorf("expected views up, got %d", res.Views)
		return
	}
	// err
	mockCollection.EXPECT().
		FindOne(ctx, gomock.Any(), gomock.Any()).Return(mockSingleResult)
	mockSingleResult.EXPECT().
		Decode(gomock.Any()).SetArg(0, expectedPost).
		Return(errors.New("db error"))

	res, err = repo.FindByID(objID)
	if err == nil {
		t.Errorf("expected err, got nil")
		return
	}
	if res != nil {
		t.Errorf("expected nil for post")
		return
	}
}

func TestPostsRepoAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollection := NewMockIMongoCollection(ctrl)
	mockInsertResult := NewMockIMongoInsertOneResult(ctrl)
	repo := &PostsRepo{
		Coll: mockCollection,
	}
	post := &post.Post{
		Title: "tested", Text: "test text",
		Type: "text", Category: "music", Views: 0,
	}

	newID := bson.NewObjectID()
	mockCollection.EXPECT().
		InsertOne(gomock.Any(), post).
		Return(mockInsertResult, nil)
	mockInsertResult.EXPECT().
		InsertedId().
		Return(newID)

	insertedP, err := repo.Add(post)
	if err != nil {
		t.Errorf("unexpected error, got %v", err)
		return
	}
	if insertedP.ID != newID {
		t.Errorf("expected ID %v, got %v", newID, insertedP.ID)
		return
	}
}

func TestPostsRepoGetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCollection := NewMockIMongoCollection(ctrl)
	mockCursor := NewMockIMongoCursor(ctrl)

	repo := &PostsRepo{Coll: mockCollection}
	mockCollection.EXPECT().
		Find(gomock.Any(), gomock.Any()).
		Return(mockCursor, nil)
	mockCursor.EXPECT().
		All(gomock.Any(), gomock.Any()).
		Return(nil)

	posts, err := repo.GetAll()
	if err != nil {
		t.Errorf("unexpected error, got %v", err)
		return
	}
	if posts == nil {
		t.Errorf("expected no nil")
		return
	}
	//err db
	mockCollection.EXPECT().
		Find(gomock.Any(), bson.M{}).
		Return(nil, errors.New("db error"))

	posts, err = repo.GetAll()
	if err == nil {
		t.Errorf("expected err, got nil")
		return
	}
	if posts != nil {
		t.Errorf("expected nil for posts")
		return
	}
}

func TestPostsRepoAddComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollection := NewMockIMongoCollection(ctrl)
	mockSingleResult := NewMockIMongoSingleResult(ctrl)
	repo := &PostsRepo{
		Coll: mockCollection,
	}

	postID := bson.NewObjectID()
	commAdd := &comment.Comment{
		ID:   bson.NewObjectID(),
		Body: "testcomment",
	}
	expectedPost := &post.Post{
		ID:    postID,
		Title: "testpost",
	}

	mockCollection.EXPECT().
		FindOneAndUpdate(
			gomock.Any(),
			bson.M{"_id": postID},
			gomock.Any(),
			gomock.Any(),
		).Return(mockSingleResult)
	mockSingleResult.EXPECT().
		Decode(gomock.Any()).SetArg(0, expectedPost).Return(nil)

	updatedPost, err := repo.AddComment(commAdd, postID)
	if err != nil {
		t.Errorf("unexpected error, got %v", err)
		return
	}
	if updatedPost.ID != expectedPost.ID {
		t.Errorf("expected post ID %v, got %v", expectedPost.ID, updatedPost.ID)
		return
	}
}

func TestPostsRepoFindByAuthor(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollection := NewMockIMongoCollection(ctrl)
	mockCursor := NewMockIMongoCursor(ctrl)

	repo := &PostsRepo{Coll: mockCollection}
	login := "tested"
	mockCollection.EXPECT().
		Find(gomock.Any(), bson.M{"author.login": login}).Return(mockCursor, nil)
	mockCursor.EXPECT().All(gomock.Any(), gomock.Any()).Return(nil)

	post, err := repo.FindByAuthor(login)
	if err != nil {
		t.Errorf("unexpected error, got %v", err)
		return
	}
	if post == nil {
		t.Errorf("expected no nil")
		return
	}
}

func TestPostsRepoUpVote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCollection := NewMockIMongoCollection(ctrl)
	mockSingleResult := NewMockIMongoSingleResult(ctrl)
	mockUpdateResult := NewMockIMongoSingleResult(ctrl)

	repo := &PostsRepo{Coll: mockCollection}
	oid := bson.NewObjectID()
	userID := 1
	postOld := &post.Post{
		ID: oid, Title: "tested", Text: "test text",
		Votes: []*post.Vote{{UserID: userID, Vote: -1}}, Score: -1,
	}
	expectedUp := &post.Post{
		ID: oid, Title: "tested", Text: "test text",
		Votes: []*post.Vote{{UserID: userID, Vote: 1}}, Score: 1,
	}

	mockCollection.EXPECT().
		FindOne(gomock.Any(), bson.M{"_id": oid}).Return(mockSingleResult)
	mockSingleResult.EXPECT().
		Decode(gomock.Any()).
		SetArg(0, *postOld).Return(nil)
	mockCollection.EXPECT().
		FindOneAndUpdate(
			gomock.Any(),
			bson.M{"_id": oid},
			gomock.Any(),
			gomock.Any(),
		).Return(mockUpdateResult)
	mockUpdateResult.EXPECT().Err().Return(nil)
	mockUpdateResult.EXPECT().
		Decode(gomock.Any()).SetArg(0, *expectedUp).Return(nil)

	result, err := repo.Upvote(oid, userID)
	if err != nil {
		t.Errorf("unexpected error, got %v", err)
		return
	}
	if !reflect.DeepEqual(result, expectedUp) {
		t.Errorf("expected %v, got %v", expectedUp, result)
		return
	}
}
