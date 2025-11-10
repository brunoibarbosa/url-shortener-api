package oauth_provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	session "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	ErrExchangingCode    = errors.New("error exchanging code")
	ErrSearchProfileInfo = errors.New("failed to search profile")
	ErrDecodeProfileInfo = errors.New("error decoding response")
)

type GoogleOAuth struct {
	config *oauth2.Config
}

func NewGoogleOAuth(clientID, clientSecret, frontendURL string) *GoogleOAuth {
	return &GoogleOAuth{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  fmt.Sprintf("%s/auth/google/callback", frontendURL),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (g *GoogleOAuth) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

func (g *GoogleOAuth) ExchangeCode(ctx context.Context, code string) (*session.OAuthUser, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, ErrExchangingCode
	}

	client := g.config.Client(ctx, token)
	res, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, ErrSearchProfileInfo
	}
	defer res.Body.Close()

	var profile struct {
		ID        string  `json:"id"`
		Name      string  `json:"name"`
		Email     string  `json:"email"`
		AvatarURL *string `json:"picture"`
	}
	if err := json.NewDecoder(res.Body).Decode(&profile); err != nil {
		return nil, ErrDecodeProfileInfo
	}

	return &session.OAuthUser{
		ID:           profile.ID,
		Name:         profile.Name,
		Email:        profile.Email,
		AvatarURL:    profile.AvatarURL,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}
