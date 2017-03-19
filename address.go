package ie

import "github.com/intervention-engine/fhir/models"

type RestructedAddress struct{}

func (a *RestructedAddress) FromFHIR(address *models.Address) *RestructedAddress {
	return a
}
