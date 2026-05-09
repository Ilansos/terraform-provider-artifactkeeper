package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"

	akclient "github.com/artifactkeeper/terraform-provider-artifactkeeper/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = (*artifactKeeperProvider)(nil)

type artifactKeeperProvider struct {
	version string
}

type providerModel struct {
	URL                types.String `tfsdk:"url"`
	BaseURL            types.String `tfsdk:"base_url"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	Token              types.String `tfsdk:"token"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &artifactKeeperProvider{version: version}
	}
}

func (p *artifactKeeperProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "artifactkeeper"
	resp.Version = p.version
}

func (p *artifactKeeperProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Terraform provider for Artifact Keeper.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Artifact Keeper base URL. Accepts either the server root or a URL already ending in `/api/v1`. Can also be set with `ARTIFACTKEEPER_URL`.",
			},
			"base_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Alias for `url`. Do not set both.",
			},
			"username": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Username for password login. Can also be set with `ARTIFACTKEEPER_USERNAME`.",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Password for password login. Can also be set with `ARTIFACTKEEPER_PASSWORD`.",
			},
			"token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Bearer token. Can also be set with `ARTIFACTKEEPER_TOKEN`.",
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Skip TLS certificate verification for lab or development instances. Can also be set with `ARTIFACTKEEPER_INSECURE_SKIP_VERIFY`.",
			},
		},
	}
}

func (p *artifactKeeperProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	urlValue := valueOrEnv(config.URL, "ARTIFACTKEEPER_URL")
	baseURLValue := valueOrEnv(config.BaseURL, "")
	if urlValue != "" && baseURLValue != "" {
		resp.Diagnostics.AddAttributeError(path.Root("base_url"), "Conflicting provider URL configuration", "Set only one of url or base_url.")
		return
	}
	if urlValue == "" {
		urlValue = baseURLValue
	}

	username := valueOrEnv(config.Username, "ARTIFACTKEEPER_USERNAME")
	password := valueOrEnv(config.Password, "ARTIFACTKEEPER_PASSWORD")
	token := valueOrEnv(config.Token, "ARTIFACTKEEPER_TOKEN")
	insecure, err := boolValueOrEnv(config.InsecureSkipVerify, "ARTIFACTKEEPER_INSECURE_SKIP_VERIFY")
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("insecure_skip_verify"), "Invalid boolean value", err.Error())
		return
	}

	if urlValue == "" {
		resp.Diagnostics.AddAttributeError(path.Root("url"), "Missing Artifact Keeper URL", "Set url, base_url, or ARTIFACTKEEPER_URL.")
		return
	}
	if token == "" && (username == "" || password == "") {
		resp.Diagnostics.AddError("Missing Artifact Keeper credentials", "Set token/ARTIFACTKEEPER_TOKEN, or username and password via provider configuration or environment variables.")
		return
	}

	c, err := akclient.New(ctx, akclient.Config{
		BaseURL:            urlValue,
		Username:           username,
		Password:           password,
		Token:              token,
		InsecureSkipVerify: insecure,
	})
	if err != nil {
		resp.Diagnostics.AddError("Configure Artifact Keeper client", sanitizeError(err))
		return
	}

	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *artifactKeeperProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRepositoryResource,
		NewUserResource,
		NewGroupResource,
		NewPermissionResource,
		NewWebhookResource,
		NewLifecyclePolicyResource,
		NewSecurityPolicyResource,
		NewLicensePolicyResource,
	}
}

func (p *artifactKeeperProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRepositoryDataSource,
		NewRepositoriesDataSource,
		NewUserDataSource,
		NewUsersDataSource,
		NewPackageDataSource,
		NewPackagesDataSource,
	}
}

func valueOrEnv(value types.String, envName string) string {
	if !value.IsNull() && !value.IsUnknown() {
		return value.ValueString()
	}
	if envName == "" {
		return ""
	}
	return os.Getenv(envName)
}

func boolValueOrEnv(value types.Bool, envName string) (bool, error) {
	if !value.IsNull() && !value.IsUnknown() {
		return value.ValueBool(), nil
	}
	raw := os.Getenv(envName)
	if raw == "" {
		return false, nil
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean value", envName)
	}
	return parsed, nil
}

func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
