package handlers

import (
	"encoding/json"
	"net/http"
)

// Response is the JSON response format
type Response struct {
	Error bool `json:"error"`
	Message string `json:"message"`
	Data interface{} `json:"data"`
}

// JSON format and output a json response
func JSON(w http.ResponseWriter, code int, r *Response) {

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(r)
	if err != nil {
		w.WriteHeader(500)
		errResponse := &Response{
			Error: true, Message: "Something went wrong!",
		}
		bytes, _ := json.Marshal(errResponse)
		w.Write(bytes)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}

// UnAuthorizedResponse responds to UnAuthorized requests
// with http code 403 UnAuthorized
func UnAuthorizedResponse(w http.ResponseWriter) {
	JSON(w, 403, &Response{Error: true, Message: "Unauthorized request"})
}

// BadRequestResponse responds to unexpected errors
// when processing requests
func BadRequestResponse(w http.ResponseWriter)  {
	JSON(w, 400, &Response{Error: true, Message: "bad request"})
}

// ParseAuthorizationKey gets authentication data from
// a request.
// It checks request's header, query parameters and cookie data
func ParseAuthorizationKey(r *http.Request) string {
	key := r.Header.Get("X-Blober-ID")
	if key != "" {
		return key
	}

	query := r.URL.Query()
	key = query.Get("bloberId")
	if key != "" {
		return key
	}

	cookie, err := r.Cookie("X-Blober-ID")
	if err != nil {
		return ""
	}

	key = cookie.Value
	if key != "" {
		return key
	}

	return ""
}

type NotFoundHandler struct {}
type MethodNotAllowedHandler struct {}

func (*NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)  {
	JSON(w, http.StatusNotFound, &Response{Error: true, Message: "That address is not found on this server"})
}

func (*MethodNotAllowedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusMethodNotAllowed, &Response{Error: true, Message: "Method not allowed"})
}

