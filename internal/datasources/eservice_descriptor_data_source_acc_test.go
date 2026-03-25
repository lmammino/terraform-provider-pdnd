package datasources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccEServiceDescriptorDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()

	eserviceID := uuid.New()
	descriptorID := uuid.New()
	now := time.Now().UTC()

	fake.SeedEService(fakepdnd.StoredEService{
		ID:          eserviceID,
		ProducerID:  fake.ProducerID(),
		Name:        "Test E-Service",
		Description: "A test e-service",
		Technology:  "REST",
		Mode:        "DELIVER",
	})

	fake.SeedDescriptor(fakepdnd.StoredDescriptor{
		ID:                      descriptorID,
		EServiceID:              eserviceID,
		Version:                 "1",
		State:                   "PUBLISHED",
		AgreementApprovalPolicy: "AUTOMATIC",
		Audience:                []string{"api.test.example.com"},
		DailyCallsPerConsumer:   1000,
		DailyCallsTotal:         10000,
		VoucherLifespan:         3600,
		CreatedAt:               now,
		PublishedAt:             &now,
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_eservice_descriptor" "test" {
  eservice_id = %q
  id          = %q
}
`, eserviceID.String(), descriptorID.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_eservice_descriptor.test", "id", descriptorID.String()),
					resource.TestCheckResourceAttr("data.pdnd_eservice_descriptor.test", "eservice_id", eserviceID.String()),
					resource.TestCheckResourceAttr("data.pdnd_eservice_descriptor.test", "version", "1"),
					resource.TestCheckResourceAttr("data.pdnd_eservice_descriptor.test", "state", "PUBLISHED"),
					resource.TestCheckResourceAttr("data.pdnd_eservice_descriptor.test", "agreement_approval_policy", "AUTOMATIC"),
					resource.TestCheckResourceAttr("data.pdnd_eservice_descriptor.test", "daily_calls_per_consumer", "1000"),
					resource.TestCheckResourceAttr("data.pdnd_eservice_descriptor.test", "daily_calls_total", "10000"),
					resource.TestCheckResourceAttr("data.pdnd_eservice_descriptor.test", "voucher_lifespan", "3600"),
				),
			},
		},
	})
}
