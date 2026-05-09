package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestProviderSchema(t *testing.T) {
	t.Parallel()

	p := New("test")()
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	for _, name := range []string{"url", "base_url", "username", "password", "token", "insecure_skip_verify"} {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Fatalf("missing provider attribute %q", name)
		}
	}
}

func TestValueOrEnv(t *testing.T) {
	t.Setenv("ARTIFACTKEEPER_TEST_VALUE", "from-env")
	if got := valueOrEnv(types.StringNull(), "ARTIFACTKEEPER_TEST_VALUE"); got != "from-env" {
		t.Fatalf("valueOrEnv = %q", got)
	}
}
