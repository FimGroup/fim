package source

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth"
)

type PluginInitializer struct {
	authJwtPlugin *AuthJwtPlugin
	corsPlugin    *CorsPlugin
}

func (p PluginInitializer) InjectRouter(r chi.Router) {
	//Note: order of each plugin
	if p.corsPlugin != nil {
		r.Use(p.corsPlugin.ChiMiddleware())
	}
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
	// cors module
	{
		enableCors, ok := options["http.plugin.cors.enable"]
		if ok && enableCors == "true" {
			if plugin, err := NewCorsPlugin(options); err != nil {
				return p, err
			} else {
				p.corsPlugin = plugin
			}
		}
	}

	return p, nil
}

const (
	ConfCorsAllowAll         = "cors.allow_all"
	ConfCorsOrigin           = "cors.allowed_origin"
	ConfCorsMethod           = "cors.allowed_method"
	ConfCorsHeader           = "cors.allowed_header"
	ConfCorsExposedHeader    = "cors.exposed_header"
	ConfCorsAllowCredentials = "cors.allow_credentials"
	ConfCorsMaxAge           = "cors.max_age"
)

type CorsPlugin struct {
	cors *cors.Cors
	// 1. handling options request
}

func NewCorsPlugin(options map[string]string) (*CorsPlugin, error) {
	var allowAll string
	var originString string
	var methodString string
	var headerString string
	var exposedHeaderString string
	var allowCredentialString string
	var maxAgeString string

	OptionalOptions(options, []OptionReq{
		{ConfCorsAllowAll, &allowAll},
		{ConfCorsOrigin, &originString},
		{ConfCorsMethod, &methodString},
		{ConfCorsHeader, &headerString},
		{ConfCorsExposedHeader, &exposedHeaderString},
		{ConfCorsAllowCredentials, &allowCredentialString},
		{ConfCorsMaxAge, &maxAgeString},
	})
	if v := strings.TrimSpace(allowAll); v == "true" {
		return &CorsPlugin{
			cors: cors.AllowAll(),
		}, nil
	}

	corsOption := cors.Options{}
	if v := strings.TrimSpace(originString); v != "" {
		items := strings.Split(originString, ",")
		for i, v := range items {
			items[i] = strings.TrimSpace(v)
		}
		corsOption.AllowedOrigins = items
	}
	if v := strings.TrimSpace(methodString); v != "" {
		items := strings.Split(methodString, ",")
		for i, v := range items {
			items[i] = strings.TrimSpace(v)
		}
		corsOption.AllowedMethods = items
	}
	if v := strings.TrimSpace(headerString); v != "" {
		items := strings.Split(headerString, ",")
		for i, v := range items {
			items[i] = strings.TrimSpace(v)
		}
		corsOption.AllowedHeaders = items
	}
	if v := strings.TrimSpace(exposedHeaderString); v != "" {
		items := strings.Split(exposedHeaderString, ",")
		for i, v := range items {
			items[i] = strings.TrimSpace(v)
		}
		corsOption.ExposedHeaders = items
	}
	if v := strings.TrimSpace(allowCredentialString); v == "true" {
		corsOption.AllowCredentials = true
	}
	if v := strings.TrimSpace(maxAgeString); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		corsOption.MaxAge = i
	}
	corsOption.Debug = true /////////////////////////

	return &CorsPlugin{
		cors: cors.New(corsOption),
	}, nil
}

func (c *CorsPlugin) ChiMiddleware() func(http.Handler) http.Handler {
	return c.cors.Handler
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
