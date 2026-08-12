package openapi

import _ "embed"

// Document is the generated, bundled API contract embedded in every backend
// binary so the authenticated admin reference matches the running artifact.
//
//go:embed generated/openapi.json
var Document []byte
