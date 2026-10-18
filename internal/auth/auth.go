package auth

import (
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"time"
	"github.com/google/uuid"
	"fmt"
	"net/http"
	"strings"
)

/*
	Hash the password using the argon2id.CreateHash function.
*/
func HashPassword(password string) (string, error){
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}


/*
	Use the argon2id.ComparePasswordAndHash function to compare the password that the user entered in the HTTP request
	with the password that is stored in the database.
*/
func CheckPasswordHash(password, hash string) (bool, error) {
	passwordMatch, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return passwordMatch, nil
}

/*
	Generate a JWT token for the user with the given userID, tokenSecret, and expiration time.
	Use jwt.NewWithClaims
	Use jwt.SigningMethodHS256 as the signing method.
	Use jwt.RegisteredClaims as the claims.
	Set the Issuer to "chirpy-access"
	Set IssuedAt to the current time in UTC
	Set ExpiresAt to the current time plus the expiration time (expiresIn)
	Set the Subject to a stringified version of the user's id
	Use token.SignedString to sign the token with the secret key. Refer to the docs for an overview of the different signing methods and their respective key types.
*/
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	// Create a d the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
			"user_id": userID.String(),
			"exp":     time.Now().Add(expiresIn).Unix(),
	})

	register := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	}

	token.Claims = register

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString([]byte(tokenSecret))
}

/*
	Use the jwt.ParseWithClaims function to validate the signature of the JWT and extract the claims into a *jwt.Token struct. 
	The keyFunc callback must return the same key type ([]byte) used when the token was signed. An error will be returned if 
	the token is invalid or has expired.
	If all is well with the token, use the token.Claims interface to get access to the user's id from the claims 
	(which should be stored in the Subject field). Return the id as a uuid.UUID.
*/
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error){
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	if !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid claims")
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

/*
	Extract the Bearer token from the Authorization header of the HTTP request.
	Use the http.Header.Get method to get the value of the Authorization header.
	If the header is missing, return an error.
	If the header does not start with "Bearer ", return an error.
	Use strings.TrimPrefix to remove the "Bearer " prefix from the header value and return the token.
*/
func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization header is missing")
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("Invalid authorization header")
	}
	return strings.TrimPrefix(authHeader, "Bearer "), nil
}