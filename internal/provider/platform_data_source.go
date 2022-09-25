package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/devopzilla/guku-client-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &PlatformDataSource{}

func NewPlatformDataSource() datasource.DataSource {
	return &PlatformDataSource{}
}

// PlatformDataSource defines the data source implementation.
type PlatformDataSource struct {
	client *guku.Client
}

// PlatformDataSourceModel describes the data source data model.
type PlatformDataSourceModel struct {
	PlatformID      types.String `tfsdk:"platform_id"`
	PlatformVersion types.String `tfsdk:"platform_version"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	CatalogedDate   types.String `tfsdk:"cataloged_date"`
	MaxAPIVersion   types.String `tfsdk:"max_api_version"`
	MinAPIVersion   types.String `tfsdk:"min_api_version"`
	Services        types.Map    `tfsdk:"services"`
	Configs         types.Map    `tfsdk:"configs"`
}

var PlatformServiceModelAttrTypes = map[string]attr.Type{
	"id":                  types.StringType,
	"name":                types.StringType,
	"namespace":           types.StringType,
	"service_id":          types.StringType,
	"service_version":     types.StringType,
	"delete_dependencies": types.ListType{ElemType: types.StringType},
	"dependencies":        types.ListType{ElemType: types.StringType},
}

var PlatformConfigModelAttrTypes = map[string]attr.Type{
	"id":     types.StringType,
	"name":   types.StringType,
	"config": types.MapType{ElemType: types.StringType},
}

func (d *PlatformDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_platform"
}

func (d *PlatformDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example data source",

		Attributes: map[string]tfsdk.Attribute{
			"platform_id": {
				MarkdownDescription: "Example identifier",
				Type:                types.StringType,
				Required:            true,
			},
			"platform_version": {
				MarkdownDescription: "Example identifier",
				Type:                types.StringType,
				Required:            true,
			},
			"name": {
				MarkdownDescription: "Example identifier",
				Type:                types.StringType,
				Computed:            true,
			},
			"description": {
				MarkdownDescription: "Example identifier",
				Type:                types.StringType,
				Computed:            true,
			},
			"cataloged_date": {
				MarkdownDescription: "Example identifier",
				Type:                types.StringType,
				Computed:            true,
			},
			"max_api_version": {
				MarkdownDescription: "Example identifier",
				Type:                types.StringType,
				Computed:            true,
			},
			"min_api_version": {
				MarkdownDescription: "Example identifier",
				Type:                types.StringType,
				Computed:            true,
			},
			"services": {
				MarkdownDescription: "Example identifier",
				Type:                types.MapType{ElemType: types.ObjectType{AttrTypes: PlatformServiceModelAttrTypes}},
				Computed:            true,
			},
			"configs": {
				MarkdownDescription: "Example identifier",
				Type:                types.MapType{ElemType: types.ObjectType{AttrTypes: PlatformConfigModelAttrTypes}},
				Computed:            true,
			},
		},
	}, nil
}

func (d *PlatformDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*guku.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *PlatformDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PlatformDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	platform, err := d.client.GetPlatform(data.PlatformID.Value, data.PlatformVersion.Value)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read platform, got error: %s", err))
		return
	}
	if platform == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprint("Unable to read platform, not found"))
		return
	}

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Name = types.String{Value: platform.GetName()}
	data.Description = StringValueOrNull(platform.GetDescription())
	data.CatalogedDate = StringValueOrNull(platform.GetCatalogedDate())
	data.MaxAPIVersion = types.String{Value: platform.GetMaxAPIVersion()}
	data.MinAPIVersion = types.String{Value: platform.GetMinAPIVersion()}

	services := map[string]attr.Value{}
	for _, service := range platform.GetServices() {
		dependencies := []attr.Value{}
		for _, dep := range service.GetDependencies() {
			dependencies = append(
				dependencies,
				types.String{Value: dep},
			)
		}
		delete_dependencies := []attr.Value{}
		for _, dep := range service.GetDelete_dependencies() {
			delete_dependencies = append(
				delete_dependencies,
				types.String{Value: dep},
			)
		}

		services[service.GetName()] = types.Object{
			AttrTypes: PlatformServiceModelAttrTypes,
			Attrs: map[string]attr.Value{
				"id":                  types.String{Value: service.GetPlatformServiceID()},
				"name":                types.String{Value: service.GetName()},
				"namespace":           types.String{Value: service.GetNamespace()},
				"service_id":          types.String{Value: service.GetServiceID()},
				"service_version":     types.String{Value: service.GetServiceVersion()},
				"dependencies":        types.List{ElemType: types.StringType, Elems: dependencies},
				"delete_dependencies": types.List{ElemType: types.StringType, Elems: delete_dependencies},
			},
		}
	}

	data.Services = types.Map{
		ElemType: types.ObjectType{
			AttrTypes: PlatformServiceModelAttrTypes,
		},
		Elems: services,
	}

	configs := map[string]attr.Value{}
	for _, config := range platform.GetConfigs() {
		data := config.GetConfig()

		var parsedData map[string]string

		json.Unmarshal([]byte(data), &parsedData)

		valuesData := map[string]attr.Value{}

		for key, val := range parsedData {
			valuesData[key] = types.String{Value: val}
		}

		configs[config.GetName()] = types.Object{
			AttrTypes: PlatformConfigModelAttrTypes,
			Attrs: map[string]attr.Value{
				"id":     types.String{Value: config.GetPlatformConfigID()},
				"name":   types.String{Value: config.GetName()},
				"config": types.Map{ElemType: types.StringType, Elems: valuesData},
			},
		}
	}

	data.Configs = types.Map{
		ElemType: types.ObjectType{
			AttrTypes: PlatformConfigModelAttrTypes,
		},
		Elems: configs,
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a platform data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
