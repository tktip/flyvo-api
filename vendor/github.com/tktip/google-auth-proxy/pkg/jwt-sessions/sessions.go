package jwtsessions

import (
	"context"
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2"

	"crypto/rsa"
	"errors"
	"net/http"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/tktip/google-auth-proxy/internal/datastore"
	"github.com/tktip/google-auth-proxy/internal/httperr"
	"github.com/tktip/google-auth-proxy/internal/state"
)

var (
	// ErrRefreshTokenNotFound indicates a user authenticated without getting a refreshtoken, and
	// without having a refreshtoken stored in our database
	ErrRefreshTokenNotFound = errors.New("no refreshtoken found for user")
)

//Handler - session handler object
type Handler struct {
	signKey   *rsa.PrivateKey
	verifyKey *rsa.PublicKey

	PrivateKey string `yaml:"privateKey"`
	PublicKey  string `yaml:"publicKey"`

	Cookie struct {
		HeaderName string `yaml:"headerName"`
		Name       string `yaml:"name"`
		Secure     bool   `yaml:"secure"`
		HTTPOnly   bool   `yaml:"httpOnly"`
	} `yaml:"cookie"`

	RefreshCookie struct {
		Name     string `yaml:"name"`
		Secure   bool   `yaml:"secure"`
		HTTPOnly bool   `yaml:"httpOnly"`
		Path     string `yaml:"path"`
	} `yaml:"refreshCookie"`

	Datastore datastore.Client `yaml:"datastore"`

	State state.Handler `yaml:"state"`
}

//Init - initialize values
func (h *Handler) Init() (err error) {
	signBytes, err := ioutil.ReadFile(h.PrivateKey)
	if err != nil {
		return fmt.Errorf("read private key path: %v", err)
	}

	if h.signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes); err != nil {
		return fmt.Errorf("parse private key: %v", err)
	}

	verifyBytes, err := ioutil.ReadFile(h.PublicKey)
	if err != nil {
		return fmt.Errorf("read public key path: %v", err)
	}

	if h.verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes); err != nil {
		return fmt.Errorf("parse public key: %v", err)
	}

	if err = h.Datastore.Init(); err != nil {
		return fmt.Errorf("init datastore client: %v", err)
	}

	return nil
}

// CreateCookies fetches data about the user based on the given token, and creates two jwt token
// cookies, one with the access token, and another containing either the refresh token in the given
// oauth2-token or the refresh-token stored on the user's email
func (h *Handler) CreateCookies(tok *oauth2.Token, provider *oidc.Provider) (
	accessCookie http.Cookie, refreshCookie http.Cookie, err error,
) {
	userInfo, err := provider.UserInfo(context.TODO(), oauth2.StaticTokenSource(tok))
	if err != nil {
		err = fmt.Errorf("Failed to get userinfo: %v", err.Error())
		return
	}

	details := make(map[string]string)
	userInfo.Claims(&details)
	accessToken := GToken{
		Profile:          userInfo.Profile,
		Email:            userInfo.Email,
		EmailVerified:    userInfo.EmailVerified,
		FamilyName:       details["family_name"],
		GivenName:        details["given_name"],
		OAuthAccessToken: tok.AccessToken,
		OAuthExp:         tok.Expiry.Unix(),
	}

	accessCookie, err = accessToken.CreateCookie(h)
	if err != nil {
		err = fmt.Errorf("creating accessCookie: %v", err)
		return
	}

	refreshCookie, err = h.createRefreshCookie(tok.RefreshToken, userInfo.Email)
	if err != nil {
		err = fmt.Errorf("creating accessCookie: %v", err)
		return
	}

	return
}

func (h *Handler) createRefreshCookie(refreshTokenString string, userEmail string) (
	cook http.Cookie, err error,
) {
	if refreshTokenString == "" {
		// user did not receive a refresh token from google, this means we should have one stored
		var storedRefreshToken datastore.RefreshToken
		storedRefreshToken, err = h.Datastore.GetRefreshToken(context.TODO(), userEmail)
		if err == datastore.ErrNoSuchEntity {
			err = ErrRefreshTokenNotFound
			log.Warnf("no refresh token found for user %s", userEmail)
			return
		}
		if err != nil {
			err = fmt.Errorf("error occurred finding user's refresh-token: %v", err)
			return
		}
		// if stored refreshtoken has non zero expiration that is before now
		if !storedRefreshToken.Expires.IsZero() && storedRefreshToken.Expires.Before(time.Now()) {
			err = ErrRefreshTokenNotFound
			log.Warnf("refresh token found for user %s was non zero and expired", userEmail)
			return
		}
		refreshTokenString = storedRefreshToken.TokenString
	} else {
		err = h.Datastore.PutRefreshToken(context.TODO(), userEmail, datastore.RefreshToken{
			TokenString: refreshTokenString,
			Created:     time.Now(),
			Expires:     time.Time{},
		})
		if err != nil {
			log.Errorf("error storing refresh-token: %v", err)
			err = nil
			// TODO: figure out what to do in this situation
			// this is a problem because we have to figure out how to ask google for a new refresh
			// token the next time the user signs in
			return
		}
	}
	claims := jwt.MapClaims{
		"refresh": refreshTokenString,
	}
	t := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), claims)
	tokenString, err := t.SignedString(h.signKey)
	cook = http.Cookie{
		Name:     h.RefreshCookie.Name,
		Value:    tokenString,
		Secure:   h.RefreshCookie.Secure,
		HttpOnly: h.RefreshCookie.HTTPOnly,
		Path:     h.RefreshCookie.Path,
	}

	return
}

