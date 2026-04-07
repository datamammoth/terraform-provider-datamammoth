package datasources

import (
	"context"

	"github.com/datamammoth/terraform-provider-datamammoth/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &zonesDataSource{}

type zonesDataSource struct {
	client *client.Client
}

type zonesDataSourceModel struct {
	Zones []zoneModel `tfsdk:"zones"`
}

type zoneModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Region  types.String `tfsdk:"region"`
	Country types.String `tfsdk:"country"`
	Status  types.String `tfsdk:"status"`
}

func NewZonesDataSource() datasource.DataSource {
	return &zonesDataSource{}
}

func (d *zonesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zones"
}

func (d *zonesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists available DataMammoth availability zones.",
		Attributes: map[string]schema.Attribute{
			"zones": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of available zones.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":      schema.StringAttribute{Computed: true, Description: "Zone ID."},
						"name":    schema.StringAttribute{Computed: true, Description: "Zone display name."},
						"region":  schema.StringAttribute{Computed: true, Description: "Region (us-east, eu-west, ap-southeast)."},
						"country": schema.StringAttribute{Computed: true, Description: "Country code (US, DE, SG)."},
						"status":  schema.StringAttribute{Computed: true, Description: "Zone status (available, maintenance)."},
					},
				},
			},
		},
	}
}

func (d *zonesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type", "Expected *client.Client")
		return
	}
	d.client = c
}

func (d *zonesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	result, err := d.client.Get("/zones")
	if err != nil {
		resp.Diagnostics.AddError("Error listing zones", err.Error())
		return
	}

	dataList, _ := result["data"].([]interface{})
	var zones []zoneModel
	for _, item := range dataList {
		z, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		zone := zoneModel{}
		if id, ok := z["id"].(string); ok {
			zone.ID = types.StringValue(id)
		}
		if name, ok := z["name"].(string); ok {
			zone.Name = types.StringValue(name)
		}
		if region, ok := z["region"].(string); ok {
			zone.Region = types.StringValue(region)
		}
		if country, ok := z["country"].(string); ok {
			zone.Country = types.StringValue(country)
		}
		if status, ok := z["status"].(string); ok {
			zone.Status = types.StringValue(status)
		}
		zones = append(zones, zone)
	}

	state := zonesDataSourceModel{Zones: zones}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
