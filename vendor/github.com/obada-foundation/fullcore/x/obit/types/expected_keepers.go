package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	nft "github.com/cosmos/cosmos-sdk/x/nft"
)

// NftKeeper expected nft keeper
type NftKeeper interface {
	Update(ctx sdk.Context, token nft.NFT) error

	GetOwner(ctx sdk.Context, classID, nftID string) sdk.AccAddress

	Mint(ctx sdk.Context, token nft.NFT, receiver sdk.AccAddress) error

	Transfer(ctx sdk.Context, classID string, nftID string, receiver sdk.AccAddress) error

	GetNFTsOfClassByOwner(ctx sdk.Context, classID string, owner sdk.AccAddress) (nfts []nft.NFT)

	GetNFT(ctx sdk.Context, classID, nftID string) (nft.NFT, bool)

	SaveClass(ctx sdk.Context, class nft.Class) error

	GetClasses(ctx sdk.Context) (classes []*nft.Class)

	HasClass(ctx sdk.Context, classID string) bool
	// Methods imported from nft should be defined here
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}
