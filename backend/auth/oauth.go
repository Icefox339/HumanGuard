package auth

import (
	"context"
	"encoding/json"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthService struct {
	config *oauth2.Config
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
			AuthURL:  "http://localhost:8081/realms/master/protocol/openid-connect/auth",
			TokenURL: "http://localhost:8081/realms/master/protocol/openid-connect/token",
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
		userInfoURL = "http://localhost:8081/realms/master/protocol/openid-connect/userinfo"
	}
	
	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
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
		userInfo.ID = string(rune(githubUser.ID))
		userInfo.Email = githubUser.Email
		userInfo.Name = githubUser.Name
		if userInfo.Name == "" {
			userInfo.Name = githubUser.Login
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