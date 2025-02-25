package data

import (
	"fmt"
	"testing"
)

func TestNewDataAccess(t *testing.T) (*Access, error) {
	//func TestNewDataAccess(t *testing.T) *data.Access {

	da := NewDataAccess("host=localhost port=5432 user=postgres password=postgres1 dbname=postgres sslmode=disable")
	if da == nil {
		if t != nil {
			t.Errorf("unable to connect to database")
		}
		return nil, fmt.Errorf("unable to connect to database")
	}

	return da, nil
}
