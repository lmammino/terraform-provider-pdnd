package resources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccConsumerDelegation_Create(t *testing.T) {
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
resource "pdnd_consumer_delegation" "test" {
  eservice_id = %q
  delegate_id = %q
}
`, esID, delegateID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_consumer_delegation.test", "id"),
					resource.TestCheckResourceAttr("pdnd_consumer_delegation.test", "state", "WAITING_FOR_APPROVAL"),
					resource.TestCheckResourceAttrSet("pdnd_consumer_delegation.test", "delegator_id"),
					resource.TestCheckResourceAttrSet("pdnd_consumer_delegation.test", "created_at"),
					resource.TestCheckResourceAttrSet("pdnd_consumer_delegation.test", "submitted_at"),
				),
			},
		},
	})
}

func TestAccConsumerDelegation_ExternalAccept(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	delegateID := uuid.New()
	esID := uuid.New()
	fake.SeedEService(fakepdnd.StoredEService{
		ID: esID, ProducerID: fake.ProducerID(), Name: "Test",
		Description: "test", Technology: "REST", Mode: "DELIVER",
	})

	var delegationID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_consumer_delegation" "test" {
  eservice_id = %q
  delegate_id = %q
}
`, esID, delegateID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_consumer_delegation.test", "state", "WAITING_FOR_APPROVAL"),
					extractResourceID("pdnd_consumer_delegation.test", &delegationID),
				),
			},
			{
				PreConfig: func() {
					id := uuid.MustParse(delegationID)
					d := fake.GetDelegation("consumer", id)
					if d != nil {
						now := time.Now().UTC()
						d.State = "ACTIVE"
						d.ActivatedAt = &now
						fake.SeedDelegation("consumer", *d)
					}
				},
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_consumer_delegation.test", "state", "ACTIVE"),
					resource.TestCheckResourceAttrSet("pdnd_consumer_delegation.test", "activated_at"),
				),
			},
		},
	})
}

func TestAccConsumerDelegation_Import(t *testing.T) {
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
resource "pdnd_consumer_delegation" "test" {
  eservice_id = %q
  delegate_id = %q
}
`, esID, delegateID),
				Check: resource.TestCheckResourceAttrSet("pdnd_consumer_delegation.test", "id"),
			},
			{
				ResourceName:      "pdnd_consumer_delegation.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccConsumerDelegation_Destroy(t *testing.T) {
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
resource "pdnd_consumer_delegation" "test" {
  eservice_id = %q
  delegate_id = %q
}
`, esID, delegateID),
				Check: resource.TestCheckResourceAttrSet("pdnd_consumer_delegation.test", "id"),
			},
		},
	})
}

// extractResourceID is a test check function that extracts the ID from a resource in state.
func extractResourceID(resourceName string, target *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}
		*target = rs.Primary.ID
		return nil
	}
}
