package provider

import (
	"context"

	akclient "github.com/artifactkeeper/terraform-provider-artifactkeeper/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*permissionResource)(nil)
	_ resource.ResourceWithConfigure   = (*permissionResource)(nil)
	_ resource.ResourceWithImportState = (*permissionResource)(nil)
)

func NewPermissionResource() resource.Resource {
	return &permissionResource{}
}

type permissionResource struct {
	client *akclient.Client
}

type permissionResourceModel struct {
	ID           types.String `tfsdk:"id"`
	RepositoryID types.String `tfsdk:"repository_id"`
	UserID       types.String `tfsdk:"user_id"`
	GroupID      types.String `tfsdk:"group_id"`
	Permissions  types.Set    `tfsdk:"permissions"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func (r *permissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission"
}

func (r *permissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = clientFromResourceData(req.ProviderData, &resp.Diagnostics)
}

func (r *permissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages an Artifact Keeper repository permission.",
		Attributes: map[string]rschema.Attribute{
			"id":            rschema.StringAttribute{Computed: true},
			"repository_id": rschema.StringAttribute{Required: true},
			"user_id":       rschema.StringAttribute{Optional: true},
			"group_id":      rschema.StringAttribute{Optional: true},
			"permissions": rschema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
			"created_at": rschema.StringAttribute{Computed: true},
			"updated_at": rschema.StringAttribute{Computed: true},
		},
	}
}

func (r *permissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan permissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !validatePermissionPrincipal(plan, &resp.Diagnostics) {
		return
	}

	permission, err := r.client.CreatePermission(ctx, permissionRequestFromModel(ctx, plan, &resp.Diagnostics))
	if err != nil {
		addClientError(&resp.Diagnostics, "Create Artifact Keeper permission", err)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, permissionModelFromAPI(ctx, permission, &resp.Diagnostics))...)
}

func (r *permissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state permissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	permission, err := r.client.GetPermission(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Read Artifact Keeper permission", err)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, permissionModelFromAPI(ctx, permission, &resp.Diagnostics))...)
}

func (r *permissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan permissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || !validatePermissionPrincipal(plan, &resp.Diagnostics) {
		return
	}

	permission, err := r.client.UpdatePermission(ctx, plan.ID.ValueString(), permissionRequestFromModel(ctx, plan, &resp.Diagnostics))
	if err != nil {
		addClientError(&resp.Diagnostics, "Update Artifact Keeper permission", err)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, permissionModelFromAPI(ctx, permission, &resp.Diagnostics))...)
}

func (r *permissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state permissionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeletePermission(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		addClientError(&resp.Diagnostics, "Delete Artifact Keeper permission", err)
	}
}

func (r *permissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func validatePermissionPrincipal(model permissionResourceModel, diagnostics *diag.Diagnostics) bool {
	hasUser := !model.UserID.IsNull() && model.UserID.ValueString() != ""
	hasGroup := !model.GroupID.IsNull() && model.GroupID.ValueString() != ""
	if hasUser == hasGroup {
		diagnostics.AddError("Invalid permission principal", "Set exactly one of user_id or group_id.")
		return false
	}
	return true
}

func permissionRequestFromModel(ctx context.Context, model permissionResourceModel, diagnostics *diag.Diagnostics) akclient.CreatePermissionRequest {
	principalType := "user"
	principalID := model.UserID.ValueString()
	if !model.GroupID.IsNull() && model.GroupID.ValueString() != "" {
		principalType = "group"
		principalID = model.GroupID.ValueString()
	}

	return akclient.CreatePermissionRequest{
		PrincipalType: principalType,
		PrincipalID:   principalID,
		TargetType:    "repository",
		TargetID:      model.RepositoryID.ValueString(),
		Actions:       setStrings(ctx, model.Permissions, diagnostics),
	}
}

func permissionModelFromAPI(ctx context.Context, permission *akclient.Permission, diagnostics *diag.Diagnostics) permissionResourceModel {
	model := permissionResourceModel{
		ID:           types.StringValue(permission.ID),
		RepositoryID: types.StringValue(permission.TargetID),
		UserID:       types.StringNull(),
		GroupID:      types.StringNull(),
		Permissions:  stringSet(ctx, permission.Actions, diagnostics),
		CreatedAt:    types.StringValue(permission.CreatedAt),
		UpdatedAt:    types.StringValue(permission.UpdatedAt),
	}
	if permission.PrincipalType == "group" {
		model.GroupID = types.StringValue(permission.PrincipalID)
	} else {
		model.UserID = types.StringValue(permission.PrincipalID)
	}
	return model
}
