package datasources

import (
	"context"

	"github.com/datamammoth/terraform-provider-datamammoth/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &productsDataSource{}

type productsDataSource struct {
	client *client.Client
}

type productsDataSourceModel struct {
	Category types.String   `tfsdk:"category"`
	Products []productModel `tfsdk:"products"`
}

type productModel struct {
	ID          types.String  `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Category    types.String  `tfsdk:"category"`
	CPU         types.Int64   `tfsdk:"cpu"`
	Memory      types.Int64   `tfsdk:"memory"`
	Disk        types.Int64   `tfsdk:"disk"`
	Bandwidth   types.Int64   `tfsdk:"bandwidth"`
	PriceMonthly types.Float64 `tfsdk:"price_monthly"`
	Currency    types.String  `tfsdk:"currency"`
}

func NewProductsDataSource() datasource.DataSource {
	return &productsDataSource{}
}

func (d *productsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_products"
}

func (d *productsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists available DataMammoth products/plans.",
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				Optional:    true,
				Description: "Filter products by category (vps, dedicated, storage).",
			},
			"products": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of available products.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true, Description: "Product ID."},
						"name":          schema.StringAttribute{Computed: true, Description: "Product name."},
						"category":      schema.StringAttribute{Computed: true, Description: "Product category."},
						"cpu":           schema.Int64Attribute{Computed: true, Description: "CPU cores."},
						"memory":        schema.Int64Attribute{Computed: true, Description: "Memory in MB."},
						"disk":          schema.Int64Attribute{Computed: true, Description: "Disk in GB."},
						"bandwidth":     schema.Int64Attribute{Computed: true, Description: "Bandwidth in TB."},
						"price_monthly": schema.Float64Attribute{Computed: true, Description: "Monthly price."},
						"currency":      schema.StringAttribute{Computed: true, Description: "Price currency (USD, EUR, GBP)."},
					},
				},
			},
		},
	}
}

func (d *productsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *productsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config productsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/products"
	if !config.Category.IsNull() && !config.Category.IsUnknown() {
		path += "?category=" + config.Category.ValueString()
	}

	result, err := d.client.Get(path)
	if err != nil {
		resp.Diagnostics.AddError("Error listing products", err.Error())
		return
	}

	dataList, _ := result["data"].([]interface{})
	var products []productModel
	for _, item := range dataList {
		p, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		product := productModel{}
		if id, ok := p["id"].(string); ok {
			product.ID = types.StringValue(id)
		}
		if name, ok := p["name"].(string); ok {
			product.Name = types.StringValue(name)
		}
		if cat, ok := p["category"].(string); ok {
			product.Category = types.StringValue(cat)
		}
		if cpu, ok := p["cpu"].(float64); ok {
			product.CPU = types.Int64Value(int64(cpu))
		}
		if mem, ok := p["memory"].(float64); ok {
			product.Memory = types.Int64Value(int64(mem))
		}
		if disk, ok := p["disk"].(float64); ok {
			product.Disk = types.Int64Value(int64(disk))
		}
		if bw, ok := p["bandwidth"].(float64); ok {
			product.Bandwidth = types.Int64Value(int64(bw))
		}
		if price, ok := p["price_monthly"].(float64); ok {
			product.PriceMonthly = types.Float64Value(price)
		}
		if cur, ok := p["currency"].(string); ok {
			product.Currency = types.StringValue(cur)
		}
		products = append(products, product)
	}

	config.Products = products
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
