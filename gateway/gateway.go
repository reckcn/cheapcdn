package gateway

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ikeikeikeike/cheapcdn/lib"
	"github.com/labstack/echo"
)

var (
	h403  = http.StatusBadRequest
	rbyte = make([]byte, 1)
)

type (
	// Object is
	Object struct {
		Name   string `json:"name" form:"name" query:"name" validate:"required"`
		Object string `json:"object" form:"object" query:"object"`
	}
)

func (o *Object) buildToken(ctx echo.Context) (string, error) {
	rand.Read(rbyte)

	m := map[string]string{
		"i": ctx.RealIP(),
		"t": time.Now().UTC().Format(lib.TF),
		"_": fmt.Sprintf("%x", rbyte),
	}
	if o.Object != "" {
		m["f"] = o.Object
	}

	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	return lib.EncryptAexHex(data), nil
}

func gateway(ctx echo.Context) (err error) {
	o := new(Object)
	if err = ctx.Bind(o); err != nil {
		return ctx.String(h403, "Bad Request")
	}
	if err = valid.Struct(o); err != nil {
		return ctx.String(h403, "Bad Request")
	}

	token, err := o.buildToken(ctx)
	if err != nil {
		return ctx.String(h403, "Bad Request")
	}

	return ctx.String(http.StatusOK, token)
}