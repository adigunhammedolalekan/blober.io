package handlers

import (
	"encoding/json"
	"fmt"
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

func UnAuthorizedResponse(w http.ResponseWriter) {
	JSON(w, 403, &Response{Error: true, Message: "Unauthorized request"})
}

func ParseAuthorizationKey(r *http.Request) string {
	key := r.Header.Get("X-Blober-ID")
	if key != "" {
		return key
	}

	query := r.URL.Query()
	fmt.Println(query)
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

func BadRequestResponse(w http.ResponseWriter)  {
	JSON(w, 400, &Response{Error: true, Message: "bad request"})
}

type NotFoundHandler struct {}

type MethodNotAllowedHandler struct {}

func (*NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)  {
	JSON(w, http.StatusNotFound, &Response{Error: true, Message: "That address is not found on this server"})
}

func (*MethodNotAllowedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusMethodNotAllowed, &Response{Error: true, Message: "Method not allowed"})
}

