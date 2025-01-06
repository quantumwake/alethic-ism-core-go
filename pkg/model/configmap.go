package model

import (
	"encoding/json"
	"time"
)

type ConfigMapType string

// TODO NOT SURE WHETHER WE NEED THIS
const (
	Secret = ConfigMapType("secret")
	Config = ConfigMapType("config_map")
)

type ConfigMap struct {
	ID         string          `gorm:"column:id; type:varchar(36); default:gen_random_uuid(); primaryKey"`
	Name       string          `gorm:"column:name; type:varchar(255); not null" json:"name"`
	Type       ConfigMapType   `gorm:"column:type; type:config_type; not null" json:"type"`
	Data       json.RawMessage `gorm:"column:data; null" json:"data"`
	VaultKeyID string          `gorm:"column:vault_key_id; type:varchar(36); null" json:"vault_key_id"`
	VaultID    string          `gorm:"column:vault_id; type:varchar(36); null" json:"vault_id"`
	OwnerID    string          `gorm:"column:owner_id; type:varchar(36); null" json:"column:owner_id; type:varchar(36); null" json:"owner_id"`
	CreatedAt  time.Time       `gorm:"column:created_at; null" json:"created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at; null" json:"updated_at"`
	DeletedAt  time.Time       `gorm:"column:deleted_at; null" json:"deleted_at"`
}

func (ConfigMap) TableName() string {
	return "config_map"
}

// 	ID           string                  `gorm:"column:id;type:varchar(73);default:gen_random_uuid();primaryKey"`
//	InternalID   uint                    `gorm:"column:internal_id;autoIncrement;unique"`
//	ProcessorID  string                  `gorm:"column:processor_id;type:varchar(36)"`
//	StateID      string                  `gorm:"column:state_id;primaryKey;type:varchar(36)"`
//	Direction    ProcessorStateDirection `gorm:"column:direction;primaryKey;type:processor_state_direction"`
//	Status       ProcessorStatus         `gorm:"column:status;type:processor_status"`
//	Count        *int                    `gorm:"column:count"`
//	CurrentIndex *int                    `gorm:"column:current_index"`
//	MaximumIndex *int                    `gorm:"column:maximum_index"`
