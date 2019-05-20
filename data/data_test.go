package data

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/boreq/starlight/network/node"
)

var validateNickTestCases = []struct {
	Nick    string
	IsValid bool
	Message string
}{
	// Length tests
	{
		Nick:    "",
		IsValid: false,
		Message: "empty nicks are not valid",
	},
	{
		Nick:    "a",
		IsValid: false,
		Message: "one letter nicks are not valid",
	},
	{
		Nick:    "aa",
		IsValid: false,
		Message: "two letter nicks are not valid",
	},
	{
		Nick:    "aaa",
		IsValid: true,
		Message: "three letter nicks are valid",
	},
	{
		Nick:    strings.Repeat("a", 20),
		IsValid: true,
		Message: "twenty letter nicks are valid",
	},
	{
		Nick:    strings.Repeat("a", 21),
		IsValid: false,
		Message: "twenty-one letter nicks are not valid",
	},

	// Character tests
	{
		Nick:    "abcde",
		IsValid: true,
		Message: "nicks containing letters are valid",
	},
	{
		Nick:    "ABCDE",
		IsValid: true,
		Message: "nicks containing upper-case letters are valid",
	},
	{
		Nick:    "1bcde",
		IsValid: false,
		Message: "nicks starting with a number are not valid",
	},
	{
		Nick:    "a2345",
		IsValid: true,
		Message: "nicks starting with a letter are valid",
	},
	{
		Nick:    "a2c4e",
		IsValid: true,
		Message: "nicks containing letters and numbers are valid",
	},
	{
		Nick:    "_bcde",
		IsValid: false,
		Message: "nicks starting with a special character are not valid",
	},
	{
		Nick:    "a_-[]",
		IsValid: true,
		Message: "nicks containing a special character are valid",
	},

	// Real-life examples
	{
		Nick:    "user",
		IsValid: true,
		Message: "a real nick is valid",
	},
	{
		Nick:    "user123",
		IsValid: true,
		Message: "a real nick is valid",
	},
	{
		Nick:    "user123[laptop]",
		IsValid: true,
		Message: "a real nick is valid",
	},
	{
		Nick:    "user-name[laptop]",
		IsValid: true,
		Message: "a real nick is valid",
	},
	{
		Nick:    "user_name123",
		IsValid: true,
		Message: "a real nick is valid",
	},
	{
		Nick:    "user_name[laptop]",
		IsValid: true,
		Message: "a real nick is valid",
	},
}

func TestValidateNick(t *testing.T) {
	for _, testCase := range validateNickTestCases {
		if err := ValidateNick(testCase.Nick); err != nil {
			if testCase.IsValid {
				t.Fatalf("test case '%s': nick %s should be valid but got: %s", testCase.Message, testCase.Nick, err)
			}
		} else {
			if !testCase.IsValid {
				t.Fatalf("test case '%s': nick %s should not be valid", testCase.Message, testCase.Nick)
			}
		}
	}
}

