package auth

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"thirdai_platform/model_bazaar/schema"

	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KeycloakIdentityProvider struct {
	keycloak *gocloak.GoCloak
	db       *gorm.DB
	auditLog AuditLogger

	realm                        string
	adminUsername, adminPassword string
}

func isConflict(err error) bool {
	apiErr, ok := err.(*gocloak.APIError)
	// Keycloak returns 409 if user/realm etc already exists when creating it.
	return ok && apiErr.Code == http.StatusConflict
}

func pArg[T any](value T) *T {
	p := new(T)
	*p = value
	return p
}

var boolArg = pArg[bool]
var intArg = pArg[int]
var strArg = pArg[string]

func adminLogin(client *gocloak.GoCloak, adminUsername, adminPassword string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// The "master" realm is the default admin realm in Keycloak.
	adminToken, err := client.LoginAdmin(ctx, adminUsername, adminPassword, "master")
	if err != nil {
		return "", fmt.Errorf("error during keycloak admin login: %w", err)
	}
	return adminToken.AccessToken, nil
}

func getUserID(ctx context.Context, client *gocloak.GoCloak, adminToken, username, realmName string) (*string, error) {
	users, err := client.GetUsers(ctx, adminToken, realmName, gocloak.GetUsersParams{
		Username: &username,
		Max:      intArg(1),
		Exact:    boolArg(true),
	})
	if err != nil {
		return nil, fmt.Errorf("error retrieving user id: %w", err)
	}
	if len(users) == 1 {
		return users[0].ID, nil
	}
	return nil, nil
}

func createAdminIfNotExists(client *gocloak.GoCloak, adminToken, username, email, password, realmName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	existingUserId, err := getUserID(ctx, client, adminToken, username, realmName)
	if err != nil {
		return "", fmt.Errorf("error checking for existing admin : %w", err)
	}
	if existingUserId != nil {
		slog.Info("KEYCLOAK: admin user has already been created")
		return *existingUserId, nil
	}

	userId, err := client.CreateUser(ctx, adminToken, realmName, gocloak.User{
		Username:      &username,
		Email:         &email,
		Enabled:       boolArg(true),
		EmailVerified: boolArg(true),
		Credentials: &[]gocloak.CredentialRepresentation{
			{
				Type:      strArg("password"),
				Value:     &password,
				Temporary: boolArg(false),
			},
		},
	})

	if err != nil {
		if isConflict(err) {
			userId, err := getUserID(ctx, client, adminToken, username, realmName)
			slog.Info("KEYCLOAK: admin user has already been created")
			if err != nil {
				return "", fmt.Errorf("error retrieving existing admin after conflict creating admin: %w", err)
			}
			if userId == nil {
				return "", fmt.Errorf("no user found after conflict creating admin")
			}
			return *userId, nil
		}
		return "", fmt.Errorf("error creating new admin: %w", err)
	}

	return userId, nil
}

func assignAdminRole(client *gocloak.GoCloak, adminToken, userId, realm string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	roles, err := client.GetRealmRoles(ctx, adminToken, "master", gocloak.GetRoleParams{})
	if err != nil {
		return fmt.Errorf("error getting keycloak roles: %w", err)
	}
	for _, role := range roles {
		if *role.Name == "admin" {
			err := client.AddRealmRoleToUser(ctx, adminToken, "master", userId, []gocloak.Role{*role})
			if err != nil {
				return fmt.Errorf("error assigning admin role: %w", err)
			}
		}
	}
	return nil
}

