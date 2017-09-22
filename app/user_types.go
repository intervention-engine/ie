// Code generated by goagen v1.2.0-dirty, DO NOT EDIT.
//
// API "api": Application User Types
//
// Command:
// $ goagen
// --design=github.com/intervention-engine/ie/design
// --out=$(GOPATH)/src/github.com/intervention-engine/ie
// --version=v1.2.0-dirty

package app

import (
	"github.com/goadesign/goa"
	"time"
)

// careTeamPayload user type.
type careTeamPayload struct {
	// Care team leader
	Leader *string `form:"leader,omitempty" json:"leader,omitempty" xml:"leader,omitempty"`
	// Care team name
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
}

// Publicize creates CareTeamPayload from careTeamPayload
func (ut *careTeamPayload) Publicize() *CareTeamPayload {
	var pub CareTeamPayload
	if ut.Leader != nil {
		pub.Leader = ut.Leader
	}
	if ut.Name != nil {
		pub.Name = ut.Name
	}
	return &pub
}

// CareTeamPayload user type.
type CareTeamPayload struct {
	// Care team leader
	Leader *string `form:"leader,omitempty" json:"leader,omitempty" xml:"leader,omitempty"`
	// Care team name
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
}

// patientHuddle user type.
type patientHuddle struct {
	// patient id
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// reason for why patient is in this huddle
	Reason     *string    `form:"reason,omitempty" json:"reason,omitempty" xml:"reason,omitempty"`
	ReasonType *string    `form:"reason_type,omitempty" json:"reason_type,omitempty" xml:"reason_type,omitempty"`
	Reviewed   *bool      `form:"reviewed,omitempty" json:"reviewed,omitempty" xml:"reviewed,omitempty"`
	ReviewedAt *time.Time `form:"reviewed_at,omitempty" json:"reviewed_at,omitempty" xml:"reviewed_at,omitempty"`
}

// Publicize creates PatientHuddle from patientHuddle
func (ut *patientHuddle) Publicize() *PatientHuddle {
	var pub PatientHuddle
	if ut.ID != nil {
		pub.ID = ut.ID
	}
	if ut.Reason != nil {
		pub.Reason = ut.Reason
	}
	if ut.ReasonType != nil {
		pub.ReasonType = ut.ReasonType
	}
	if ut.Reviewed != nil {
		pub.Reviewed = ut.Reviewed
	}
	if ut.ReviewedAt != nil {
		pub.ReviewedAt = ut.ReviewedAt
	}
	return &pub
}

// PatientHuddle user type.
type PatientHuddle struct {
	// patient id
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// reason for why patient is in this huddle
	Reason     *string    `form:"reason,omitempty" json:"reason,omitempty" xml:"reason,omitempty"`
	ReasonType *string    `form:"reason_type,omitempty" json:"reason_type,omitempty" xml:"reason_type,omitempty"`
	Reviewed   *bool      `form:"reviewed,omitempty" json:"reviewed,omitempty" xml:"reviewed,omitempty"`
	ReviewedAt *time.Time `form:"reviewed_at,omitempty" json:"reviewed_at,omitempty" xml:"reviewed_at,omitempty"`
}

// schedulePatientPayload user type.
type schedulePatientPayload struct {
	// Date in YYYY-MM-dd format to schedule huddle
	Date *string `form:"date,omitempty" json:"date,omitempty" xml:"date,omitempty"`
	// Unique patient ID
	PatientID *string `form:"patient_id,omitempty" json:"patient_id,omitempty" xml:"patient_id,omitempty"`
}

// Validate validates the schedulePatientPayload type instance.
func (ut *schedulePatientPayload) Validate() (err error) {
	if ut.PatientID == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "patient_id"))
	}
	if ut.Date == nil {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "date"))
	}
	return
}

// Publicize creates SchedulePatientPayload from schedulePatientPayload
func (ut *schedulePatientPayload) Publicize() *SchedulePatientPayload {
	var pub SchedulePatientPayload
	if ut.Date != nil {
		pub.Date = *ut.Date
	}
	if ut.PatientID != nil {
		pub.PatientID = *ut.PatientID
	}
	return &pub
}

// SchedulePatientPayload user type.
type SchedulePatientPayload struct {
	// Date in YYYY-MM-dd format to schedule huddle
	Date string `form:"date" json:"date" xml:"date"`
	// Unique patient ID
	PatientID string `form:"patient_id" json:"patient_id" xml:"patient_id"`
}

