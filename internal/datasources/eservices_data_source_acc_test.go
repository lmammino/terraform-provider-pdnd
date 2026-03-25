package datasources_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccEServicesDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()

	es1ID := uuid.New()
	es2ID := uuid.New()

	fake.SeedEService(fakepdnd.StoredEService{
		ID:          es1ID,
		ProducerID:  fake.ProducerID(),
		Name:        "REST E-Service",
		Description: "A REST e-service",
		Technology:  "REST",
		Mode:        "DELIVER",
	})

	fake.SeedEService(fakepdnd.StoredEService{
		ID:          es2ID,
		ProducerID:  fake.ProducerID(),
		Name:        "SOAP E-Service",
		Description: "A SOAP e-service",
		Technology:  "SOAP",
		Mode:        "DELIVER",
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + `
data "pdnd_eservices" "test" {
  technology = "REST"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_eservices.test", "eservices.#", "1"),
					resource.TestCheckResourceAttr("data.pdnd_eservices.test", "eservices.0.technology", "REST"),
				),
			},
		},
	})
}
