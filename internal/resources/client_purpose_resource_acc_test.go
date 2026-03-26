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

func TestAccClientPurpose_Create(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	clientID := uuid.New()
	purposeID := uuid.New()
	fake.SeedClient(fakepdnd.StoredClient{
		ID: clientID, ConsumerID: fake.ConsumerID(),
		Name: "Test Client", CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_client_purpose" "test" {
  client_id  = %q
  purpose_id = %q
}
`, clientID, purposeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_client_purpose.test", "id"),
				),
			},
		},
	})
}

func TestAccClientPurpose_Delete(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	clientID := uuid.New()
	purposeID := uuid.New()
	fake.SeedClient(fakepdnd.StoredClient{
		ID: clientID, ConsumerID: fake.ConsumerID(),
		Name: "Test Client", CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(s *terraform.State) error {
			purposes := fake.GetClientPurposes(clientID)
			if purposes[purposeID] {
				return fmt.Errorf("expected purpose %s to be unlinked, but it is still linked", purposeID)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_client_purpose" "test" {
  client_id  = %q
  purpose_id = %q
}
`, clientID, purposeID),
				Check: resource.TestCheckResourceAttrSet("pdnd_client_purpose.test", "id"),
			},
		},
	})
}
