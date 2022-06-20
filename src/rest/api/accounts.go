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

func (ra *accounts) createAccount(w http.ResponseWriter, r *http.Request) {
	var newAccount account.NewAccount
	var account account.Account

	if err := render.DecodeJSON(http.MaxBytesReader(w, r.Body, hardBodyLimit), &newAccount); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't decode request data", rest.ErrDecode)
		return
	}

	account, err := ra.accountSvc.Create(newAccount)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, &account)
}

func (ra *accounts) myAccount(w http.ResponseWriter, r *http.Request) {
	var account account.Account

	claims, err := auth.GetClaims(r.Context())
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, nil, "", rest.ErrDecode)
		return
	}

	account, err = ra.accountSvc.Find(claims.Id)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, &account)
}
