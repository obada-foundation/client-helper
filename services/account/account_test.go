// nolint:all
package account_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/google/uuid"
	"github.com/mustafaturan/bus/v3"
	"github.com/obada-foundation/client-helper/auth"
	"github.com/obada-foundation/client-helper/events"
	"github.com/obada-foundation/client-helper/services"
	"github.com/obada-foundation/client-helper/services/account"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nolint:misspell //ignore misspelling
const defaultMnemonic string = "radio distance sweet artefact attack liar until video army raccoon green error ceiling size spread burst galaxy bottom cave rubber setup west address must"

const defaultAddress = "obada1yxxnd624tgwqm3eyv5smdvjrrydfh9h943qptg"

const defaultPubKey string = "020A8D4D4F9920F80F70C1CDE78926929CD1254038B9AF79984CAC65FCE620AB36"

const secondMnemonic string = "more team rail confirm alley design random bridge drill abuse power airport hero ridge lake carry error never swear panther napkin icon army banana"

const secondAddress string = "obada1e520k92tleyft5rep76m5pa9h6etqjvrfac8he"

const defaultObadaPrivateKey string = `
-----BEGIN TENDERMINT PRIVATE KEY-----
type: secp256k1
kdf: bcrypt
salt: 85D3E663D8C4024F67CA5A48C20177F0
        
Bvp92n5ChMmbgutjU1eH17x2fAD2i0Gw4ie2bRwkrJR/zy7UZezgpT0OqwoEJJQk
p1OrbxmEKrKBl16PpWUGWZ5DnD3OkgqzxmVbCf0=
=ekQv
-----END TENDERMINT PRIVATE KEY-----
`

const secondPrivateKey string = `
-----BEGIN TENDERMINT PRIVATE KEY-----
salt: 838A66523BEE7B3C40D6B245EA83E418
type: secp256k1
kdf: bcrypt

+KI+TpyXkU6Mh1+p9VkD33WRANRyDM7c+kuXRdfz43xapG9dl6KsLl+ye9kGTCHM
GQ6Zf8VMpvoLRz5o4VI8FBXPk1Sk96LMMpGCvPM=
=LgD4
-----END TENDERMINT PRIVATE KEY-----
`

func makeProfile(t *testing.T, svc *account.Service, ctx context.Context) (services.Profile, error) {
	profileID := auth.GetUserID(ctx)

	np := services.NewProfile{
		ID:    profileID,
		Email: "jon.doe@supermail.com",
	}

	return svc.RegisterProfile(ctx, np)
}

