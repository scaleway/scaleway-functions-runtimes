package authentication

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	fixturePrivateKey       *rsa.PrivateKey
	fixturePublicKey        *rsa.PublicKey
	fixturePublicKeyEncoded string
	fixtureTokenApplication string
	fixtureTokenNamespace   string
	fixtureApplicationID    = "app-id"
	fixtureNamespaceID      = "namespace-id"
	fixtureIssuer           = "scaleway"
	fixtureSubject          = "token"
	fixtureService          = "functions"
	fixtureExpirationDate   = time.Now().Add(time.Hour)
)

// ==== Test Set Up - Initialize public key, and generate test token ==== //

func TestMain(m *testing.M) {
	setUpPublicKey()
	setUpTestToken()
	os.Exit(m.Run())
}

func setUpPublicKey() {
	var err error
	fixturePrivateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatalf("Unable to generate private key, got error: %v", err)
	}
	fixturePublicKey = &fixturePrivateKey.PublicKey

	var pemPrivateBlock = &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(fixturePublicKey),
	}
	var buffer bytes.Buffer
	err = pem.Encode(&buffer, pemPrivateBlock)
	if err != nil {
		log.Fatalf("Unable to encode public key, got error: %v", err)
	}

	fixturePublicKeyEncoded = buffer.String()
}

func setUpTestToken() {
	var err error
	functionClaims := []ApplicationClaim{
		{
			ApplicationID: fixtureApplicationID,
		},
	}

	namespaceClaim := []ApplicationClaim{
		{
			NamespaceID: fixtureNamespaceID,
		},
	}

	appClaims := Claims{
		functionClaims,
		jwt.StandardClaims{
			Issuer:    fixtureIssuer,
			Subject:   fixtureSubject,
			Audience:  fixtureService,
			ExpiresAt: fixtureExpirationDate.Unix(),
			NotBefore: time.Now().Unix(),
			IssuedAt:  time.Now().Unix(),
			Id:        "test",
		},
	}

	namespaceClaims := Claims{
		namespaceClaim,
		jwt.StandardClaims{
			Issuer:    fixtureIssuer,
			Subject:   fixtureSubject,
			Audience:  fixtureService,
			ExpiresAt: fixtureExpirationDate.Unix(),
			NotBefore: time.Now().Unix(),
			IssuedAt:  time.Now().Unix(),
			Id:        "test-namespace",
		},
	}

	tokenApplication := jwt.NewWithClaims(jwt.SigningMethodRS256, appClaims)
	tokenNamespace := jwt.NewWithClaims(jwt.SigningMethodRS256, namespaceClaims)
	// sign token
	fixtureTokenApplication, err = tokenApplication.SignedString(fixturePrivateKey)
	if err != nil {
		log.Fatalf("Unable to sign application test token, got error: %v", err)
	}
	fixtureTokenNamespace, err = tokenNamespace.SignedString(fixturePrivateKey)
	if err != nil {
		log.Fatalf("Unable to sign namespace test token, got error: %v", err)
	}
}

// ==== Test Helpers ==== //

func setUpEnvironmentVariables() {
	os.Setenv("SCW_PUBLIC", "false")
	os.Setenv("SCW_PUBLIC_KEY", fixturePublicKeyEncoded)
	os.Setenv("SCW_APPLICATION_ID", fixtureApplicationID)
	os.Setenv("SCW_NAMESPACE_ID", fixtureNamespaceID)
}

func setUpAndTestAuthentication(token string) error {
	setUpEnvironmentVariables()
	return testAuthentication(token)
}

func testAuthentication(token string) error {
	req := &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("SCW_FUNCTIONS_TOKEN", token)
	return Authenticate(req)
}

// ==== Test ==== //

func TestAuthenticate(t *testing.T) {
	t.Run("function is public", func(t *testing.T) {
		os.Setenv("SCW_PUBLIC", "true")
		req := &http.Request{}
		if err := Authenticate(req); err != nil {
			t.Errorf("Authenticate(), received error %v", err)
		}
	})

	t.Run("request token not provided", func(t *testing.T) {
		os.Setenv("SCW_PUBLIC", "false")
		req := &http.Request{}
		if err := Authenticate(req); err != errorEmptyRequestToken {
			t.Errorf("Authenticate(), received error %v, expected %v", err, errorEmptyRequestToken)
		}
	})

	t.Run("missing public key", func(t *testing.T) {
		os.Setenv("SCW_PUBLIC", "false")
		req := &http.Request{
			Header: http.Header{},
		}
		req.Header.Set("SCW_FUNCTIONS_TOKEN", "test-token")
		if err := Authenticate(req); err != errorInvalidPublicKey {
			t.Errorf("Authenticate(), received error %v, expected %v", err, errorInvalidPublicKey)
		}
	})

	t.Run("invalid public key", func(t *testing.T) {
		os.Setenv("SCW_PUBLIC", "false")
		os.Setenv("SCW_PUBLIC_KEY", "invalid public key")
		req := &http.Request{
			Header: http.Header{},
		}
		req.Header.Set("SCW_FUNCTIONS_TOKEN", "test-token")
		if err := Authenticate(req); err != errorInvalidPublicKey {
			t.Errorf("Authenticate(), received error %v, expected %v", err, errorInvalidPublicKey)
		}
	})

	t.Run("valid authentication for Application ID", func(t *testing.T) {
		if err := setUpAndTestAuthentication(fixtureTokenApplication); err != nil {
			t.Errorf("Authenticate(), received error %v", err)
		}
	})

	t.Run("valid authentication for Namespace ID", func(t *testing.T) {
		if err := setUpAndTestAuthentication(fixtureTokenNamespace); err != nil {
			t.Errorf("Authenticate(), received error %v", err)
		}
	})

	t.Run("claims do not match injected application ID", func(t *testing.T) {
		setUpEnvironmentVariables()
		os.Setenv("SCW_APPLICATION_ID", "another-app-id")
		if err := testAuthentication(fixtureTokenApplication); err != errorInvalidClaims {
			t.Errorf("Authenticate(), got error %v, expected %v", err, errorInvalidClaims)
		}
	})

	t.Run("claims do not match injected namespace ID", func(t *testing.T) {
		setUpEnvironmentVariables()
		os.Setenv("SCW_NAMESPACE_ID", "another-namespace-id")
		if err := testAuthentication(fixtureTokenNamespace); err != errorInvalidClaims {
			t.Errorf("Authenticate(), got error %v, expected %v", err, errorInvalidClaims)
		}
	})
}
