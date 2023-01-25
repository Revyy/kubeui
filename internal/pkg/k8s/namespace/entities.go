package namespace

// List represents an immutable list of namespaces.
type List struct {
	namespaces []string
}

// NameSpaces returns the namespaces.
func (ns List) NameSpaces() []string {
	return ns.namespaces
}
