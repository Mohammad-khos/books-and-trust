package middleware

import (
	"books-and-trust/services/api-gateway/util"
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"
)

type AdminMiddleware struct {
	FilePath string
	Logger   *zap.SugaredLogger
	Admins   map[string]bool
}

func NewAdminMiddleware(path string, logger *zap.SugaredLogger) *AdminMiddleware {
	return &AdminMiddleware{
		FilePath: path,
		Logger:   logger,
	}
}

func (m *appMiddlewareHub) AdminsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userID, ok := r.Context().Value("user_id").(string)
		if !ok || userID == "" {
			util.ForbiddenErr(w, r, m.admin.Logger, "unauthorized - user id missing")
			return
		}
		testMode := true
		if !testMode && !m.admin.Admins[userID] {
			util.ForbiddenErr(w, r, m.admin.Logger, "forbidden - admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *AdminMiddleware) LoadAdmins() error {
	data, err := os.ReadFile(a.FilePath)
	if err != nil {
		return err
	}

	a.Admins = make(map[string]bool)
	admins := strings.Split(string(data), "\n")
	for _, line := range admins {
		id := strings.TrimSpace(line)
		if id != "" {
			a.Admins[id] = true
		}
	}
	return nil
}
