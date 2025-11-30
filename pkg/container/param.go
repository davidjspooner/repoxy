package container

// param represents the parsed pieces of a Docker registry request.
type param struct {
	name   string
	tag    string
	uuid   string
	digest string
}
