package api

import (
	"net/http"
)

func HandleAndRoute() {
	http.HandleFunc("/checkAttributes", CheckAttributes)
}
