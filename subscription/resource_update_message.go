package subscription

type ResourceUpdateMessage struct {
	PatientID string
	Timestamp string
}

func NewResourceUpdateMessage(patientID, timestamp string) ResourceUpdateMessage {
	return ResourceUpdateMessage{PatientID: patientID, Timestamp: timestamp}
}
