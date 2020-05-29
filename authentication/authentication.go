package authentication

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"net/http"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
)

// ApplicationClaim represents the claims related to an application
// composed of either NamespaceID or ApplicationID of the linked JWT
type ApplicationClaim struct {
	NamespaceID   string `json:"namespace_id"`
	ApplicationID string `json:"application_id"`
}

// Claims represents a custom JWT claims with a list of applications
type Claims struct {
	ApplicationsClaims []ApplicationClaim `json:"application_claim"`
	jwt.StandardClaims
}

var (
	errorInvalidClaims      = errors.New("invalid claims")
	errorInvalidPublicKey   = errors.New("invalid public key")
	errorEmptyRequestToken  = errors.New("authentication token was not provided in the request")
	errorInvalidApplication = errors.New("application ID was not provided")
	errorInvalidNamespace   = errors.New("namespace ID was not provided")
)

// ENV should not change during runtime
var (
	isPublicFunction bool
	publicKey        *rsa.PublicKey
	applicationID    string
	namespaceID      string
)

func init() {
	initEnv()
}

func initEnv() {
	isPublicFunction = os.Getenv("SCW_PUBLIC") == "true"
	applicationID = os.Getenv("SCW_APPLICATION_ID")
	namespaceID = os.Getenv("SCW_NAMESPACE_ID")

	publicKeyPem := os.Getenv("SCW_PUBLIC_KEY")
	if publicKeyPem == "" {
		return
	}

	block, _ := pem.Decode([]byte(publicKeyPem))
	if block == nil {
		return
	}

	parsedKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		// Print additional error
		log.Print(err.Error())
		return
	}
	publicKey = parsedKey
}

// Authenticate incoming request based on multiple factors:
// - 1: Whether the function's privacy has been set to private, if public, just leave this middleware
// - 2: Get the public key injected in this function runtime (done automatically by Scaleway)
// - 3: Check whether a Token has been sent via a specific Headers reserved by Scaleway
// - 4: Parse the incoming JWT with the public key
// - 5: Check the "Application Claims" linked to the JWT
// - 6: Both FunctionID and NamespaceID are injected via environment variables by Scaleway
// ---  so we have to check the authenticity of the incoming token by comparing the claims
func Authenticate(req *http.Request) (err error) {
	if isPublicFunction {
		return
	}

	// Check that request holds an authentication token
	requestToken := req.Header.Get("SCW_FUNCTIONS_TOKEN")
	if requestToken == "" {
		return errorEmptyRequestToken
	}

	if publicKey == nil {
		return errorInvalidPublicKey
	}

	// Parse JWT and retrieve claims
	claims := &Claims{}
	_, err = jwt.ParseWithClaims(requestToken, claims, func(*jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if err != nil {
		return
	}

	if len(claims.ApplicationsClaims) == 0 {
		return errorInvalidClaims
	}
	applicationClaims := claims.ApplicationsClaims[0]

	if applicationID == "" {
		return errorInvalidApplication
	} else if namespaceID == "" {
		return errorInvalidNamespace
	}

	// Check that the token's claims match with the injected Application or Namespace ID (depending on the scope of the token)
	if applicationClaims.NamespaceID != namespaceID && applicationClaims.ApplicationID != applicationID {
		err = errorInvalidClaims
	}
	return
}
