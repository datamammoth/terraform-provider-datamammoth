package provider

import (
	"context"
	"os"

	"github.com/datamammoth/terraform-provider-datamammoth/internal/client"
	"github.com/datamammoth/terraform-provider-datamammoth/internal/datasources"
	"github.com/datamammoth/terraform-provider-datamammoth/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dataMammothProvider struct{ version string }
type dataMammothProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &dataMammothProvider{version: version} }
}

func (p *dataMammothProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "datamammoth"
	resp.Version = p.version
}

func (p *dataMammothProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage DataMammoth cloud infrastructure.",
		Attributes: map[string]schema.Attribute{
			"api_key":  schema.StringAttribute{Optional: true, Sensitive: true, Description: "API key. Falls back to DM_API_KEY env var."},
			"base_url": schema.StringAttribute{Optional: true, Description: "API base URL. Defaults to https://app.datamammoth.com/api/v2"},
		},
	}
}

func (p *dataMammothProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config dataMammothProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiKey := config.APIKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("DM_API_KEY")
	}
	if apiKey == "" {
		resp.Diagnostics.AddError("Missing API key", "Set api_key or DM_API_KEY env var")
		return
	}
	baseURL := config.BaseURL.ValueString()
	c := client.New(apiKey, baseURL)
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *dataMammothProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewServerResource,
		resources.NewSnapshotResource,
		resources.NewWebhookResource,
	}
}

func (p *dataMammothProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewProductsDataSource,
		datasources.NewZonesDataSource,
	}
}
