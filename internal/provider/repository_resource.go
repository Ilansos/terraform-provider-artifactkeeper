package provider

import (
	"context"
	"strings"

	akclient "github.com/artifactkeeper/terraform-provider-artifactkeeper/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*repositoryResource)(nil)
	_ resource.ResourceWithConfigure   = (*repositoryResource)(nil)
	_ resource.ResourceWithImportState = (*repositoryResource)(nil)
)

func NewRepositoryResource() resource.Resource {
	return &repositoryResource{}
}

type repositoryResource struct {
	client *akclient.Client
}

type repositoryResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Key                    types.String `tfsdk:"key"`
	Name                   types.String `tfsdk:"name"`
	Format                 types.String `tfsdk:"format"`
	RepoType               types.String `tfsdk:"repo_type"`
	Description            types.String `tfsdk:"description"`
	Public                 types.Bool   `tfsdk:"public"`
	QuotaBytes             types.Int64  `tfsdk:"quota_bytes"`
	Repositories           types.List   `tfsdk:"repositories"`
	ScanEnabled            types.Bool   `tfsdk:"scan_enabled"`
	ScanOnUpload           types.Bool   `tfsdk:"scan_on_upload"`
	ScanOnProxy            types.Bool   `tfsdk:"scan_on_proxy"`
	BlockOnViolation       types.Bool   `tfsdk:"block_on_violation"`
	SeverityThreshold      types.String `tfsdk:"severity_threshold"`
	UpstreamURL            types.String `tfsdk:"upstream_url"`
	IndexUpstreamURL       types.String `tfsdk:"index_upstream_url"`
	UpstreamAuthType       types.String `tfsdk:"upstream_auth_type"`
	UpstreamUsername       types.String `tfsdk:"upstream_username"`
	UpstreamPassword       types.String `tfsdk:"upstream_password"`
	UpstreamAuthConfigured types.Bool   `tfsdk:"upstream_auth_configured"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	SizeBytes              types.Int64  `tfsdk:"size_bytes"`
}

func (r *repositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (r *repositoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = clientFromResourceData(req.ProviderData, &resp.Diagnostics)
}

func (r *repositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages an Artifact Keeper repository.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Repository UUID assigned by Artifact Keeper.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": rschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Stable repository key used in Artifact Keeper API paths.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": rschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name.",
			},
			"format": rschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Repository format, such as docker, npm, maven, or generic.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repo_type": rschema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("local"),
				MarkdownDescription: "Repository type. Defaults to `local`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": rschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Repository description.",
			},
			"public": rschema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the repository is public.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"quota_bytes": rschema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Optional repository quota in bytes.",
			},
			"repositories": rschema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Ordered member repository keys for virtual repositories. The provider maps list order to Artifact Keeper virtual member priorities.",
			},
			"scan_enabled": rschema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether repository security scanning is enabled.",
			},
			"scan_on_upload": rschema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether artifacts are scanned when uploaded.",
			},
			"scan_on_proxy": rschema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether artifacts are scanned when proxied from an upstream repository.",
			},
			"block_on_violation": rschema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether policy violations block repository operations.",
			},
			"severity_threshold": rschema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Minimum severity threshold for blocking or policy evaluation. Valid values: `Low`, `Medium`, `High`, `Critical`.",
			},
			"upstream_url": rschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Upstream URL for remote repositories.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"index_upstream_url": rschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional separate upstream index URL for formats such as Cargo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"upstream_auth_type": rschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Upstream authentication type for remote repositories, such as `basic` or `bearer`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"upstream_username": rschema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Username for upstream basic authentication. Artifact Keeper does not return this value.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"upstream_password": rschema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Password or token for upstream authentication. Artifact Keeper treats this value as write-only.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"upstream_auth_configured": rschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether upstream authentication is configured for this repository.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": rschema.StringAttribute{
				Computed: true,
			},
			"size_bytes": rschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Repository storage usage in bytes.",
			},
		},
	}
}

func (r *repositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan repositoryResourceModel
	var config repositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !validateRepositoryMembers(plan, &resp.Diagnostics) {
		return
	}
	if !validateRepositorySecurityConfig(config, &resp.Diagnostics) {
		return
	}

	public := plan.Public.ValueBool()
	repository, err := r.client.CreateRepository(ctx, akclient.CreateRepositoryRequest{
		Key:              plan.Key.ValueString(),
		Name:             plan.Name.ValueString(),
		Format:           plan.Format.ValueString(),
		RepoType:         plan.RepoType.ValueString(),
		Description:      optionalString(plan.Description),
		IsPublic:         &public,
		QuotaBytes:       optionalInt64(plan.QuotaBytes),
		MemberRepos:      createVirtualMembersFromList(ctx, plan.Repositories, &resp.Diagnostics),
		UpstreamURL:      optionalString(plan.UpstreamURL),
		IndexUpstreamURL: optionalString(plan.IndexUpstreamURL),
		UpstreamAuthType: optionalString(plan.UpstreamAuthType),
		UpstreamUsername: optionalString(plan.UpstreamUsername),
		UpstreamPassword: optionalString(plan.UpstreamPassword),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "Create Artifact Keeper repository", err)
		return
	}

	state := repositoryModelFromAPI(repository, &plan)
	if repositorySecurityConfigured(config) {
		security, err := r.client.UpsertRepositorySecurity(ctx, repository.Key, repositorySecurityRequestFromModel(plan))
		if err != nil {
			addClientError(&resp.Diagnostics, "Configure Artifact Keeper repository security", err)
			return
		}
		applyRepositorySecurity(&state, security, plan.SeverityThreshold)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *repositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state repositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repository, err := r.client.GetRepository(ctx, state.Key.ValueString())
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Read Artifact Keeper repository", err)
		return
	}

	newState := repositoryModelFromAPI(repository, &state)
	if repository.RepoType == "virtual" {
		members, err := r.client.GetVirtualMembers(ctx, repository.Key)
		if err != nil {
			addClientError(&resp.Diagnostics, "Read Artifact Keeper virtual repository members", err)
			return
		}
		newState.Repositories = stringList(ctx, virtualMemberKeys(members.Members), &resp.Diagnostics)
	}
	security, err := r.client.GetRepositorySecurity(ctx, repository.Key)
	if err != nil {
		addClientError(&resp.Diagnostics, "Read Artifact Keeper repository security", err)
		return
	}
	if security.Config != nil {
		applyRepositorySecurity(&newState, security.Config, state.SeverityThreshold)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *repositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan repositoryResourceModel
	var config repositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !validateRepositoryMembers(plan, &resp.Diagnostics) {
		return
	}
	if !validateRepositorySecurityConfig(config, &resp.Diagnostics) {
		return
	}

	public := plan.Public.ValueBool()
	repository, err := r.client.UpdateRepository(ctx, plan.Key.ValueString(), akclient.UpdateRepositoryRequest{
		Name:        optionalString(plan.Name),
		Description: optionalString(plan.Description),
		IsPublic:    &public,
		QuotaBytes:  optionalInt64(plan.QuotaBytes),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "Update Artifact Keeper repository", err)
		return
	}

	state := repositoryModelFromAPI(repository, &plan)
	if plan.RepoType.ValueString() == "virtual" && !plan.Repositories.IsNull() && !plan.Repositories.IsUnknown() {
		members, err := r.client.UpdateVirtualMembers(ctx, plan.Key.ValueString(), virtualMemberPrioritiesFromList(ctx, plan.Repositories, &resp.Diagnostics))
		if err != nil {
			addClientError(&resp.Diagnostics, "Update Artifact Keeper virtual repository members", err)
			return
		}
		state.Repositories = stringList(ctx, virtualMemberKeys(members.Members), &resp.Diagnostics)
	}
	if repositorySecurityConfigured(config) {
		security, err := r.client.UpsertRepositorySecurity(ctx, plan.Key.ValueString(), repositorySecurityRequestFromModel(plan))
		if err != nil {
			addClientError(&resp.Diagnostics, "Configure Artifact Keeper repository security", err)
			return
		}
		applyRepositorySecurity(&state, security, plan.SeverityThreshold)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *repositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state repositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRepository(ctx, state.Key.ValueString()); err != nil && !isNotFound(err) {
		addClientError(&resp.Diagnostics, "Delete Artifact Keeper repository", err)
	}
}

func (r *repositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}

func repositoryModelFromAPI(repository *akclient.Repository, prior *repositoryResourceModel) repositoryResourceModel {
	model := repositoryResourceModel{
		ID:                     types.StringValue(repository.ID),
		Key:                    types.StringValue(repository.Key),
		Name:                   types.StringValue(repository.Name),
		Format:                 types.StringValue(repository.Format),
		RepoType:               types.StringValue(repository.RepoType),
		Description:            stringPtrValue(repository.Description),
		Public:                 types.BoolValue(repository.IsPublic),
		QuotaBytes:             int64PtrValue(repository.QuotaBytes),
		Repositories:           types.ListNull(types.StringType),
		ScanEnabled:            types.BoolNull(),
		ScanOnUpload:           types.BoolNull(),
		ScanOnProxy:            types.BoolNull(),
		BlockOnViolation:       types.BoolNull(),
		SeverityThreshold:      types.StringNull(),
		UpstreamURL:            stringPtrValue(repository.UpstreamURL),
		IndexUpstreamURL:       types.StringNull(),
		UpstreamAuthType:       stringPtrValue(repository.UpstreamAuthType),
		UpstreamUsername:       types.StringNull(),
		UpstreamPassword:       types.StringNull(),
		UpstreamAuthConfigured: types.BoolValue(repository.UpstreamAuthConfigured),
		CreatedAt:              types.StringValue(repository.CreatedAt),
		UpdatedAt:              types.StringValue(repository.UpdatedAt),
		SizeBytes:              types.Int64Value(repository.StorageUsedBytes),
	}
	if prior != nil {
		model.IndexUpstreamURL = prior.IndexUpstreamURL
		model.Repositories = prior.Repositories
		model.ScanEnabled = prior.ScanEnabled
		model.ScanOnUpload = prior.ScanOnUpload
		model.ScanOnProxy = prior.ScanOnProxy
		model.BlockOnViolation = prior.BlockOnViolation
		model.SeverityThreshold = prior.SeverityThreshold
		model.UpstreamUsername = prior.UpstreamUsername
		model.UpstreamPassword = prior.UpstreamPassword
		if model.UpstreamAuthType.IsNull() {
			model.UpstreamAuthType = prior.UpstreamAuthType
		}
		if model.UpstreamURL.IsNull() {
			model.UpstreamURL = prior.UpstreamURL
		}
	}
	return model
}

func validateRepositoryMembers(model repositoryResourceModel, diagnostics *diag.Diagnostics) bool {
	if !model.Repositories.IsNull() && model.RepoType.ValueString() != "virtual" {
		diagnostics.AddAttributeError(path.Root("repositories"), "Invalid repository members", "The repositories attribute is only valid when repo_type is virtual.")
		return false
	}
	return true
}

func validateRepositorySecurityConfig(model repositoryResourceModel, diagnostics *diag.Diagnostics) bool {
	if !repositorySecurityConfigured(model) {
		return true
	}

	missing := []string{}
	if model.ScanEnabled.IsNull() || model.ScanEnabled.IsUnknown() {
		missing = append(missing, "scan_enabled")
	}
	if model.ScanOnUpload.IsNull() || model.ScanOnUpload.IsUnknown() {
		missing = append(missing, "scan_on_upload")
	}
	if model.ScanOnProxy.IsNull() || model.ScanOnProxy.IsUnknown() {
		missing = append(missing, "scan_on_proxy")
	}
	if model.BlockOnViolation.IsNull() || model.BlockOnViolation.IsUnknown() {
		missing = append(missing, "block_on_violation")
	}
	if model.SeverityThreshold.IsNull() || model.SeverityThreshold.IsUnknown() {
		missing = append(missing, "severity_threshold")
	}
	if len(missing) > 0 {
		diagnostics.AddError("Incomplete repository security configuration", "When configuring repository security, set all of scan_enabled, scan_on_upload, scan_on_proxy, block_on_violation, and severity_threshold.")
		return false
	}

	switch model.SeverityThreshold.ValueString() {
	case "Low", "Medium", "High", "Critical", "low", "medium", "high", "critical":
		return true
	default:
		diagnostics.AddAttributeError(path.Root("severity_threshold"), "Invalid severity threshold", "Use one of: Low, Medium, High, Critical.")
		return false
	}
}

func repositorySecurityConfigured(model repositoryResourceModel) bool {
	return explicitlyConfigured(model.ScanEnabled) ||
		explicitlyConfigured(model.ScanOnUpload) ||
		explicitlyConfigured(model.ScanOnProxy) ||
		explicitlyConfigured(model.BlockOnViolation) ||
		explicitlyConfigured(model.SeverityThreshold)
}

func explicitlyConfigured(value any) bool {
	switch v := value.(type) {
	case types.Bool:
		return !v.IsNull() && !v.IsUnknown()
	case types.String:
		return !v.IsNull() && !v.IsUnknown()
	default:
		return false
	}
}

func repositorySecurityRequestFromModel(model repositoryResourceModel) akclient.UpsertScanConfigRequest {
	return akclient.UpsertScanConfigRequest{
		ScanEnabled:            model.ScanEnabled.ValueBool(),
		ScanOnUpload:           model.ScanOnUpload.ValueBool(),
		ScanOnProxy:            model.ScanOnProxy.ValueBool(),
		BlockOnPolicyViolation: model.BlockOnViolation.ValueBool(),
		SeverityThreshold:      severityThresholdForAPI(model.SeverityThreshold.ValueString()),
	}
}

func applyRepositorySecurity(model *repositoryResourceModel, security *akclient.ScanConfig, preferredSeverity types.String) {
	model.ScanEnabled = types.BoolValue(security.ScanEnabled)
	model.ScanOnUpload = types.BoolValue(security.ScanOnUpload)
	model.ScanOnProxy = types.BoolValue(security.ScanOnProxy)
	model.BlockOnViolation = types.BoolValue(security.BlockOnPolicyViolation)
	if !preferredSeverity.IsNull() && !preferredSeverity.IsUnknown() && strings.EqualFold(preferredSeverity.ValueString(), security.SeverityThreshold) {
		model.SeverityThreshold = preferredSeverity
		return
	}
	model.SeverityThreshold = types.StringValue(severityThresholdForTerraform(security.SeverityThreshold))
}

func severityThresholdForAPI(value string) string {
	return strings.ToLower(value)
}

func severityThresholdForTerraform(value string) string {
	switch strings.ToLower(value) {
	case "low":
		return "Low"
	case "medium":
		return "Medium"
	case "high":
		return "High"
	case "critical":
		return "Critical"
	default:
		return value
	}
}

func createVirtualMembersFromList(ctx context.Context, list types.List, diagnostics *diag.Diagnostics) []akclient.CreateVirtualMemberInput {
	keys := listStrings(ctx, list, diagnostics)
	members := make([]akclient.CreateVirtualMemberInput, 0, len(keys))
	for i, key := range keys {
		members = append(members, akclient.CreateVirtualMemberInput{
			RepoKey:  key,
			Priority: i + 1,
		})
	}
	return members
}

func virtualMemberPrioritiesFromList(ctx context.Context, list types.List, diagnostics *diag.Diagnostics) []akclient.VirtualMemberPriority {
	keys := listStrings(ctx, list, diagnostics)
	members := make([]akclient.VirtualMemberPriority, 0, len(keys))
	for i, key := range keys {
		members = append(members, akclient.VirtualMemberPriority{
			MemberKey: key,
			Priority:  i + 1,
		})
	}
	return members
}

func virtualMemberKeys(members []akclient.VirtualMember) []string {
	out := make([]string, 0, len(members))
	for _, member := range members {
		out = append(out, member.MemberRepoKey)
	}
	return out
}

func int64PtrValue(ptr *int64) types.Int64 {
	if ptr == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*ptr)
}
