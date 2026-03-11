package auth

import "net/http"

type AuthStyle int

const (
	AuthStyleBearer AuthStyle = iota
	AuthStyleBasic
)

type Provider interface {
	AuthStyle() AuthStyle
	Token() (string, error)
	ApplyAuth(req *http.Request) error
}
