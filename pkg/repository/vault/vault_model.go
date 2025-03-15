package vault

import (
	"encoding/json"
	"time"
)

// VaultType is the type of vault to use for storing secrets and configuration.
type VaultType string

// ConfigMapType is the type of configuration map to use for storing configuration.
type ConfigMapType string

// Vault types for storing secrets and configuration.
const (
	VaultAWS        = VaultType("aws")
	VaultLocal      = VaultType("local")
	ConfigMapSecret = ConfigMapType("secret")
	ConfigMapConfig = ConfigMapType("config_map")
)

// Vault is a secret storage mechanism. It can be used to store secrets and configuration.
type Vault struct {
	// UUID v4 string (36 characters) with dashes
	ID string `gorm:"column:id; type:varchar(36); default:gen_random_uuid(); primaryKey"`

	// Name of the vault (e.g. "my-vault")
	Name string `gorm:"column:name; type:varchar(255); not null" json:"name"`

	// Type of vault to use for storage (e.g. local, aws secret manager, hashicorp vault, custom, etc.)
	Type VaultType `gorm:"column:type; type:config_type; not null" json:"type"`

	// JSON object containing metadata about the vault (e.g. {"region": "us-west-2"})
	Metadata map[string]any `json:"metadata"`

	// ISO timestamp (e.g. "2021-01-01T00:00:00Z")
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`

	// ISO timestamp (e.g. "2021-01-01T00:00:00Z")
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName sets the table name for the Vault struct
func (Vault) TableName() string {
	return "vault"
}

// ConfigMap is a configuration map storage mechanism. It can be used to store configuration maps.
type ConfigMap struct {
	ID         *string         `gorm:"not null; column:id; type:varchar(36); primaryKey"`
	Name       string          `gorm:"not null; column:name; type:varchar(255); not null" json:"name"`
	Type       ConfigMapType   `gorm:"not null; column:type; type:config_type; not null" json:"type"`
	Data       json.RawMessage `gorm:"not null; column:data; null" json:"data"`
	VaultKeyID *string         `gorm:"    null; column:vault_key_id; type:varchar(36); null" json:"vault_key_id; omitempty"`
	VaultID    *string         `gorm:"    null; column:vault_id; type:varchar(36); null" json:"vault_id; omitempty"`
	OwnerID    string          `gorm:"not null; column:owner_id; type:varchar(36); null" json:"column:owner_id; type:varchar(36); null" json:"owner_id"`
	CreatedAt  time.Time       `gorm:"column:created_at; null" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at; null" json:"updated_at; omitempty"`
	DeletedAt  time.Time       `gorm:"column:deleted_at; null" json:"deleted_at; omitempty"`
}

// TableName sets the table name for the ConfigMap struct
func (ConfigMap) TableName() string {
	return "config_map"
}