func createRealm(client *gocloak.GoCloak, adminToken, realmName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	serverInfo, err := client.GetServerInfo(ctx, adminToken)
	if err != nil {
		return fmt.Errorf("error getting keycloak server info: %w", err)
	}

	args := gocloak.RealmRepresentation{
		Realm:                        &realmName,
		Enabled:                      boolArg(true),
		IdentityProviders:            &[]interface{}{},
		DefaultRoles:                 &[]string{"user"},
		RegistrationAllowed:          boolArg(true),
		ResetPasswordAllowed:         boolArg(true),
		AccessCodeLifespan:           intArg(1500),
		VerifyEmail:                  boolArg(true), // Require email verification for new users
		AccessTokenLifespan:          intArg(1500),  // Access token lifespan (in seconds)
		PasswordPolicy:               strArg("length(8) and digits(1) and lowerCase(1) and upperCase(1) and specialChars(1)"),
		BruteForceProtected:          boolArg(true),
		MaxFailureWaitSeconds:        intArg(900),
		MinimumQuickLoginWaitSeconds: intArg(60),
		WaitIncrementSeconds:         intArg(60),
		QuickLoginCheckMilliSeconds:  pArg(int64(1000)),
		MaxDeltaTimeSeconds:          intArg(43200),
		FailureFactor:                intArg(30),
		SMTPServer: &map[string]string{
			"host":     "smtp.sendgrid.net",
			"port":     "465",
			"from":     "platform@thirdai.com",
			"replyTo":  "platform@thirdai.com",
			"ssl":      "true",
			"starttls": "true",
			"auth":     "true",
			"user":     "apikey",
			"password": "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",
		},
	}

	if serverInfo.Themes != nil {
		for _, theme := range serverInfo.Themes.Login {
			if theme.Name == "custom-theme" {
				args.LoginTheme = strArg("custom-theme")
				args.AccountTheme = strArg("custom-theme")
				args.AdminTheme = strArg("custom-theme")
				args.EmailTheme = strArg("custom-theme")
				args.DisplayName = &realmName
				args.DisplayNameHTML = strArg("<div class='kc-logo-text'><span>Keycloak</span></div>")
			}
		}
	}

	_, err = client.CreateRealm(ctx, adminToken, args)
	if err != nil {
		if isConflict(err) {
			slog.Info(fmt.Sprintf("KEYCLOAK: realm '%v' has already been created", realmName))
			return nil // Ok if realm already exists
		}
		return fmt.Errorf("error creating realm: %w", err)
	}
	return nil
}

func createClient(client *gocloak.GoCloak, adminToken, realm string, redirectUrls []string, rootUrl string) error {
	clientName := "thirdai-platform-login"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	clients, err := client.GetClients(ctx, adminToken, realm, gocloak.GetClientsParams{
		ClientID: &clientName,
	})
	if err != nil {
		return fmt.Errorf("error listing existing clients for realm: %w", err)
	}
	if len(clients) == 1 {
		slog.Info(fmt.Sprintf("KEYCLOAK: client '%v' already exists for realm '%v'", clientName, realm))
		return nil
	}

	_, err = client.CreateClient(ctx, adminToken, realm, gocloak.Client{
		ClientID:                  &clientName,
		Enabled:                   boolArg(true),
		PublicClient:              boolArg(true),    // Public client that doesn't require a secret for authentication.
		RedirectURIs:              &redirectUrls,    // URIs where the client will redirect after authentication.
		RootURL:                   &rootUrl,         // Root URL for the client application.
		BaseURL:                   strArg("/login"), // Base URL for the client application.
		DirectAccessGrantsEnabled: boolArg(true),    // Direct grants like password flow are disabled.
		ServiceAccountsEnabled:    boolArg(false),   // Service accounts are disabled.
		StandardFlowEnabled:       boolArg(true),    // Standard authorization code flow is enabled.
		ImplicitFlowEnabled:       boolArg(false),   // Implicit flow is disabled.
		FullScopeAllowed:          boolArg(false),   // Limit access to only allowed scopes.
		DefaultClientScopes:       &[]string{"profile", "email", "openid", "roles"},
		OptionalClientScopes:      &[]string{"offline_access", "microprofile-jwt"},
		ProtocolMappers: &[]gocloak.ProtocolMapperRepresentation{
			{
				Name:            strArg("auidience resolve"),            // Protocol mappers adjust tokens for clients.
				Protocol:        strArg("openid-connect"),               // The OIDC protocol used for authentication.
				ProtocolMapper:  strArg("oidc-audience-resolve-mapper"), // Mapper to add audience claim in tokens.
				ConsentRequired: boolArg(false),
				Config:          &map[string]string{},
			},
		},
		WebOrigins: &redirectUrls,
	})
	if err != nil {
		if isConflict(err) {
			slog.Info(fmt.Sprintf("KEYCLOAK: client '%v' has already been created for realm '%v'", clientName, realm))
			return nil
		}
		return fmt.Errorf("error creating realm client: %w", err)
	}
	return nil
}

