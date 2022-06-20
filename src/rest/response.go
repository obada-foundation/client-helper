package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// JSON key/value
type JSON map[string]interface{}

// RenderJSON sends data as json
func RenderJSON(w http.ResponseWriter, data interface{}) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}
