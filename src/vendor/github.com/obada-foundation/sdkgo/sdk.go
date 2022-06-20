package sdkgo

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/obada-foundation/sdkgo/properties"
)

const obadaReleaseDateUTC = 1644864441

// Sdk OBADA SDK
type Sdk struct {
	logger   *log.Logger
	debug    bool
	validate *validator.Validate
}

// NewSdk creates a new OBADA SDK instance
func NewSdk(logger *log.Logger, debug bool) (*Sdk, error) {
	v, err := initializeValidator()

	if err != nil {
		return nil, err
	}

	return &Sdk{
		logger:   logger,
		debug:    debug,
		validate: v,
	}, nil
}

func initializeValidator() (*validator.Validate, error) {
	v := validator.New()

	if err := v.RegisterValidation("min-modified-on", validateMinModifiedOn); err != nil {
		return v, err
	}

	return v, nil
}

func validateMinModifiedOn(fl validator.FieldLevel) bool {
	return fl.Field().Int() >= obadaReleaseDateUTC
}

// NewObitID creates new obit id
func (sdk *Sdk) NewObitID(dto ObitIDDto) (properties.ObitID, error) {
	var obitID properties.ObitID

	if sdk.debug {
		sdk.logger.Printf("NewObitID(%q, %q, %q)", dto.SerialNumberHash, dto.Manufacturer, dto.PartNumber)
	}

	err := sdk.validate.Struct(dto)
	if err != nil {
		return obitID, err
	}

	snProp, err := properties.NewStringProperty(
		"Making serialNumberHash",
		dto.SerialNumberHash,
		sdk.logger,
		sdk.debug,
	)

	if err != nil {
		return obitID, err
	}

	mnProp, err := properties.NewStringProperty(
		"Making manufacturer",
		dto.Manufacturer,
		sdk.logger,
		sdk.debug,
	)

	if err != nil {
		return obitID, err
	}

	pnProp, err := properties.NewStringProperty(
		"Making partNumber",
		dto.PartNumber,
		sdk.logger,
		sdk.debug,
	)

	if err != nil {
		return obitID, err
	}

	obitID, err = properties.NewObitIDProperty(snProp, mnProp, pnProp, sdk.logger, sdk.debug)

	if err != nil {
		return obitID, err
	}

	return obitID, nil
}
