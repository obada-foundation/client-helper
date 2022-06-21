package api

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/obada-foundation/client-helper/rest"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/system/auth"
	"go.uber.org/zap"
)

type accounts struct {
	accountSvc *account.Service
	logger     *zap.SugaredLogger
}

func (ra *accounts) create(w http.ResponseWriter, r *http.Request) {
	var newAccount account.NewAccount

	claims, err := auth.GetClaims(r.Context())
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, ra.logger)
		return
	}

	if err := render.DecodeJSON(http.MaxBytesReader(w, r.Body, hardBodyLimit), &newAccount); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't decode request data", rest.ErrDecode, ra.logger)
		return
	}

	newAccount.ID = claims.UserID

	account, err := ra.accountSvc.Create(newAccount)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, ra.logger)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, &account)
}

func (ra *accounts) balance(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.GetClaims(r.Context())
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, ra.logger)
		return
	}

	balance, err := ra.accountSvc.Balance(claims.UserID)
	if err != nil {
		switch err {
		case account.ErrAccountNotExists:
			rest.SendErrorJSON(w, r, http.StatusNotFound, err, "", rest.ErrDecode, ra.logger)
		default:
			rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, ra.logger)
		}
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &balance)
}