type KeycloakArgs struct {
	KeycloakServerUrl string

	KeycloakAdminUsername string
	KeycloakAdminPassword string

	AdminUsername string
	AdminEmail    string
	AdminPassword string

	PublicHostname  string
	PrivateHostname string

	SslLogin bool

	Verbose bool
}

func NewKeycloakIdentityProvider(db *gorm.DB, auditLog AuditLogger, args KeycloakArgs) (IdentityProvider, error) {
	realm := "ThirdAI-Platform"

	client := gocloak.NewClient(args.KeycloakServerUrl)
	restyClient := client.RestyClient()
	restyClient.SetDebug(args.Verbose) // Adds logging for every request

	if args.SslLogin {
		cert, err := tls.LoadX509KeyPair("/model_bazaar/certs/traefik.crt", "/model_bazaar/certs/traefik.key")
		if err != nil {
			return nil, fmt.Errorf("error loading cert: %w", err)
		}
		restyClient.SetCertificates(cert)
	} else {
		restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	adminToken, err := adminLogin(client, args.KeycloakAdminUsername, args.KeycloakAdminPassword)
	if err != nil {
		slog.Error("KEYCLOAK: admin login failed", "error", err)
		return nil, err
	}
	slog.Info("KEYCLOAK: admin login successful")

	err = createRealm(client, adminToken, realm)
	if err != nil {
		slog.Error("KEYCLOAK: realm creation failed", "error", err)
		return nil, err
	}
	slog.Info("KEYCLOAK: realm creation successful")

	redirectUrls := []string{
		fmt.Sprintf("http://%v/*", args.PublicHostname),
		fmt.Sprintf("https://%v/*", args.PublicHostname),
		fmt.Sprintf("http://%v:80/*", args.PublicHostname),
		fmt.Sprintf("https://%v:80/*", args.PublicHostname),
		fmt.Sprintf("http://%v/*", args.PrivateHostname),
		fmt.Sprintf("https://%v/*", args.PrivateHostname),
		fmt.Sprintf("http://%v:80/*", args.PrivateHostname),
		fmt.Sprintf("https://%v:80/*", args.PrivateHostname),
		"http://localhost/*",
		"https://localhost/*",
		"http://localhost:80/*",
		"https://localhost:80/*",
		"http://127.0.0.1/*",
		"https://127.0.0.1/*",
		"*",
	}
	err = createClient(client, adminToken, realm, redirectUrls, args.KeycloakServerUrl)
	if err != nil {
		slog.Error("KEYCLOAK: client creation failed", "error", err)
		return nil, err
	}
	slog.Info("KEYCLOAK: client creation successful")

	// Create new admin user in master realm
	masterUserId, err := createAdminIfNotExists(client, adminToken, args.AdminUsername, args.AdminEmail, args.AdminPassword, "master")
	if err != nil {
		slog.Error("KEYCLOAK: new admin creation failed", "realm", "master", "error", err)
		return nil, err
	}
	slog.Info("KEYCLOAK: new admin creation successful")

	err = assignAdminRole(client, adminToken, masterUserId, "master")
	if err != nil {
		slog.Error("KEYCLOAK: admin role assignment failed", "error", err)
		return nil, err
	}
	slog.Info("KEYCLOAK: admin role assignment successful")

	// Create new admin user in platform realm
	userId, err := createAdminIfNotExists(client, adminToken, args.AdminUsername, args.AdminEmail, args.AdminPassword, realm)
	if err != nil {
		slog.Error("KEYCLOAK: new admin creation failed", "realm", realm, "error", err)
		return nil, err
	}
	slog.Info("KEYCLOAK: new admin creation successful")

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid '%v' returned from keycloak: %w", userId, err)
	}

	err = addInitialAdminToDb(db, userUUID, args.AdminUsername, args.AdminEmail, nil)
	if err != nil {
		slog.Error("KEYCLOAK: adding new admin to db failed", "error", err)
		return nil, err
	}
	slog.Info("KEYCLOAK: adding new admin to db successful")

	return &KeycloakIdentityProvider{
		keycloak:      client,
		db:            db,
		auditLog:      auditLog,
		realm:         realm,
		adminUsername: args.AdminUsername,
		adminPassword: args.AdminPassword,
	}, nil
}

