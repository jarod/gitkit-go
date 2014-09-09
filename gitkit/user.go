package gitkit

type ProviderUserInfo struct {
	ProviderID  string `json:"providerId"`
	DisplayName string `json:"displayName"`
	PhotoURL    string `json:"photoUrl"`
	FederatedID string `json:"federatedId"`
}

// User GtkkitUser
type User struct {
	LocalID           string             `json:"localId"`
	Email             string             `json:"email"`
	EmailVerified     bool               `json:"emailVerified"`
	DisplayName       string             `json:"displayName"`
	ProviderUserInfo  []ProviderUserInfo `json:"providerUserInfo"`
	PhotoURL          string             `json:"photoUrl"`
	PasswordHash      []byte             `json:"passwordHash"`
	Salt              []byte             `json:"salt"`
	Version           int                `json:"version"`
	PasswordUpdatedAt float64            `json:"passwordUpdatedAt"`
}
