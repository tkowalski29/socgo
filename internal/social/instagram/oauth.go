package instagram

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tkowalski/socgo/internal/data/config"
	"github.com/tkowalski/socgo/internal/data/oauth"
)

type OAuth struct {
	*oauth.BaseProviderAuth
}

func NewOAuth(cfg *config.Config) *OAuth {
	return &OAuth{
		BaseProviderAuth: oauth.NewBaseProviderAuth(cfg),
	}
}

func (i *OAuth) GetProviderType() oauth.ProviderType {
	return oauth.ProviderTypeInstagram
}

func (i *OAuth) GetConnectURL(userID string, providerName string, providerConfig *config.ProviderInstance) (string, error) {
	metadata := oauth.SupportedProviders[oauth.ProviderTypeInstagram]

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", providerConfig.ClientID)
	params.Add("redirect_uri", i.GetRedirectURI(oauth.ProviderTypeInstagram))
	params.Add("scope", strings.Join(metadata.Scopes, " "))
	params.Add("state", fmt.Sprintf("%s:%s", userID, providerName))

	authURL := metadata.AuthURL + "?" + params.Encode()

	// Log the OAuth URL details
	log.Printf("🔗 Instagram OAuth URL being generated:")
	log.Printf("   Client ID: %s", providerConfig.ClientID)
	log.Printf("   Redirect URI: %s", i.GetRedirectURI(oauth.ProviderTypeInstagram))
	log.Printf("   Scopes: %s", strings.Join(metadata.Scopes, " "))
	log.Printf("   Auth URL: %s", metadata.AuthURL)
	log.Printf("   Final URL: %s", authURL)

	return authURL, nil
}

