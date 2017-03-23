package ie

import "github.com/intervention-engine/fhir/models"

type RestructedAddress struct {
	Street     []string `json:"street"`
	City       string   `json:"city"`
	State      string   `json:"state"`
	PostalCode string   `json:"postalCode"`
}

func (a *RestructedAddress) FromFHIR(address *models.Address) *RestructedAddress {
	a.Street = address.Line
	a.City = address.City
	a.State = address.State
	a.PostalCode = address.PostalCode
	return a
}
