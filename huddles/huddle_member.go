package huddles

import "github.com/intervention-engine/fhir/models"

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
