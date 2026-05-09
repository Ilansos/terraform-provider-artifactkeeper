package provider

import (
	"context"

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
	_ resource.Resource                = (*licensePolicyResource)(nil)
	_ resource.ResourceWithConfigure   = (*licensePolicyResource)(nil)
	_ resource.ResourceWithImportState = (*licensePolicyResource)(nil)
)

func NewLicensePolicyResource() resource.Resource {
	return &licensePolicyResource{}
}

type licensePolicyResource struct {
	client *akclient.Client
}

type licensePolicyResourceModel struct {
	ID              types.String `tfsdk:"id"`
	RepositoryID    types.String `tfsdk:"repository_id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	AllowedLicenses types.Set    `tfsdk:"allowed_licenses"`
	DeniedLicenses  types.Set    `tfsdk:"denied_licenses"`
	AllowUnknown    types.Bool   `tfsdk:"allow_unknown"`
	Action          types.String `tfsdk:"action"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (r *licensePolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license_policy"
}

func (r *licensePolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = clientFromResourceData(req.ProviderData, &resp.Diagnostics)
}

func (r *licensePolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages an Artifact Keeper SBOM license compliance policy under `/api/v1/sbom/license-policies`.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "License policy UUID.",
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
				MarkdownDescription: "License policy name. The API upserts policies by name/scope, so changing this replaces the Terraform resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": rschema.StringAttribute{
				Optional: true,
			},
			"allowed_licenses": rschema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Set of SPDX license identifiers that are explicitly allowed. Use an empty set for deny-list mode.",
			},
			"denied_licenses": rschema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Set of SPDX license identifiers that are explicitly denied. Use an empty set for allow-list mode.",
			},
			"allow_unknown": rschema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether artifacts with unknown or unrecognized licenses are allowed.",
			},
			"action": rschema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("block"),
				MarkdownDescription: "Policy action when a license violation is found. Supported values: `block`, `warn`, `allow`.",
			},
			"enabled": rschema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
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

func (r *licensePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan licensePolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !validateLicensePolicy(plan, &resp.Diagnostics) {
		return
	}

	policy, err := r.client.UpsertLicensePolicy(ctx, licensePolicyRequestFromModel(ctx, plan, &resp.Diagnostics))
	if resp.Diagnostics.HasError() {
		return
	}
	if err != nil {
		addClientError(&resp.Diagnostics, "Create Artifact Keeper license policy", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, licensePolicyModelFromAPI(ctx, policy, &plan, &resp.Diagnostics))...)
}

func (r *licensePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state licensePolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetLicensePolicy(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Read Artifact Keeper license policy", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, licensePolicyModelFromAPI(ctx, policy, &state, &resp.Diagnostics))...)
}

func (r *licensePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan licensePolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !validateLicensePolicy(plan, &resp.Diagnostics) {
		return
	}

	policy, err := r.client.UpsertLicensePolicy(ctx, licensePolicyRequestFromModel(ctx, plan, &resp.Diagnostics))
	if resp.Diagnostics.HasError() {
		return
	}
	if err != nil {
		addClientError(&resp.Diagnostics, "Update Artifact Keeper license policy", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, licensePolicyModelFromAPI(ctx, policy, &plan, &resp.Diagnostics))...)
}

func (r *licensePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state licensePolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteLicensePolicy(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		addClientError(&resp.Diagnostics, "Delete Artifact Keeper license policy", err)
	}
}

func (r *licensePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func validateLicensePolicy(model licensePolicyResourceModel, diagnostics *diag.Diagnostics) bool {
	switch model.Action.ValueString() {
	case "block", "warn", "allow":
		return true
	default:
		diagnostics.AddAttributeError(path.Root("action"), "Invalid license policy action", "Use one of: block, warn, allow.")
		return false
	}
}

func licensePolicyRequestFromModel(ctx context.Context, model licensePolicyResourceModel, diagnostics *diag.Diagnostics) akclient.UpsertLicensePolicyRequest {
	return akclient.UpsertLicensePolicyRequest{
		RepositoryID:    optionalString(model.RepositoryID),
		Name:            model.Name.ValueString(),
		Description:     optionalString(model.Description),
		AllowedLicenses: setStrings(ctx, model.AllowedLicenses, diagnostics),
		DeniedLicenses:  setStrings(ctx, model.DeniedLicenses, diagnostics),
		AllowUnknown:    model.AllowUnknown.ValueBool(),
		Action:          model.Action.ValueString(),
		Enabled:         model.Enabled.ValueBool(),
	}
}

func licensePolicyModelFromAPI(ctx context.Context, policy *akclient.LicensePolicy, prior *licensePolicyResourceModel, diagnostics *diag.Diagnostics) licensePolicyResourceModel {
	return licensePolicyResourceModel{
		ID:              types.StringValue(policy.ID),
		RepositoryID:    stringPtrValue(policy.RepositoryID),
		Name:            types.StringValue(policy.Name),
		Description:     stringPtrValue(policy.Description),
		AllowedLicenses: stringSet(ctx, policy.AllowedLicenses, diagnostics),
		DeniedLicenses:  stringSet(ctx, policy.DeniedLicenses, diagnostics),
		AllowUnknown:    types.BoolValue(policy.AllowUnknown),
		Action:          types.StringValue(policy.Action),
		Enabled:         types.BoolValue(policy.Enabled),
		CreatedAt:       types.StringValue(policy.CreatedAt),
		UpdatedAt:       stringPtrValue(policy.UpdatedAt),
	}
}
