package runtime

// Config provides minimal config to spin up the server. The env var name is prefixed with split words (e.g. ServerURL = RISER_SERVER_URL)
type Config struct {
	/* Required  */
	ServerURL    string `split_words:"true" required:"true"`
	ServerApikey string `split_words:"true" required:"true"`
	Environment  string `split_words:"true" required:"true"`

	/* Optional */
	ServerPingSeconds               int    `split_words:"true" default:"10"`
	SealedSecretEnabled             bool   `split_words:"true" default:"true"`
	SealedsecretControllerName      string `split_words:"true" default:"sealed-secrets-controller"`
	SealedsecretNamespace           string `split_words:"true" default:"kube-system"`
	SealedsecretCertRefreshDuration string `split_words:"true" default:"24h"`
}
