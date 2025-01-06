package model

import "time"

type VaultType string

const (
	VaultKMS   = VaultType("kms")
	VaultLocal = VaultType("local")
)

type Vault struct {
	ID   string    `gorm:"column:id; type:varchar(73); default:gen_random_uuid(); primaryKey"`
	Name string    `gorm:"column:name; type:varchar(255); not null" json:"name"`
	Type VaultType `gorm:"column:type; type:config_type; not null" json:"type"`
	//OwnerID  string         `gorm:"column:owner_id; type:varchar(36); null" json:"column:owner_id; type:varchar(36); null" json:"owner_id"`
	Metadata map[string]any `json:"metadata"`

	// ISO timestamp
	CreatedAt time.Time `json:"createdAt"`

	// ISO timestamp
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Vault) TableName() string {
	return "config_map"
}
