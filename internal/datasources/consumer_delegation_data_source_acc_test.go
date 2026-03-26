package datasources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccConsumerDelegationDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	delegationID := uuid.New()
	delegateID := uuid.New()
	esID := uuid.New()
	now := time.Now().UTC()

	fake.SeedDelegation("consumer", fakepdnd.StoredDelegation{
		ID:          delegationID,
		DelegatorID: fake.ConsumerID(),
		DelegateID:  delegateID,
		EServiceID:  esID,
		State:       "ACTIVE",
		CreatedAt:   now,
		SubmittedAt: now,
		ActivatedAt: &now,
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_consumer_delegation" "test" {
  id = %q
}
`, delegationID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_consumer_delegation.test", "id", delegationID.String()),
					resource.TestCheckResourceAttr("data.pdnd_consumer_delegation.test", "state", "ACTIVE"),
					resource.TestCheckResourceAttr("data.pdnd_consumer_delegation.test", "delegate_id", delegateID.String()),
					resource.TestCheckResourceAttr("data.pdnd_consumer_delegation.test", "delegator_id", fake.ConsumerID().String()),
					resource.TestCheckResourceAttr("data.pdnd_consumer_delegation.test", "eservice_id", esID.String()),
					resource.TestCheckResourceAttrSet("data.pdnd_consumer_delegation.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.pdnd_consumer_delegation.test", "submitted_at"),
					resource.TestCheckResourceAttrSet("data.pdnd_consumer_delegation.test", "activated_at"),
				),
			},
		},
	})
}
