package datasources_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccProducerDelegationsDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	now := time.Now().UTC()

	fake.SeedDelegation("producer", fakepdnd.StoredDelegation{
		ID:          uuid.New(),
		DelegatorID: fake.ProducerID(),
		DelegateID:  uuid.New(),
		EServiceID:  uuid.New(),
		State:       "ACTIVE",
		CreatedAt:   now,
		SubmittedAt: now,
		ActivatedAt: &now,
	})
	fake.SeedDelegation("producer", fakepdnd.StoredDelegation{
		ID:          uuid.New(),
		DelegatorID: fake.ProducerID(),
		DelegateID:  uuid.New(),
		EServiceID:  uuid.New(),
		State:       "WAITING_FOR_APPROVAL",
		CreatedAt:   now,
		SubmittedAt: now,
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + `
data "pdnd_producer_delegations" "test" {}
`,
				Check: resource.TestCheckResourceAttr("data.pdnd_producer_delegations.test", "delegations.#", "2"),
			},
		},
	})
}
