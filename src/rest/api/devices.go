package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/obada-foundation/client-helper/rest"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/device"
	"github.com/obada-foundation/client-helper/system/auth"
	"go.uber.org/zap"

	"net/http"
)

type deviceGroup struct {
	deviceSvc  *device.Service
	accountSvc *account.Account
	logger     *zap.SugaredLogger
}

func (dgroup *deviceGroup) save(w http.ResponseWriter, r *http.Request) {
	var saveRequest device.SaveDevice

	claims, err := auth.GetClaims(r.Context())
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, dgroup.logger)
		return
	}

	if err := render.DecodeJSON(http.MaxBytesReader(w, r.Body, hardBodyLimit), &saveRequest); err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't decode request data", rest.ErrDecode, dgroup.logger)
		return
	}

	device, err := dgroup.deviceSvc.Save(r.Context(), claims.UserID, saveRequest)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "can't save device", rest.ErrDecode, dgroup.logger)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &device)
}

func (dgroup *deviceGroup) get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	ctx := r.Context()

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, ErrBadRequest, "", rest.ErrDecode, dgroup.logger)
		return
	}

	device, err := dgroup.deviceSvc.Get(ctx, claims.UserID, key)

	if err != nil {
		rest.SendErrorJSON(w, r, http.StatusBadRequest, err, "", rest.ErrDecode, dgroup.logger)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, &device)
}
