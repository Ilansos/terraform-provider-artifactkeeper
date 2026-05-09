package provider

import (
	"context"
	"encoding/json"
	"strings"

	akclient "github.com/artifactkeeper/terraform-provider-artifactkeeper/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*securityPolicyResource)(nil)
	_ resource.ResourceWithConfigure   = (*securityPolicyResource)(nil)
	_ resource.ResourceWithImportState = (*securityPolicyResource)(nil)
)

func NewSecurityPolicyResource() resource.Resource {
	return &securityPolicyResource{}
}

type securityPolicyResource struct {
	client *akclient.Client
}

type securityPolicyResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	RepositoryID       types.String `tfsdk:"repository_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	MaxSeverity        types.String `tfsdk:"max_severity"`
	BlockUnscanned     types.Bool   `tfsdk:"block_unscanned"`
	BlockOnFail        types.Bool   `tfsdk:"block_on_fail"`
	RequireSignature   types.Bool   `tfsdk:"require_signature"`
	MaxArtifactAgeDays types.Int64  `tfsdk:"max_artifact_age_days"`
	MinStagingHours    types.Int64  `tfsdk:"min_staging_hours"`
	Rules              types.String `tfsdk:"rules"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

func (r *securityPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_policy"
}

func (r *securityPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = clientFromResourceData(req.ProviderData, &resp.Diagnostics)
}

func (r *securityPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages an Artifact Keeper security policy under `/api/v1/security/policies`.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Security policy UUID.",
			},
			"repository_id": rschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Repository UUID to scope this policy. Omit for a global policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": rschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Security policy name.",
			},
			"description": rschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description kept in Terraform state for documentation. The current security policy API does not persist descriptions.",
				DeprecationMessage:  "The current Artifact Keeper security policy API does not persist descriptions.",
			},
			"enabled": rschema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"max_severity": rschema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("high"),
				MarkdownDescription: "Maximum allowed vulnerability severity threshold. Valid values are `low`, `medium`, `high`, and `critical`.",
			},
			"block_unscanned": rschema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Block artifacts that do not have scan results.",
			},
			"block_on_fail": rschema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Block when policy evaluation or scanning fails.",
			},
			"require_signature": rschema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Require artifacts to have a valid signature.",
			},
			"max_artifact_age_days": rschema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Maximum allowed artifact age in days.",
			},
			"min_staging_hours": rschema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Minimum staging duration in hours before an artifact can pass the policy.",
			},
			"rules": rschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Deprecated compatibility field for the prose documentation's free-form rules examples. The current OpenAPI-backed endpoint does not accept rules, so this value is kept in Terraform state only and is not sent to Artifact Keeper.",
				DeprecationMessage:  "The current Artifact Keeper security policy API does not accept free-form rules. Use max_severity, block_unscanned, block_on_fail, require_signature, max_artifact_age_days, and min_staging_hours.",
			},
			"created_at": rschema.StringAttribute{
				Computed: true,
			},
			"updated_at": rschema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *securityPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !validateSecurityPolicy(plan, &resp.Diagnostics) {
		return
	}

	if !validateDeprecatedSecurityPolicyJSON(plan, &resp.Diagnostics) {
		return
	}

	addDeprecatedSecurityPolicyWarnings(plan, &resp.Diagnostics)

	policy, err := r.client.CreateSecurityPolicy(ctx, akclient.CreateSecurityPolicyRequest{
		RepositoryID:       optionalString(plan.RepositoryID),
		Name:               plan.Name.ValueString(),
		MaxSeverity:        severityThresholdForAPI(plan.MaxSeverity.ValueString()),
		BlockUnscanned:     plan.BlockUnscanned.ValueBool(),
		BlockOnFail:        plan.BlockOnFail.ValueBool(),
		RequireSignature:   plan.RequireSignature.ValueBool(),
		MaxArtifactAgeDays: intPointerFromInt64(plan.MaxArtifactAgeDays),
		MinStagingHours:    intPointerFromInt64(plan.MinStagingHours),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "Create Artifact Keeper security policy", err)
		return
	}

	if plan.Enabled.ValueBool() != policy.Enabled {
		policy, err = r.client.UpdateSecurityPolicy(ctx, policy.ID, securityPolicyUpdateRequestFromModel(plan))
		if err != nil {
			addClientError(&resp.Diagnostics, "Set Artifact Keeper security policy enabled state after create", err)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, securityPolicyModelFromAPI(policy, &plan))...)
}