func TestService(t *testing.T) {
	_, service, ctx, deferFn := createTestService(t)
	defer deferFn()

	ctx = auth.SetClaims(ctx, auth.Claims{
		UserID: "1",
	})

	np := services.NewProfile{
		ID:    uuid.New().String(),
		Email: "jon.doe@supermail.com",
	}

	t.Log("Test user profile creation")
	createdProfile, err := service.RegisterProfile(ctx, np)
	require.NoError(t, err, "Cannot register new account")
	assert.Equal(t, np.ID, createdProfile.ID)
	assert.Equal(t, np.Email, createdProfile.Email)

	t.Log("Test that user profile won't be created if already exists")
	{
		_, er := service.RegisterProfile(ctx, np)
		require.ErrorIs(t, er, account.ErrProfileExists)
	}

	t.Log("Test wallet creation fails with invalid mnemonic")
	{
		_, er := service.NewWallet(ctx, "invalid seed", false)
		require.ErrorIs(t, er, account.ErrInvalidMnemonic)
	}

	t.Log("Test user profile HD wallet creation")
	{
		_, er := service.NewWallet(ctx, defaultMnemonic, false)
		require.NoError(t, er, "Cannot create user profile HD wallet")
	}

	t.Log("Test fetching a wallet info")
	{
		wallet, er := service.GetWallet(ctx)
		require.NoError(t, er, "Cannot fetch wallet info")

		assert.Equal(t, defaultMnemonic, wallet.Mnemonic)
		assert.Equal(t, uint(0), wallet.AccountIndex)
	}

	t.Log("Test that user profile wallet won't be created if already exists")
	{
		_, er := service.NewWallet(ctx, defaultMnemonic, false)
		require.ErrorIs(t, er, account.ErrWalletExists)
	}

	t.Log("Test find user profile by context")
	{
		p, er := service.GetProfile(ctx)
		require.NoError(t, er, "Cannot find a profile that was previostly created")
		assert.Equal(t, createdProfile, p)
	}

	t.Log("Test fetching wallet account index with one provate key created")
	{
		walletAccountIndex, er := service.GetWalletAccountIndex(ctx)
		require.NoError(t, er)
		assert.Equal(t, uint(0), walletAccountIndex)
	}

	t.Log("Test fetching profile accounts (OBADA addresses)")
	{
		profileAccounts, er := service.GetProfileAccounts(ctx)
		require.NoError(t, er)

		hdAccounts := profileAccounts.HDAccounts

		assert.Equal(t, 1, len(hdAccounts))
		assert.Equal(t, 0, len(profileAccounts.ImportedAccounts))

		assert.Equal(t, defaultAddress, hdAccounts[0].Address)
		assert.NotEmpty(t, hdAccounts[0].Balance)
		assert.Equal(t, uint(0), hdAccounts[0].NFTsCount)
		assert.Equal(t, defaultPubKey, hdAccounts[0].PublicKey)
	}

	t.Log("Test that adding new OBADA account creation will fail if there account with no transactions")
	{
		_, err = service.NewAccount(ctx, account.Account{})
		require.ErrorIs(t, err, account.ErrAccountHasZeroTx)

		profileAccounts, er := service.GetProfileAccounts(ctx)
		require.NoError(t, er)

		assert.Equal(t, 1, len(profileAccounts.HDAccounts))
		assert.Equal(t, 0, len(profileAccounts.ImportedAccounts))
	}

	t.Log("Test it will not import HD wallet if it already exists")
	{
		err = service.ImportWallet(ctx, defaultMnemonic, false)
		require.ErrorIs(t, err, account.ErrWalletExists)
	}
}

func TestService_NewWallet(t *testing.T) {
	_, service, ctx, deferFn := createTestService(t)
	defer deferFn()

	t.Log("create wallet from first user")
	{
		_, err := service.NewWallet(ctx, defaultMnemonic, true)
		require.NoError(t, err)

		_, err = service.GetProfileAccounts(ctx)
		require.NoError(t, err)
	}

	t.Log("create wallet from second user")
	{
		secondCtx := auth.SetClaims(context.Background(), auth.Claims{
			UserID: "4",
		})

		_, err := makeProfile(t, service, secondCtx)
		require.NoError(t, err)

		_, err = service.NewWallet(secondCtx, defaultMnemonic, true)
		require.ErrorIs(t, err, account.ErrAccountExists)

		_, err = service.NewWallet(secondCtx, secondMnemonic, true)
		require.NoError(t, err)

		accounts, err := service.GetProfileAccounts(secondCtx)
		require.NoError(t, err)

		assert.Equal(t, 1, len(accounts.HDAccounts))
	}
}

func TestService_ExportWallet(t *testing.T) {
	_, svc, ctx, deferFn := createTestService(t)
	defer deferFn()

	_, err := svc.NewWallet(ctx, defaultMnemonic, false)
	require.NoError(t, err, "Cannot create user profile HD wallet")

	t.Log("Test account export")
	{
		exportedAccount, err := svc.ExportAccount(ctx, defaultAddress, "")
		require.NoError(t, err)

		assert.True(t, strings.Contains(exportedAccount, "BEGIN TENDERMINT PRIVATE KEY"))
	}
}

