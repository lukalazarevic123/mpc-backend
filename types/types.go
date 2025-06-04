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

type InvitationMessage struct {
	OrganizationID   int    `json:"organization_id"`
	OrganizationName string `json:"organization_name"`
	Message          string `json:"message"`
}

// TransactionRequest is the payload when initiating a transaction.
type TransactionRequest struct {
	OrganizationName string `json:"organization_name"`
	Initiator        string `json:"initiator"`
}

// TransactionNotification is sent to all members of an organization.
type TransactionNotification struct {
	OrganizationID int    `json:"organization_id"`
	Initiator      string `json:"initiator"`
	Details        string `json:"details"`
	Message        string `json:"message"`
}

// TransactionConfirmationRequest is the payload for confirming a transaction.
type TransactionConfirmationRequest struct {
	OrganizationName string `json:"organization_name"`
	Address          string `json:"address"`
}
