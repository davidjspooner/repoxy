package docker

// param represents a reference to a Terraform provider by namespace and name.
type param struct {
	name   string
	tag    string
	uuid   string
	digest string
}
