package oauth_provider

import (
	"context"
	"encoding/json"
	"fmt"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

func (g *GoogleOAuth) ExchangeCode(ctx context.Context, code string) (*domain.OAuthUser, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("falha ao trocar código por token: %w", err)
	}

	client := g.config.Client(ctx, token)
	res, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar perfil: %w", err)
	}
	defer res.Body.Close()

	var profile struct {
		ID        string  `json:"id"`
		Name      string  `json:"name"`
		Email     string  `json:"email"`
		AvatarURL *string `json:"picture"`
	}
	if err := json.NewDecoder(res.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	return &domain.OAuthUser{
		ID:           profile.ID,
		Name:         profile.Name,
		Email:        profile.Email,
		AvatarURL:    profile.AvatarURL,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}