func getToken(r *http.Request) (string, error) {
	if token := jwtauth.TokenFromHeader(r); token != "" {
		return token, nil
	}
	if token := jwtauth.TokenFromCookie(r); token != "" {
		return token, nil
	}
	return "", fmt.Errorf("unable to find auth token")
}

func (auth *KeycloakIdentityProvider) middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			token, err := getToken(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			userInfo, err := auth.keycloak.GetUserInfo(ctx, token, auth.realm)
			if err != nil {
				http.Error(w, fmt.Sprintf("unable to verify token with keycloak: %v", err), http.StatusUnauthorized)
				return
			}

			if userInfo.Sub == nil {
				http.Error(w, "user identifier missing in keycloak response", http.StatusInternalServerError)
				return
			}

			userUUID, err := uuid.Parse(*userInfo.Sub)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid uuid '%v' returned from keycloak: %v", *userInfo.Sub, err), http.StatusInternalServerError)
				return
			}

			user, err := schema.GetUser(userUUID, auth.db)
			if err != nil {
				if errors.Is(err, schema.ErrUserNotFound) {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				slog.Error("unable to find user from keycloak id", "keycloak_id", *userInfo.Sub, "error", err)
				http.Error(w, fmt.Sprintf("unable to find user %v: %v", *userInfo.Sub, schema.ErrDbAccessFailed), http.StatusInternalServerError)
				return
			}

			reqCtx := r.Context()
			reqCtx = context.WithValue(reqCtx, UserRequestContextKey, user)
			next.ServeHTTP(w, r.WithContext(reqCtx))
		}

		return http.HandlerFunc(handler)
	}
}

func (auth *KeycloakIdentityProvider) AuthMiddleware() chi.Middlewares {
	return chi.Middlewares{auth.middleware(), auth.auditLog.Middleware}
}

func (auth *KeycloakIdentityProvider) AllowDirectSignup() bool {
	return false
}

func (auth *KeycloakIdentityProvider) LoginWithEmail(email, password string) (LoginResult, error) {
	return LoginResult{}, fmt.Errorf("login with email is not supported for this identity provider")
}

func (auth *KeycloakIdentityProvider) LoginWithToken(accessToken string) (LoginResult, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userInfo, err := auth.keycloak.GetUserInfo(ctx, accessToken, auth.realm)
	if err != nil {
		slog.Error("failed to get user info from keycloak", "error", err)
		return LoginResult{}, fmt.Errorf("failed to authenticate user with keycloak: %w", err)
	}

	if userInfo.Sub == nil || userInfo.Email == nil || userInfo.PreferredUsername == nil {
		slog.Error("invalid user info from keycloak, missing required fields", "userInfo", userInfo)
		return LoginResult{}, fmt.Errorf("invalid user info from keycloak, missing required fields")
	}

	userId, err := uuid.Parse(*userInfo.Sub)
	if err != nil {
		slog.Error("invalid uuid returned from keycloak", "uuid", *userInfo.Sub, "error", err)
		return LoginResult{}, fmt.Errorf("invalid uuid '%v' returned from keycloak: %w", *userInfo.Sub, err)
	}

	var user schema.User

	err = auth.db.Transaction(func(txn *gorm.DB) error {

		findUserResult := txn.Limit(1).Find(&user, "email = ?", *userInfo.Email)
		if findUserResult.Error != nil {
			slog.Error("sql error checking for existing user in keycloak identity provider", "email", *userInfo.Email, "error", findUserResult.Error)
			return schema.ErrDbAccessFailed
		}

		if findUserResult.RowsAffected != 1 {
			user = schema.User{
				Id:       userId,
				Username: *userInfo.PreferredUsername,
				Email:    *userInfo.Email,
				IsAdmin:  false,
			}

			createUserResult := txn.Create(&user)
			if createUserResult.Error != nil {
				slog.Error("sql error creating new user in keycloak identity provider", "error", createUserResult.Error)
				return schema.ErrDbAccessFailed
			}
		}
		return nil
	})

	if err != nil {
		return LoginResult{}, fmt.Errorf("error logging in user: %w", err)
	}

	return LoginResult{UserId: user.Id, AccessToken: accessToken}, nil
}

