package minio

import (
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"time"

	"github.com/ikeikeikeike/cheapcdn/lib"
	"github.com/labstack/echo"
)

const (
	authScheme = "Bearer"
	parameter  = "cheapkey"
)

func keyAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// Extract and verify key
			key, err := extractor(ctx)
			if err != nil {
				return echo.NewHTTPError(http.StatusForbidden)
			}
			if !validator(ctx, key) {
				return echo.ErrUnauthorized
			}

			return next(ctx)
		}
	}
}

// _, b := ctx.(*lib.CacheContext).Store.Get(key)
// if !b { return false }
func validator(ctx echo.Context, key string) bool {
	var st struct {
		File string `json:"f"`
		IP   string `json:"i"`
		Time string `json:"t"`
	}

	bytes := lib.DecryptAexHex(key)
	if err := json.Unmarshal(bytes, &st); err != nil {
		return false
	}

	if st.File != filepath.Base(ctx.Request().URL.Path) {
		return false
	}
	if st.IP != ctx.RealIP() {
		return false
	}
	t1, err := time.Parse(lib.TF, st.Time)
	if err != nil {
		return false
	}
	t2 := time.Now().UTC()
	if t1.Add(1*time.Hour).UnixNano() < t2.UnixNano() {
		return false
	}

	return true
}

func extractor(ctx echo.Context) (string, error) {
	token, err := extractHeader(ctx)
	if err == nil {
		return token, nil
	}

	token, err = extractParam(ctx)
	if err == nil {
		return token, nil
	}

	return "", errors.New("Missing token")
}

func extractHeader(ctx echo.Context) (string, error) {
	auth := ctx.Request().Header.Get(echo.HeaderAuthorization)
	if auth == "" {
		return "", errors.New("Missing token in request header")
	}

	length := len(authScheme)

	if len(auth) > length+1 && auth[:length] == authScheme {
		ctx.Request().Header.Del(echo.HeaderAuthorization)
		return auth[length+1:], nil
	}

	return "", errors.New("Invalid token in the request header")
}

func extractParam(ctx echo.Context) (string, error) {
	key := ctx.QueryParam(parameter)
	if key == "" {
		return "", errors.New("Missing token in the query string")
	}

	return key, nil
}
