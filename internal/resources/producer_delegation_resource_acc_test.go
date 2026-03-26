package resources_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccProducerDelegation_Create(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	delegateID := uuid.New()
	esID := uuid.New()
	fake.SeedEService(fakepdnd.StoredEService{
		ID: esID, ProducerID: fake.ProducerID(), Name: "Test",
		Description: "test", Technology: "REST", Mode: "DELIVER",
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_producer_delegation" "test" {
  eservice_id = %q
  delegate_id = %q
}
`, esID, delegateID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_producer_delegation.test", "id"),
					resource.TestCheckResourceAttr("pdnd_producer_delegation.test", "state", "WAITING_FOR_APPROVAL"),
					resource.TestCheckResourceAttrSet("pdnd_producer_delegation.test", "delegator_id"),
					resource.TestCheckResourceAttrSet("pdnd_producer_delegation.test", "created_at"),
					resource.TestCheckResourceAttrSet("pdnd_producer_delegation.test", "submitted_at"),
				),
			},
		},
	})
}

func TestAccProducerDelegation_Import(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	delegateID := uuid.New()
	esID := uuid.New()
	fake.SeedEService(fakepdnd.StoredEService{
		ID: esID, ProducerID: fake.ProducerID(), Name: "Test",
		Description: "test", Technology: "REST", Mode: "DELIVER",
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_producer_delegation" "test" {
  eservice_id = %q
  delegate_id = %q
}
`, esID, delegateID),
				Check: resource.TestCheckResourceAttrSet("pdnd_producer_delegation.test", "id"),
			},
			{
				ResourceName:      "pdnd_producer_delegation.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
