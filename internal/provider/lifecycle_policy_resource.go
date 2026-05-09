package provider

import (
	"context"
	"encoding/json"

	akclient "github.com/artifactkeeper/terraform-provider-artifactkeeper/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*lifecyclePolicyResource)(nil)
	_ resource.ResourceWithConfigure   = (*lifecyclePolicyResource)(nil)
	_ resource.ResourceWithImportState = (*lifecyclePolicyResource)(nil)
)

func NewLifecyclePolicyResource() resource.Resource {
	return &lifecyclePolicyResource{}
}

type lifecyclePolicyResource struct {
	client *akclient.Client
}

type lifecyclePolicyResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	RepositoryID        types.String `tfsdk:"repository_id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	PolicyType          types.String `tfsdk:"policy_type"`
	Config              types.String `tfsdk:"config"`
	Priority            types.Int64  `tfsdk:"priority"`
	LastRunAt           types.String `tfsdk:"last_run_at"`
	LastRunItemsRemoved types.Int64  `tfsdk:"last_run_items_removed"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func (r *lifecyclePolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lifecycle_policy"
}

func (r *lifecyclePolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = clientFromResourceData(req.ProviderData, &resp.Diagnostics)
}

func (r *lifecyclePolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages an Artifact Keeper lifecycle policy.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
			},
			"repository_id": rschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Repository UUID to scope this policy. Omit for a global policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": rschema.StringAttribute{
				Required: true,
			},
			"description": rschema.StringAttribute{
				Optional: true,
			},
			"enabled": rschema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"policy_type": rschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Lifecycle policy type. Supported values: `max_age_days`, `max_versions`, `no_downloads_days`, `tag_pattern_keep`, `tag_pattern_delete`, `size_quota_bytes`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config": rschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Policy-specific JSON configuration object. Examples: `{ \"days\": 90 }`, `{ \"keep\": 5 }`, `{ \"pattern\": \"^release-\" }`, `{ \"quota_bytes\": 10737418240 }`.",
			},
			"priority": rschema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
			"last_run_at": rschema.StringAttribute{
				Computed: true,
			},
			"last_run_items_removed": rschema.Int64Attribute{
				Computed: true,
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

func (r *lifecyclePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan lifecyclePolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !validateLifecyclePolicy(plan, &resp.Diagnostics) {
		return
	}

	config, ok := lifecyclePolicyConfig(plan.Config, &resp.Diagnostics)
	if !ok {
		return
	}

	enabled := plan.Enabled.ValueBool()
	policy, err := r.client.CreateLifecyclePolicy(ctx, akclient.CreateLifecyclePolicyRequest{
		RepositoryID: optionalString(plan.RepositoryID),
		Name:         plan.Name.ValueString(),
		Description:  optionalString(plan.Description),
		Enabled:      &enabled,
		PolicyType:   plan.PolicyType.ValueString(),
		Config:       config,
		Priority:     int(plan.Priority.ValueInt64()),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "Create Artifact Keeper lifecycle policy", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, lifecyclePolicyModelFromAPI(policy, &plan))...)
}

func (r *lifecyclePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state lifecyclePolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetLifecyclePolicy(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Read Artifact Keeper lifecycle policy", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, lifecyclePolicyModelFromAPI(policy, &state))...)
}

func (r *lifecyclePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan lifecyclePolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !validateLifecyclePolicy(plan, &resp.Diagnostics) {
		return
	}

	config, ok := lifecyclePolicyConfig(plan.Config, &resp.Diagnostics)
	if !ok {
		return
	}

	policy, err := r.client.UpdateLifecyclePolicy(ctx, plan.ID.ValueString(), akclient.UpdateLifecyclePolicyRequest{
		Name:        plan.Name.ValueString(),
		Description: optionalString(plan.Description),
		Enabled:     plan.Enabled.ValueBool(),
		Config:      config,
		Priority:    int(plan.Priority.ValueInt64()),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "Update Artifact Keeper lifecycle policy", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, lifecyclePolicyModelFromAPI(policy, &plan))...)
}

func (r *lifecyclePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state lifecyclePolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLifecyclePolicy(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		addClientError(&resp.Diagnostics, "Delete Artifact Keeper lifecycle policy", err)
	}
}

func (r *lifecyclePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func validateLifecyclePolicy(model lifecyclePolicyResourceModel, diagnostics *diag.Diagnostics) bool {
	switch model.PolicyType.ValueString() {
	case "max_age_days", "max_versions", "no_downloads_days", "tag_pattern_keep", "tag_pattern_delete", "size_quota_bytes":
	default:
		diagnostics.AddAttributeError(path.Root("policy_type"), "Invalid lifecycle policy type", "Use one of: max_age_days, max_versions, no_downloads_days, tag_pattern_keep, tag_pattern_delete, size_quota_bytes.")
		return false
	}

	switch model.PolicyType.ValueString() {
	case "max_versions", "size_quota_bytes":
		if model.RepositoryID.IsNull() || model.RepositoryID.IsUnknown() || model.RepositoryID.ValueString() == "" {
			diagnostics.AddAttributeError(path.Root("repository_id"), "Repository ID required", "The max_versions and size_quota_bytes policy types must be scoped to a repository.")
			return false
		}
	}

	return true
}

func lifecyclePolicyConfig(value types.String, diagnostics *diag.Diagnostics) (json.RawMessage, bool) {
	var decoded map[string]any
	if err := json.Unmarshal([]byte(value.ValueString()), &decoded); err != nil {
		diagnostics.AddAttributeError(path.Root("config"), "Invalid JSON config", "Config must be a JSON object: "+err.Error())
		return nil, false
	}

	canonical, err := canonicalJSONString(value.ValueString())
	if err != nil {
		diagnostics.AddAttributeError(path.Root("config"), "Invalid JSON config", err.Error())
		return nil, false
	}

	return json.RawMessage(canonical), true
}

func lifecyclePolicyModelFromAPI(policy *akclient.LifecyclePolicy, prior *lifecyclePolicyResourceModel) lifecyclePolicyResourceModel {
	config := string(policy.Config)
	if prior != nil && !prior.Config.IsNull() && !prior.Config.IsUnknown() && jsonStringsEqual(prior.Config.ValueString(), config) {
		config = prior.Config.ValueString()
	} else if canonical, err := canonicalJSONString(config); err == nil {
		config = canonical
	}

	return lifecyclePolicyResourceModel{
		ID:                  types.StringValue(policy.ID),
		RepositoryID:        stringPtrValue(policy.RepositoryID),
		Name:                types.StringValue(policy.Name),
		Description:         stringPtrValue(policy.Description),
		Enabled:             types.BoolValue(policy.Enabled),
		PolicyType:          types.StringValue(policy.PolicyType),
		Config:              types.StringValue(config),
		Priority:            types.Int64Value(int64(policy.Priority)),
		LastRunAt:           stringPtrValue(policy.LastRunAt),
		LastRunItemsRemoved: int64PtrValue(policy.LastRunItemsRemoved),
		CreatedAt:           types.StringValue(policy.CreatedAt),
		UpdatedAt:           types.StringValue(policy.UpdatedAt),
	}
}
