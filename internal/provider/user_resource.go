package provider

import (
	"context"

	akclient "github.com/artifactkeeper/terraform-provider-artifactkeeper/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*userResource)(nil)
	_ resource.ResourceWithConfigure   = (*userResource)(nil)
	_ resource.ResourceWithImportState = (*userResource)(nil)
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *akclient.Client
}

type userResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	Email     types.String `tfsdk:"email"`
	Password  types.String `tfsdk:"password"`
	Role      types.String `tfsdk:"role"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = clientFromResourceData(req.ProviderData, &resp.Diagnostics)
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages an Artifact Keeper user.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{Computed: true},
			"username": rschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": rschema.StringAttribute{Required: true},
			"password": rschema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Initial password. Artifact Keeper does not return passwords, so the provider never reads this value back from the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": rschema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("user"),
				MarkdownDescription: "`user` or `admin`. This maps to the API `is_admin` flag.",
			},
			"created_at": rschema.StringAttribute{Computed: true},
		},
	}
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.Role.ValueString() != "user" && plan.Role.ValueString() != "admin" {
		resp.Diagnostics.AddAttributeError(path.Root("role"), "Invalid role", "Role must be either user or admin.")
		return
	}

	admin := adminFromRole(plan.Role.ValueString())
	user, err := r.client.CreateUser(ctx, akclient.CreateUserRequest{
		Username: plan.Username.ValueString(),
		Email:    plan.Email.ValueString(),
		Password: optionalString(plan.Password),
		IsAdmin:  &admin,
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "Create Artifact Keeper user", err)
		return
	}

	state := userModelFromAPI(user)
	state.Password = plan.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Read Artifact Keeper user", err)
		return
	}

	newState := userModelFromAPI(user)
	newState.Password = state.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan.Role.ValueString() != "user" && plan.Role.ValueString() != "admin" {
		resp.Diagnostics.AddAttributeError(path.Root("role"), "Invalid role", "Role must be either user or admin.")
		return
	}

	admin := adminFromRole(plan.Role.ValueString())
	user, err := r.client.UpdateUser(ctx, plan.ID.ValueString(), akclient.UpdateUserRequest{
		Email:   optionalString(plan.Email),
		IsAdmin: &admin,
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "Update Artifact Keeper user", err)
		return
	}

	state := userModelFromAPI(user)
	state.Password = plan.Password
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteUser(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		addClientError(&resp.Diagnostics, "Delete Artifact Keeper user", err)
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func userModelFromAPI(user *akclient.User) userResourceModel {
	return userResourceModel{
		ID:        types.StringValue(user.ID),
		Username:  types.StringValue(user.Username),
		Email:     types.StringValue(user.Email),
		Password:  types.StringNull(),
		Role:      types.StringValue(roleFromAdmin(user.IsAdmin)),
		CreatedAt: types.StringValue(user.CreatedAt),
	}
}
