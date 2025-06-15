package tf

// param represents a reference to a Terraform provider by namespace and name.
type param struct {
	namespace string
	name      string
	version   string
	tail      string
}