func (auth *KeycloakIdentityProvider) checkExistingUsers(adminToken, field string, params gocloak.GetUsersParams) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	params.Max = intArg(1) // Limit number of results
	users, err := auth.keycloak.GetUsers(ctx, adminToken, auth.realm, params)
	if err != nil {
		return fmt.Errorf("unable to get users: %w", err)
	}

	if len(users) > 0 {
		if field == "username" {
			return ErrUsernameAlreadyInUse
		}
		return ErrEmailAlreadyInUse
	}

	return nil
}

func (auth *KeycloakIdentityProvider) CreateUser(username, email, password string) (uuid.UUID, error) {
	adminToken, err := adminLogin(auth.keycloak, auth.adminUsername, auth.adminPassword)
	if err != nil {
		return uuid.Nil, err
	}

	existingUsername := auth.checkExistingUsers(adminToken, "username", gocloak.GetUsersParams{Username: &username})
	if existingUsername != nil {
		return uuid.Nil, existingUsername
	}

	existingEmail := auth.checkExistingUsers(adminToken, "email", gocloak.GetUsersParams{Email: &email})
	if existingEmail != nil {
		return uuid.Nil, existingEmail
	}

	trueArg := true
	falseArg := false
	passwordKey := "password"

	keycloakUser := gocloak.User{
		Username:      &username,
		Email:         &email,
		Enabled:       &trueArg,
		EmailVerified: &trueArg,
		Credentials: &[]gocloak.CredentialRepresentation{{
			Type: &passwordKey, Value: &password, Temporary: &falseArg,
		}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userId, err := auth.keycloak.CreateUser(ctx, adminToken, auth.realm, keycloakUser)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error creating new user in keycloak: %w", err)
	}

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid '%v' returned from keycloak: %w", userId, err)
	}

	user := schema.User{
		Id:       userUUID,
		Username: username,
		Email:    email,
		IsAdmin:  false,
	}

	result := auth.db.Create(&user)
	if result.Error != nil {
		slog.Error("sql error creating user in keycloak identity provider", "error", result.Error)
		return uuid.Nil, schema.ErrDbAccessFailed
	}

	return userUUID, nil
}

func (auth *KeycloakIdentityProvider) VerifyUser(userId uuid.UUID) error {
	adminToken, err := adminLogin(auth.keycloak, auth.adminUsername, auth.adminPassword)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userIdStr := userId.String()
	verified := true
	err = auth.keycloak.UpdateUser(ctx, adminToken, auth.realm, gocloak.User{
		ID:            &userIdStr,
		EmailVerified: &verified,
	})
	if err != nil {
		slog.Error("failed to verify user with keycloak", "user_id", userId, "error", err)
		return fmt.Errorf("failed to verify user with keycloak: %w", err)
	}

	return nil
}

func (auth *KeycloakIdentityProvider) DeleteUser(userId uuid.UUID) error {
	adminToken, err := adminLogin(auth.keycloak, auth.adminUsername, auth.adminPassword)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = auth.keycloak.DeleteUser(ctx, adminToken, auth.realm, userId.String())
	if err != nil {
		slog.Error("failed to delete user with keycloak", "user_id", userId, "error", err)
		return fmt.Errorf("failed to delete user with keycloak: %w", err)
	}

	return nil
}

func (auth *KeycloakIdentityProvider) GetTokenExpiration(r *http.Request) (time.Time, error) {
	authToken, err := getToken(r)
	if err != nil {
		return time.Time{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tokenInfo, _, err := auth.keycloak.DecodeAccessToken(ctx, authToken, auth.realm)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to verify token with keycloak: %w", err)
	}

	exp, err := tokenInfo.Claims.GetExpirationTime()
	if err != nil {
		return time.Time{}, fmt.Errorf("error getting token expiration: %w", err)
	}
	if exp == nil {
		return time.Time{}, fmt.Errorf("no token expiration found")
	}

	return exp.Time, nil
}
