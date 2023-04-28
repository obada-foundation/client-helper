package accounts

import (
	"context"
	"fmt"
	"net/http"

	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/obada-foundation/client-helper/services/blockchain"
	"github.com/obada-foundation/client-helper/system/validate"
	"github.com/obada-foundation/client-helper/system/web"
)

// Handlers contains all needed dependencies
type Handlers struct {
	AccountSvc    *account.Service
	BlockchainSvc *blockchain.Service
}

// Account returns a single account
func (h Handlers) Account(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	address := web.Param(r, "address")

	acc, err := h.AccountSvc.GetProfileAccount(ctx, address)
	if err != nil {
		return fmt.Errorf("unable to fetch profile account by address %s: %w", address, err)
	}

	return web.Respond(ctx, w, acc, http.StatusOK)
}

// Accounts returns a list of profile accounts
func (h Handlers) Accounts(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	profileAccounts, err := h.AccountSvc.GetProfileAccounts(ctx)
	if err != nil {
		return fmt.Errorf("unable to fetch profile accounts: %w", err)
	}

	return web.Respond(ctx, w, profileAccounts, http.StatusOK)
}

// Register a new user profile
func (h Handlers) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var newProfile services.NewProfile

	if err := web.Decode(r, &newProfile); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	newProfile.ID = auth.GetUserID(ctx)

	profile, err := h.AccountSvc.RegisterProfile(ctx, newProfile)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, profile, http.StatusCreated)
}

// NewWalletRequest request body for creating a new wallet
type NewWalletRequest struct {
	Mnemonic string `json:"mnemonic"`
	Force    bool   `json:"force"`
}

// NewWallet creates a new HD wallet
func (h Handlers) NewWallet(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req NewWalletRequest

	if err := web.Decode(r, &req); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	if req.Mnemonic == "" {
		return validate.FieldErrors{
			validate.FieldError{
				Field: "mnemonic",
				Error: "mnemonic is required",
			},
		}
	}

	if _, err := h.AccountSvc.NewWallet(ctx, req.Mnemonic, req.Force); err != nil {
		return err
	}

	return web.RespondWithNoContent(ctx, w, http.StatusCreated)
}

// AccountRequest request body for creating a new account
type AccountRequest struct {
	AccountName string `json:"account_name"`
}

// NewAccount creates a new OBADA account from HD wallet
func (h Handlers) NewAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req AccountRequest

	if err := web.Decode(r, &req); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	acc := account.Account{
		Name: req.AccountName,
	}

	if _, err := h.AccountSvc.NewAccount(ctx, acc); err != nil {
		return err
	}

	return web.RespondWithNoContent(ctx, w, http.StatusCreated)
}

// UpdateAccount updates an account
func (h Handlers) UpdateAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req AccountRequest

	address := web.Param(r, "address")

	if err := web.Decode(r, &req); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	if err := h.AccountSvc.UpdateAccountName(ctx, address, req.AccountName); err != nil {
		return err
	}

	return web.RespondWithNoContent(ctx, w, http.StatusNoContent)
}

// DeleteAccount deletes an exported account
func (h Handlers) DeleteAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	address := web.Param(r, "address")

	if err := h.AccountSvc.DeleteAccount(ctx, address); err != nil {
		return err
	}

	return web.RespondWithNoContent(ctx, w, http.StatusNoContent)
}

// ImportWalletRequest request body for importing an existing mnemonic phrase
type ImportWalletRequest struct {
	Mnemonic string `json:"mnemonic"`
	Force    bool   `json:"force"`
}

// ImportWallet imports an existing mnemonic phrase
func (h Handlers) ImportWallet(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req ImportWalletRequest

	if err := web.Decode(r, &req); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	if req.Mnemonic == "" {
		return validate.FieldErrors{
			validate.FieldError{
				Field: "mnemonic",
				Error: "mnemonic is required",
			},
		}
	}

	if err := h.AccountSvc.ImportWallet(ctx, req.Mnemonic, req.Force); err != nil {
		return err
	}

	return web.RespondWithNoContent(ctx, w, http.StatusCreated)
}

// ImportAccountRequest request body for importing an existing account
type ImportAccountRequest struct {
	PrivateKey  string `json:"private_key"`
	Passphrase  string `json:"passphrase"`
	AccountName string `json:"account_name"`
}

// ImportAccount imports an existing private key
func (h Handlers) ImportAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req ImportAccountRequest

	if err := web.Decode(r, &req); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	if req.PrivateKey == "" {
		return validate.FieldErrors{
			validate.FieldError{
				Field: "private_key",
				Error: "private key is required",
			},
		}
	}

	acc := account.Account{
		Name: req.AccountName,
	}

	if err := h.AccountSvc.ImportAccount(ctx, req.PrivateKey, req.Passphrase, acc); err != nil {
		return err
	}

	return web.RespondWithNoContent(ctx, w, http.StatusCreated)
}

// ExportAccountRequest request body for exporting an account
type ExportAccountRequest struct {
	Address    string `json:"address"`
	Passphrase string `json:"passphrase"`
}

// ExportAccountResponse response body for exporting an account
type ExportAccountResponse struct {
	PrivateKey string `json:"private_key"`
}

// ExportAccount exports armored private key
func (h Handlers) ExportAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req ExportAccountRequest

	if err := web.Decode(r, &req); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	exportedAccount, err := h.AccountSvc.ExportAccount(ctx, req.Address, req.Passphrase)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, &ExportAccountResponse{PrivateKey: exportedAccount}, http.StatusOK)
}

// GenerateMnemonicResponse response body for generating a new mnemonic phrase
type GenerateMnemonicResponse struct {
	Mnemonic string `json:"mnemonic"`
}

// NewMnemonic generates a new mnemonic phrase
func (h Handlers) NewMnemonic(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	mnemonic, err := account.NewMnemonic()
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, &GenerateMnemonicResponse{Mnemonic: mnemonic}, http.StatusOK)
}

// Mnemonic shows am existing mnemonic phrase
func (h Handlers) Mnemonic(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	wallet, err := h.AccountSvc.GetWallet(ctx)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, &GenerateMnemonicResponse{Mnemonic: wallet.Mnemonic}, http.StatusOK)
}

// SendCoinsRequest requestion body for seding OBD
type SendCoinsRequest struct {
	RecipientAddress string `json:"recipient_address"`
	Amount           string `json:"amount"`
	Denom            string `json:"denom"`
}

// SendCoins sends coins to a recipient address
func (h Handlers) SendCoins(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	address := web.Param(r, "address")

	var req SendCoinsRequest

	if err := web.Decode(r, &req); err != nil {
		return fmt.Errorf("unable to decode request data: %w", err)
	}

	acc, err := h.AccountSvc.GetProfileAccount(ctx, address)
	if err != nil {
		return err
	}

	privKey, err := h.AccountSvc.GetAccountPrivateKey(ctx, address)
	if err != nil {
		return err
	}

	amount := fmt.Sprintf("%s%s", req.Amount, req.Denom)

	if err := h.BlockchainSvc.Send(ctx, acc, req.RecipientAddress, amount, privKey); err != nil {
		return err
	}

	return web.RespondWithNoContent(ctx, w, http.StatusCreated)
}