func TestService_ImportWallet(t *testing.T) {
	eventbus, svc, ctx, deferFn := createTestService(t)
	defer deferFn()

	t.Log("Test import fails with invalid mnemonic")
	{
		err := svc.ImportWallet(ctx, "invalid seed", false)
		require.ErrorIs(t, err, account.ErrInvalidMnemonic)
	}

	t.Log("Test HD wallet import")
	{
		err := svc.ImportWallet(ctx, defaultMnemonic, false)
		require.NoError(t, err)
	}

	t.Log("Test that HD wallet cannot be imported if already exists")
	{
		err := svc.ImportWallet(ctx, secondMnemonic, false)
		require.ErrorIs(t, err, account.ErrWalletExists)
	}

	t.Log("Test that HD wallet import with force flag and events fired")
	{
		accountDeletedfn := func(ctx context.Context, e bus.Event) {
			t.Run(fmt.Sprintf("receives %q event", events.AccountDeleted), func(t *testing.T) {
				assert := assert.New(t)
				assert.Equal("afakeid", e.ID)
				assert.Equal(events.AccountDeleted, e.Topic)
				assert.Equal(defaultAddress, e.Data)
				assert.True(e.OccurredAt.Before(time.Now()))
			})
		}

		eventbus.RegisterHandler(
			events.AccountDeletedHandler,
			bus.Handler{Handle: accountDeletedfn, Matcher: events.AccountDeleted},
		)

		accountCreatedfn := func(ctx context.Context, e bus.Event) {
			t.Run(fmt.Sprintf("receives %q event", events.AccountCreated), func(t *testing.T) {
				assert := assert.New(t)
				assert.Equal("afakeid", e.ID)
				assert.Equal(events.AccountCreated, e.Topic)
				assert.Equal(defaultAddress, e.Data)
				assert.True(e.OccurredAt.Before(time.Now()))
			})
		}

		eventbus.RegisterHandler(
			events.AccountCreatedHandler,
			bus.Handler{Handle: accountCreatedfn, Matcher: events.AccountCreated},
		)

		err := svc.ImportWallet(ctx, defaultMnemonic, true)
		require.NoError(t, err)

		eventbus.DeregisterHandler(events.AccountDeletedHandler)
		eventbus.DeregisterHandler(events.AccountCreatedHandler)
	}
}

func TestService_ImportAccount(t *testing.T) {
	eventbus, svc, ctx, deferFn := createTestService(t)
	defer deferFn()

	t.Log("Test account import")
	{
		profileAccounts, err := svc.GetProfileAccounts(ctx)
		require.NoError(t, err)

		assert.Equal(t, 0, len(profileAccounts.ImportedAccounts))
		assert.Equal(t, 0, len(profileAccounts.HDAccounts))

		fn := func(ctx context.Context, e bus.Event) {
			t.Run(fmt.Sprintf("receives %q event", events.AccountCreated), func(t *testing.T) {
				assert := assert.New(t)
				assert.Equal("afakeid", e.ID)
				assert.Equal(events.AccountCreated, e.Topic)
				assert.Equal(defaultAddress, e.Data)
				assert.True(e.OccurredAt.Before(time.Now()))
			})
		}

		eventbus.RegisterHandler(
			events.AccountCreatedHandler,
			bus.Handler{Handle: fn, Matcher: events.AccountCreated},
		)

		err = svc.ImportAccount(ctx, defaultObadaPrivateKey, "", account.Account{
			Name: "test",
		})
		require.NoError(t, err)

		profileAccounts, err = svc.GetProfileAccounts(ctx)
		require.NoError(t, err)

		assert.Equal(t, 1, len(profileAccounts.ImportedAccounts))
		assert.Equal(t, 0, len(profileAccounts.HDAccounts))

		acc := profileAccounts.ImportedAccounts[0]
		assert.Equal(t, "test", acc.Name)
		assert.Equal(t, defaultAddress, acc.Address)

		eventbus.DeregisterHandler(events.AccountCreatedHandler)
	}

	t.Log("Test that HD accounts and imported accounts are combined")
	{
		err := svc.ImportWallet(ctx, secondMnemonic, false)
		require.NoError(t, err)

		profileAccounts, err := svc.GetProfileAccounts(ctx)
		require.NoError(t, err)

		assert.Equal(t, 1, len(profileAccounts.ImportedAccounts))
		assert.Equal(t, 1, len(profileAccounts.HDAccounts))
	}

	t.Log("Test that HD account was not deleted")
	{
		err := svc.DeleteAccount(ctx, secondAddress)
		require.ErrorIs(t, account.ErrHDAccountDelete, err)

		profileAccounts, err := svc.GetProfileAccounts(ctx)
		require.NoError(t, err)

		assert.Equal(t, 1, len(profileAccounts.ImportedAccounts))
		assert.Equal(t, 1, len(profileAccounts.HDAccounts))
	}

	t.Log("Test that imported account was deleted")
	{
		fn := func(ctx context.Context, e bus.Event) {
			t.Run(fmt.Sprintf("receives %q event", events.AccountDeleted), func(t *testing.T) {
				assert := assert.New(t)
				assert.Equal("afakeid", e.ID)
				assert.Equal(events.AccountDeleted, e.Topic)
				assert.Equal(defaultAddress, e.Data)
				assert.True(e.OccurredAt.Before(time.Now()))
			})
		}

		eventbus.RegisterHandler(
			events.AccountDeletedHandler,
			bus.Handler{Handle: fn, Matcher: events.AccountDeleted},
		)

		err := svc.DeleteAccount(ctx, defaultAddress)
		require.NoError(t, err)

		profileAccounts, err := svc.GetProfileAccounts(ctx)
		require.NoError(t, err)

		assert.Equal(t, 0, len(profileAccounts.ImportedAccounts))
		assert.Equal(t, 1, len(profileAccounts.HDAccounts))
	}
}

