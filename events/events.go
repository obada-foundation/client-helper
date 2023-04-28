package events

const (
	// AccountDeleted is the event name for when an account is deleted
	AccountDeleted = "account.deleted"

	// AccountCreated is the event name for when an account is created
	AccountCreated = "account.created"

	// DeviceSaved is the event name for when a device is saved
	DeviceSaved = "device.saved"

	// NftTransfered is the event name for when an nft is transferred
	NftTransfered = "nft.transfered"

	// NftMinted is the event name for when an nft is minted
	NftMinted = "nft.minted"

	// NftMetadataUpdated is the event name for when an nft metadata is updated
	NftMetadataUpdated = "nft.metadata.updated"

	// AccountDeletedHandler is the handler for the account deleted event
	AccountDeletedHandler = "handlers:" + AccountDeleted

	// AccountCreatedHandler is the handler for the account created event
	AccountCreatedHandler = "handlers:" + AccountCreated

	// DeviceSavedHandler is the event handler for when a device is saved
	DeviceSavedHandler = "handlers:" + DeviceSaved

	// NftTransferedHandler is the event handler for when an nft is transferred
	NftTransferedHandler = "handlers:" + NftTransfered

	// NftMintedHandler is the event handler for when an nft is minted
	NftMintedHandler = "handlers:" + NftMinted

	// NftMetadataUpdatedHandler is the event handler for when an nft metadata is updated
	NftMetadataUpdatedHandler = "handlers:" + NftMetadataUpdated
)
