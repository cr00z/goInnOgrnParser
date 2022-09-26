package innogrn

import (
	"testing"
)

func TestCheckINN(t *testing.T) {
	tests := []struct {
		name string
		inn  string
		want bool
	}{
		{"INN 7830002293", "7830002293", true},
		{"INN 1111111111", "1111111111", false},
		{"INN 500100732259", "500100732259", true},
		{"INN 111111111111", "111111111111", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CheckINN(test.inn)
			if got != test.want {
				t.Errorf("%s: got %t, want %t", test.name, got, test.want)
			}
		})
	}
}

func TestCheckOGRN(t *testing.T) {
	tests := []struct {
		name string
		ogrn string
		want bool
	}{
		{"OGRN 1037739010891", "1037739010891", true},
		{"OGRN 1035006110083", "1035006110083", true},
		{"OGRN 1111111111111", "1111111111111", false},
		{"OGRN 304500116000157", "304500116000157", true},
		{"OGRN 304463210700212", "304463210700212", true},
		{"OGRN 222222222222222", "222222222222222", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := CheckOGRN(test.ogrn)
			if got != test.want {
				t.Errorf("%s: got %t, want %t", test.name, got, test.want)
			}
		})
	}
}
