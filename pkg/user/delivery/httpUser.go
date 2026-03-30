package delivery

import (
	"net/http"
	"redditclone/pkg/hash"
	"redditclone/pkg/helpers"
	"redditclone/pkg/middleware"
	"redditclone/pkg/session"
	"redditclone/pkg/user"

	"go.uber.org/zap"
)

type UserRepoInterface interface {
	CheckUser(string, string) (*user.User, error)
	Create(string, string) (int, error)
}

type SessManagerInterface interface {
	CreateSess(*session.Session) (*session.SessionToken, error)
	CleanExpiredSessions(int) error
}

type UserHandler struct {
	UserRepo UserRepoInterface
	Session  SessManagerInterface
}

type AuthForm struct {
	Login    string `json:"username" valid:"required,length(1|32)"`
	Password string `json:"password" valid:"required,length(8|72)"`
}

func (uh *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	u := &AuthForm{}
	err := helpers.UnmarshalAndValidate(r, u)
	ctx := r.Context()
	loger := middleware.GetLogger(ctx)
	if err != nil {
		loger.Error("request err", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	password, err := hash.HashPass(u.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userId, errDb := uh.UserRepo.Create(password, u.Login)
	if errDb != nil {
		switch errDb {
		case user.ErrUserExist:
			loger.Error("user exist:", zap.Error(errDb))
			http.Error(w, "login alredy exist", http.StatusConflict)
			return
		default:
			loger.Error("db err in create user:", zap.Error(errDb))
			http.Error(w, errDb.Error(), http.StatusInternalServerError)
			return
		}
	}
	token, err := uh.Session.CreateSess(&session.Session{UserID: userId, Login: u.Login})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := map[string]string{
		"token": token.Token,
	}
	helpers.WriteJson(w, r, http.StatusCreated, res)
	loger.Info("Created token for: ", zap.Int("User ID", userId))
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	u := &AuthForm{}
	err := helpers.UnmarshalAndValidate(r, u)
	ctx := r.Context()
	loger := middleware.GetLogger(ctx)
	if err != nil {
		loger.Error("request err", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	findedUser, err := uh.UserRepo.CheckUser(u.Login, u.Password)
	if err != nil {
		switch err {
		case user.ErrUserNotFound:
			loger.Error("login error, no user found", zap.Error(err))
			http.Error(w, "user by this login doesnt exist", http.StatusNotFound)
			return
		case user.ErrBadPass:
			loger.Error("bad password", zap.Error(err))
			http.Error(w, "invalid password", http.StatusUnauthorized)
			return
		default:
			loger.Error("DB err:", zap.Error(err))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	token, err := uh.Session.CreateSess(&session.Session{UserID: findedUser.ID, Login: findedUser.Login})
	if err != nil {
		loger.Error("jwt err:", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := map[string]string{
		"token": token.Token,
	}
	redisErr := uh.Session.CleanExpiredSessions(findedUser.ID)
	if redisErr != nil {
		loger.Error("redis error at clean expired sessions ", zap.Int("user id at login ", findedUser.ID))
	}
	helpers.WriteJson(w, r, http.StatusOK, res)
	loger.Info("logged user", zap.Int("UserID", findedUser.ID))
}
