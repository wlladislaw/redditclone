package pkg

import "net/http"

func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./template/index.html")
}
