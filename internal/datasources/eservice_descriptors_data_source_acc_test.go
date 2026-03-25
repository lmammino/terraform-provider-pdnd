package datasources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccEServiceDescriptorsDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()

	eserviceID := uuid.New()
	desc1ID := uuid.New()
	desc2ID := uuid.New()
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
		ID:                      desc1ID,
		EServiceID:              eserviceID,
		Version:                 "1",
		State:                   "DRAFT",
		AgreementApprovalPolicy: "AUTOMATIC",
		Audience:                []string{"api.test.example.com"},
		DailyCallsPerConsumer:   1000,
		DailyCallsTotal:         10000,
		VoucherLifespan:         3600,
		CreatedAt:               now,
	})

	fake.SeedDescriptor(fakepdnd.StoredDescriptor{
		ID:                      desc2ID,
		EServiceID:              eserviceID,
		Version:                 "2",
		State:                   "PUBLISHED",
		AgreementApprovalPolicy: "AUTOMATIC",
		Audience:                []string{"api.test.example.com"},
		DailyCallsPerConsumer:   2000,
		DailyCallsTotal:         20000,
		VoucherLifespan:         7200,
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
data "pdnd_eservice_descriptors" "test" {
  eservice_id = %q
  state       = "PUBLISHED"
}
`, eserviceID.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_eservice_descriptors.test", "descriptors.#", "1"),
					resource.TestCheckResourceAttr("data.pdnd_eservice_descriptors.test", "descriptors.0.state", "PUBLISHED"),
				),
			},
		},
	})
}