// findClaims finds the user's jwt token and then the token's claims.
// Returns an error if the tokens fails parsing or is not valid (due to expiration)
func (h *Handler) findValidClaims(r *http.Request) (claims jwt.MapClaims, hErr httperr.Error) {
	// check if we have a cookie with out tokenName
	tokenCookie, err := r.Cookie(h.Cookie.Name)
	if err != nil {
		hErr = httperr.Newf(http.StatusUnauthorized, "get cookie in GetSession: %v", err)
		return
	}

	token, err := jwt.Parse(tokenCookie.Value, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		if token.Method.Alg() != jwt.SigningMethodRS256.Name {
			return nil, errors.New("invalid signing method")
		}
		return h.verifyKey, nil
	})

	if err != nil {
		hErr = httperr.NewText(
			http.StatusUnauthorized,
			fmt.Errorf("parsing jwt token: %v", err),
			"error occurred parsing jwt token",
		)
		return
	}
	if !token.Valid { // may still be invalid
		hErr = httperr.Newf(http.StatusUnauthorized, "token is not valid")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		hErr = httperr.Newf(http.StatusUnauthorized, "token claims type assertion failed")
		return
	}
	return
}

// GetSession returns data on a user's JWT session.
func (h *Handler) GetSession(r *http.Request) (tok GToken, hErr httperr.Error) {
	claims, hErr := h.findValidClaims(r)
	if hErr != nil {
		hErr = hErr.WithCall("find claims")
		return
	}

	gauthEntry, ok := claims["gauth"]
	if !ok {
		hErr = httperr.Newf(http.StatusUnauthorized, "get gauth claim: gauth claim not present")
		return
	}

	err := mapstructure.Decode(gauthEntry, &tok)
	if err != nil {
		hErr = httperr.NewText(
			http.StatusUnauthorized,
			fmt.Errorf("decode token gauth claim: %v", err),
			"invalid token: failed decoding gauth claim value")
		return
	}

	jtiInter, exists := claims["jti"]
	if !exists || jtiInter == nil {
		hErr = httperr.Newf(http.StatusUnauthorized,
			"invalid token: no jti claim value")
		return
	}

	jti, jtiOfTypeString := jtiInter.(string)
	if !exists || !jtiOfTypeString || jti == "" {
		hErr = httperr.Newf(http.StatusUnauthorized, "parsing claims: empty or invalid jti")
		return
	}

	isBlacklisted, err := h.State.IsBlacklisted(jti)
	if isBlacklisted {
		hErr = httperr.Newf(http.StatusUnauthorized, "check blacklisted: user is blacklisted")
		err = errors.New("Token is invalid")
		return
	}

	tok.JTI = jti
	return
}

// GetRefreshToken validates and returns a user's refresh-token string
func (h *Handler) GetRefreshToken(r *http.Request) (tokenString string, err error) {
	// check if we have a cookie with out tokenName
	tokenCookie, err := r.Cookie(h.RefreshCookie.Name)
	if err != nil {
		err = fmt.Errorf("get cookie in GetRefreshToken: %v", err)
		return
	}
	// TODO decide: should we blacklist refresh tokens?

	token, err := jwt.Parse(tokenCookie.Value, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return h.verifyKey, nil
	})

	if err != nil {
		err = fmt.Errorf("error parsing jwt: %v", err)
		return
	}
	if !token.Valid { // but may still be invalid
		err = errors.New("token was invalid")
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	tokenVal, ok := claims["refresh"]
	if !ok {
		err = errors.New("invalid token: refresh claim not present")
		return
	}

	tokenString, ok = tokenVal.(string)
	if !ok {
		err = errors.New("failed type-asserting claim value to string")
	}

	return
}
