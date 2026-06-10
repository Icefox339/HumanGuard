package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthService struct {
	config   *oauth2.Config
	provider string
}

func NewOAuthService(provider, clientID, clientSecret, redirectURL string) *OAuthService {
	var endpoint oauth2.Endpoint
	scopes := []string{"openid", "profile", "email"}

	switch provider {
	case "google":
		endpoint = google.Endpoint
		scopes = []string{"openid", "profile", "email"}
	case "github":
		endpoint = oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		}
		scopes = []string{"user:email"}
	default:
		endpoint = oauth2.Endpoint{
			AuthURL:  getEnv("KEYCLOAK_REDIRECT_AUTH_URL", "http://localhost:8081/realms/master/protocol/openid-connect/auth"),
			TokenURL: getEnv("KEYCLOAK_TOKEN_URL", "http://localhost:8081/realms/master/protocol/openid-connect/token"),
		}
	}

	return &OAuthService{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     endpoint,
		},
		provider: provider,
	}
}

func (o *OAuthService) GetAuthURL(state string) string {
	return o.config.AuthCodeURL(state)
}

func (o *OAuthService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return o.config.Exchange(ctx, code)
}

func (o *OAuthService) GetUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := o.config.Client(ctx, token)

	var userInfoURL string
	switch o.provider {
	case "google":
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "github":
		userInfoURL = "https://api.github.com/user"
	default:
		userInfoURL = getEnv("KEYCLOAK_INFO_URL", "http://localhost:8081/realms/master/protocol/openid-connect/userinfo")
	}

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var userInfo OAuthUserInfo

	switch o.provider {
	case "google":
		var googleUser struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
			return nil, err
		}
		userInfo.ID = googleUser.ID
		userInfo.Email = googleUser.Email
		userInfo.Name = googleUser.Name

	case "github":
		var githubUser struct {
			ID    int    `json:"id"`
			Login string `json:"login"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
			return nil, err
		}
		userInfo.ID = strconv.Itoa(githubUser.ID)
		userInfo.Email = githubUser.Email
		userInfo.Name = githubUser.Name
		if userInfo.Name == "" {
			userInfo.Name = githubUser.Login
		}
		if userInfo.Email == "" {
			emailReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
			if err != nil {
				return nil, err
			}
			emailResp, err := client.Do(emailReq)
			if err != nil {
				return nil, err
			}
			defer func() { _ = emailResp.Body.Close() }()

			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
				return nil, err
			}

			for _, item := range emails {
				if item.Primary && item.Verified {
					userInfo.Email = item.Email
					break
				}
			}
			if userInfo.Email == "" && len(emails) > 0 {
				userInfo.Email = emails[0].Email
			}
		}
		if userInfo.Email == "" {
			return nil, fmt.Errorf("github account has no accessible email")
		}

	default:
		var kcUser struct {
			Sub   string `json:"sub"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&kcUser); err != nil {
			return nil, err
		}
		userInfo.ID = kcUser.Sub
		userInfo.Email = kcUser.Email
		userInfo.Name = kcUser.Name
	}

	return &userInfo, nil
}

type OAuthUserInfo struct {
	ID    string
	Email string
	Name  string
}

// getEnv - получение переменной окружения с дефолтом
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
