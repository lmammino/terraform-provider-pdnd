package datasources_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccTenantsDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	now := time.Now().UTC()

	fake.SeedTenant(fakepdnd.StoredTenant{
		ID:             uuid.New(),
		Name:           "Tenant One",
		Kind:           "PA",
		ExternalOrigin: "IPA",
		ExternalValue:  "t1",
		CreatedAt:      now,
	})
	fake.SeedTenant(fakepdnd.StoredTenant{
		ID:             uuid.New(),
		Name:           "Tenant Two",
		Kind:           "PRIVATE",
		ExternalOrigin: "IPA",
		ExternalValue:  "t2",
		CreatedAt:      now,
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + `
data "pdnd_tenants" "test" {}
`,
				Check: resource.TestCheckResourceAttr("data.pdnd_tenants.test", "tenants.#", "2"),
			},
		},
	})
}
