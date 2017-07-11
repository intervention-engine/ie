package appt

// Support sorting by huddle priority by using the following sort rules:
// - First check for dueness.  Patients who are due or overdue go first (more overdue before less overdue).
// - If neither is overdue, or they are overdue the same, sort from earliest ideal huddle to latest ideal huddle.
// - If they have the same ideal huddle, sub-sort by risk score (highest goes first).
// - If they have same risk score too, sub-sort by furthest allowed huddle (implies urgency).
// - If they have the same furthest allowed huddle too, sort by the last huddle (oldest first).
// - If they have same last huddle too, just sort by the patients id.
type priority struct {
	patients  []itinerary
	huddleIdx int
}

func (p priority) Len() int {
	return len(p.patients)
}
func (p priority) Swap(i, j int) {
	p.patients[i], p.patients[j] = p.patients[j], p.patients[i]
}
func (p priority) Less(i, j int) bool {
	// First look at "dueness" by seeing if the furthest allowed huddle is this one or before.
	// If either or both are due, and have different furthest allowed huddles, the more overdue one goes first.
	if p.patients[i].FurthestAllowedHuddle != p.patients[j].FurthestAllowedHuddle {
		iDue := p.patients[i].FurthestAllowedHuddle != nil && *p.patients[i].FurthestAllowedHuddle <= p.huddleIdx
		jDue := p.patients[j].FurthestAllowedHuddle != nil && *p.patients[j].FurthestAllowedHuddle <= p.huddleIdx
		if iDue && jDue && *p.patients[i].FurthestAllowedHuddle != *p.patients[j].FurthestAllowedHuddle {
			return *p.patients[i].FurthestAllowedHuddle < *p.patients[j].FurthestAllowedHuddle
		} else if iDue && !jDue {
			return true
		} else if !iDue && jDue {
			return false
		}
	}

	// If neither is due, or they are overdue the same, compare by the next ideal huddle.
	// If one is nil (which shouldn't happen) assume they should go last.
	if p.patients[i].NextIdealHuddle != p.patients[j].NextIdealHuddle {
		if p.patients[i].NextIdealHuddle == nil {
			return false
		} else if p.patients[j].NextIdealHuddle == nil {
			return true
		} else if *p.patients[i].NextIdealHuddle != *p.patients[j].NextIdealHuddle {
			return *p.patients[i].NextIdealHuddle < *p.patients[j].NextIdealHuddle
		}
	}

	// If they have the same next ideal huddle, compare by the score (higher goes first).
	// If one is nil assume they should go last.
	if p.patients[i].Score != p.patients[j].Score {
		if p.patients[i].Score == nil {
			return false
		} else if p.patients[j].Score == nil {
			return true
		} else if *p.patients[i].Score != *p.patients[j].Score {
			return *p.patients[i].Score > *p.patients[j].Score
		}
	}

	// If they have the same risk score too, compare by the furthest allowed huddle.
	// If one is nil (which shouldn't happen) assume they should go last.
	if p.patients[i].FurthestAllowedHuddle != p.patients[j].FurthestAllowedHuddle {
		if p.patients[i].FurthestAllowedHuddle == nil {
			return false
		} else if p.patients[j].FurthestAllowedHuddle == nil {
			return true
		} else if *p.patients[i].FurthestAllowedHuddle != *p.patients[j].FurthestAllowedHuddle {
			return *p.patients[i].FurthestAllowedHuddle < *p.patients[j].FurthestAllowedHuddle
		}
	}

	// If they have the same furthest allowed huddle, try comparing by the last huddle.
	// If one is nil, they've never had a huddle, so they should be prefered first over someone who has.
	if p.patients[i].LastHuddle != p.patients[j].LastHuddle {
		if p.patients[i].LastHuddle == nil {
			return true
		} else if p.patients[j].LastHuddle == nil {
			return false
		} else if *p.patients[i].LastHuddle != *p.patients[j].LastHuddle {
			return *p.patients[i].LastHuddle < *p.patients[j].LastHuddle
		}
	}

	// OK, they're equal for all intents and purposes, but compare by ID so we have consistent order
	return p.patients[i].ID < p.patients[j].ID
}