package types

import (
	"context"

	"cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NftKeeper defines the expected nft keeper
type NftKeeper interface {
	// Methods imported from nft should be defined here
	Update(ctx context.Context, token nft.NFT) error

	// BatchUpdate defines a method for updating a batch of exist nfts
	// Note: When the upper module uses this method, it needs to authenticate nft
	BatchUpdate(ctx context.Context, tokens []nft.NFT) error

	// GetOwner returns the owner of the specified NFT
	GetOwner(ctx context.Context, classID, nftID string) sdk.AccAddress

	// Mint mints a new NFT
	Mint(ctx context.Context, token nft.NFT, receiver sdk.AccAddress) error

	// BatchMint defines a method for minting a batch of nfts
	BatchMint(ctx context.Context, tokens []nft.NFT, receiver sdk.AccAddress) error

	// Burn burns a NFT
	Transfer(ctx context.Context, classID string, nftID string, receiver sdk.AccAddress) error

	// BatchTransfer defines a method for sending a batch of nfts from one account to another account from a specific classID.
	// Note: When the upper module uses this method, it needs to authenticate nft
	BatchTransfer(ctx context.Context,
		classID string,
		nftIDs []string,
		receiver sdk.AccAddress,
	) error

	// GetNFTs returns all NFTs
	GetNFTsOfClassByOwner(ctx context.Context, classID string, owner sdk.AccAddress) (nfts []nft.NFT)

	// GetNFT returns the specified NFT
	GetNFT(ctx context.Context, classID, nftID string) (nft.NFT, bool)

	// GetClass returns the specified class
	SaveClass(ctx context.Context, class nft.Class) error

	// GetClass returns the specified class
	GetClasses(ctx context.Context) (classes []*nft.Class)

	// HasClass returns whether the specified class exists
	HasClass(ctx context.Context, classID string) bool
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}
