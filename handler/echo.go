package handler

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func Echo(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	fmt.Fprint(response, vars["echoString"])
}
