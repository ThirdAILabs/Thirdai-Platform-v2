package deployment

import (
	"fmt"
	"net/http"
	"thirdai_platform/client"
	"thirdai_platform/model_bazaar/services"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

type Permissions struct {
	ModelBazaarEndpoint string
	ModelId             uuid.UUID
}

func (p *Permissions) GetModelPermissions(token string) (services.ModelPermissions, error) {
	client := client.NewModelClient(p.ModelBazaarEndpoint, token, p.ModelId)
	return client.GetPermissions()
}

func (p *Permissions) ModelPermissionsCheck(permission_type string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token := jwtauth.TokenFromHeader(r)
			if token != "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			modelPermissions, err := p.GetModelPermissions(token)
			if err != nil {
				http.Error(w, "Failed to retrieve permissions", http.StatusInternalServerError)
				return
			}

			hasPermission := (permission_type == "read" && modelPermissions.Read) || (permission_type == "write" && modelPermissions.Write)

			if hasPermission {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, fmt.Sprintf("not authorized for %v actions on model %v", permission_type, p.ModelId), http.StatusUnauthorized)
		}
		return http.HandlerFunc(hfn)
	}
}
