package provider

import (
	"context"

	akclient "github.com/artifactkeeper/terraform-provider-artifactkeeper/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type configuredDataSource struct {
	client *akclient.Client
}

func (d *configuredDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = clientFromResourceData(req.ProviderData, &resp.Diagnostics)
}

type repositoryDataSource struct{ configuredDataSource }
type repositoriesDataSource struct{ configuredDataSource }
type userDataSource struct{ configuredDataSource }
type usersDataSource struct{ configuredDataSource }
type packageDataSource struct{ configuredDataSource }
type packagesDataSource struct{ configuredDataSource }

var (
	_ datasource.DataSource              = (*repositoryDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*repositoryDataSource)(nil)
	_ datasource.DataSource              = (*repositoriesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*repositoriesDataSource)(nil)
	_ datasource.DataSource              = (*userDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*userDataSource)(nil)
	_ datasource.DataSource              = (*usersDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*usersDataSource)(nil)
	_ datasource.DataSource              = (*packageDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*packageDataSource)(nil)
	_ datasource.DataSource              = (*packagesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*packagesDataSource)(nil)
)

func NewRepositoryDataSource() datasource.DataSource   { return &repositoryDataSource{} }
func NewRepositoriesDataSource() datasource.DataSource { return &repositoriesDataSource{} }
func NewUserDataSource() datasource.DataSource         { return &userDataSource{} }
func NewUsersDataSource() datasource.DataSource        { return &usersDataSource{} }
func NewPackageDataSource() datasource.DataSource      { return &packageDataSource{} }
func NewPackagesDataSource() datasource.DataSource     { return &packagesDataSource{} }

type repositoryDataSourceModel struct {
	Key         types.String `tfsdk:"key"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Format      types.String `tfsdk:"format"`
	RepoType    types.String `tfsdk:"repo_type"`
	Description types.String `tfsdk:"description"`
	Public      types.Bool   `tfsdk:"public"`
	QuotaBytes  types.Int64  `tfsdk:"quota_bytes"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	SizeBytes   types.Int64  `tfsdk:"size_bytes"`
}

type repositoriesDataSourceModel struct {
	Format       types.String                `tfsdk:"format"`
	Type         types.String                `tfsdk:"type"`
	Query        types.String                `tfsdk:"query"`
	Repositories []repositoryDataSourceModel `tfsdk:"repositories"`
}

func (d *repositoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (d *repositoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Reads an Artifact Keeper repository by key.",
		Attributes:          repositoryDataSourceAttributes(true),
	}
}

func (d *repositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config repositoryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	repository, err := d.client.GetRepository(ctx, config.Key.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Read Artifact Keeper repository data source", err)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, repositoryDataSourceModelFromAPI(repository))...)
}

func (d *repositoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repositories"
}

func (d *repositoriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Lists Artifact Keeper repositories.",
		Attributes: map[string]dschema.Attribute{
			"format": dschema.StringAttribute{Optional: true},
			"type":   dschema.StringAttribute{Optional: true},
			"query":  dschema.StringAttribute{Optional: true},
			"repositories": dschema.ListNestedAttribute{
				Computed: true,
				NestedObject: dschema.NestedAttributeObject{
					Attributes: repositoryDataSourceAttributes(false),
				},
			},
		},
	}
}

