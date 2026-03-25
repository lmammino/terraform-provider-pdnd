package datasources_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccCertifiedAttributesDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()

	createdAt := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID:          uuid.New(),
		Name:        "Comuni italiani",
		Description: "Certified attribute for Comuni italiani",
		Code:        "L6_COMUNI",
		Origin:      "IPA",
		CreatedAt:   createdAt,
	})

	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID:          uuid.New(),
		Name:        "Comuni montani",
		Description: "Certified attribute for Comuni montani",
		Code:        "L6_COMUNI_MONTANI",
		Origin:      "IPA",
		CreatedAt:   createdAt,
	})

	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID:          uuid.New(),
		Name:        "Regioni",
		Description: "Certified attribute for Regioni",
		Code:        "L2_REGIONI",
		Origin:      "IPA",
		CreatedAt:   createdAt,
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + `
data "pdnd_certified_attributes" "test" {
  name = "Comuni"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_certified_attributes.test", "attributes.#", "2"),
				),
			},
		},
	})
}
