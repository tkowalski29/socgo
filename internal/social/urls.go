package social

import "fmt"

// BuildExternalURL generates external URL for a given provider type and post ID
func BuildExternalURL(providerType, providerName, postID string) string {
	if postID == "" {
		return ""
	}

	switch providerType {
	case "facebook":
		return fmt.Sprintf("https://www.facebook.com/%s", postID)
	case "instagram":
		return fmt.Sprintf("https://www.instagram.com/p/%s/", postID)
	case "tiktok":
		return fmt.Sprintf("https://www.tiktok.com/@%s/video/%s", providerName, postID)
	default:
		return ""
	}
}
