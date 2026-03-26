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

func TestAccClientKey_Create(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	clientID := uuid.New()
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
resource "pdnd_client_key" "test" {
  client_id = %q
  key       = "dGVzdC1wdWJsaWMta2V5LWRhdGE="
  use       = "SIG"
  alg       = "RS256"
  name      = "Test Key"
}
`, clientID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_client_key.test", "kid"),
					resource.TestCheckResourceAttrSet("pdnd_client_key.test", "kty"),
					resource.TestCheckResourceAttrSet("pdnd_client_key.test", "id"),
				),
			},
		},
	})
}

func TestAccClientKey_Delete(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	clientID := uuid.New()
	fake.SeedClient(fakepdnd.StoredClient{
		ID: clientID, ConsumerID: fake.ConsumerID(),
		Name: "Test Client", CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(s *terraform.State) error {
			keys := fake.GetClientKeys(clientID)
			if len(keys) > 0 {
				return fmt.Errorf("expected no keys, but found %d", len(keys))
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_client_key" "test" {
  client_id = %q
  key       = "dGVzdC1wdWJsaWMta2V5LWRhdGE="
  use       = "SIG"
  alg       = "RS256"
  name      = "Test Key"
}
`, clientID),
				Check: resource.TestCheckResourceAttrSet("pdnd_client_key.test", "kid"),
			},
		},
	})
}

func TestAccClientKey_Import(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	clientID := uuid.New()
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
resource "pdnd_client_key" "test" {
  client_id = %q
  key       = "dGVzdC1wdWJsaWMta2V5LWRhdGE="
  use       = "SIG"
  alg       = "RS256"
  name      = "Test Key"
}
`, clientID),
			},
			{
				ResourceName: "pdnd_client_key.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					for _, rs := range s.RootModule().Resources {
						if rs.Type == "pdnd_client_key" {
							return fmt.Sprintf("%s/%s",
								rs.Primary.Attributes["client_id"],
								rs.Primary.Attributes["kid"]), nil
						}
					}
					return "", fmt.Errorf("resource not found")
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key", "use", "alg", "name"},
			},
		},
	})
}
