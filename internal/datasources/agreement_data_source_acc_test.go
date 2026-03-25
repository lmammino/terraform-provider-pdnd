package datasources_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/provider"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

var testPrivateKeyPEM string

func init() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("failed to generate test RSA key: %v", err))
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	testPrivateKeyPEM = string(pemBytes)
}

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"pdnd": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func testAccProviderConfig(serverURL string) string {
	return fmt.Sprintf(`
provider "pdnd" {
  base_url         = %q
  access_token     = "test-access-token"
  dpop_private_key = <<-EOT
%sEOT
  dpop_key_id      = "test-key-id"
}
`, serverURL, testPrivateKeyPEM)
}

func TestAccAgreementDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()

	agreementID := uuid.New()
	eserviceID := uuid.New()
	descriptorID := uuid.New()

	fake.SeedAgreement(fakepdnd.StoredAgreement{
		ID:           agreementID,
		EServiceID:   eserviceID,
		DescriptorID: descriptorID,
		ProducerID:   fake.ProducerID(),
		ConsumerID:   fake.ConsumerID(),
		State:        "ACTIVE",
		CreatedAt:    time.Now(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_agreement" "test" {
  id = %q
}
`, agreementID.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_agreement.test", "id", agreementID.String()),
					resource.TestCheckResourceAttr("data.pdnd_agreement.test", "state", "ACTIVE"),
					resource.TestCheckResourceAttr("data.pdnd_agreement.test", "eservice_id", eserviceID.String()),
					resource.TestCheckResourceAttr("data.pdnd_agreement.test", "descriptor_id", descriptorID.String()),
					resource.TestCheckResourceAttr("data.pdnd_agreement.test", "producer_id", fake.ProducerID().String()),
					resource.TestCheckResourceAttr("data.pdnd_agreement.test", "consumer_id", fake.ConsumerID().String()),
				),
			},
		},
	})
}
