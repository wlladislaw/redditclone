package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"redditclone/pkg/post"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"testing"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/mock/gomock"
)

type Case struct {
	name        string
	method      string
	url         string
	urlVars     map[string]string
	body        []byte
	session     *session.Session
	mockST      func()
	handlerFunc func(w http.ResponseWriter, r *http.Request)
	status      int
}

func TestPostsHandler(t *testing.T) {
	testPost := &post.Post{
		ID: bson.NewObjectID(), Title: "tested", Text: "test text", Type: "text",
		Category: "music", Author: &user.User{ID: 1, Login: "user1"}}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	st := NewMockPostRepoInterface(ctrl)
	service := &PostsHandler{
		PostsRepo: st,
	}

	validSess := &session.Session{UserID: testPost.Author.ID, Login: testPost.Author.Login}
	otherSess := &session.Session{UserID: 77, Login: "otheruser"}
	validID := testPost.ID.Hex()
	invalidID := "badId"

	cases := []Case{
		{
			name:        "List posts",
			method:      http.MethodGet,
			url:         "/api/posts/",
			handlerFunc: service.List,
			mockST: func() {
				st.EXPECT().GetAll().Return([]*post.Post{testPost}, nil)
			},
			status: http.StatusOK,
		},
		{
			name:        "List by login",
			method:      http.MethodGet,
			url:         "/api/user/" + testPost.Author.Login,
			urlVars:     map[string]string{"USER_LOGIN": testPost.Author.Login},
			handlerFunc: service.ListByLogin,
			mockST: func() {
				st.EXPECT().FindByAuthor(testPost.Author.Login).Return([]*post.Post{testPost}, nil)
			},
			status: http.StatusOK,
		},
		{
			name:        "List by category",
			method:      http.MethodGet,
			url:         "/api/posts/" + testPost.Category,
			urlVars:     map[string]string{"CATEGORY_NAME": testPost.Category},
			handlerFunc: service.ListByCategory,
			mockST: func() {
				st.EXPECT().FindByCategory(testPost.Category).Return([]*post.Post{testPost}, nil)
			},
			status: http.StatusOK,
		},
		{
			name:        "Get by valid ID",
			method:      http.MethodGet,
			url:         "/api/posts/" + validID,
			urlVars:     map[string]string{"POST_ID": validID},
			handlerFunc: service.GetByID,
			mockST: func() {
				st.EXPECT().FindByID(testPost.ID).Return(testPost, nil)
			},
			status: http.StatusOK,
		},
		{
			name:        "Get by invalid ID",
			method:      http.MethodGet,
			url:         "/api/posts/" + invalidID,
			urlVars:     map[string]string{"POST_ID": invalidID},
			handlerFunc: service.GetByID,
			status:      http.StatusBadRequest,
		},
		{
			name:        "Add post with auth",
			method:      http.MethodPost,
			url:         "/api/posts",
			session:     validSess,
			body:        jsonMarshal(testPost, t),
			handlerFunc: service.AddPost,
			mockST: func() {
				st.EXPECT().Add(gomock.Any()).Return(testPost, nil)
			},
			status: http.StatusCreated,
		},
		{
			name:        "Addpost no auth",
			method:      http.MethodPost,
			url:         "/api/posts",
			body:        jsonMarshal(testPost, t),
			handlerFunc: service.AddPost,
			status:      http.StatusUnauthorized,
		},
		{
			name:        "Delete post by owner",
			method:      http.MethodDelete,
			url:         "/api/post/" + validID,
			urlVars:     map[string]string{"POST_ID": validID},
			session:     validSess,
			handlerFunc: service.Delete,
			mockST: func() {
				st.EXPECT().FindByID(testPost.ID).Return(testPost, nil)
				st.EXPECT().Destroy(testPost.ID).Return(nil)
			},
			status: http.StatusOK,
		},
		{
			name:        "Delete post by other user",
			method:      http.MethodDelete,
			url:         "/api/post/" + validID,
			urlVars:     map[string]string{"POST_ID": validID},
			session:     otherSess,
			handlerFunc: service.Delete,
			mockST: func() {
				st.EXPECT().FindByID(testPost.ID).Return(testPost, nil)
			},
			status: http.StatusForbidden,
		},
		{
			name:        "Delete no session",
			method:      http.MethodDelete,
			url:         "/api/post/" + validID,
			urlVars:     map[string]string{"POST_ID": validID},
			handlerFunc: service.Delete,
			status:      http.StatusUnauthorized,
		},
		{
			name:        "Upvote valid",
			method:      http.MethodGet,
			url:         "/api/post/" + validID + "/upvote",
			urlVars:     map[string]string{"POST_ID": validID},
			session:     validSess,
			handlerFunc: service.Upvote,
			mockST: func() {
				st.EXPECT().Upvote(testPost.ID, validSess.UserID).Return(testPost, nil)
			},
			status: http.StatusOK,
		},
		{
			name:        "Upvote invalid postID",
			method:      http.MethodGet,
			url:         "/api/post/" + invalidID + "/upvote",
			urlVars:     map[string]string{"POST_ID": invalidID},
			handlerFunc: service.Upvote,
			status:      http.StatusBadRequest,
		},
		{
			name:        "Upvote no auth",
			method:      http.MethodGet,
			url:         "/api/post/" + validID + "/upvote",
			urlVars:     map[string]string{"POST_ID": validID},
			handlerFunc: service.Upvote,
			status:      http.StatusUnauthorized,
		},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			if item.mockST != nil {
				item.mockST()
			}
			req := httptest.NewRequest(item.method, item.url, bytes.NewReader(item.body))
			if item.urlVars != nil {
				req = mux.SetURLVars(req, item.urlVars)
			}
			if item.session != nil {
				req = req.WithContext(context.WithValue(req.Context(), session.SessKey, item.session))
			}
			w := httptest.NewRecorder()
			item.handlerFunc(w, req)

			resp := w.Result()
			if resp.StatusCode != item.status {
				t.Errorf(" --- expected status %d, got %d", item.status, resp.StatusCode)
				return
			}
		})
	}
}

func jsonMarshal(val interface{}, t *testing.T) []byte {
	data, err := json.Marshal(val)
	if err != nil {
		t.Errorf("failed to pack json: %v", err)
	}
	return data
}
