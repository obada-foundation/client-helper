package device

import (
	"github.com/obada-foundation/client-helper/services"
)

// DeviceSaved is the event that is fired when a device is saved
type DeviceSaved struct { //nolint:revive // need refactroing
	Device    services.Device `json:"device"`
	ProfileID string          `json:"profile_id"`
}
