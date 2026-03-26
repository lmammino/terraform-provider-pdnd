package datasources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccClientKeysDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	clientID := uuid.New()

	fake.SeedClient(fakepdnd.StoredClient{
		ID:         clientID,
		ConsumerID: fake.ConsumerID(),
		Name:       "Test Client",
		CreatedAt:  time.Now().UTC(),
	})

	fake.SeedClientKey(clientID, fakepdnd.StoredClientKey{
		Kid: "test-kid-001",
		Kty: "RSA",
		Alg: "RS256",
		Use: "SIG",
		Name: "Key One",
		Key:  "dGVzdC1rZXktMQ==",
	})
	fake.SeedClientKey(clientID, fakepdnd.StoredClientKey{
		Kid: "test-kid-002",
		Kty: "RSA",
		Alg: "RS256",
		Use: "SIG",
		Name: "Key Two",
		Key:  "dGVzdC1rZXktMg==",
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_client_keys" "test" {
  client_id = %q
}
`, clientID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_client_keys.test", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.pdnd_client_keys.test", "keys.0.kid", "test-kid-001"),
				),
			},
		},
	})
}