func (r *securityPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetSecurityPolicy(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Read Artifact Keeper security policy", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, securityPolicyModelFromAPI(policy, &state))...)
}

func (r *securityPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan securityPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !validateSecurityPolicy(plan, &resp.Diagnostics) {
		return
	}

	if !validateDeprecatedSecurityPolicyJSON(plan, &resp.Diagnostics) {
		return
	}

	addDeprecatedSecurityPolicyWarnings(plan, &resp.Diagnostics)

	policy, err := r.client.UpdateSecurityPolicy(ctx, plan.ID.ValueString(), securityPolicyUpdateRequestFromModel(plan))
	if err != nil {
		addClientError(&resp.Diagnostics, "Update Artifact Keeper security policy", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, securityPolicyModelFromAPI(policy, &plan))...)
}

func (r *securityPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSecurityPolicy(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		addClientError(&resp.Diagnostics, "Delete Artifact Keeper security policy", err)
	}
}

func (r *securityPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func validateSecurityPolicy(model securityPolicyResourceModel, diagnostics *diag.Diagnostics) bool {
	switch strings.ToLower(model.MaxSeverity.ValueString()) {
	case "low", "medium", "high", "critical":
		return true
	default:
		diagnostics.AddAttributeError(path.Root("max_severity"), "Invalid max severity", "Use one of: low, medium, high, critical.")
		return false
	}
}

func validateDeprecatedSecurityPolicyJSON(model securityPolicyResourceModel, diagnostics *diag.Diagnostics) bool {
	if model.Rules.IsNull() || model.Rules.IsUnknown() {
		return true
	}
	var decoded []any
	if err := json.Unmarshal([]byte(model.Rules.ValueString()), &decoded); err != nil {
		diagnostics.AddAttributeError(path.Root("rules"), "Invalid JSON rules", "Rules must be a JSON array: "+err.Error())
		return false
	}
	return true
}

func addDeprecatedSecurityPolicyWarnings(model securityPolicyResourceModel, diagnostics *diag.Diagnostics) {
	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		diagnostics.AddWarning("Security policy description is not sent to Artifact Keeper", "The current security policy API does not persist descriptions, so Terraform keeps this value only in state.")
	}
	if !model.Rules.IsNull() && !model.Rules.IsUnknown() {
		diagnostics.AddWarning("Security policy rules are not sent to Artifact Keeper", "The current security policy API does not accept free-form rules. Configure supported controls with max_severity, block_unscanned, block_on_fail, require_signature, max_artifact_age_days, and min_staging_hours.")
	}
}

func securityPolicyUpdateRequestFromModel(model securityPolicyResourceModel) akclient.UpdateSecurityPolicyRequest {
	return akclient.UpdateSecurityPolicyRequest{
		Name:               model.Name.ValueString(),
		MaxSeverity:        severityThresholdForAPI(model.MaxSeverity.ValueString()),
		BlockUnscanned:     model.BlockUnscanned.ValueBool(),
		BlockOnFail:        model.BlockOnFail.ValueBool(),
		Enabled:            model.Enabled.ValueBool(),
		RequireSignature:   model.RequireSignature.ValueBool(),
		MaxArtifactAgeDays: intPointerFromInt64(model.MaxArtifactAgeDays),
		MinStagingHours:    intPointerFromInt64(model.MinStagingHours),
	}
}

func intPointerFromInt64(value types.Int64) *int {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := int(value.ValueInt64())
	return &v
}

func securityPolicyModelFromAPI(policy *akclient.SecurityPolicy, prior *securityPolicyResourceModel) securityPolicyResourceModel {
	maxSeverity := severityThresholdForTerraform(policy.MaxSeverity)
	if prior != nil && !prior.MaxSeverity.IsNull() && !prior.MaxSeverity.IsUnknown() && strings.EqualFold(prior.MaxSeverity.ValueString(), policy.MaxSeverity) {
		maxSeverity = prior.MaxSeverity.ValueString()
	}

	model := securityPolicyResourceModel{
		ID:                 types.StringValue(policy.ID),
		RepositoryID:       stringPtrValue(policy.RepositoryID),
		Name:               types.StringValue(policy.Name),
		Description:        types.StringNull(),
		Enabled:            types.BoolValue(policy.Enabled),
		MaxSeverity:        types.StringValue(maxSeverity),
		BlockUnscanned:     types.BoolValue(policy.BlockUnscanned),
		BlockOnFail:        types.BoolValue(policy.BlockOnFail),
		RequireSignature:   types.BoolValue(policy.RequireSignature),
		MaxArtifactAgeDays: intPtrValue(policy.MaxArtifactAgeDays),
		MinStagingHours:    intPtrValue(policy.MinStagingHours),
		Rules:              types.StringNull(),
		CreatedAt:          types.StringValue(policy.CreatedAt),
		UpdatedAt:          types.StringValue(policy.UpdatedAt),
	}

	if prior != nil {
		model.Description = prior.Description
		model.Rules = prior.Rules
	}

	return model
}

func intPtrValue(ptr *int) types.Int64 {
	if ptr == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*ptr))
}
