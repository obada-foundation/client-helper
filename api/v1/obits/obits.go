package obits

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"

	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/device"
	"github.com/obada-foundation/client-helper/system/web"
	"github.com/obada-foundation/registry/api/pb/v1/diddoc"
	"github.com/obada-foundation/registry/client"
)

// Handlers holds dependencies
type Handlers struct {
	DeviceSvc  *device.Service
	AccountSvc *account.Service
	Registry   client.Client
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

// BatchSave saves a batch of obits into local database
func (h Handlers) BatchSave(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var batchSaveRequest services.BatchSaveDevice

	if err := web.Decode(r, &batchSaveRequest); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	numCPU := runtime.NumCPU()
	errs := make(chan error, numCPU)
	results := make(chan services.Device, numCPU)

	privKey, err := h.AccountSvc.GetAccountPrivateKey(ctx, batchSaveRequest.Address)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i, saveRequest := range batchSaveRequest.Obits {
		wg.Add(1)
		go func(i int, saveRequest services.SaveDevice) {
			defer wg.Done()
			saveRequest.Address = batchSaveRequest.Address
			d, err := h.DeviceSvc.Save(ctx, saveRequest, privKey)
			if err != nil {
				errs <- err
				return
			}
			results <- d
		}(i, saveRequest)
	}

	go func() {
		wg.Wait()
		close(errs)
		close(results)
	}()

	devices := make([]services.Device, 0, len(batchSaveRequest.Obits))
	for d := range results {
		devices = append(devices, d)
	}

	// Check if any errors occurred during the parallel execution
	for err := range errs {
		if err != nil {
			return err
		}
	}

	return web.Respond(ctx, w, devices, http.StatusOK)
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

// History returns Obit history of changes
func (h Handlers) History(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	key := web.Param(r, "key")

	d, err := h.DeviceSvc.GetByUSN(ctx, key)
	if err != nil {
		return err
	}

	msg := &diddoc.GetMetadataHistoryRequest{
		Did: d.DID,
	}

	resp, err := h.Registry.GetMetadataHistory(ctx, msg)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, resp.GetMetadataHistory(), http.StatusOK)
}
