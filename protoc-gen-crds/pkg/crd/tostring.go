package crd

// ToString serializes a set of CRDs into a single text stream.
func ToString(defs []*ResourceDefinition) string {
	content := ""
	for _, d := range defs {
		if content != "" {
			content += "---\n"
		}
		content += d.String()
	}

	return content
}
