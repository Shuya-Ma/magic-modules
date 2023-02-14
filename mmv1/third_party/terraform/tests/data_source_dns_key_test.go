package google

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceDNSKeys_basic(t *testing.T) {
	t.Parallel()

	dnsZoneName := fmt.Sprintf("tf-dnskey-test-%s", randString(t, 10))

	var kskDigest1, kskDigest2, zskPubKey1, zskPubKey2, kskAlg1, kskAlg2 string

	vcrTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDNSManagedZoneDestroyProducerFramework(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion450(),
				Config:            testAccDataSourceDNSKeysConfig(dnsZoneName, "on"),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceDNSKeysDSRecordCheck("data.google_dns_keys.foo_dns_key"),
					resource.TestCheckResourceAttr("data.google_dns_keys.foo_dns_key", "key_signing_keys.#", "1"),
					resource.TestCheckResourceAttr("data.google_dns_keys.foo_dns_key", "zone_signing_keys.#", "1"),
					resource.TestCheckResourceAttr("data.google_dns_keys.foo_dns_key_id", "key_signing_keys.#", "1"),
					resource.TestCheckResourceAttr("data.google_dns_keys.foo_dns_key_id", "zone_signing_keys.#", "1"),
					testExtractResourceAttr("data.google_dns_keys.foo_dns_key", "key_signing_keys.0.digests.0.digest", &kskDigest1),
					testExtractResourceAttr("data.google_dns_keys.foo_dns_key_id", "zone_signing_keys.0.public_key", &zskPubKey1),
					testExtractResourceAttr("data.google_dns_keys.foo_dns_key_id", "key_signing_keys.0.algorithm", &kskAlg1),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(t),
				Config:                   testAccDataSourceDNSKeysConfig(dnsZoneName, "on"),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceDNSKeysDSRecordCheck("data.google_dns_keys.foo_dns_key"),
					resource.TestCheckResourceAttr("data.google_dns_keys.foo_dns_key", "key_signing_keys.#", "1"),
					resource.TestCheckResourceAttr("data.google_dns_keys.foo_dns_key", "zone_signing_keys.#", "1"),
					testExtractResourceAttr("data.google_dns_keys.foo_dns_key", "key_signing_keys.0.digests.0.digest", &kskDigest2),
					testExtractResourceAttr("data.google_dns_keys.foo_dns_key_id", "zone_signing_keys.0.public_key", &zskPubKey2),
					testExtractResourceAttr("data.google_dns_keys.foo_dns_key_id", "key_signing_keys.0.algorithm", &kskAlg2),
					testCheckAttributeValuesEqual(&kskDigest1, &kskDigest2),
					testCheckAttributeValuesEqual(&zskPubKey1, &zskPubKey2),
					testCheckAttributeValuesEqual(&kskAlg1, &kskAlg2),
				),
			},
		},
	})
}

func TestAccDataSourceDNSKeys_noDnsSec(t *testing.T) {
	t.Parallel()

	dnsZoneName := fmt.Sprintf("tf-dnskey-test-%s", randString(t, 10))

	vcrTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckDNSManagedZoneDestroyProducerFramework(t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: providerVersion450(),
				Config:            testAccDataSourceDNSKeysConfig(dnsZoneName, "off"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.google_dns_keys.foo_dns_key", "key_signing_keys.#", "0"),
					resource.TestCheckResourceAttr("data.google_dns_keys.foo_dns_key", "zone_signing_keys.#", "0"),
				),
			},
			{
				ProtoV5ProviderFactories: protoV5ProviderFactories(t),
				Config:                   testAccDataSourceDNSKeysConfig(dnsZoneName, "off"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.google_dns_keys.foo_dns_key", "key_signing_keys.#", "0"),
					resource.TestCheckResourceAttr("data.google_dns_keys.foo_dns_key", "zone_signing_keys.#", "0"),
				),
			},
		},
	})
}

func testAccDataSourceDNSKeysDSRecordCheck(datasourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		if ds.Primary.Attributes["key_signing_keys.0.ds_record"] == "" {
			return fmt.Errorf("DS record not found in data source")
		}

		return nil
	}
}

func testAccDataSourceDNSKeysConfig(dnsZoneName, dnssecStatus string) string {
	return fmt.Sprintf(`
resource "google_dns_managed_zone" "foo" {
  name     = "%s"
  dns_name = "dnssec.tf-test.club."

  dnssec_config {
    state         = "%s"
    non_existence = "nsec3"
  }
}

data "google_dns_keys" "foo_dns_key" {
  managed_zone = google_dns_managed_zone.foo.name
}

data "google_dns_keys" "foo_dns_key_id" {
  managed_zone = google_dns_managed_zone.foo.id
}
`, dnsZoneName, dnssecStatus)
}