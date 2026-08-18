package handler

import "net/http"

// HandleUpload is a stub for a future endpoint to directly upload files
func (a *API) HandleUpload(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement direct file upload endpoint
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error": "not implemented"}`))
}
