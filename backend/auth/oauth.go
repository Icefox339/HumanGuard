package auth

import (
	"context"
	"encoding/json"

	"golang.org/x/oauth2"
)

type OAuthService struct {
	config *oauth2.Config
}

func NewOAuthService(clientID, clientSecret, redirectURL, keycloakURL string) *OAuthService {
	return &OAuthService{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  keycloakURL + "/realms/master/protocol/openid-connect/auth",
				TokenURL: keycloakURL + "/realms/master/protocol/openid-connect/token",
			},
		},
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
	resp, err := client.Get("http://localhost:8081/realms/master/protocol/openid-connect/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:    userInfo.Sub,
		Email: userInfo.Email,
		Name:  userInfo.Name,
	}, nil
}

type OAuthUserInfo struct {
	ID    string
	Email string
	Name  string
}
