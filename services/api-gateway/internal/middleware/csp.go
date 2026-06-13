package middleware

import (
	"fmt"
	"net/http"
)

type CSPMiddleware struct {
	ImgSrc    string
	ScriptSrc string
	StyleSrc  string
}

func NewCSPMIddleware(
	imgSrc string,
	scriptSrc string,
	styleSrc string,
) *CSPMiddleware {
	return &CSPMiddleware{
		ImgSrc:    imgSrc,
		ScriptSrc: scriptSrc,
		StyleSrc:  styleSrc,
	}
}

func (m *appMiddlewareHub) CspMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := fmt.Sprintf("img-src '%s'; script-src '%s'; style-src '%s';", m.csp.ImgSrc, m.csp.ScriptSrc, m.csp.StyleSrc)
		w.Header().Set("Content-Security-Policy", value)
		next.ServeHTTP(w, r)
	})
}
