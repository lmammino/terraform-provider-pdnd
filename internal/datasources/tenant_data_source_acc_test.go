package datasources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccTenantDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	tenantID := uuid.New()

	fake.SeedTenant(fakepdnd.StoredTenant{
		ID:             tenantID,
		Name:           "Test Tenant",
		Kind:           "PA",
		ExternalOrigin: "IPA",
		ExternalValue:  "test123",
		CreatedAt:      time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_tenant" "test" {
  id = %q
}
`, tenantID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_tenant.test", "id", tenantID.String()),
					resource.TestCheckResourceAttr("data.pdnd_tenant.test", "name", "Test Tenant"),
					resource.TestCheckResourceAttr("data.pdnd_tenant.test", "kind", "PA"),
					resource.TestCheckResourceAttrSet("data.pdnd_tenant.test", "created_at"),
				),
			},
		},
	})
}
