package huddles

import "github.com/intervention-engine/fhir/models"

// Huddle provides convenient functions on a Group to get access to extended huddle data fields
type Huddle models.Group

// IsHuddle checks the Group's code to ensure it has the proper Huddle code
func (h *Huddle) IsHuddle() bool {
	return h.Code.MatchesCode("http://interventionengine.org/fhir/cs/huddle", "HUDDLE")
}

// ActiveDateTime returns the huddle's active datetime (or nil if there is not one)
func (h *Huddle) ActiveDateTime() *models.FHIRDateTime {
	activeDT := findExtension(h.Extension, "http://interventionengine.org/fhir/extension/group/activeDateTime")
	if activeDT != nil {
		return activeDT.ValueDateTime
	}
	return nil
}

// Leader returns the huddle's leader (or nil if there is not one)
func (h *Huddle) Leader() *models.Reference {
	leader := findExtension(h.Extension, "http://interventionengine.org/fhir/extension/group/leader")
	if leader != nil {
		return leader.ValueReference
	}
	return nil
}

// HuddleMembers returns a slice of HuddleMembers associated to this huddle
func (h *Huddle) HuddleMembers() []HuddleMember {
	members := make([]HuddleMember, len(h.Member))
	for i := range h.Member {
		members[i] = HuddleMember(h.Member[i])
	}
	return members
}

// FindHuddleMember returns the huddle member with the specified ID (or nil if the patient is not in the huddle)
func (h *Huddle) FindHuddleMember(patientID string) *HuddleMember {
	for i := range h.Member {
		if h.Member[i].Entity.ReferencedID == patientID {
			hm := HuddleMember(h.Member[i])
			return &hm
		}
	}
	return nil
}

// HuddleMember provides convenient functions on a GroupMemberComponent to get access to extended huddle data fields
type HuddleMember models.GroupMemberComponent

// ID returns the members ID
func (h *HuddleMember) ID() string {
	if h.Entity == nil {
		return ""
	}
	return h.Entity.ReferencedID
}

// Reason returns the reason the member was added to the huddle (or nil if the reason isn't set)
func (h *HuddleMember) Reason() *models.CodeableConcept {
	reason := findExtension(h.Extension, "http://interventionengine.org/fhir/extension/group/member/reason")
	if reason != nil {
		return reason.ValueCodeableConcept
	}
	return nil
}

// ReasonIsManuallyAdded indicates if the member reason is due to the patient being manually added to the huddle
func (h *HuddleMember) ReasonIsManuallyAdded() bool {
	reason := h.Reason()
	return reason != nil && reason.MatchesCode("http://interventionengine.org/fhir/cs/huddle-member-reason", "MANUAL_ADDITION")
}

// ReasonIsRecentEncounter indicates if the member reason is due to a recent significant encounter
func (h *HuddleMember) ReasonIsRecentEncounter() bool {
	reason := h.Reason()
	return reason != nil && reason.MatchesCode("http://interventionengine.org/fhir/cs/huddle-member-reason", "RECENT_ENCOUNTER")
}

// ReasonIsRiskScore indicates if the member reason is due to the patient's current risk score
func (h *HuddleMember) ReasonIsRiskScore() bool {
	reason := h.Reason()
	return reason != nil && reason.MatchesCode("http://interventionengine.org/fhir/cs/huddle-member-reason", "RISK_SCORE")
}

// ReasonIsRollOver indicates if the member reason is due to roll over from a previous huddle
func (h *HuddleMember) ReasonIsRollOver() bool {
	reason := h.Reason()
	return reason != nil && reason.MatchesCode("http://interventionengine.org/fhir/cs/huddle-member-reason", "ROLLOVER")
}

// Reviewed returns the date that the member was reviewed for this huddle (or nil if they haven't been reviewed)
func (h *HuddleMember) Reviewed() *models.FHIRDateTime {
	reviewed := findExtension(h.Extension, "http://interventionengine.org/fhir/extension/group/member/reviewed")
	if reviewed != nil {
		return reviewed.ValueDateTime
	}
	return nil
}

func findExtension(ext []models.Extension, extURL string) *models.Extension {
	for i := range ext {
		if ext[i].Url == extURL {
			return &ext[i]
		}
	}
	return nil
}
