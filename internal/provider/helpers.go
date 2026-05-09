package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	akclient "github.com/artifactkeeper/terraform-provider-artifactkeeper/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func clientFromResourceData(data any, diagnostics *diag.Diagnostics) *akclient.Client {
	if data == nil {
		diagnostics.AddError("Missing client", "The provider client was not configured.")
		return nil
	}
	c, ok := data.(*akclient.Client)
	if !ok {
		diagnostics.AddError("Unexpected client type", fmt.Sprintf("Expected *client.Client, got %T.", data))
		return nil
	}
	return c
}

func addClientError(diagnostics *diag.Diagnostics, summary string, err error) {
	var apiErr *akclient.APIError
	if errors.As(err, &apiErr) && apiErr.RetryAfter != nil {
		diagnostics.AddError(summary, fmt.Sprintf("%s. Retry after %s.", apiErr.Error(), apiErr.RetryAfter.Round(time.Second)))
		return
	}
	diagnostics.AddError(summary, sanitizeError(err))
}

func isNotFound(err error) bool {
	var apiErr *akclient.APIError
	return errors.As(err, &apiErr) && apiErr.IsNotFound()
}

func optionalString(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueString()
	return &v
}

func optionalInt64(value types.Int64) *int64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueInt64()
	return &v
}

func setStrings(ctx context.Context, set types.Set, diagnostics *diag.Diagnostics) []string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var values []string
	diagnostics.Append(set.ElementsAs(ctx, &values, false)...)
	sort.Strings(values)
	return values
}

func stringSet(ctx context.Context, values []string, diagnostics *diag.Diagnostics) types.Set {
	sort.Strings(values)
	out, diags := types.SetValueFrom(ctx, types.StringType, values)
	diagnostics.Append(diags...)
	return out
}

func listStrings(ctx context.Context, list types.List, diagnostics *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var values []string
	diagnostics.Append(list.ElementsAs(ctx, &values, false)...)
	return values
}

func stringList(ctx context.Context, values []string, diagnostics *diag.Diagnostics) types.List {
	out, diags := types.ListValueFrom(ctx, types.StringType, values)
	diagnostics.Append(diags...)
	return out
}

func stringPtrValue(ptr *string) types.String {
	if ptr == nil {
		return types.StringNull()
	}
	return types.StringValue(*ptr)
}

func roleFromAdmin(isAdmin bool) string {
	if isAdmin {
		return "admin"
	}
	return "user"
}

func adminFromRole(role string) bool {
	return role == "admin"
}

func canonicalJSONString(raw string) (string, error) {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return "", err
	}
	return strings.TrimSpace(buffer.String()), nil
}

func jsonStringsEqual(a, b string) bool {
	canonicalA, err := canonicalJSONString(a)
	if err != nil {
		return false
	}
	canonicalB, err := canonicalJSONString(b)
	if err != nil {
		return false
	}
	return canonicalA == canonicalB
}

func diffStrings(desired, current []string) (add []string, remove []string) {
	desiredSet := map[string]struct{}{}
	currentSet := map[string]struct{}{}
	for _, value := range desired {
		desiredSet[value] = struct{}{}
	}
	for _, value := range current {
		currentSet[value] = struct{}{}
	}
	for value := range desiredSet {
		if _, ok := currentSet[value]; !ok {
			add = append(add, value)
		}
	}
	for value := range currentSet {
		if _, ok := desiredSet[value]; !ok {
			remove = append(remove, value)
		}
	}
	sort.Strings(add)
	sort.Strings(remove)
	return add, remove
}
