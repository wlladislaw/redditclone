package delivery

import (
	"context"
	"errors"
	"strings"

	"net/http"
	"redditclone/pkg/helpers"
	"redditclone/pkg/middleware"
	"redditclone/pkg/post"
	"redditclone/pkg/session"
	"redditclone/pkg/user"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/v2/bson"

	"go.uber.org/zap"
)

type PostRepoInterface interface {
	GetAll() ([]*post.Post, error)
	FindByCategory(string) ([]*post.Post, error)
	FindByAuthor(string) ([]*post.Post, error)
	FindByID(bson.ObjectID) (*post.Post, error)
	Add(*post.Post) (*post.Post, error)
	Destroy(bson.ObjectID) error
	Upvote(bson.ObjectID, int) (*post.Post, error)
	Downvote(bson.ObjectID, int) (*post.Post, error)
	Unvote(bson.ObjectID, int) (*post.Post, error)
}
type PostsHandler struct {
	PostsRepo PostRepoInterface
}

func (ph *PostsHandler) List(w http.ResponseWriter, r *http.Request) {
	posts, err := ph.PostsRepo.GetAll()
	ctx := r.Context()
	if err != nil {
		middleware.GetLogger(ctx).Error("List posts db err:", zap.Error(err))
		http.Error(w, "internal db error ", http.StatusInternalServerError)
		return
	}

	helpers.WriteJson(w, r, http.StatusOK, posts)
}

func (ph *PostsHandler) ListByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryName := vars["CATEGORY_NAME"]
	posts, err := ph.PostsRepo.FindByCategory(categoryName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.WriteJson(w, r, http.StatusOK, &posts)
}

func (ph *PostsHandler) ListByLogin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	login := vars["USER_LOGIN"]
	posts, err := ph.PostsRepo.FindByAuthor(login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.WriteJson(w, r, http.StatusOK, &posts)
}

func (ph *PostsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	oid, err := helpers.GetOIdFromParams(r, "POST_ID")
	ctx := r.Context()
	if err != nil {
		middleware.GetLogger(ctx).Error("request err:", zap.Error(err))
		http.Error(w, "invalid post ID in params", http.StatusBadRequest)
		return
	}

	post, errDB := ph.PostsRepo.FindByID(oid)
	if errDB != nil {
		http.Error(w, "error DB", http.StatusInternalServerError)
		return
	}
	if post == nil {
		http.Error(w, "post by this ID not found", http.StatusNotFound)
		return
	}
	helpers.WriteJson(w, r, http.StatusOK, &post)
}

func (ph *PostsHandler) AddPost(w http.ResponseWriter, r *http.Request) {
	p := &post.Post{}
	ctx := r.Context()

	err := helpers.UnmarshalAndValidate(r, p)
	if err != nil {
		middleware.GetLogger(ctx).Error("json or validation err", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	errTextUrlValid := checkTextUrl(p)
	if errTextUrlValid != nil {
		http.Error(w, errTextUrlValid.Error(), http.StatusBadRequest)
		return
	}

	sess, err := session.GetSessFromCtx(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	p.Author = &user.User{ID: sess.UserID, Login: sess.Login}
	post, errDb := ph.PostsRepo.Add(p)
	if errDb != nil {
		http.Error(w, errDb.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteJson(w, r, http.StatusCreated, &post)
	middleware.GetLogger(ctx).Info("Insert post :", zap.String("ID", p.ID.Hex()))
}

func (ph *PostsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := helpers.GetOIdFromParams(r, "POST_ID")
	ctx := r.Context()
	if err != nil {
		middleware.GetLogger(ctx).Error("request err :", zap.Error(err))
		http.Error(w, "invalid post ID in params", http.StatusBadGateway)
		return
	}
	sess, err := session.GetSessFromCtx(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	post, errDB := ph.PostsRepo.FindByID(id)
	if errDB != nil {
		http.Error(w, "error DB", http.StatusInternalServerError)
		return
	}
	if post == nil {
		http.Error(w, "post by this ID not found", http.StatusNotFound)
		return
	}
	if post.Author.ID != sess.UserID {
		middleware.GetLogger(ctx).Error("try delete not own post:", zap.String("ID", id.Hex()))
		http.Error(w, "havent creds for this operation", http.StatusForbidden)
		return
	}

	errDel := ph.PostsRepo.Destroy(id)
	if errDel != nil {
		middleware.GetLogger(ctx).Error("cant destroy post:", zap.String("ID", id.Hex()))
		http.Error(w, errDel.Error(), http.StatusNotFound)
		return
	}

	resSuccess := map[string]string{"message": "success"}

	helpers.WriteJson(w, r, http.StatusOK, &resSuccess)
	middleware.GetLogger(ctx).Info("Delete post: ", zap.String("ID", id.Hex()))
}

func (ph *PostsHandler) Upvote(w http.ResponseWriter, r *http.Request) {
	ph.vote(w, r, 1)
}

func (ph *PostsHandler) Downvote(w http.ResponseWriter, r *http.Request) {
	ph.vote(w, r, -1)
}

func (ph *PostsHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	ph.vote(w, r, 0)
}

func (ph *PostsHandler) vote(w http.ResponseWriter, r *http.Request, vote int) {
	var post *post.Post
	var errDb error
	ctx := context.Background()
	id, err := helpers.GetOIdFromParams(r, "POST_ID")
	if err != nil {
		middleware.GetLogger(ctx).Error("request err :", zap.Error(err))
		http.Error(w, "invalid post ID in params", http.StatusBadRequest)
		return
	}
	sess, err := session.GetSessFromCtx(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	switch vote {
	case 1:
		post, errDb = ph.PostsRepo.Upvote(id, sess.UserID)
	case -1:
		post, errDb = ph.PostsRepo.Downvote(id, sess.UserID)
	case 0:
		post, errDb = ph.PostsRepo.Unvote(id, sess.UserID)
	}
	if errDb != nil {
		http.Error(w, errDb.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteJson(w, r, http.StatusOK, post)
	middleware.GetLogger(ctx).Info("vote on post:", zap.String("ID", post.ID.Hex()))
}

func checkTextUrl(post *post.Post) error {
	if post.Type == "text" {
		if post.Text == "" {
			return errors.New("text is required field")
		}
		if len(strings.Trim(post.Text, " ")) < 4 {
			return errors.New("must be at least 4 characters long")
		}
	} else {
		if post.URL == "" {
			return errors.New("URL is required field")
		}
		if !govalidator.IsURL(post.URL) {
			return errors.New("invalid URL")
		}
	}
	return nil
}
