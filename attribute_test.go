package radius

import (
	"testing"
)

func TestNewUserPassword_length(t *testing.T) {
	tbl := []struct {
		Password      string
		EncodedLength int
	}{
		{"", 16},
		{"abc", 16},
		{"0123456789abcde", 16},
		{"0123456789abcdef", 16},
		{"0123456789abcdef0", 16 * 2},
		{"0123456789abcdef0123456789abcdef0123456789abcdef", 16 * 3},
	}

	secret := []byte(`12345`)
	ra := []byte(`0123456789abcdef`)

	for _, x := range tbl {
		attr, err := NewUserPassword([]byte(x.Password), secret, ra)
		if err != nil {
			t.Fatal(err)
		}
		if len(attr) != x.EncodedLength {
			t.Fatalf("expected encoded length of %#v = %d, got %d", x.Password, x.EncodedLength, len(attr))
		}
	}
}