func TestService_GetProfileAccount(t *testing.T) {
	t.Log("Testing single account")
	_, service, ctx, deferFn := createTestService(t)
	defer deferFn()

	_, err := service.NewWallet(ctx, defaultMnemonic, false)
	require.NoError(t, err, "Cannot create user profile HD wallet")

	acc, err := service.GetProfileAccount(ctx, defaultAddress)
	require.NoError(t, err)

	assert.NotEmpty(t, acc.Balance)
	assert.Equal(t, uint(0), acc.NFTsCount)
	assert.Equal(t, "", acc.Name)
	assert.Equal(t, defaultAddress, acc.Address)
	assert.Equal(t, defaultPubKey, acc.PublicKey)

	privKey, err := service.GetAccountPrivateKey(ctx, defaultAddress)
	require.NoError(t, err)

	assert.Equal(t, defaultPubKey, fmt.Sprintf("%X", privKey.PubKey().Bytes()))
}

func TestService_GetProfileAccounts(t *testing.T) {
	_, service, ctx, deferFn := createTestService(t)
	defer deferFn()

	t.Log("Testing empty accounts")
	{
		profileAccounts, err := service.GetProfileAccounts(ctx)
		require.NoError(t, err)

		assert.Equal(t, 0, len(profileAccounts.ImportedAccounts))
		assert.Equal(t, 0, len(profileAccounts.HDAccounts))
	}
}

func TestService_ImportWallet2(t *testing.T) {
	_, service, ctx, deferFn := createTestService(t)
	defer deferFn()

	t.Log("Import wallet from first user")
	{
		accounts, err := service.GetProfileAccounts(ctx)
		require.NoError(t, err)

		assert.Equal(t, 0, len(accounts.HDAccounts))
		assert.Equal(t, 0, len(accounts.ImportedAccounts))

		err = service.ImportWallet(ctx, defaultMnemonic, true)
		require.NoError(t, err)

		accounts, err = service.GetProfileAccounts(ctx)
		require.NoError(t, err)

		assert.Equal(t, 1, len(accounts.HDAccounts))
		assert.Equal(t, 0, len(accounts.ImportedAccounts))
	}

	t.Log("Import wallet from second user")
	{
		secondCtx := auth.SetClaims(context.Background(), auth.Claims{
			UserID: "4",
		})

		_, err := makeProfile(t, service, secondCtx)
		require.NoError(t, err)

		accounts, err := service.GetProfileAccounts(secondCtx)
		require.NoError(t, err)

		assert.Equal(t, 0, len(accounts.HDAccounts))
		assert.Equal(t, 0, len(accounts.ImportedAccounts))

		err = service.ImportWallet(secondCtx, defaultMnemonic, true)
		require.ErrorIs(t, err, account.ErrAccountExists)

		accounts, err = service.GetProfileAccounts(secondCtx)
		require.NoError(t, err)

		assert.Equal(t, 0, len(accounts.HDAccounts))
		assert.Equal(t, 0, len(accounts.ImportedAccounts))

		err = service.ImportWallet(secondCtx, secondMnemonic, true)
		require.NoError(t, err)

		accounts, err = service.GetProfileAccounts(secondCtx)
		require.NoError(t, err)

		assert.Equal(t, 1, len(accounts.HDAccounts))
		assert.Equal(t, 0, len(accounts.ImportedAccounts))
	}
}

