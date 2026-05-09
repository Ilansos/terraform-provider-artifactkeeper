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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*webhookResource)(nil)
	_ resource.ResourceWithConfigure   = (*webhookResource)(nil)
	_ resource.ResourceWithImportState = (*webhookResource)(nil)
)

func NewWebhookResource() resource.Resource {
	return &webhookResource{}
}

type webhookResource struct {
	client *akclient.Client
}

type webhookResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	URL             types.String `tfsdk:"url"`
	Events          types.Set    `tfsdk:"events"`
	RepositoryID    types.String `tfsdk:"repository_id"`
	Secret          types.String `tfsdk:"secret"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	CreatedAt       types.String `tfsdk:"created_at"`
	LastTriggeredAt types.String `tfsdk:"last_triggered_at"`
}

func (r *webhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *webhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = clientFromResourceData(req.ProviderData, &resp.Diagnostics)
}

func (r *webhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages an Artifact Keeper webhook. The current API does not expose a webhook update endpoint, so most changes replace the resource.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{Computed: true},
			"name": rschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": rschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"events": rschema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"repository_id": rschema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"secret": rschema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Webhook signing secret. Artifact Keeper does not return this value.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": rschema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"created_at":        rschema.StringAttribute{Computed: true},
			"last_triggered_at": rschema.StringAttribute{Computed: true},
		},
	}
}

func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhook, err := r.client.CreateWebhook(ctx, akclient.CreateWebhookRequest{
		Name:         plan.Name.ValueString(),
		URL:          plan.URL.ValueString(),
		Events:       setStrings(ctx, plan.Events, &resp.Diagnostics),
		RepositoryID: optionalString(plan.RepositoryID),
		Secret:       optionalString(plan.Secret),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "Create Artifact Keeper webhook", err)
		return
	}
	if !plan.Enabled.ValueBool() {
		if err := r.client.DisableWebhook(ctx, webhook.ID); err != nil {
			addClientError(&resp.Diagnostics, "Disable Artifact Keeper webhook", err)
			return
		}
		webhook.IsEnabled = false
	}

	state := webhookModelFromAPI(ctx, webhook, &resp.Diagnostics)
	state.Secret = plan.Secret
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhook, err := r.client.GetWebhook(ctx, state.ID.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Read Artifact Keeper webhook", err)
		return
	}

	newState := webhookModelFromAPI(ctx, webhook, &resp.Diagnostics)
	newState.Secret = state.Secret
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webhookResourceModel
	var state webhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Enabled.ValueBool() && !state.Enabled.ValueBool() {
		if err := r.client.EnableWebhook(ctx, state.ID.ValueString()); err != nil {
			addClientError(&resp.Diagnostics, "Enable Artifact Keeper webhook", err)
			return
		}
	}
	if !plan.Enabled.ValueBool() && state.Enabled.ValueBool() {
		if err := r.client.DisableWebhook(ctx, state.ID.ValueString()); err != nil {
			addClientError(&resp.Diagnostics, "Disable Artifact Keeper webhook", err)
			return
		}
	}

	webhook, err := r.client.GetWebhook(ctx, state.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Read Artifact Keeper webhook after update", err)
		return
	}
	newState := webhookModelFromAPI(ctx, webhook, &resp.Diagnostics)
	newState.Secret = plan.Secret
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteWebhook(ctx, state.ID.ValueString()); err != nil && !isNotFound(err) {
		addClientError(&resp.Diagnostics, "Delete Artifact Keeper webhook", err)
	}
}

func (r *webhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func webhookModelFromAPI(ctx context.Context, webhook *akclient.Webhook, diagnostics *diag.Diagnostics) webhookResourceModel {
	return webhookResourceModel{
		ID:              types.StringValue(webhook.ID),
		Name:            types.StringValue(webhook.Name),
		URL:             types.StringValue(webhook.URL),
		Events:          stringSet(ctx, webhook.Events, diagnostics),
		RepositoryID:    stringPtrValue(webhook.RepositoryID),
		Secret:          types.StringNull(),
		Enabled:         types.BoolValue(webhook.IsEnabled),
		CreatedAt:       types.StringValue(webhook.CreatedAt),
		LastTriggeredAt: stringPtrValue(webhook.LastTriggeredAt),
	}
}
