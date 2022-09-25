package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/devopzilla/guku-client-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ClusterResource{}
var _ resource.ResourceWithImportState = &ClusterResource{}

func NewClusterResource() resource.Resource {
	return &ClusterResource{}
}

// ClusterResource defines the resource implementation.
type ClusterResource struct {
	client *guku.Client
}

// ClusterResourceModel describes the resource data model.
type ClusterResourceModel struct {
	ClusterID types.String `tfsdk:"id"`

	Name       types.String `tfsdk:"name"`
	Token      types.String `tfsdk:"token"`
	ApiVersion types.String `tfsdk:"api_version"`
	Ca         types.String `tfsdk:"ca"`
	Server     types.String `tfsdk:"server"`
	Context    types.String `tfsdk:"context"`
}

func (r *ClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *ClusterResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cluster resource",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Computed:            true,
				MarkdownDescription: "Example identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
			"name": {
				MarkdownDescription: "Example configurable attribute",
				Required:            true,
				Type:                types.StringType,
			},
			"token": {
				MarkdownDescription: "Example configurable attribute",
				Required:            true,
				Type:                types.StringType,
				Sensitive:           true,
			},
			"api_version": {
				MarkdownDescription: "Example configurable attribute",
				Required:            true,
				Type:                types.StringType,
			},
			"ca": {
				MarkdownDescription: "Example configurable attribute",
				Optional:            true,
				Type:                types.StringType,
			},
			"server": {
				MarkdownDescription: "Example configurable attribute",
				Optional:            true,
				Type:                types.StringType,
			},
			"context": {
				MarkdownDescription: "Example configurable attribute",
				Optional:            true,
				Type:                types.StringType,
			},
		},
	}, nil
}

func (r *ClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*guku.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *guku.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ClusterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// data.Context.Value = MinifyJSONString(data.Context.Value)

	cluster, err := r.client.CreateCluster(
		data.Name.Value,
		data.Server.Value,
		data.Ca.Value,
		data.Token.Value,
		data.ApiVersion.Value,
		data.Context.Value,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create cluster, got error: %s", err))
		return
	}

	time.Sleep(time.Second * 40)

	data.ClusterID = types.String{Value: cluster.GetClusterID()}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a cluster")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ClusterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.GetCluster(
		data.ClusterID.Value,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read cluster, got error: %s", err))
		return
	}

	data.Name = types.String{Value: cluster.GetName()}
	data.ApiVersion = types.String{Value: cluster.GetApiVersion()}

	data.Ca = StringValueOrNull(cluster.GetCa())
	data.Server = StringValueOrNull(cluster.GetServer())
	data.Context = StringValueOrNull(cluster.GetContext())

	data.Context.Value = MinifyJSONString(data.Context.Value)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ClusterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// if !data.Context.IsNull() {
	// 	data.Context.Value = MinifyJSONString(data.Context.Value)
	// }

	_, err := r.client.UpdateCluster(
		data.ClusterID.Value,
		ValueStringOrNull(data.Name),
		ValueStringOrNull(data.Server),
		ValueStringOrNull(data.Ca),
		ValueStringOrNull(data.Token),
		ValueStringOrNull(data.ApiVersion),
		ValueStringOrNull(data.Context),
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update cluster, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "updated a cluster")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ClusterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteCluster(
		data.ClusterID.Value,
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete cluster, got error: %s", err))
		return
	}

	time.Sleep(time.Second * 30)

	tflog.Trace(ctx, "deleted a cluster")
}

func (r *ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
