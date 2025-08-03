package oauth

import (
	"time"
)

type ProviderType string

const (
	ProviderTypeTikTok    ProviderType = "tiktok"
	ProviderTypeInstagram ProviderType = "instagram"
	ProviderTypeFacebook  ProviderType = "facebook"
	ProviderTypeLinkedIn  ProviderType = "linkedin"
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

type AccountInfo struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Username        string `json:"username,omitempty"`
	Type            string `json:"type"`
	Category        string `json:"category,omitempty"`
	PageAccessToken string `json:"page_access_token,omitempty"`
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
		Scopes:      []string{"user.info.basic", "user.info.profile", "video.publish", "video.upload"},
		// Scopes:      []string{"user.info.profile"},
		RedirectURI: "/oauth/callback/tiktok",
	},
	ProviderTypeInstagram: {
		Name:        "Instagram",
		Type:        ProviderTypeInstagram,
		AuthURL:     "https://www.facebook.com/v18.0/dialog/oauth",
		TokenURL:    "https://graph.facebook.com/v18.0/oauth/access_token",
		UserInfoURL: "https://graph.facebook.com/v18.0/me/accounts",
		Scopes:      []string{"instagram_basic", "instagram_content_publish", "pages_show_list", "pages_read_engagement", "pages_manage_posts", "business_management", "instagram_manage_insights"},
		RedirectURI: "/oauth/callback/instagram",
	},
	ProviderTypeFacebook: {
		Name:        "Facebook",
		Type:        ProviderTypeFacebook,
		AuthURL:     "https://www.facebook.com/v18.0/dialog/oauth",
		TokenURL:    "https://graph.facebook.com/v18.0/oauth/access_token",
		UserInfoURL: "https://graph.facebook.com/me",
		Scopes:      []string{"public_profile", "email", "pages_show_list", "pages_read_engagement", "pages_manage_posts", "business_management"},
		RedirectURI: "/oauth/callback/facebook",
	},
	ProviderTypeLinkedIn: {
		Name:        "LinkedIn",
		Type:        ProviderTypeLinkedIn,
		AuthURL:     "https://www.linkedin.com/oauth/v2/authorization",
		TokenURL:    "https://www.linkedin.com/oauth/v2/accessToken",
		UserInfoURL: "https://api.linkedin.com/v2/userinfo",
		Scopes:      []string{"openid", "profile", "w_member_social", "email"},
		RedirectURI: "/oauth/callback/linkedin",
	},
}
