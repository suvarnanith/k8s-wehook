package mutate

import "encoding/json"

const (
	environmentLabelKey          = "environment"
	environmentLabelDefaultValue = "default"
)

type jsonPatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// objectMetaOnly lets us decode just the labels field regardless of which
// kind the object is.
type objectMetaOnly struct {
	Metadata struct {
		Labels map[string]string `json:"labels"`
	} `json:"metadata"`
}

// buildEnvironmentLabelPatch returns a JSON Patch (RFC 6902) that adds the
// environmentLabelKey=environmentLabelDefaultValue label when it's missing,
// or nil if the object already declares that label.
func buildEnvironmentLabelPatch(raw []byte) ([]byte, error) {
	var obj objectMetaOnly
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}

	if _, ok := obj.Metadata.Labels[environmentLabelKey]; ok {
		return nil, nil
	}

	var op jsonPatchOperation

	if obj.Metadata.Labels == nil {
		op = jsonPatchOperation{
			Op:    "add",
			Path:  "/metadata/labels",
			Value: map[string]string{environmentLabelKey: environmentLabelDefaultValue},
		}
	} else {
		op = jsonPatchOperation{
			Op:    "add",
			Path:  "/metadata/labels/" + environmentLabelKey,
			Value: environmentLabelDefaultValue,
		}
	}

	return json.Marshal([]jsonPatchOperation{op})
}
