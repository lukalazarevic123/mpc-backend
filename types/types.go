package types

type Participant struct {
	Address string `json:"address"`
}

type CreateOrganizationRequest struct {
	Name         string        `json:"name"`
	Threshold    int           `json:"threshold"`
	Participants []Participant `json:"participants"`
}

type Organization struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	Threshold    int           `json:"threshold"`
	Participants []Participant `json:"participants"`
}