const identity = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpgIBAAKCAQEAuO7bjXbSoVO3HRkhMiPQwCpCTdUCvZPgM0wwrAi6L0ex+Y5b
6d1iCks9EsPcP+pNDvhYuPn3UE1jfw4mZIMqJpTXJaN5cEnyF2wdRTP8hpsDepz6
76yoUOd8oFmEJfSZEGfYuOrYB7mAx2bE53WYW03Mbhx7p8qaUkgySzMTDDuduD4f
RbATRkw1f8Ki38yRn2NRaZFNFh3mHOWVliHp1ER4/FkyDYc7RCNTollcqBuqEkUJ
vJjehGAU73IYp8Dqb4xr9qO2dmT4CPoeT4icNec2Y8aEcRdW4Fo4XgyzXlwVtxRe
lCgtfnaza0wHoqTJHf/MFH5P4C5/JiO+1gnOHQIDAQABAoIBAQCH4DfAQMWRcwjf
gE87n8UI7AO7W/6fe78G8bvxKphhlLPXQBmYQuh917oPx4hUDbqAfUfy4PYtMi8g
cy0SPK0Dm+hX5zyanDobq3v2FLQ90jdEJ4LYBmvExdBzoFHP8V9lBmfdte7z/f/4
gjG6PlSrAQZrANJ5/gpU2mbZibU9DnAX4TJviP2gZV87Q5D66vVLaus9tWnAE819
A9JtF2R+vBer+32KECSNkQz0cvHu4+TzIJbLi5AYnlHDPx0ke7/b+99yqnUg+LRe
737u2vEInV/DcdDMruDBh5WZf/VuBETuHGoVqwtGWuzQT9Q7YiTyVWdAMNjYgbcA
gWn0upYBAoGBAM1tKpv1ApF31rPP262UAUr/HUtTZNOwqVJh/G3rrz7qAPNZIPiJ
W2hT8od72K8ZH9aE9gfnjuScjnswM5GKCEb5sB/b1GHoF0b8d8y0itdnGxV7FI3+
ddEGh8SrDUHQwM4ImGj5PDU/HABdwREi6QBlK4SsPBoSinhKQOOF+y0tAoGBAOZ2
HJo8ahEBGUG3M+sjlJQgwvKEwN8EwCtaL6InLz76l8lOOVKJwVTM8cBsxBmJl/u/
+jAriVcSJYDmaVr9xmVSlOkONdVd9MV3VC7Ssr4yScQZjldCMFD7QBqhqGLj7/zw
swmKpbOtsRuJae8U0RdBN39kY0CDnxh/dkJrjRqxAoGBAMoI7poZ2t/Es+V+rXhG
kwr2YxI9P3GvUqgSdJiK7nz62dp7syCcvsiZn3K+S/rRw+1QMUTO6UtP6hWf72fZ
EJD1atG6e2ObRqFrFku+2LpGzm1O8oVAWREt0gOLk2tCaw13iKXdUeiwW9LEYmh/
JBdeaPGAD1A5IfRyWuUqVUE1AoGBAMC0YpZVjhtJ3+SjXDZyOfriqiBAAUZ6onWd
o9bjDQ6MW/9n+Waa6Z4PANb2G8N+2icYEAvXW7AC7HksMUx0h0CSHRIDX+BaACJd
9XZxmCSRyDzBYdR09BHDBYc/RZ3rGvFWE18XIBduVXnBHWNc9LmNPuq29ocriAzk
B+7iH8sBAoGBAM1KDkO8xYimvNnNiVUd931GvvCzF5DG+91QS8VDDFwXP8HvK3kL
Uw3M3qBafFl0BxvcM2/OqotjbBFVSa8xiQLgGodk2Z3KsftEKt2LnzPmgt4aj1D7
Rj+QkQ34gbWRCI9qpOacVysy3ddsQQKOvklRoHvWKNXeidcB3EBFRDjD
-----END RSA PRIVATE KEY-----
`

func makeIdentity() *node.Identity {
	iden, err := node.LoadIdentity([]byte(identity))
	if err != nil {
		panic(err)
	}
	return iden
}

func makeValidNickData() *NickData {
	iden := makeIdentity()

	pubKeyBytes, err := iden.PubKey.Bytes()
	if err != nil {
		panic(err)
	}

	rv := &NickData{
		Id:        iden.Id,
		Nick:      "nick",
		Time:      time.Now(),
		PublicKey: pubKeyBytes,
	}
	return withValidSignature(rv)
}

func withValidSignature(nickData *NickData) *NickData {
	iden := makeIdentity()

	data := nickData.GetDataToSign()
	signature, err := iden.PrivKey.Sign(data, SigningHash)
	if err != nil {
		panic(err)
	}

	nickData.Signature = signature
	return nickData
}

func TestNickDataValidateValid(t *testing.T) {
	nickData := makeValidNickData()
	if err := nickData.Validate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestNickDataValidateInvalidId(t *testing.T) {
	nickData := makeValidNickData()
	nickData.Id = nil
	nickData = withValidSignature(nickData)

	if err := nickData.Validate(); err == nil {
		t.Fatal("expected an error")
	} else {
		t.Log(err)
		if !strings.Contains(err.Error(), "id is invalid") {
			t.Fatal(err)
		}
	}
}

func TestNickDataValidateIdNotMatchingThePublicKey(t *testing.T) {
	nickData := makeValidNickData()
	nickData.Id[len(nickData.Id)-1] = 0
	nickData = withValidSignature(nickData)

	if err := nickData.Validate(); err == nil {
		t.Fatal("expected an error")
	} else {
		t.Log(err)
		if !strings.Contains(err.Error(), "does not match the public key") {
			t.Fatal(err)
		}
	}
}

func TestNickDataValidateInvalidNick(t *testing.T) {
	nickData := makeValidNickData()
	nickData.Nick = ""
	nickData = withValidSignature(nickData)

	if err := nickData.Validate(); err == nil {
		t.Fatal("expected an error")
	} else {
		t.Log(err)
		if !strings.Contains(err.Error(), "invalid nick") {
			t.Fatal(err)
		}
	}
}

func TestNickDataValidateInvalidTime(t *testing.T) {
	nickData := makeValidNickData()
	nickData.Time = time.Time{}
	nickData = withValidSignature(nickData)

	if err := nickData.Validate(); err == nil {
		t.Fatal("expected an error")
	} else {
		t.Log(err)
		if !strings.Contains(err.Error(), "time") {
			t.Fatal(err)
		}
	}
}

func TestNickDataValidateMissingPublicKey(t *testing.T) {
	nickData := makeValidNickData()
	nickData.PublicKey = nil
	nickData = withValidSignature(nickData)

	if err := nickData.Validate(); err == nil {
		t.Fatal("expected an error")
	} else {
		t.Log(err)
		if !strings.Contains(err.Error(), "read the public key") {
			t.Fatal(err)
		}
	}
}

func TestNickDataValidateInvalidPublicKey(t *testing.T) {
	nickData := makeValidNickData()
	nickData.PublicKey[0] = 0
	nickData = withValidSignature(nickData)

	if err := nickData.Validate(); err == nil {
		t.Fatal("expected an error")
	} else {
		t.Log(err)
		if !strings.Contains(err.Error(), "read the public key") {
			t.Fatal(err)
		}
	}
}

func TestNickDataValidateMissingSignature(t *testing.T) {
	nickData := makeValidNickData()
	nickData.Signature = nil

	if err := nickData.Validate(); err == nil {
		t.Fatal("expected an error")
	} else {
		t.Log(err)
		if !strings.Contains(err.Error(), "signature") {
			t.Fatal(err)
		}
	}
}

func TestNickDataValidateInvalidSignature(t *testing.T) {
	nickData := makeValidNickData()
	nickData.Signature[0] = 0

	if err := nickData.Validate(); err == nil {
		t.Fatal("expected an error")
	} else {
		t.Log(err)
		if !strings.Contains(err.Error(), "signature") {
			t.Fatal(err)
		}
	}
}

type cleanupFunc func()

func makeBoltRepository(t *testing.T) (*BoltRepository, cleanupFunc) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}

	dirCleanup := func() {
		os.RemoveAll(dir)
	}

	boltDatabasePath := filepath.Join(dir, "database.bolt")

	b, err := NewBoltRepository(boltDatabasePath)
	if err != nil {
		dirCleanup()
		t.Fatal(err)
	}

	cleanup := func() {
		b.Close()
		dirCleanup()
	}
	return b, cleanup
}

func TestBoltRepositoryGetEmpty(t *testing.T) {
	b, cleanup := makeBoltRepository(t)
	defer cleanup()

	iden := makeIdentity()

	result, err := b.Get(iden.Id)

	if result != nil {
		t.Fatalf("result should be nil, got: %s", result)
	}

	if err != nil {
		t.Fatalf("get error should be nil, got: %s", err)
	}
}

func TestBoltRepositoryPutGet(t *testing.T) {
	b, cleanup := makeBoltRepository(t)
	defer cleanup()

	nickData := makeValidNickData()

	if err := b.Put(nickData); err != nil {
		t.Fatalf("put error should be nil, got: %s", err)
	}

	if data, err := b.Get(nickData.Id); err != nil {
		t.Fatalf("get error should be nil, got: %s", err)
	} else {
		if err := data.Validate(); err != nil {
			t.Fatalf("retrieved data should be valid, got: %s", err)
		}
	}
}

func TestBoltRepositoryPutInvalid(t *testing.T) {
	b, cleanup := makeBoltRepository(t)
	defer cleanup()

	nickData := makeValidNickData()
	nickData.Nick = ""

	if err := b.Put(nickData); err != InvalidNickDataErr {
		t.Fatalf("expected %s, got: %s", InvalidNickDataErr, err)
	}
}

func TestBoltRepositoryPutOlder(t *testing.T) {
	b, cleanup := makeBoltRepository(t)
	defer cleanup()

	nickData := makeValidNickData()
	nickData.Time = time.Date(1990, 1, 1, 1, 1, 1, 1, time.UTC)
	nickData = withValidSignature(nickData)

	if err := b.Put(nickData); err != nil {
		t.Fatalf("put error: %s", err)
	}

	nickData = makeValidNickData()
	nickData.Time = time.Date(1989, 1, 1, 1, 1, 1, 1, time.UTC)
	nickData = withValidSignature(nickData)

	if err := b.Put(nickData); err != NewerNickDataPresentErr {
		t.Fatalf("expected %s, got: %s", NewerNickDataPresentErr, err)
	}
}

func TestBoltRepositoryListEmpty(t *testing.T) {
	// given
	b, cleanup := makeBoltRepository(t)
	defer cleanup()

	// when
	result, err := b.List()

	// tehn
	require.NoError(t, err, "error should be nil")
	require.NotNil(t, result, "result should not be nil")
	require.Empty(t, result, "result should be an empty slice")
}

func TestBoltRepositoryListOneElement(t *testing.T) {
	// given
	b, cleanup := makeBoltRepository(t)
	defer cleanup()

	nickData := makeValidNickData()

	// when
	err := b.Put(nickData)
	require.NoError(t, err, "put should not fail")

	result, err := b.List()

	// tehn
	require.NoError(t, err, "error should be nil")
	require.Equal(t, 1, len(result), "shouild return a single result")
}
