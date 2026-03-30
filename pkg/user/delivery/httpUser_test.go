package delivery

import (
	"bytes"
	"encoding/json"
	"errors"

	"net/http"
	"net/http/httptest"
	"redditclone/pkg/session"
	"redditclone/pkg/user"

	"testing"

	gomock "go.uber.org/mock/gomock"
)

func TestUserHandlerSignUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := NewMockUserRepoInterface(ctrl)
	mockSess := NewMockSessManagerInterface(ctrl)

	service := &UserHandler{
		UserRepo: st,
		Session:  mockSess,
	}
	form := AuthForm{
		Login:    "user",
		Password: "pass1231",
	}

	result := &session.SessionToken{Token: "testToken"}
	st.EXPECT().Create(gomock.Any(), form.Login).Return(1, nil)
	mockSess.EXPECT().CreateSess(gomock.Any()).Return(result, nil)
	body, _ := json.Marshal(form)

	req := httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	service.SignUp(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status created, got %d", resp.StatusCode)
		return
	}
	var resT map[string]string
	errDecode := json.NewDecoder(resp.Body).Decode(&resT)
	if errDecode != nil {
		t.Errorf("err test signup:decode res json")
		return
	}
	if resT["token"] != "testToken" {
		t.Errorf("expected token 'testToken', got %s", resT["token"])
		return
	}

	//err json parse form
	badJson := []byte(`{v`)
	req = httptest.NewRequest("POST", "/api/register", bytes.NewReader(badJson))
	w = httptest.NewRecorder()
	service.SignUp(w, req)
	resp = w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
		return
	}

	//err exist user
	st.EXPECT().Create(gomock.Any(), form.Login).Return(-1, user.ErrUserExist)
	req = httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	w = httptest.NewRecorder()
	service.SignUp(w, req)

	respErrUExist := w.Result()
	defer respErrUExist.Body.Close()

	if respErrUExist.StatusCode != http.StatusConflict {
		t.Errorf("expected status 409, got %d", respErrUExist.StatusCode)
		return
	}
	//err DB
	st.EXPECT().Create(gomock.Any(), form.Login).Return(-1, errors.New("internal error"))
	req = httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	w = httptest.NewRecorder()
	service.SignUp(w, req)
	resp = w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
		return
	}
	//err session create
	st.EXPECT().Create(gomock.Any(), form.Login).Return(1, nil)
	mockSess.EXPECT().CreateSess(gomock.Any()).Return(nil, errors.New("session error"))
	req = httptest.NewRequest("POST", "/api/register", bytes.NewReader(body))
	w = httptest.NewRecorder()
	service.SignUp(w, req)
	resp = w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
		return
	}

}

func TestUserHandlerLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	st := NewMockUserRepoInterface(ctrl)
	mockSess := NewMockSessManagerInterface(ctrl)

	service := &UserHandler{
		UserRepo: st,
		Session:  mockSess,
	}
	form := AuthForm{
		Login:    "user",
		Password: "pass1231",
	}

	result := &session.SessionToken{Token: "testToken"}
	st.EXPECT().CheckUser(form.Login, form.Password).Return(&user.User{ID: 1, Login: form.Login}, nil)
	mockSess.EXPECT().CreateSess(gomock.Any()).Return(result, nil)
	body, _ := json.Marshal(form)

	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	service.Login(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
		return
	}
	//err json parse form
	badJson := []byte(`{v`)
	req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(badJson))
	w = httptest.NewRecorder()
	service.Login(w, req)
	resp = w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
		return
	}

	//err bad pass
	st.EXPECT().CheckUser(form.Login, form.Password).Return(nil, user.ErrBadPass)
	req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	w = httptest.NewRecorder()
	service.Login(w, req)
	resp = w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 bad password, got %d", resp.StatusCode)
		return
	}

	//err user not found
	st.EXPECT().CheckUser(form.Login, form.Password).Return(nil, user.ErrUserNotFound)
	req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	w = httptest.NewRecorder()
	service.Login(w, req)
	resp = w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
		return
	}

	//err DB
	st.EXPECT().CheckUser(form.Login, form.Password).Return(nil, errors.New("internal error"))
	req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	w = httptest.NewRecorder()
	service.Login(w, req)
	resp = w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
		return
	}

	//err session create
	st.EXPECT().CheckUser(form.Login, form.Password).Return(&user.User{ID: 1, Login: form.Login}, nil)
	mockSess.EXPECT().CreateSess(gomock.Any()).Return(nil, errors.New("session error"))
	req = httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	w = httptest.NewRecorder()
	service.Login(w, req)
	resp = w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
		return
	}

}
