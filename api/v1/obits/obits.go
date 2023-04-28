package obits

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/device"
	"github.com/obada-foundation/client-helper/system/web"
)

// Handlers holds dependencies
type Handlers struct {
	DeviceSvc  *device.Service
	AccountSvc *account.Service
}

// Obit returns an obit by USN or DID
func (h Handlers) Obit(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	key := web.Param(r, "key")

	d, err := h.DeviceSvc.Get(ctx, key)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, d, http.StatusOK)
}

// Save saves an obit into local database
func (h Handlers) Save(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var saveRequest services.SaveDevice

	if err := web.Decode(r, &saveRequest); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	privKey, err := h.AccountSvc.GetAccountPrivateKey(ctx, saveRequest.Address)
	if err != nil {
		return err
	}

	d, err := h.DeviceSvc.Save(ctx, saveRequest, privKey)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, d, http.StatusOK)
}

// Search returns a list of obits by giver query
func (h Handlers) Search(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	q := web.Query(r, "q")

	devices := make([]services.Device, 0)
	var err error

	if q == "" {
		devices, err = h.DeviceSvc.GetByUser(ctx)
		if err != nil {
			return err
		}

		return web.Respond(ctx, w, devices, http.StatusOK)
	}

	if strings.Contains(q, "obada") {
		devices, err = h.DeviceSvc.GetByAddress(ctx, q)
		if err != nil {
			return err
		}
	}

	return web.Respond(ctx, w, devices, http.StatusOK)
}