func TestService_HasAccount(t *testing.T) {
	_, service, ctx, deferFn := createTestService(t)
	defer deferFn()

	t.Log("Testing has account with empty accounts")
	{
		ok := service.HasAccount(ctx, defaultAddress)
		assert.False(t, ok)
	}

	t.Log("Testing has account with default account")
	{
		err := service.ImportWallet(ctx, defaultMnemonic, false)
		require.NoError(t, err)

		ok := service.HasAccount(ctx, defaultAddress)
		assert.True(t, ok)
	}

	t.Log("Testing has account with imported account")
	{
		err := service.ImportAccount(ctx, secondPrivateKey, "", account.Account{
			Name: "test",
		})
		require.NoError(t, err)

		ok := service.HasAccount(ctx, "obada1hdyluqekzmfnnenvqszcc5hrgy5lurmr6vnd0h")
		assert.True(t, ok)
	}

	t.Log("Testing has no access to other user accounts")
	{
		ctx = auth.SetClaims(ctx, auth.Claims{
			UserID: "4",
		})

		_, err := makeProfile(t, service, ctx)
		require.NoError(t, err)

		ok := service.HasAccount(ctx, "obada1hdyluqekzmfnnenvqszcc5hrgy5lurmr6vnd0h")
		assert.False(t, ok)

		ok = service.HasAccount(ctx, defaultAddress)
		assert.False(t, ok)
	}
}

func TestService_UpdateAccountName(t *testing.T) {
	_, service, ctx, deferFn := createTestService(t)
	defer deferFn()

	t.Log("Testing renaming not existing address")
	{
		err := service.UpdateAccountName(ctx, defaultAddress, "test")
		require.ErrorIs(t, account.ErrAccountNotExists, err)
	}

	t.Log("Testing renaming HD account")
	{
		_, err := service.NewWallet(ctx, secondMnemonic, false)
		require.NoError(t, err, "Cannot create user profile HD wallet")

		acc, err := service.GetProfileAccount(ctx, secondAddress)
		require.NoError(t, err)

		assert.Equal(t, "", acc.Name)

		err = service.UpdateAccountName(ctx, secondAddress, "test")
		require.NoError(t, err)

		acc, err = service.GetProfileAccount(ctx, secondAddress)
		require.NoError(t, err)

		assert.Equal(t, "test", acc.Name)
	}

	t.Log("Testing imported account")
	{
		err := service.ImportAccount(ctx, defaultObadaPrivateKey, "", account.Account{})
		require.NoError(t, err)

		acc, err := service.GetProfileAccount(ctx, defaultAddress)
		require.NoError(t, err)

		assert.Equal(t, "", acc.Name)

		err = service.UpdateAccountName(ctx, defaultAddress, "test")
		require.NoError(t, err)

		acc, err = service.GetProfileAccount(ctx, defaultAddress)
		require.NoError(t, err)

		assert.Equal(t, "test", acc.Name)
	}
}

func TestService_GetProfileByAddress(t *testing.T) {
	_, service, ctx, deferFn := createTestService(t)
	defer deferFn()

	t.Log("Testing getting profile ID by imported account address")
	{
		err := service.ImportAccount(ctx, defaultObadaPrivateKey, "", account.Account{})
		require.NoError(t, err)

		profileID, err := service.GetProfileByAddress(defaultAddress)
		require.NoError(t, err)

		assert.Equal(t, "3", profileID)
	}
	t.Log("Testing getting profile ID by imported second address")
	{
		err := service.ImportWallet(ctx, secondMnemonic, false)
		require.NoError(t, err)

		profileID, err := service.GetProfileByAddress(secondAddress)
		require.NoError(t, err)

		assert.Equal(t, "3", profileID)
	}
	t.Log("Testing getting profile ID by not existing address")
	{
		_, er := service.GetProfileByAddress("obada1zm2fhq2y425ua7rmprjfpr2u46kjvvsleteyvp")
		require.ErrorIs(t, sdkerrors.ErrKeyNotFound, er)

	}
}
