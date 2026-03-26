package datasources_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccClientsDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	now := time.Now().UTC()

	fake.SeedClient(fakepdnd.StoredClient{
		ID:         uuid.New(),
		ConsumerID: fake.ConsumerID(),
		Name:       "Client One",
		CreatedAt:  now,
	})
	fake.SeedClient(fakepdnd.StoredClient{
		ID:         uuid.New(),
		ConsumerID: fake.ConsumerID(),
		Name:       "Client Two",
		CreatedAt:  now,
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + `
data "pdnd_clients" "test" {}
`,
				Check: resource.TestCheckResourceAttr("data.pdnd_clients.test", "clients.#", "2"),
			},
		},
	})
}