func (i *OAuth) ExchangeCodeForToken(code string, providerConfig *config.ProviderInstance) (*oauth.ProviderConfig, error) {
	metadata := oauth.SupportedProviders[oauth.ProviderTypeInstagram]

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", providerConfig.ClientID)
	data.Set("client_secret", providerConfig.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", i.GetRedirectURI(oauth.ProviderTypeInstagram))

	resp, err := http.PostForm(metadata.TokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)

	return &oauth.ProviderConfig{
		AccessToken: tokenResponse.AccessToken,
		TokenType:   tokenResponse.TokenType,
		ExpiresAt:   expiresAt,
	}, nil
}

func (i *OAuth) GetUserInfo(accessToken string) (*oauth.UserInfo, error) {
	req, err := http.NewRequest("GET", "https://graph.facebook.com/v18.0/me", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("access_token", accessToken)
	q.Add("fields", "id,name")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user info request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var userInfo oauth.UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info response: %w", err)
	}

	return &userInfo, nil
}

func (i *OAuth) GetAvailableAccounts(accessToken string) ([]oauth.AccountInfo, error) {
	log.Printf("🔍 Getting available Instagram accounts...")

	// First, let's check what permissions we have
	log.Printf("🔐 Checking permissions for access token...")
	permissions, err := i.checkPermissions(accessToken)
	if err != nil {
		log.Printf("❌ Error checking permissions: %v", err)
	} else {
		log.Printf("✅ Permissions: %s", permissions)
	}

	// Check user info to see what account we're using
	log.Printf("👤 Checking user info...")
	userInfo, err := i.GetUserInfo(accessToken)
	if err != nil {
		log.Printf("❌ Error getting user info: %v", err)
	} else {
		log.Printf("✅ User: %s (ID: %s)", userInfo.Name, userInfo.ID)
	}

	// Check if this is a business profile
	log.Printf("🏢 Checking profile type...")
	profileType, err := i.checkProfileType(accessToken)
	if err != nil {
		log.Printf("❌ Error checking profile type: %v", err)
	} else {
		log.Printf("✅ Profile type: %s", profileType)
		if profileType == "BUSINESS" {
			log.Printf("⚠️  WARNING: You're using a Business Profile!")
			log.Printf("💡 Business profiles have limited API access")
			log.Printf("💡 This might cause issues with Facebook Pages access")
			log.Printf("💡 Consider switching back to personal profile for testing")
		}
	}

	// Check if we can access Facebook Pages with this token
	log.Printf("📄 Checking Facebook Pages access...")
	pages, err := i.getFacebookPages(accessToken)
	if err != nil {
		log.Printf("❌ Error getting Facebook pages: %v", err)
	} else {
		log.Printf("📊 Found %d Facebook pages", len(pages))
		for i, page := range pages {
			log.Printf("  %d. ID: %s, Name: %s, Category: %s", i+1, page.ID, page.Name, page.Category)
		}
	}

	// Try a different approach - check if we can access pages directly
	log.Printf("🔍 Trying alternative Facebook Pages check...")
	altPages, err := i.getFacebookPagesAlternative(accessToken)
	if err != nil {
		log.Printf("❌ Error with alternative Facebook pages check: %v", err)
	} else {
		log.Printf("📊 Alternative method found %d Facebook pages", len(altPages))
		for i, page := range altPages {
			log.Printf("  %d. ID: %s, Name: %s", i+1, page.ID, page.Name)
		}
	}

	// Check if this is the same account that has Facebook Pages
	log.Printf("🔍 Checking if this account has Facebook Pages...")
	if len(pages) == 0 && len(altPages) == 0 {
		log.Printf("⚠️  WARNING: This Facebook account has no Facebook Pages!")
		log.Printf("💡 But you mentioned you have Facebook Page 'Test1'")
		log.Printf("💡 This suggests you might be using a different Facebook account")
		log.Printf("")
		log.Printf("🔧 POSSIBLE SOLUTIONS:")
		log.Printf("   1. Check if you're logged into the correct Facebook account")
		log.Printf("   2. The Facebook Page 'Test1' might belong to a different account")
		log.Printf("   3. Try logging out and logging into the account that owns 'Test1'")
		log.Printf("")
		log.Printf("📋 TO VERIFY:")
		log.Printf("   1. Go to https://www.facebook.com/pages")
		log.Printf("   2. Check if you see 'Test1' page")
		log.Printf("   3. If not, you're using a different account")
		log.Printf("   4. If yes, there might be a permission issue")
	}

	// Get Facebook Pages that can have Instagram Business Accounts connected
	req, err := http.NewRequest("GET", "https://graph.facebook.com/v18.0/me/accounts", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("access_token", accessToken)
	q.Add("fields", "id,name,category,instagram_business_account,access_token")
	req.URL.RawQuery = q.Encode()

	log.Printf("📡 Making request to: %s", req.URL.String())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Instagram accounts request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("Instagram accounts request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("📄 Raw API response: %s", string(bodyBytes))

	var response struct {
		Data []struct {
			ID                       string `json:"id"`
			Name                     string `json:"name"`
			Category                 string `json:"category"`
			AccessToken              string `json:"access_token"`
			InstagramBusinessAccount struct {
				ID string `json:"id"`
			} `json:"instagram_business_account"`
		} `json:"data"`
	}

	if err := json.NewDecoder(strings.NewReader(string(bodyBytes))).Decode(&response); err != nil {
		log.Printf("❌ Failed to decode Instagram accounts response: %v", err)
		return nil, fmt.Errorf("failed to decode Instagram accounts response: %w", err)
	}

	log.Printf("📊 Found %d Facebook pages via Instagram endpoint", len(response.Data))

	// If no Facebook Pages found, provide detailed instructions
	if len(response.Data) == 0 {
		log.Printf("❌ NO FACEBOOK PAGES FOUND via Instagram endpoint!")
		log.Printf("💡 This could mean:")
		log.Printf("   1. Missing 'pages_show_list' permission")
		log.Printf("   2. Using wrong access token")
		log.Printf("   3. Facebook Pages not accessible via this token")
		log.Printf("   4. Token is for a different Facebook account")
		log.Printf("   5. This Facebook account has no Facebook Pages")
		log.Printf("   6. Instagram connection changed token permissions")
		log.Printf("   7. Using Business Profile with limited API access")
		log.Printf("")
		log.Printf("🔧 TROUBLESHOOTING:")
		log.Printf("   1. Check if you granted 'pages_show_list' permission")
		log.Printf("   2. Try reconnecting your Instagram account")
		log.Printf("   3. Make sure you're using the same Facebook account")
		log.Printf("   4. Check if your Facebook Pages are public")
		log.Printf("   5. Create a Facebook Page if you don't have one")
		log.Printf("   6. Check if Instagram connection affected permissions")
		log.Printf("   7. Switch from Business Profile to Personal Profile")
		log.Printf("")
		log.Printf("📋 REQUIRED PERMISSIONS:")
		log.Printf("   - pages_show_list (to see your Facebook Pages)")
		log.Printf("   - instagram_basic (to access Instagram)")
		log.Printf("   - instagram_content_publish (to publish content)")
		log.Printf("   - pages_manage_posts (to publish via Facebook Pages)")
		log.Printf("")
		log.Printf("🔗 MANUAL CHECK:")
		log.Printf("   1. Go to https://www.facebook.com/pages")
		log.Printf("   2. Check if you see your Facebook Pages")
		log.Printf("   3. If not, you might be using a different Facebook account")
		log.Printf("   4. If no pages exist, create one:")
		log.Printf("      - Click 'Create Page'")
		log.Printf("      - Choose 'Business or Brand'")
		log.Printf("      - Enter page name (e.g., 'My Business')")
		log.Printf("      - Choose category (e.g., 'Software')")
		log.Printf("      - Complete page creation")
		log.Printf("")
		log.Printf("📱 AFTER CREATING FACEBOOK PAGE:")
		log.Printf("   1. Go to your new Facebook Page")
		log.Printf("   2. Settings > Instagram")
		log.Printf("   3. Connect your Instagram Business Account")
		log.Printf("   4. Try connecting Instagram again in this app")
		log.Printf("")
		log.Printf("🔍 POSSIBLE ISSUE:")
		log.Printf("   If you connected Instagram to Facebook Page and now pages disappeared,")
		log.Printf("   this might be a token permission issue. Try:")
		log.Printf("   1. Disconnect Instagram from Facebook Page")
		log.Printf("   2. Reconnect Instagram in this app")
		log.Printf("   3. Then connect Instagram to Facebook Page again")
		log.Printf("")
		log.Printf("🏢 BUSINESS PROFILE ISSUE:")
		log.Printf("   If you're using a Business Profile, try:")
		log.Printf("   1. Switch back to Personal Profile")
		log.Printf("   2. Reconnect Instagram in this app")
		log.Printf("   3. Then connect Instagram to Facebook Page")
	}

	var accounts []oauth.AccountInfo
	for i, page := range response.Data {
		log.Printf("📄 Page %d: ID=%s, Name=%s, Category=%s", i+1, page.ID, page.Name, page.Category)
		log.Printf("📄 Page %d: InstagramBusinessAccount.ID=%s", i+1, page.InstagramBusinessAccount.ID)

		if page.InstagramBusinessAccount.ID != "" {
			log.Printf("✅ Found Instagram Business Account: %s for page: %s", page.InstagramBusinessAccount.ID, page.Name)
			accounts = append(accounts, oauth.AccountInfo{
				ID:              page.InstagramBusinessAccount.ID,
				Name:            page.Name,
				Type:            "instagram",
				Category:        page.Category,
				PageAccessToken: page.AccessToken,
			})
		} else {
			log.Printf("❌ No Instagram Business Account found for page: %s", page.Name)
			log.Printf("💡 This means Instagram is not connected to this Facebook Page")
			log.Printf("💡 To fix this:")
			log.Printf("   1. Go to https://www.facebook.com/pages")
			log.Printf("   2. Select page: %s", page.Name)
			log.Printf("   3. Go to Settings > Instagram")
			log.Printf("   4. Connect your Instagram Business Account")
		}
	}

	// If no Instagram Business Accounts found, provide detailed troubleshooting
	if len(accounts) == 0 {
		log.Printf("⚠️  No Instagram Business Accounts found!")
		log.Printf("")
		log.Printf("🔧 STEP-BY-STEP SOLUTION:")
		log.Printf("")
		log.Printf("1. VERIFY YOUR INSTAGRAM ACCOUNT TYPE:")
		log.Printf("   - Open Instagram app")
		log.Printf("   - Go to Profile → Settings → Account")
		log.Printf("   - Check if it shows 'Switch to Personal Account' (means you're Business)")
		log.Printf("   - If it shows 'Switch to Business Account', click it!")
		log.Printf("")
		log.Printf("2. VERIFY FACEBOOK PAGE CONNECTION:")
		log.Printf("   - Go to your Facebook Page 'Test1'")
		log.Printf("   - Settings → Instagram")
		log.Printf("   - Should show connected Instagram account: @tkowalski8")
		log.Printf("   - If not connected, click 'Connect Account'")
		log.Printf("")
		log.Printf("3. VERIFY PERMISSIONS:")
		log.Printf("   - Your Facebook App needs these permissions:")
		log.Printf("   - instagram_basic, instagram_content_publish")
		log.Printf("   - pages_show_list, pages_manage_posts")
		log.Printf("")
		log.Printf("4. TEST WITH FACEBOOK PAGE:")
		log.Printf("   - Since Facebook works, try connecting via Facebook first")
		log.Printf("   - Facebook Page should show Instagram Business Account")
		log.Printf("")
		log.Printf("💡 After fixing above, reconnect Instagram in this app")

		// For debugging purposes, try to get basic user info
		userInfo, err := i.GetUserInfo(accessToken)
		if err == nil {
			log.Printf("📊 Current user: %s (%s)", userInfo.Name, userInfo.ID)
			log.Printf("💡 This is the Facebook account you're using")
			log.Printf("💡 Make sure this Facebook account owns Page 'Test1'")
		}

		// Return empty accounts array instead of creating a dummy account
		return accounts, nil
	}

	log.Printf("📊 Total Instagram accounts found: %d", len(accounts))
	return accounts, nil
}

// checkPermissions checks what permissions the access token has
func (i *OAuth) checkPermissions(accessToken string) (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/me/permissions?access_token=%s", accessToken)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var response struct {
		Data []struct {
			Permission string `json:"permission"`
			Status     string `json:"status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	permissions := ""
	for _, perm := range response.Data {
		if perm.Status == "granted" {
			permissions += perm.Permission + ", "
		}
	}

	return permissions, nil
}

// checkProfileType checks if the user is using a business profile
func (i *OAuth) checkProfileType(accessToken string) (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/me?fields=id,name,account_type&access_token=%s", accessToken)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var response struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		AccountType string `json:"account_type"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return response.AccountType, nil
}

// getFacebookPages tries to get Facebook Pages using a different approach
func (i *OAuth) getFacebookPages(accessToken string) ([]struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}, error) {
	url := fmt.Sprintf("https://graph.facebook.com/v18.0/me/accounts?fields=id,name,category&access_token=%s", accessToken)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var response struct {
		Data []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// getFacebookPagesAlternative tries a different approach to get Facebook Pages
func (i *OAuth) getFacebookPagesAlternative(accessToken string) ([]struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}, error) {
	// Try using the user's ID directly
	userInfo, err := i.GetUserInfo(accessToken)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/accounts?fields=id,name&access_token=%s", userInfo.ID, accessToken)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var response struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

// SaveAllAccounts saves all available Instagram accounts as separate providers
func (i *OAuth) SaveAllAccounts(userID string, providerName string, token *oauth.ProviderConfig, saveFunc func(userID string, providerType oauth.ProviderType, providerName string, config *oauth.ProviderConfig) error) error {
	accounts, err := i.GetAvailableAccounts(token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get available accounts: %w", err)
	}

	if len(accounts) == 0 {
		return fmt.Errorf("no Instagram accounts available")
	}

	// Save each account as a separate provider
	for _, account := range accounts {
		// Use PageAccessToken for Instagram Business Accounts, fallback to main token for basic accounts
		accessToken := token.AccessToken
		if account.PageAccessToken != "" {
			accessToken = account.PageAccessToken
			log.Printf("Using PageAccessToken for Instagram Business Account: %s", account.ID)
		} else {
			log.Printf("Using main access token for basic Instagram account: %s", account.ID)
		}

		accountToken := &oauth.ProviderConfig{
			AccessToken: accessToken,
			TokenType:   token.TokenType,
			ExpiresAt:   token.ExpiresAt,
			UserInfo: &oauth.UserInfo{
				ID:   account.ID,
				Name: account.Name,
			},
		}

		// Use account name as provider name
		accountProviderName := account.Name
		if accountProviderName == "" {
			accountProviderName = fmt.Sprintf("%s_%s", providerName, account.ID)
		}

		if err := saveFunc(userID, i.GetProviderType(), accountProviderName, accountToken); err != nil {
			return fmt.Errorf("failed to save Instagram account %s: %w", account.ID, err)
		}
	}

	return nil
}
