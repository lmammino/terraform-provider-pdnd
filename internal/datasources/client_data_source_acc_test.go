package datasources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccClientDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	clientID := uuid.New()

	fake.SeedClient(fakepdnd.StoredClient{
		ID:         clientID,
		ConsumerID: fake.ConsumerID(),
		Name:       "Test Client",
		CreatedAt:  time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_client" "test" {
  id = %q
}
`, clientID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_client.test", "id", clientID.String()),
					resource.TestCheckResourceAttr("data.pdnd_client.test", "name", "Test Client"),
					resource.TestCheckResourceAttr("data.pdnd_client.test", "consumer_id", fake.ConsumerID().String()),
					resource.TestCheckResourceAttrSet("data.pdnd_client.test", "created_at"),
				),
			},
		},
	})
}
