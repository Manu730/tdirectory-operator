package hndlr

import (
	"errors"
	"net/http"
	"tdirectory/common"
	"tdirectory/db"
)

const (
	SUCCESS          = "SUCCESS"
	FAILURE          = "FAILURE"
	NONE             = "NONE"
	MAX_REST_PAYLOAD = 1024 * 1024
)

type resp struct {
	Result          string `json:"Result"`
	RespCode        string `json:"RespCode"`
	RespDescription string `json:"RespDescription"`
}

type getuseroutput struct {
	resp
	user common.User `json: "user"`
}

func validateUser(user common.User) error {
	if user.Name == "" {
		return errors.New("USER_NAME_EMPTY")
	} else if user.Email == "" {
		return errors.New("USER_EMAIL_EMPTY")
	} else if user.Phone == "" {
		return errors.New("USER_PHONE_EMPTy")
	} else if user.Address == "" {
		return errors.New("USER_ADDRESS_EMPTY")
	}
	return nil
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	var e error
	var usr common.User
	var out resp
	if e = getInput(w, r, &usr); e != nil {
		updateOutAndSendResp(w, r, FAILURE, "INVALID_INPUT", e.Error(), 422, &out)
		return
	}
	e = validateUser(usr)
	if e != nil {
		updateOutAndSendResp(w, r, FAILURE, "INVALID_INPUT", e.Error(), 422, &out)
		return
	}
	usr.ID = common.SecureRandomAlphaString(16)
	e = db.CreateUser(usr)
	if e != nil {
		updateOutAndSendResp(w, r, FAILURE, "INTERNAL_ERROR", e.Error(), 500, &out)
		return
	}
	updateOutAndSendResp(w, r, SUCCESS, NONE, NONE, 200, &out)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var e error
	var usr common.User
	var out resp
	if e = getInput(w, r, &usr); e != nil {
		updateOutAndSendResp(w, r, FAILURE, "INVALID_INPUT", e.Error(), 422, &out)
		return
	}
	if usr.ID == "" {
		updateOutAndSendResp(w, r, FAILURE, "INVALID_INPUT", "USER_ID_EMPTY", 422, &out)
	}
	e = validateUser(usr)
	if e != nil {
		updateOutAndSendResp(w, r, FAILURE, "INVALID_INPUT", e.Error(), 422, &out)
		return
	}
	e = db.UpdateUser(usr)
	if e != nil {
		updateOutAndSendResp(w, r, FAILURE, "INTERNAL_ERROR", e.Error(), 500, &out)
		return
	}
	updateOutAndSendResp(w, r, SUCCESS, NONE, NONE, 200, &out)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	var out getuseroutput
	usr, e := db.GetUser(id)
	if e != nil {
		updateOutAndSendResp(w, r, FAILURE, "INVALID_INPUT", e.Error(), 422, &out)
		return
	}
	out.user = *usr
	updateOutAndSendResp(w, r, SUCCESS, NONE, NONE, 200, &out)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	var out resp
	if e := db.DeleteUser(id); e != nil {
		updateOutAndSendResp(w, r, FAILURE, "INVALID_INPUT", e.Error(), 422, &out)
		return
	}
	updateOutAndSendResp(w, r, SUCCESS, NONE, NONE, 200, &out)
}
