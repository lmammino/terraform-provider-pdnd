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
				Description: "Access token for PDND API authentication (manual mode). Mutually exclusive with client_id/purpose_id.",
				Optional:    true,
				Sensitive:   true,
			},
			"client_id": schema.StringAttribute{
				Description: "PDND client UUID for automatic token generation. Must be used together with purpose_id.",
				Optional:    true,
			},
			"purpose_id": schema.StringAttribute{
				Description: "PDND purpose UUID for automatic token generation. Must be used together with client_id.",
				Optional:    true,
			},
			"token_endpoint": schema.StringAttribute{
				Description: "PDND authorization server token endpoint (default: https://auth.interop.pagopa.it/token.oauth2)",
				Optional:    true,
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
	timeout := time.Duration(timeoutS) * time.Second

	// Determine auth mode
	hasAccessToken := !config.AccessToken.IsNull() && !config.AccessToken.IsUnknown() && config.AccessToken.ValueString() != ""
	hasClientID := !config.ClientID.IsNull() && !config.ClientID.IsUnknown() && config.ClientID.ValueString() != ""
	hasPurposeID := !config.PurposeID.IsNull() && !config.PurposeID.IsUnknown() && config.PurposeID.ValueString() != ""

	var tokenProvider client.TokenProvider

	if hasAccessToken && (hasClientID || hasPurposeID) {
		resp.Diagnostics.AddError("Invalid auth config",
			"Cannot set both 'access_token' and 'client_id'/'purpose_id'. Use either manual token or auto-token mode.")
		return
	}

	if hasAccessToken {
		// Manual mode
		tokenProvider = client.NewStaticTokenProvider(config.AccessToken.ValueString())
	} else if hasClientID && hasPurposeID {
		// Auto-token mode
		tokenEndpoint := "https://auth.interop.pagopa.it/token.oauth2"
		if !config.TokenEndpoint.IsNull() && !config.TokenEndpoint.IsUnknown() {
			tokenEndpoint = config.TokenEndpoint.ValueString()
		}
		tokenProvider = client.NewAutoTokenProvider(
			config.ClientID.ValueString(),
			config.PurposeID.ValueString(),
			tokenEndpoint,
			proofGen,
			&http.Client{Timeout: timeout},
		)
	} else if hasClientID || hasPurposeID {
		resp.Diagnostics.AddError("Incomplete auth config",
			"Both 'client_id' and 'purpose_id' are required for automatic token generation.")
		return
	} else {
		resp.Diagnostics.AddError("Missing auth config",
			"Either 'access_token' or both 'client_id' and 'purpose_id' must be provided.")
		return
	}

	// Build transport chain: DPoP -> Retry -> http.Client
	dpopTransport := &client.DPoPTransport{
		Base:          http.DefaultTransport,
		TokenProvider: tokenProvider,
		ProofGen:      proofGen,
	}

	retryTransport := &client.RetryTransport{
		Base: dpopTransport,
	}

	httpClient := &http.Client{
		Transport: retryTransport,
		Timeout:   timeout,
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
	attributesAPI := api.NewAttributesClient(genClient)
	descriptorAttributesAPI := api.NewDescriptorAttributesClient(genClient)
	descriptorDocumentsAPI := api.NewDescriptorDocumentsClient(genClient)
	purposesAPI := api.NewPurposesClient(genClient)
	delegationsAPI := api.NewDelegationsClient(genClient)
	clientsAPI := api.NewClientsClient(genClient)

	// Store provider data
	pd := &providerdata.ProviderData{
		AgreementsAPI:           agreementsAPI,
		EServicesAPI:            eservicesAPI,
		AttributesAPI:           attributesAPI,
		DescriptorAttributesAPI: descriptorAttributesAPI,
		DescriptorDocumentsAPI:  descriptorDocumentsAPI,
		PurposesAPI:             purposesAPI,
		DelegationsAPI:          delegationsAPI,
		ClientsAPI:              clientsAPI,
	}

	resp.DataSourceData = pd
	resp.ResourceData = pd
}

func (p *pdndProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewAgreementResource,
		resources.NewEServiceResource,
		resources.NewEServiceDescriptorResource,
		resources.NewDescriptorCertifiedAttributesResource,
		resources.NewDescriptorDeclaredAttributesResource,
		resources.NewDescriptorVerifiedAttributesResource,
		resources.NewDescriptorDocumentResource,
		resources.NewDescriptorInterfaceResource,
		resources.NewPurposeResource,
		resources.NewConsumerDelegationResource,
		resources.NewProducerDelegationResource,
		resources.NewClientKeyResource,
		resources.NewClientPurposeResource,
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
		datasources.NewCertifiedAttributeDataSource,
		datasources.NewCertifiedAttributesDataSource,
		datasources.NewDeclaredAttributeDataSource,
		datasources.NewDeclaredAttributesDataSource,
		datasources.NewVerifiedAttributeDataSource,
		datasources.NewVerifiedAttributesDataSource,
		datasources.NewPurposeDataSource,
		datasources.NewPurposesDataSource,
		datasources.NewConsumerDelegationDataSource,
		datasources.NewConsumerDelegationsDataSource,
		datasources.NewProducerDelegationDataSource,
		datasources.NewProducerDelegationsDataSource,
		datasources.NewClientDataSource,
		datasources.NewClientsDataSource,
		datasources.NewClientKeysDataSource,
	}
}