func (d *repositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config repositoriesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := d.client.ListRepositories(ctx, akclient.RepositoryListOptions{
		PerPage: 100,
		Format:  valueOrEnv(config.Format, ""),
		Type:    valueOrEnv(config.Type, ""),
		Query:   valueOrEnv(config.Query, ""),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "List Artifact Keeper repositories", err)
		return
	}
	config.Repositories = make([]repositoryDataSourceModel, 0, len(result.Items))
	for _, repository := range result.Items {
		repository := repository
		config.Repositories = append(config.Repositories, repositoryDataSourceModelFromAPI(&repository))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

func repositoryDataSourceAttributes(keyRequired bool) map[string]dschema.Attribute {
	keyAttr := dschema.StringAttribute{Computed: true}
	if keyRequired {
		keyAttr = dschema.StringAttribute{Required: true}
	}
	return map[string]dschema.Attribute{
		"key":         keyAttr,
		"id":          dschema.StringAttribute{Computed: true},
		"name":        dschema.StringAttribute{Computed: true},
		"format":      dschema.StringAttribute{Computed: true},
		"repo_type":   dschema.StringAttribute{Computed: true},
		"description": dschema.StringAttribute{Computed: true},
		"public":      dschema.BoolAttribute{Computed: true},
		"quota_bytes": dschema.Int64Attribute{Computed: true},
		"created_at":  dschema.StringAttribute{Computed: true},
		"updated_at":  dschema.StringAttribute{Computed: true},
		"size_bytes":  dschema.Int64Attribute{Computed: true},
	}
}

func repositoryDataSourceModelFromAPI(repository *akclient.Repository) repositoryDataSourceModel {
	return repositoryDataSourceModel{
		Key:         types.StringValue(repository.Key),
		ID:          types.StringValue(repository.ID),
		Name:        types.StringValue(repository.Name),
		Format:      types.StringValue(repository.Format),
		RepoType:    types.StringValue(repository.RepoType),
		Description: stringPtrValue(repository.Description),
		Public:      types.BoolValue(repository.IsPublic),
		QuotaBytes:  int64PtrValue(repository.QuotaBytes),
		CreatedAt:   types.StringValue(repository.CreatedAt),
		UpdatedAt:   types.StringValue(repository.UpdatedAt),
		SizeBytes:   types.Int64Value(repository.StorageUsedBytes),
	}
}

type userDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	CreatedAt types.String `tfsdk:"created_at"`
}

type usersDataSourceModel struct {
	Search types.String          `tfsdk:"search"`
	Users  []userDataSourceModel `tfsdk:"users"`
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Attributes: userDataSourceAttributes(true),
	}
}

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config userDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	user, err := d.client.GetUser(ctx, config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Read Artifact Keeper user data source", err)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, userDataSourceModelFromAPI(user))...)
}

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Attributes: map[string]dschema.Attribute{
			"search": dschema.StringAttribute{Optional: true},
			"users": dschema.ListNestedAttribute{
				Computed: true,
				NestedObject: dschema.NestedAttributeObject{
					Attributes: userDataSourceAttributes(false),
				},
			},
		},
	}
}

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config usersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := d.client.ListUsers(ctx, akclient.UserListOptions{
		PerPage: 100,
		Search:  valueOrEnv(config.Search, ""),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "List Artifact Keeper users", err)
		return
	}
	config.Users = make([]userDataSourceModel, 0, len(result.Items))
	for _, user := range result.Items {
		user := user
		config.Users = append(config.Users, userDataSourceModelFromAPI(&user))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

func userDataSourceAttributes(idRequired bool) map[string]dschema.Attribute {
	idAttr := dschema.StringAttribute{Computed: true}
	if idRequired {
		idAttr = dschema.StringAttribute{Required: true}
	}
	return map[string]dschema.Attribute{
		"id":         idAttr,
		"username":   dschema.StringAttribute{Computed: true},
		"email":      dschema.StringAttribute{Computed: true},
		"role":       dschema.StringAttribute{Computed: true},
		"created_at": dschema.StringAttribute{Computed: true},
	}
}

func userDataSourceModelFromAPI(user *akclient.User) userDataSourceModel {
	return userDataSourceModel{
		ID:        types.StringValue(user.ID),
		Username:  types.StringValue(user.Username),
		Email:     types.StringValue(user.Email),
		Role:      types.StringValue(roleFromAdmin(user.IsAdmin)),
		CreatedAt: types.StringValue(user.CreatedAt),
	}
}

type packageDataSourceModel struct {
	ID            types.String          `tfsdk:"id"`
	RepositoryKey types.String          `tfsdk:"repository_key"`
	Name          types.String          `tfsdk:"name"`
	Version       types.String          `tfsdk:"version"`
	Format        types.String          `tfsdk:"format"`
	SizeBytes     types.Int64           `tfsdk:"size_bytes"`
	DownloadCount types.Int64           `tfsdk:"download_count"`
	Description   types.String          `tfsdk:"description"`
	CreatedAt     types.String          `tfsdk:"created_at"`
	UpdatedAt     types.String          `tfsdk:"updated_at"`
	Versions      []packageVersionModel `tfsdk:"versions"`
}

type packageVersionModel struct {
	Version       types.String `tfsdk:"version"`
	SizeBytes     types.Int64  `tfsdk:"size_bytes"`
	DownloadCount types.Int64  `tfsdk:"download_count"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

type packagesDataSourceModel struct {
	RepositoryKey types.String             `tfsdk:"repository_key"`
	Format        types.String             `tfsdk:"format"`
	Search        types.String             `tfsdk:"search"`
	Packages      []packageDataSourceModel `tfsdk:"packages"`
}

func (d *packageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package"
}

func (d *packageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{Attributes: packageDataSourceAttributes(true)}
}

func (d *packageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config packageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pkg, err := d.client.GetPackage(ctx, config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Read Artifact Keeper package data source", err)
		return
	}
	versions, err := d.client.GetPackageVersions(ctx, config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Read Artifact Keeper package versions", err)
		return
	}
	state := packageDataSourceModelFromAPI(pkg)
	state.Versions = packageVersionModelsFromAPI(versions.Versions)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (d *packagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_packages"
}

func (d *packagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Attributes: map[string]dschema.Attribute{
			"repository_key": dschema.StringAttribute{Optional: true},
			"format":         dschema.StringAttribute{Optional: true},
			"search":         dschema.StringAttribute{Optional: true},
			"packages": dschema.ListNestedAttribute{
				Computed: true,
				NestedObject: dschema.NestedAttributeObject{
					Attributes: packageDataSourceAttributes(false),
				},
			},
		},
	}
}

func (d *packagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config packagesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, err := d.client.ListPackages(ctx, akclient.PackageListOptions{
		PerPage:       100,
		RepositoryKey: valueOrEnv(config.RepositoryKey, ""),
		Format:        valueOrEnv(config.Format, ""),
		Search:        valueOrEnv(config.Search, ""),
	})
	if err != nil {
		addClientError(&resp.Diagnostics, "List Artifact Keeper packages", err)
		return
	}
	config.Packages = make([]packageDataSourceModel, 0, len(result.Items))
	for _, pkg := range result.Items {
		pkg := pkg
		config.Packages = append(config.Packages, packageDataSourceModelFromAPI(&pkg))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

func packageDataSourceAttributes(idRequired bool) map[string]dschema.Attribute {
	idAttr := dschema.StringAttribute{Computed: true}
	if idRequired {
		idAttr = dschema.StringAttribute{Required: true}
	}
	return map[string]dschema.Attribute{
		"id":             idAttr,
		"repository_key": dschema.StringAttribute{Computed: true},
		"name":           dschema.StringAttribute{Computed: true},
		"version":        dschema.StringAttribute{Computed: true},
		"format":         dschema.StringAttribute{Computed: true},
		"size_bytes":     dschema.Int64Attribute{Computed: true},
		"download_count": dschema.Int64Attribute{Computed: true},
		"description":    dschema.StringAttribute{Computed: true},
		"created_at":     dschema.StringAttribute{Computed: true},
		"updated_at":     dschema.StringAttribute{Computed: true},
		"versions": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"version":        dschema.StringAttribute{Computed: true},
					"size_bytes":     dschema.Int64Attribute{Computed: true},
					"download_count": dschema.Int64Attribute{Computed: true},
					"created_at":     dschema.StringAttribute{Computed: true},
					"updated_at":     dschema.StringAttribute{Computed: true},
				},
			},
		},
	}
}

func packageDataSourceModelFromAPI(pkg *akclient.Package) packageDataSourceModel {
	return packageDataSourceModel{
		ID:            types.StringValue(pkg.ID),
		RepositoryKey: types.StringValue(pkg.RepositoryKey),
		Name:          types.StringValue(pkg.Name),
		Version:       types.StringValue(pkg.Version),
		Format:        types.StringValue(pkg.Format),
		SizeBytes:     types.Int64Value(pkg.SizeBytes),
		DownloadCount: types.Int64Value(pkg.DownloadCount),
		Description:   stringPtrValue(pkg.Description),
		CreatedAt:     types.StringValue(pkg.CreatedAt),
		UpdatedAt:     types.StringValue(pkg.UpdatedAt),
		Versions:      nil,
	}
}

func packageVersionModelsFromAPI(versions []akclient.PackageVersion) []packageVersionModel {
	out := make([]packageVersionModel, 0, len(versions))
	for _, version := range versions {
		out = append(out, packageVersionModel{
			Version:       types.StringValue(version.Version),
			SizeBytes:     types.Int64Value(version.SizeBytes),
			DownloadCount: types.Int64Value(version.DownloadCount),
			CreatedAt:     types.StringValue(version.CreatedAt),
			UpdatedAt:     types.StringValue(version.UpdatedAt),
		})
	}
	return out
}
