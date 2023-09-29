package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"webApp/pkg/data"

	"github.com/golang-jwt/jwt/v4"
)

var jwtTokenExpiry = time.Minute * 15
var refreshTokenExpiry = time.Hour * 24

type TokenPairs struct {
	Token string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	Username string `json:"name"`
	jwt.RegisteredClaims
}

func (app *application) getTokenFromHeaderandVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {
	// add a header
	w.Header().Add("Vary", "Authorization")

	// get the authorization header
	authHeader := r.Header.Get("Authorization")

	// sanity check
	if authHeader == "" {
		return "", nil, errors.New("no auth header")
	}

	// split the header on spaces
	headerPars := strings.Split(authHeader, " ")
	if len(headerPars) != 2 {
		return "", nil, errors.New("invalid auth header")
	}

	// check to see if we have the word "Bearer"
	if headerPars[0] != "Bearer" {
		return "", nil, errors.New("unauthorized: no Bearer")
	}

	token := headerPars[1]

	// declare an empty Claims variable
	claims := &Claims{}

	// parse the token with our claims (we read into claims), using our secret (from the receiver)
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		// validate the signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil,fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(app.JWTSecret), nil
	})

	// check for an error: note that this catches expired tokens as well
	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return "", nil, errors.New("expired token")
		}
		return "", nil, err
	}

	// make sure that we issued this token
	if claims.Issuer != app.Domain {
		return "", nil, errors.New("incorrect issuer")
	}

	// valid tokens
	return token, claims, nil
}

func (app *application) generateTokenPair(user *data.User) (TokenPairs, error) {
	// create the token.
	token := jwt.New(jwt.SigningMethodHS256)

	// set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = fmt.Sprint(user.ID)
	claims["aud"] = app.Domain
	claims["iss"] = app.Domain
	if user.IsAdmin == 1 {
		claims["admin"] = true
	} else {
		claims["admin"] = false
	}

	// set the expiry
	claims["exp"] = time.Now().Add(jwtTokenExpiry).Unix()

	// create the signed token
	signedAccessToken, err := token.SignedString([]byte(app.JWTSecret))
	if err != nil {
		return TokenPairs{}, err
	}

	// create the refresh token
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["sub"] = fmt.Sprint(user.ID)
	// set expiry: must longer than hwt expiry
	refreshTokenClaims["exp"] = time.Now().Add(refreshTokenExpiry).Unix()

	// create signedd refresh token
	signedRefreshToken, err := refreshToken.SignedString([]byte(app.JWTSecret))
	if err != nil {
		return TokenPairs{}, err
	}

	var TokenPairs = TokenPairs{
		Token: signedAccessToken,
		RefreshToken: signedRefreshToken,
	}

	return TokenPairs, nil
}