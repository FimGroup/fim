package source

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
)

type PluginInitializer struct {
	authJwtPlugin *AuthJwtPlugin
}

func (p PluginInitializer) InjectRouter(r chi.Router) {
	if p.authJwtPlugin != nil {
		r.Use(p.authJwtPlugin.ChiMiddleware()...)
	}
}

func InitializePlugin(options map[string]string) (PluginInitializer, error) {
	p := PluginInitializer{}

	// auth jwt module
	{
		enableAuthJwt, ok := options["http.plugin.authjwt.enable"]
		if ok && enableAuthJwt == "true" {
			if plugin, err := NewAuthJwtPlugin(options); err != nil {
				return p, err
			} else {
				p.authJwtPlugin = plugin
			}
		}
	}

	return p, nil
}

type CorsPlugin struct {
	//TODO 1. handling options request
}

const (
	ConfJwtAlg       = "auth.jwt.alg"
	ConfJwtSignKey   = "auth.jwt.signkey"
	ConfJwtVerifyKey = "auth.jwt.verifykey"
)

type AuthJwtPlugin struct {
	auth *jwtauth.JWTAuth
	// builtin plugin -> hardcoded in http connector
	// 1. http header validation & key loading from resource manager(configure)
	// 2. Unauthorized response: http status code and body(no body currently)
	// 3. in/out parameter
}

func NewAuthJwtPlugin(options map[string]string) (*AuthJwtPlugin, error) {
	var alg string
	var signKey string
	if err := MandatoryOptions(options, []OptionReq{
		{ConfJwtAlg, &alg},
		{ConfJwtSignKey, &signKey},
	}); err != nil {
		return nil, err
	}
	var verifyKey string
	OptionalOptions(options, []OptionReq{
		{ConfJwtVerifyKey, &verifyKey},
	})

	if len(alg) == 0 {
		return nil, errors.New("auth.jwt.alg is emtpy")
	}
	var verifyKeyData interface{}
	if len(verifyKey) > 0 {
		verifyKeyData = []byte(verifyKey)
	} else {
		verifyKeyData = nil
	}

	tokenAuth := jwtauth.New(alg, []byte(signKey), verifyKeyData)

	return &AuthJwtPlugin{
		auth: tokenAuth,
	}, nil
}

func (a *AuthJwtPlugin) ChiMiddleware() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		jwtauth.Verifier(a.auth),
		jwtauth.Authenticator,
	}
}

func (a *AuthJwtPlugin) JwtAuthData(r *http.Request) (interface{}, error) {
	token, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return nil, err
	}
	var _ = token
	return claims, nil
}

type OptionReq struct {
	Key string
	Val *string
}

func MandatoryOptions(options map[string]string, pairs []OptionReq) error {
	for _, v := range pairs {
		val, ok := options[v.Key]
		if !ok {
			return errors.New("option=" + v.Key + " is not found")
		}
		*v.Val = val
	}
	return nil
}

func OptionalOptions(options map[string]string, pairs []OptionReq) {
	for _, v := range pairs {
		val, ok := options[v.Key]
		if ok {
			*v.Val = val
		}
	}
}
