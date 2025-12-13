package openrouterfx

// Config holds the OpenRouter client configuration.
// Token is the OpenRouter API key (required).
// AppName and AppURL are optional metadata sent with requests.
type Config struct {
	Token string

	AppName string
	AppURL  string
}
