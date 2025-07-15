package oauth

import (
	"time"
)

type ProviderType string

const (
	ProviderTypeTikTok    ProviderType = "tiktok"
	ProviderTypeInstagram ProviderType = "instagram"
	ProviderTypeFacebook  ProviderType = "facebook"
)

type ProviderConfig struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
	UserInfo     *UserInfo `json:"user_info,omitempty"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

type ProviderMetadata struct {
	Name        string       `json:"name"`
	Type        ProviderType `json:"type"`
	AuthURL     string       `json:"auth_url"`
	TokenURL    string       `json:"token_url"`
	UserInfoURL string       `json:"user_info_url"`
	Scopes      []string     `json:"scopes"`
	RedirectURI string       `json:"redirect_uri"`
}

var SupportedProviders = map[ProviderType]ProviderMetadata{
	ProviderTypeTikTok: {
		Name:        "TikTok",
		Type:        ProviderTypeTikTok,
		AuthURL:     "https://www.tiktok.com/v2/auth/authorize/",
		TokenURL:    "https://open.tiktokapis.com/v2/oauth/token/",
		UserInfoURL: "https://open.tiktokapis.com/v2/user/info/",
		Scopes:      []string{"user.info.basic", "user.info.profile", "user.info.stats"},
		RedirectURI: "/oauth/callback/tiktok",
	},
	ProviderTypeInstagram: {
		Name:        "Instagram",
		Type:        ProviderTypeInstagram,
		AuthURL:     "https://api.instagram.com/oauth/authorize",
		TokenURL:    "https://api.instagram.com/oauth/access_token",
		UserInfoURL: "https://graph.instagram.com/me",
		Scopes:      []string{"user_profile", "user_media"},
		RedirectURI: "/oauth/callback/instagram",
	},
	ProviderTypeFacebook: {
		Name:        "Facebook",
		Type:        ProviderTypeFacebook,
		AuthURL:     "https://www.facebook.com/v18.0/dialog/oauth",
		TokenURL:    "https://graph.facebook.com/v18.0/oauth/access_token",
		UserInfoURL: "https://graph.facebook.com/me",
		Scopes:      []string{"public_profile", "email", "pages_show_list", "pages_read_engagement"},
		RedirectURI: "/oauth/callback/facebook",
	},
}
