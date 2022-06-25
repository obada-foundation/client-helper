package api

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/obada-foundation/client-helper/rest"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/system/auth"
	"go.uber.org/zap"
)

type accountGroup struct {
	accountSvc *account.Service
	logger     *zap.SugaredLogger
}

func (agrp *accountGroup) create(w http.ResponseWriter, r *http.Request) {
	var newAccount account.NewAccount

	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, agrp.logger)
		return
	}

	if err := render.DecodeJSON(http.MaxBytesReader(w, r.Body, hardBodyLimit), &newAccount); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't decode request data", rest.ErrDecode, agrp.logger)
		return
	}

	newAccount.ID = userID

	account, err := agrp.accountSvc.Create(r.Context(), newAccount)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, agrp.logger)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, &account)
}

func (agrp *accountGroup) balance(w http.ResponseWriter, r *http.Request) {
	balance, err := agrp.accountSvc.Balance(r.Context())
	if err != nil {
		switch err {
		case account.ErrAccountNotExists:
			rest.SendErrorJSON(w, r, http.StatusNotFound, err, "", rest.ErrDecode, agrp.logger)
		default:
			rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, agrp.logger)
		}
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &balance)
}
