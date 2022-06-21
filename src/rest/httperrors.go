package rest

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"

	"github.com/go-chi/render"
	"go.uber.org/zap"
)

// All error codes for UI mapping and translation
const (
	ErrInternal        = 0 // any internal error
	ErrCommentNotFound = 1 // can't find comment
	ErrDecode          = 2 // failed to unmarshal incoming request
	ErrNoAccess        = 3 // rejected by auth
)

// SendErrorJSON makes {error: blah, details: blah} json body and responds with error code
func SendErrorJSON(w http.ResponseWriter, r *http.Request, httpStatusCode int, err error, details string, errCode int, logger *zap.SugaredLogger) {
	if logger != nil {
		logger.Error(err)
	}

	log.Printf("[WARN] %s", errDetailsMsg(r, httpStatusCode, err, details, errCode))
	render.Status(r, httpStatusCode)
	render.JSON(w, r, JSON{"error": err.Error(), "details": details, "code": errCode})
}

func errDetailsMsg(r *http.Request, httpStatusCode int, err error, details string, errCode int) string {
	uinfoStr := ""
	q := r.URL.String()
	if qun, e := url.QueryUnescape(q); e == nil {
		q = qun
	}

	srcFileInfo := ""
	if pc, file, line, ok := runtime.Caller(2); ok {
		fnameElems := strings.Split(file, "/")
		funcNameElems := strings.Split(runtime.FuncForPC(pc).Name(), "/")
		srcFileInfo = fmt.Sprintf("[%s:%d %s]", strings.Join(fnameElems[len(fnameElems)-3:], "/"),
			line, funcNameElems[len(funcNameElems)-1])
	}

	return fmt.Sprintf("%s - %v - %d (%d) - %s%s - %s",
		details, err, httpStatusCode, errCode, uinfoStr, q, srcFileInfo)
}