// Validate validates the SchedulePatientPayload type instance.
func (ut *SchedulePatientPayload) Validate() (err error) {
	if ut.PatientID == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "patient_id"))
	}
	if ut.Date == "" {
		err = goa.MergeErrors(err, goa.MissingAttributeError(`response`, "date"))
	}
	return
}

// activeElement user type.
type activeElement struct {
	// Name of the Condition/Medication/Allergy
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Start Date of
	StartDate *time.Time `form:"start_date,omitempty" json:"start_date,omitempty" xml:"start_date,omitempty"`
}

// Publicize creates ActiveElement from activeElement
func (ut *activeElement) Publicize() *ActiveElement {
	var pub ActiveElement
	if ut.Name != nil {
		pub.Name = ut.Name
	}
	if ut.StartDate != nil {
		pub.StartDate = ut.StartDate
	}
	return &pub
}

// ActiveElement user type.
type ActiveElement struct {
	// Name of the Condition/Medication/Allergy
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Start Date of
	StartDate *time.Time `form:"start_date,omitempty" json:"start_date,omitempty" xml:"start_date,omitempty"`
}

// address user type.
type address struct {
	// City Name
	City *string `form:"city,omitempty" json:"city,omitempty" xml:"city,omitempty"`
	// Postal Code
	PostalCode *string `form:"postal_code,omitempty" json:"postal_code,omitempty" xml:"postal_code,omitempty"`
	// State Name
	State *string `form:"state,omitempty" json:"state,omitempty" xml:"state,omitempty"`
	// Street Name
	Street []string `form:"street,omitempty" json:"street,omitempty" xml:"street,omitempty"`
}

// Publicize creates Address from address
func (ut *address) Publicize() *Address {
	var pub Address
	if ut.City != nil {
		pub.City = ut.City
	}
	if ut.PostalCode != nil {
		pub.PostalCode = ut.PostalCode
	}
	if ut.State != nil {
		pub.State = ut.State
	}
	if ut.Street != nil {
		pub.Street = ut.Street
	}
	return &pub
}

// Address user type.
type Address struct {
	// City Name
	City *string `form:"city,omitempty" json:"city,omitempty" xml:"city,omitempty"`
	// Postal Code
	PostalCode *string `form:"postal_code,omitempty" json:"postal_code,omitempty" xml:"postal_code,omitempty"`
	// State Name
	State *string `form:"state,omitempty" json:"state,omitempty" xml:"state,omitempty"`
	// Street Name
	Street []string `form:"street,omitempty" json:"street,omitempty" xml:"street,omitempty"`
}

// name user type.
type name struct {
	// Family Name
	Family *string `form:"family,omitempty" json:"family,omitempty" xml:"family,omitempty"`
	// Full Name
	Full *string `form:"full,omitempty" json:"full,omitempty" xml:"full,omitempty"`
	// Given Name
	Given *string `form:"given,omitempty" json:"given,omitempty" xml:"given,omitempty"`
	// Middle Initial
	MiddleInitial *string `form:"middle_initial,omitempty" json:"middle_initial,omitempty" xml:"middle_initial,omitempty"`
}

// Publicize creates Name from name
func (ut *name) Publicize() *Name {
	var pub Name
	if ut.Family != nil {
		pub.Family = ut.Family
	}
	if ut.Full != nil {
		pub.Full = ut.Full
	}
	if ut.Given != nil {
		pub.Given = ut.Given
	}
	if ut.MiddleInitial != nil {
		pub.MiddleInitial = ut.MiddleInitial
	}
	return &pub
}

// Name user type.
type Name struct {
	// Family Name
	Family *string `form:"family,omitempty" json:"family,omitempty" xml:"family,omitempty"`
	// Full Name
	Full *string `form:"full,omitempty" json:"full,omitempty" xml:"full,omitempty"`
	// Given Name
	Given *string `form:"given,omitempty" json:"given,omitempty" xml:"given,omitempty"`
	// Middle Initial
	MiddleInitial *string `form:"middle_initial,omitempty" json:"middle_initial,omitempty" xml:"middle_initial,omitempty"`
}

// nextHuddle user type.
type nextHuddle struct {
	CareTeamName *string    `form:"care_team_name,omitempty" json:"care_team_name,omitempty" xml:"care_team_name,omitempty"`
	HuddleDate   *time.Time `form:"huddle_date,omitempty" json:"huddle_date,omitempty" xml:"huddle_date,omitempty"`
	HuddleID     *string    `form:"huddle_id,omitempty" json:"huddle_id,omitempty" xml:"huddle_id,omitempty"`
	Reason       *string    `form:"reason,omitempty" json:"reason,omitempty" xml:"reason,omitempty"`
	ReasonType   *string    `form:"reason_type,omitempty" json:"reason_type,omitempty" xml:"reason_type,omitempty"`
	Reviewed     *bool      `form:"reviewed,omitempty" json:"reviewed,omitempty" xml:"reviewed,omitempty"`
	ReviewedAt   *time.Time `form:"reviewed_at,omitempty" json:"reviewed_at,omitempty" xml:"reviewed_at,omitempty"`
}

