package session

import (
	"context"
	"errors"
	"fmt"
	"redditclone/pkg/helpers"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionManager struct {
	redisDb *redis.Client
}
type SessionToken struct {
	Token string
}
type Session struct {
	UserID int
	Login  string
}
type sessKey string

const (
	sessPrefix         = "user_sessions_"
	SessKey    sessKey = "token"
	ttlHours           = 130
)

var ErrAuth = errors.New("session didnt find")

func NewSessionManager(rclient *redis.Client) *SessionManager {
	return &SessionManager{
		redisDb: rclient,
	}
}

func (sm *SessionManager) CreateSess(sess *Session) (*SessionToken, error) {
	userSess := sessPrefix + strconv.Itoa(sess.UserID)
	sessId := helpers.RandBytesHex(16)
	token, err := CreateJWT(sess.UserID, sess.Login, sessId)
	if err != nil {
		return nil, err
	}
	ex := time.Now().Add(ttlHours * time.Hour).Unix()
	errRedis := sm.redisDb.HSet(context.Background(), userSess, sessId, ex).Err()
	if errRedis != nil {
		return nil, errRedis
	}

	return &SessionToken{token}, nil
}

func (sm *SessionManager) Check(token string) (*Session, error) {
	jwtClaims, err := CheckJWT(token)
	if err != nil {
		return nil, err
	}
	sessId := jwtClaims.SessionId
	userID := jwtClaims.User.ID
	rdsKey := sessPrefix + strconv.Itoa(userID)
	_, errRedis := sm.redisDb.HGet(context.Background(), rdsKey, sessId).Result()
	if errRedis != nil {
		return nil, errRedis
	}
	sess := &Session{UserID: userID, Login: jwtClaims.User.Login}
	return sess, nil
}

func (sm *SessionManager) CleanExpiredSessions(userId int) error {
	userSess := sessPrefix + strconv.Itoa(userId)
	sessions, err := sm.redisDb.HGetAll(context.Background(), userSess).Result()
	if err != nil {
		return err
	}
	currTime := time.Now().Unix()
	for sessId, exTime := range sessions {
		exp, err := strconv.ParseInt(exTime, 10, 64)
		if err != nil {
			fmt.Printf("err parse session at clean method %v", err)
			continue
		}
		if exp < currTime {
			err := sm.redisDb.HDel(context.Background(), userSess, sessId).Err()
			if err != nil {
				fmt.Printf("cant clean session: %v"+sessId, err)
				return err
			}
		}
	}
	return nil
}

func GetSessFromCtx(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value(SessKey).(*Session)
	if !ok || sess == nil {
		return nil, ErrAuth
	}
	return sess, nil
}
