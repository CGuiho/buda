package schemas_test

import (
	"encoding/json"
	"io/fs"
	"strings"
	"testing"

	"github.com/CGuiho/buda/examples"
	"github.com/CGuiho/buda/internal/config"
	"github.com/CGuiho/buda/schemas"
)

func TestEmbeddedSchemasPresentAndValid(t *testing.T) {
	for _, name := range []string{"buda.schema.json", "buda.global.schema.json"} {
		data, err := fs.ReadFile(schemas.FS, name)
		if err != nil {
			t.Fatalf("ReadFile(%q) = %v", name, err)
		}
		var parsed map[string]any
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("Unmarshal(%q) = %v", name, err)
		}
		if parsed["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
			t.Fatalf("%s $schema = %v", name, parsed["$schema"])
		}
		id, ok := parsed["$id"].(string)
		if !ok || !strings.Contains(id, "v0.2.0") {
			t.Fatalf("%s $id = %v", name, id)
		}
	}
}

func TestEmbeddedExamplesValidateAgainstRuntimeConfig(t *testing.T) {
	projectData, err := fs.ReadFile(examples.FS, "buda.example.yaml")
	if err != nil {
		t.Fatalf("read buda.example.yaml: %v", err)
	}
	projectConfig, err := config.DecodeProject(strings.NewReader(string(projectData)))
	if err != nil {
		t.Fatalf("DecodeProject(buda.example.yaml): %v", err)
	}
	if projectConfig.WikiID != "example-wiki" || projectConfig.Schema != config.CurrentSchema {
		t.Fatalf("projectConfig = %+v", projectConfig)
	}

	globalData, err := fs.ReadFile(examples.FS, "buda.global.example.yaml")
	if err != nil {
		t.Fatalf("read buda.global.example.yaml: %v", err)
	}
	globalConfig, err := config.DecodeGlobal(strings.NewReader(string(globalData)))
	if err != nil {
		t.Fatalf("DecodeGlobal(buda.global.example.yaml): %v", err)
	}
	if globalConfig.Schema != config.CurrentSchema {
		t.Fatalf("globalConfig = %+v", globalConfig)
	}

	merged, err := config.Merge(globalConfig, projectConfig)
	if err != nil {
		t.Fatalf("Merge examples: %v", err)
	}
	if err := merged.Validate(); err != nil {
		t.Fatalf("Validate merged examples: %v", err)
	}
}

func TestDecodingRejectsMissingSchema(t *testing.T) {
	noSchemaProject := `wiki_id: test
bundle: knowledge
`
	if _, err := config.DecodeProject(strings.NewReader(noSchemaProject)); err == nil {
		t.Fatal("DecodeProject unexpectedly accepted missing schema")
	}

	noSchemaGlobal := `bundle: knowledge
derived: .buda
`
	if _, err := config.DecodeGlobal(strings.NewReader(noSchemaGlobal)); err == nil {
		t.Fatal("DecodeGlobal unexpectedly accepted missing schema")
	}
}