// Publicize creates NextHuddle from nextHuddle
func (ut *nextHuddle) Publicize() *NextHuddle {
	var pub NextHuddle
	if ut.CareTeamName != nil {
		pub.CareTeamName = ut.CareTeamName
	}
	if ut.HuddleDate != nil {
		pub.HuddleDate = ut.HuddleDate
	}
	if ut.HuddleID != nil {
		pub.HuddleID = ut.HuddleID
	}
	if ut.Reason != nil {
		pub.Reason = ut.Reason
	}
	if ut.ReasonType != nil {
		pub.ReasonType = ut.ReasonType
	}
	if ut.Reviewed != nil {
		pub.Reviewed = ut.Reviewed
	}
	if ut.ReviewedAt != nil {
		pub.ReviewedAt = ut.ReviewedAt
	}
	return &pub
}

// NextHuddle user type.
type NextHuddle struct {
	CareTeamName *string    `form:"care_team_name,omitempty" json:"care_team_name,omitempty" xml:"care_team_name,omitempty"`
	HuddleDate   *time.Time `form:"huddle_date,omitempty" json:"huddle_date,omitempty" xml:"huddle_date,omitempty"`
	HuddleID     *string    `form:"huddle_id,omitempty" json:"huddle_id,omitempty" xml:"huddle_id,omitempty"`
	Reason       *string    `form:"reason,omitempty" json:"reason,omitempty" xml:"reason,omitempty"`
	ReasonType   *string    `form:"reason_type,omitempty" json:"reason_type,omitempty" xml:"reason_type,omitempty"`
	Reviewed     *bool      `form:"reviewed,omitempty" json:"reviewed,omitempty" xml:"reviewed,omitempty"`
	ReviewedAt   *time.Time `form:"reviewed_at,omitempty" json:"reviewed_at,omitempty" xml:"reviewed_at,omitempty"`
}

// pie user type.
type pie struct {
	// Individual Pie sli
	Slices []*pieSlice `form:"slices,omitempty" json:"slices,omitempty" xml:"slices,omitempty"`
}

// Publicize creates Pie from pie
func (ut *pie) Publicize() *Pie {
	var pub Pie
	if ut.Slices != nil {
		pub.Slices = make([]*PieSlice, len(ut.Slices))
		for i2, elem2 := range ut.Slices {
			pub.Slices[i2] = elem2.Publicize()
		}
	}
	return &pub
}

// Pie user type.
type Pie struct {
	// Individual Pie sli
	Slices []*PieSlice `form:"slices,omitempty" json:"slices,omitempty" xml:"slices,omitempty"`
}

// pieSlice user type.
type pieSlice struct {
	// Maximum possible value
	MaxValue *int `form:"maxValue,omitempty" json:"maxValue,omitempty" xml:"maxValue,omitempty"`
	// Risk Category Name
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Risk Category Value
	Value *int `form:"value,omitempty" json:"value,omitempty" xml:"value,omitempty"`
	// Weight On Overall Risk Value
	Weight *int `form:"weight,omitempty" json:"weight,omitempty" xml:"weight,omitempty"`
}

// Publicize creates PieSlice from pieSlice
func (ut *pieSlice) Publicize() *PieSlice {
	var pub PieSlice
	if ut.MaxValue != nil {
		pub.MaxValue = ut.MaxValue
	}
	if ut.Name != nil {
		pub.Name = ut.Name
	}
	if ut.Value != nil {
		pub.Value = ut.Value
	}
	if ut.Weight != nil {
		pub.Weight = ut.Weight
	}
	return &pub
}

// PieSlice user type.
type PieSlice struct {
	// Maximum possible value
	MaxValue *int `form:"maxValue,omitempty" json:"maxValue,omitempty" xml:"maxValue,omitempty"`
	// Risk Category Name
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Risk Category Value
	Value *int `form:"value,omitempty" json:"value,omitempty" xml:"value,omitempty"`
	// Weight On Overall Risk Value
	Weight *int `form:"weight,omitempty" json:"weight,omitempty" xml:"weight,omitempty"`
}
