package types

import (
	"fmt"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		TaList: []Ta{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated ID in ta
	taIdMap := make(map[uint64]bool)
	taCount := gs.GetTaCount()
	for _, elem := range gs.TaList {
		if _, ok := taIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for ta")
		}
		if elem.Id >= taCount {
			return fmt.Errorf("ta id should be lower or equal than the last id")
		}
		taIdMap[elem.Id] = true
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
