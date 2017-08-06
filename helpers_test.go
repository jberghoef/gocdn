package main

import (
	"testing"
	"net/http"
)

func TestVerifyContentType(t *testing.T) {
	tables := []struct {
		x string
		y bool
	}{
		{"text/css", true},
		{"text/javascript", true},
		{"image/png", true},
		{"image/jpeg", true},
		{"image/pong", false},
		{"image/jpg", false},
	}

	mockHeader := http.Header{}
	mockHeader.Add("Content-Type", "")
	for _, table := range tables {
		mockHeader.Set("Content-Type", table.x)
		result := verifyContentType(mockHeader)
		if result != table.y {
			t.Errorf("Content type verification returned unexpected result [%t] for value [%s]. Expected [%t].",
				result, table.x, table.y)
		}
	}
}

func TestDefineCacheControl(t *testing.T) {
	tables := []struct {
		x string
		i bool
		r bool
		m int
	}{
		{"private", true, false, 3600},
		{"no-store", true, false, 3600},
		{"private, no-store", true, false, 3600},
		{"no-cache", false, true, 3600},
		{"must-revalidate", false, true, 3600},
		{"no-cache, must-revalidate", false, true, 3600},
		{"private, no-cache", true, true, 3600},
		{"no-store, must-revalidate", true, true, 3600},
		{"max-age:7200", false, false, 7200},
		{"private, max-age:7200", true, false, 7200},
		{"no-cache, max-age:7200", false, true, 7200},
		{"private, no-cache, max-age:7200", true, true, 7200},
	}

	mockHeader := http.Header{}
	mockHeader.Add("Cache-Control", "")
	for _, table := range tables {
		mockHeader.Set("Cache-Control", table.x)
		i, r, m := defineCacheControl(mockHeader)
		if i != table.i || r != table.r || m != table.m {
			t.Errorf("Cache control definition returned unexpected result [%t, %t, %d] for value [%s]. Expected [%t, %t, %d].",
				i, r, m, table.x, table.i, table.r, table.m)
		}
	}
}

func TestCreateHash(t *testing.T) {
	tables := []struct {
		x string
		y string
	}{
		{"http://www.google.nl", "3d04d47edbf1d492cd15bf9fa1950b4f"},
		{"http://www.google.com", "ed646a3334ca891fd3467db131372140"},
	}

	for _, table := range tables {
		result := createHash(table.x)
		if result != table.y {
			t.Errorf("Create hash returned unexpected result [%s] for value [%s]. Expected [%s].",
				result, table.x, table.y)
		}
	}
}
