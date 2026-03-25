package providerdata

import "github.com/lmammino/terraform-provider-pdnd/internal/client/api"

// ProviderData holds the configured API clients for use by resources and data sources.
type ProviderData struct {
	AgreementsAPI api.AgreementsAPI
	EServicesAPI  api.EServicesAPI
}
