package datasources_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccEServiceDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()

	eserviceID := uuid.New()

	fake.SeedEService(fakepdnd.StoredEService{
		ID:          eserviceID,
		ProducerID:  fake.ProducerID(),
		Name:        "Test E-Service",
		Description: "A test e-service for data source testing",
		Technology:  "REST",
		Mode:        "DELIVER",
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_eservice" "test" {
  id = %q
}
`, eserviceID.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_eservice.test", "id", eserviceID.String()),
					resource.TestCheckResourceAttr("data.pdnd_eservice.test", "name", "Test E-Service"),
					resource.TestCheckResourceAttr("data.pdnd_eservice.test", "description", "A test e-service for data source testing"),
					resource.TestCheckResourceAttr("data.pdnd_eservice.test", "technology", "REST"),
					resource.TestCheckResourceAttr("data.pdnd_eservice.test", "mode", "DELIVER"),
					resource.TestCheckResourceAttr("data.pdnd_eservice.test", "producer_id", fake.ProducerID().String()),
				),
			},
		},
	})
}
