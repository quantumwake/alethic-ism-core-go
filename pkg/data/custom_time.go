package data

import (
	"encoding/json"
	"time"
)

//
//type CustomTime struct {
//	time.Time
//}
//
//func (ct *CustomTime) UnmarshalJSON(b []byte) error {
//	var raw string
//	err := json.Unmarshal(b, &raw)
//	if err != nil {
//		return err
//	}
//
//	// Try parsing with different formats
//	formats := []string{
//		time.RFC3339,
//		"2006-01-02T15:04:05.999999",
//		"2006-01-02T15:04:05",
//		"2006-01-02 15:04:05",
//		"2006-01-02",
//	}
//
//	for _, format := range formats {
//		t, err := time.Parse(format, raw)
//		if err == nil {
//			*ct = CustomTime{t}
//			return nil
//		}
//	}
//
//	return fmt.Errorf("unable to parse time: %s", raw)
//}
//
//func (u *Usage) UnmarshalJSON(data []byte) error {
//	type Alias Usage
//	aux := &struct {
//		*Alias
//	}{
//		Alias: (*Alias)(u),
//	}
//	if err := json.Unmarshal(data, &aux); err != nil {
//		return err
//	}
//	return nil
//}

// Custom JSON unmarshaler for time.Time
func (u *Usage) UnmarshalJSON(data []byte) error {
	type Alias Usage
	aux := &struct {
		TransactionTime string `json:"transaction_time"`
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	u.TransactionTime, err = time.Parse("2006-01-02T15:04:05.999999", aux.TransactionTime)
	if err != nil {
		return err
	}
	return nil
}
