package nubio

type Config struct {
	Address string `json:"address"` // Local HTTP server address.
	Profile string `json:"profile"` // Path to JSON file where profile data is stored.
	Domain  string `json:"domain"`  // Public domain name used to host the site.
}
