package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/client"
	"github.com/lmammino/terraform-provider-pdnd/internal/client/api"
	generated "github.com/lmammino/terraform-provider-pdnd/internal/client/generated"
	"github.com/lmammino/terraform-provider-pdnd/internal/datasources"
	"github.com/lmammino/terraform-provider-pdnd/internal/providerdata"
	"github.com/lmammino/terraform-provider-pdnd/internal/resources"
)

var _ provider.Provider = &pdndProvider{}

type pdndProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &pdndProvider{version: version}
	}
}

func (p *pdndProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pdnd"
	resp.Version = p.version
}

func (p *pdndProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for PDND Interoperability API v3",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Description: "Base URL of the PDND API (e.g., https://api.interop.pagopa.it/v3)",
				Required:    true,
			},
			"access_token": schema.StringAttribute{
				Description: "Access token for PDND API authentication",
				Required:    true,
				Sensitive:   true,
			},
			"dpop_private_key": schema.StringAttribute{
				Description: "PEM-encoded private key for DPoP proof generation",
				Required:    true,
				Sensitive:   true,
			},
			"dpop_key_id": schema.StringAttribute{
				Description: "Key ID for the DPoP private key",
				Required:    true,
			},
			"request_timeout_s": schema.Int64Attribute{
				Description: "Request timeout in seconds (default: 30)",
				Optional:    true,
			},
		},
	}
}

func (p *pdndProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config pdndProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate base_url
	baseURL := config.BaseURL.ValueString()
	if baseURL == "" {
		resp.Diagnostics.AddError("Invalid Configuration", "base_url must not be empty")
		return
	}
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", fmt.Sprintf("base_url is not a valid URL: %s", err))
		return
	}

	// Validate access_token
	accessToken := config.AccessToken.ValueString()
	if accessToken == "" {
		resp.Diagnostics.AddError("Invalid Configuration", "access_token must not be empty")
		return
	}

	// Validate dpop_key_id
	dpopKeyID := config.DPoPKeyID.ValueString()
	if dpopKeyID == "" {
		resp.Diagnostics.AddError("Invalid Configuration", "dpop_key_id must not be empty")
		return
	}

	// Parse DPoP private key
	dpopPrivateKey := config.DPoPPrivateKey.ValueString()
	proofGen, err := client.NewDPoPProofGenerator([]byte(dpopPrivateKey), dpopKeyID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid DPoP Private Key",
			fmt.Sprintf("Failed to parse dpop_private_key: %s", err),
		)
		return
	}

	// Set timeout
	var timeoutS int64 = 30
	if !config.RequestTimeoutS.IsNull() && !config.RequestTimeoutS.IsUnknown() {
		timeoutS = config.RequestTimeoutS.ValueInt64()
		if timeoutS <= 0 {
			resp.Diagnostics.AddError("Invalid Configuration", "request_timeout_s must be greater than 0")
			return
		}
	}

	// Build transport chain: DPoP -> Retry -> http.Client
	dpopTransport := &client.DPoPTransport{
		Base:        http.DefaultTransport,
		AccessToken: accessToken,
		ProofGen:    proofGen,
	}

	retryTransport := &client.RetryTransport{
		Base: dpopTransport,
	}

	httpClient := &http.Client{
		Transport: retryTransport,
		Timeout:   time.Duration(timeoutS) * time.Second,
	}

	// Create generated client
	genClient, err := generated.NewClientWithResponses(baseURL, generated.WithHTTPClient(httpClient))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create API Client",
			fmt.Sprintf("Could not create PDND API client: %s", err),
		)
		return
	}

	// Create API wrappers
	agreementsAPI := api.NewAgreementsClient(genClient)
	eservicesAPI := api.NewEServicesClient(genClient)

	// Store provider data
	pd := &providerdata.ProviderData{
		AgreementsAPI: agreementsAPI,
		EServicesAPI:  eservicesAPI,
	}

	resp.DataSourceData = pd
	resp.ResourceData = pd
}

func (p *pdndProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewAgreementResource,
		resources.NewEServiceResource,
		resources.NewEServiceDescriptorResource,
	}
}

func (p *pdndProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewAgreementDataSource,
		datasources.NewAgreementsDataSource,
		datasources.NewAgreementPurposesDataSource,
		datasources.NewEServiceDataSource,
		datasources.NewEServicesDataSource,
		datasources.NewEServiceDescriptorDataSource,
		datasources.NewEServiceDescriptorsDataSource,
	}
}
