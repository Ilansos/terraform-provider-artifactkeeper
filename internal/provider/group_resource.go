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
	_ resource.Resource                = (*groupResource)(nil)
	_ resource.ResourceWithConfigure   = (*groupResource)(nil)
	_ resource.ResourceWithImportState = (*groupResource)(nil)
)

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

type groupResource struct {
	client *akclient.Client
}

type groupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	UserIDs     types.Set    `tfsdk:"user_ids"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	MemberCount types.Int64  `tfsdk:"member_count"`
}

func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = clientFromResourceData(req.ProviderData, &resp.Diagnostics)
}

func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages an Artifact Keeper group.",
		Attributes: map[string]rschema.Attribute{
			"id":          rschema.StringAttribute{Computed: true},
			"name":        rschema.StringAttribute{Required: true},
			"description": rschema.StringAttribute{Optional: true},
			"user_ids": rschema.SetAttribute{
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "User UUIDs that should be group members.",
			},
			"created_at":   rschema.StringAttribute{Computed: true},
			"updated_at":   rschema.StringAttribute{Computed: true},
			"member_count": rschema.Int64Attribute{Computed: true},
		},
	}
}

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.CreateGroup(ctx, akclient.CreateGroupRequest{
		Name:        plan.Name.ValueString(),
		Description: optionalString(plan.Description),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "Create Artifact Keeper group", err)
		return
	}

	userIDs := setStrings(ctx, plan.UserIDs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.AddGroupMembers(ctx, group.ID, userIDs); err != nil {
		addClientError(&resp.Diagnostics, "Add Artifact Keeper group members", err)
		return
	}

	group, err = r.client.GetGroup(ctx, group.ID)
	if err != nil {
		addClientError(&resp.Diagnostics, "Read Artifact Keeper group after create", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, groupModelFromAPI(ctx, group, &resp.Diagnostics))...)
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetGroup(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Read Artifact Keeper group", err)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, groupModelFromAPI(ctx, group, &resp.Diagnostics))...)
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan groupResourceModel
	var state groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.UpdateGroup(ctx, state.ID.ValueString(), akclient.CreateGroupRequest{
		Name:        plan.Name.ValueString(),
		Description: optionalString(plan.Description),
	}); err != nil {
		addClientError(&resp.Diagnostics, "Update Artifact Keeper group", err)
		return
	}

	current, err := r.client.GetGroup(ctx, state.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Read Artifact Keeper group members", err)
		return
	}

	var currentIDs []string
	for _, member := range current.Members {
		currentIDs = append(currentIDs, member.UserID)
	}
	add, remove := diffStrings(setStrings(ctx, plan.UserIDs, &resp.Diagnostics), currentIDs)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RemoveGroupMembers(ctx, state.ID.ValueString(), remove); err != nil {
		addClientError(&resp.Diagnostics, "Remove Artifact Keeper group members", err)
		return
	}
	if err := r.client.AddGroupMembers(ctx, state.ID.ValueString(), add); err != nil {
		addClientError(&resp.Diagnostics, "Add Artifact Keeper group members", err)
		return
	}

	group, err := r.client.GetGroup(ctx, state.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Read Artifact Keeper group after update", err)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, groupModelFromAPI(ctx, group, &resp.Diagnostics))...)
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteGroup(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		addClientError(&resp.Diagnostics, "Delete Artifact Keeper group", err)
	}
}

func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func groupModelFromAPI(ctx context.Context, group *akclient.Group, diagnostics *diag.Diagnostics) groupResourceModel {
	userIDs := make([]string, 0, len(group.Members))
	for _, member := range group.Members {
		userIDs = append(userIDs, member.UserID)
	}
	return groupResourceModel{
		ID:          types.StringValue(group.ID),
		Name:        types.StringValue(group.Name),
		Description: stringPtrValue(group.Description),
		UserIDs:     stringSet(ctx, userIDs, diagnostics),
		CreatedAt:   types.StringValue(group.CreatedAt),
		UpdatedAt:   types.StringValue(group.UpdatedAt),
		MemberCount: types.Int64Value(group.MemberCount),
	}
}
