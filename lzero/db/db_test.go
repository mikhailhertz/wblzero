package db

import (
	"testing"
)

func TestCreateEmpty(t *testing.T) {
	db, err := Create("", nil)
	if err == nil || db != nil {
		t.Fatalf("Didn't error on Create with bad parameters")
	}
}

func TestCreateValid(t *testing.T) {
	db, err := Create("postgresql://postgres:beijing@localhost/wb", make(chan string))
	if err != nil || db == nil {
		t.Fatalf("Error on Create with valid parameters")
	}
}

// TODO: test other stuff
