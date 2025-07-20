package data

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type JSON map[string]any

// Scan value into Jsonb, implements sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSON)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*j = make(JSON)
			return nil
		}
		return json.Unmarshal(v, j)
	case string:
		if v == "" {
			*j = make(JSON)
			return nil
		}
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan type %T into JSON", value)
	}
}

// Value return json value, implement driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if j == nil || len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

// GormDataType returns the data type for gorm
func (JSON) GormDataType() string {
	return "json"
}

// GormDBDataType returns the data type for the database
func (JSON) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "JSONB"
	case "mysql":
		return "JSON"
	case "sqlite":
		return "JSON"
	default:
		return "TEXT"
	}
}
