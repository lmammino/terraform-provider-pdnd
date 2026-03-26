package resources_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccDescriptorDocument_Create(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, descriptorID := seedEServiceWithDraftDescriptor(fake)

	ts := fake.Start()
	defer ts.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-doc.yaml")
	if err := os.WriteFile(filePath, []byte("openapi: 3.0.0\ninfo:\n  title: Test"), 0644); err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_document" "test" {
  eservice_id   = %q
  descriptor_id = %q
  pretty_name   = "Test Document"
  file_path     = %q
  content_type  = "application/yaml"
}
`, eserviceID, descriptorID, filePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_document.test", "document_id"),
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_document.test", "name"),
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_document.test", "pretty_name"),
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_document.test", "content_type"),
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_document.test", "created_at"),
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_document.test", "file_hash"),
				),
			},
		},
	})
}

func TestAccDescriptorDocument_Delete(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, descriptorID := seedEServiceWithDraftDescriptor(fake)

	ts := fake.Start()
	defer ts.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-doc.yaml")
	if err := os.WriteFile(filePath, []byte("openapi: 3.0.0\ninfo:\n  title: Test"), 0644); err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(s *terraform.State) error {
			docs := fake.GetDocuments(eserviceID, descriptorID)
			if len(docs) > 0 {
				return fmt.Errorf("expected no documents, but found %d", len(docs))
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_document" "test" {
  eservice_id   = %q
  descriptor_id = %q
  pretty_name   = "Test Document"
  file_path     = %q
  content_type  = "application/yaml"
}
`, eserviceID, descriptorID, filePath),
				Check: resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_document.test", "document_id"),
			},
		},
	})
}

func TestAccDescriptorDocument_Import(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, descriptorID := seedEServiceWithDraftDescriptor(fake)

	ts := fake.Start()
	defer ts.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-doc.yaml")
	if err := os.WriteFile(filePath, []byte("openapi: 3.0.0\ninfo:\n  title: Test"), 0644); err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_document" "test" {
  eservice_id   = %q
  descriptor_id = %q
  pretty_name   = "Test Document"
  file_path     = %q
  content_type  = "application/yaml"
}
`, eserviceID, descriptorID, filePath),
			},
			{
				ResourceName: "pdnd_eservice_descriptor_document.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					for _, rs := range s.RootModule().Resources {
						if rs.Type == "pdnd_eservice_descriptor_document" {
							return fmt.Sprintf("%s/%s/%s",
								rs.Primary.Attributes["eservice_id"],
								rs.Primary.Attributes["descriptor_id"],
								rs.Primary.Attributes["document_id"]), nil
						}
					}
					return "", fmt.Errorf("resource not found")
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"file_path", "file_hash"},
			},
		},
	})
}
