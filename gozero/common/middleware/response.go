package middleware

import (
	"encoding/json"
	"net/http"
)

const (
	CodeOK           = 200
	CodeUnauthorized = 401
	CodeServerError  = 500
)

type response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, code int, msg string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}
