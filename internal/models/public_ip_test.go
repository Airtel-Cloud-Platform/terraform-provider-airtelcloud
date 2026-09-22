package models

import (
	"encoding/json"
	"testing"
)

func TestCreatePublicIPRequest_MarshalsNullPortID(t *testing.T) {
	b, err := json.Marshal(CreatePublicIPRequest{Name: "temporal-pip", Description: "new pip", PortID: nil})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	want := `{"name":"temporal-pip","description":"new pip","port_id":null}`
	if string(b) != want {
		t.Fatalf("Marshal() = %s, want %s", b, want)
	}
}
