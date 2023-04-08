package webhandler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ok(w http.ResponseWriter, object interface{}) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	res, err := convJson(object)
	if err != nil {
		badRequest(w, err)
		return
	}
	w.Write([]byte(*res))
}

func badRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = err.Error()
	res, _ := convJson(resp) // if this fails we are doomed
	w.Write([]byte(*res))
}

func convJson(object interface{}) (*string, error) {
	b, err := json.Marshal(object)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return nil, err
	}
	ret := string(b)

	return &ret, nil
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func getArg(r *http.Request, param string) string {
	return r.URL.Query().Get(param)
}

func missingParam(w http.ResponseWriter, param string) {
	badRequest(w, fmt.Errorf("missing parameter: %s", param))
}
