package hndlr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime/debug"
)

func getInput(w http.ResponseWriter, r *http.Request, input interface{}) error {
	if input == nil {
		return errors.New("PROVIDE_DATA_TYPE")
	}
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, MAX_REST_PAYLOAD))
	if err != nil {
		fmt.Println("Error reading request body", err.Error())
		return err
	}

	if err = r.Body.Close(); err != nil {
		fmt.Println("Error in closing request body", err.Error())
		return err
	}

	if err = json.Unmarshal(body, input); err != nil {
		return getUnmarshalError(err)
	}

	return nil
}

func getUnmarshalError(e error) error {
	if un, ok := e.(*json.UnmarshalTypeError); ok {
		return errors.New("Input " + un.Value + " for field " + un.Field + " is incorrect.")
	}
	return e
}

func updateOutAndSendResp(w http.ResponseWriter, r *http.Request, result, respCode, description string, headercode int, out interface{}) {
	o := reflect.ValueOf(out).Elem()
	f1 := o.FieldByName("Result")
	if f1.IsValid() {
		f1.SetString(result)
	} else {
		debug.PrintStack()
		panic("someone using this who isn't supposed to use")
	}
	f2 := o.FieldByName("RespCode")
	if f2.IsValid() {
		f2.SetString(respCode)
	} else {
		debug.PrintStack()
		panic("someone using this who isn't supposed to use")
	}
	f3 := o.FieldByName("RespDescription")
	if f3.IsValid() {
		f3.SetString(description)
	} else {
		debug.PrintStack()
		panic("someone using this who isn't supposed to use")
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(headercode)
	if err := json.NewEncoder(w).Encode(out); err != nil {
	}
}
