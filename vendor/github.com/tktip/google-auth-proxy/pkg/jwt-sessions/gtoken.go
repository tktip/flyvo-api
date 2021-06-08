package jwtsessions

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var (
	// ErrTokenExpired indicates an expired token
	ErrTokenExpired = errors.New("Expired or invalid token. No way to continue")
)

//JwtConfig - jwt config
type JwtConfig struct {
	Client struct {
		ID     string
		Secret string
	}

	Issuer string
}

// GToken contains the user's google oauth2-token and info about the user such as email and name
type GToken struct {
	Profile       string `json:"profile" mapstructure:"profile"`
	Email         string `json:"email" mapstructure:"email"`
	EmailVerified bool   `json:"email_verified" mapstructure:"email_verified"`

	FamilyName string `json:"family_name" mapstructure:"family_name"`
	GivenName  string `json:"given_name" mapstructure:"given_name"`

	OAuthAccessToken string `json:"gauth-access" mapstructure:"gauth-access"`
	OAuthExp         int64  `json:"gauth-access-exp" mapstructure:"gauth-access-exp"`
	JTI              string `json:"jti" mapstructure:"jti"`
}

//AsJSON - converts a person to json PANICS IS MARSHAL FAILS
func (tok *GToken) AsJSON() string {
	data, err := json.Marshal(tok)
	if err != nil {
		panic(err.Error())
	}
	return string(data)
}

// OAuthTokenExpired returns true if OAuthExp is expired
func (tok *GToken) OAuthTokenExpired() (isExpired bool, err error) {
	t := time.Unix(tok.OAuthExp, 0)
	if t.IsZero() { // invalid expiry
		isExpired = true
		return
	}

	return time.Now().After(t), nil
}

// CreateCookie fetches data about the user based on the token, and creates a jwt token cookie
// containing both user data and the google oauth2-token
func (tok *GToken) CreateCookie(h *Handler) (
	cook http.Cookie, err error,
) {

	claims := jwt.MapClaims{
		"gauth": tok,
		"jti":   uuid.New().String(),
		"exp":   time.Now().Add(time.Minute * 10).Unix(),
	}
	t := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), claims)
	tokenString, err := t.SignedString(h.signKey)
	cook = http.Cookie{
		Name:     h.Cookie.Name,
		Value:    tokenString,
		Secure:   h.Cookie.Secure,
		HttpOnly: h.Cookie.HTTPOnly,
	}

	return
}
