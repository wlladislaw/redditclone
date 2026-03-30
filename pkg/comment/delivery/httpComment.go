package delivery

import (
	"encoding/json"

	"net/http"
	"redditclone/pkg/comment"
	"redditclone/pkg/helpers"
	"redditclone/pkg/post"
	"redditclone/pkg/session"
	"redditclone/pkg/user"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"
)

type PostRepoInterface interface {
	AddComment(comment *comment.Comment, id bson.ObjectID) (*post.Post, error)
	DestroyComment(idComm, idPost bson.ObjectID) (*post.Post, error)
}

type CommentHandler struct {
	Logger    *zap.Logger
	PostsRepo PostRepoInterface
}

type ReqData struct {
	Comment string `json:"comment" valid:"required,length(1|2000)"`
}

func (ch *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	id, err := helpers.GetOIdFromParams(r, "POST_ID")
	if err != nil {
		ch.Logger.Error("request err :", zap.Error(err))
		http.Error(w, "invalid post ID in params", http.StatusBadRequest)
		return
	}

	reqData := &ReqData{}
	errReq := helpers.UnmarshalAndValidate(r, reqData)
	if errReq != nil {
		ch.Logger.Error("request err", zap.Error(errReq))
		http.Error(w, errReq.Error(), http.StatusBadRequest)
		return
	}

	sess, err := session.GetSessFromCtx(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	comment := &comment.Comment{}
	comment.Author = user.User{ID: sess.UserID, Login: sess.Login}
	comment.Body = reqData.Comment

	post, err := ch.PostsRepo.AddComment(comment, id)
	if err != nil {
		http.Error(w, "error DB", http.StatusInternalServerError)
		return
	}

	if post == nil {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}
	res, err := json.Marshal(post)
	if err != nil {
		http.Error(w, "cant pack json", http.StatusInternalServerError)
	}

	ch.Logger.Info("Insert at post :",
		zap.String("ID post", post.ID.Hex()),
		zap.String("comment", reqData.Comment))

	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}

func (ch *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	idPost, err := helpers.GetOIdFromParams(r, "POST_ID")
	if err != nil {
		ch.Logger.Error("request err :", zap.Error(err))
		http.Error(w, "invalid post ID in params", http.StatusBadRequest)
		return
	}
	oidComment, err := helpers.GetOIdFromParams(r, "COMMENT_ID")
	if err != nil {
		ch.Logger.Error("request err :", zap.Error(err))
		http.Error(w, "invalid comment ID in params", http.StatusBadRequest)
		return
	}

	post, err := ch.PostsRepo.DestroyComment(oidComment, idPost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if post == nil {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	res, err := json.Marshal(post)
	if err != nil {
		http.Error(w, "cant pack json", http.StatusInternalServerError)
	}

	ch.Logger.Info("Delete comment by:",
		zap.String("ID", oidComment.Hex()),
	)

	w.Write(res)
}
