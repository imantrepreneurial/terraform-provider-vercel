package vercel_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vercel/terraform-provider-vercel/client"
)

func testAccDNSRecordDestroy(n, teamID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		_, err := testClient().GetDNSRecord(context.TODO(), rs.Primary.ID, teamID)

		var apiErr client.APIError
		if err == nil {
			return fmt.Errorf("Found project but expected it to have been deleted")
		}
		if err != nil && errors.As(err, &apiErr) {
			if apiErr.StatusCode == 404 {
				return nil
			}
			return fmt.Errorf("Unexpected error checking for deleted project: %s", apiErr)
		}

		return err
	}
}

func testAccDNSRecordExists(n, teamID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		_, err := testClient().GetDNSRecord(context.TODO(), rs.Primary.ID, teamID)
		return err
	}
}

func TestAcc_DNSRecord(t *testing.T) {
	t.Parallel()
	nameSuffix := acctest.RandString(16)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccDNSRecordDestroy("vercel_dns_record.a", ""),
			testAccDNSRecordDestroy("vercel_dns_record.aaaa", ""),
			testAccDNSRecordDestroy("vercel_dns_record.alias", ""),
			testAccDNSRecordDestroy("vercel_dns_record.caa", ""),
			testAccDNSRecordDestroy("vercel_dns_record.cname", ""),
			testAccDNSRecordDestroy("vercel_dns_record.mx", ""),
			testAccDNSRecordDestroy("vercel_dns_record.srv", ""),
			testAccDNSRecordDestroy("vercel_dns_record.txt", ""),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSRecordConfig(testDomain(), nameSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDNSRecordExists("vercel_dns_record.a", ""),
					testAccDNSRecordExists("vercel_dns_record.aaaa", ""),
					testAccDNSRecordExists("vercel_dns_record.alias", ""),
					testAccDNSRecordExists("vercel_dns_record.caa", ""),
					testAccDNSRecordExists("vercel_dns_record.cname", ""),
					testAccDNSRecordExists("vercel_dns_record.mx", ""),
					testAccDNSRecordExists("vercel_dns_record.srv", ""),
					testAccDNSRecordExists("vercel_dns_record.srv_no_target", ""),
					testAccDNSRecordExists("vercel_dns_record.txt", ""),
				),
			},
			{
				Config: testAccDNSRecordConfigUpdated(testDomain(), nameSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccDNSRecordExists("vercel_dns_record.a", ""),
					testAccDNSRecordExists("vercel_dns_record.aaaa", ""),
					testAccDNSRecordExists("vercel_dns_record.alias", ""),
					testAccDNSRecordExists("vercel_dns_record.caa", ""),
					testAccDNSRecordExists("vercel_dns_record.cname", ""),
					testAccDNSRecordExists("vercel_dns_record.mx", ""),
					testAccDNSRecordExists("vercel_dns_record.srv", ""),
					testAccDNSRecordExists("vercel_dns_record.txt", ""),
				),
			},
		},
	})
}

func testAccDNSRecordConfig(testDomain, nameSuffix string) string {
	return fmt.Sprintf(`
resource "vercel_dns_record" "a" {
  domain = "%[1]s"
  name  = "test-acc-%[2]s-a-record"
  type  = "A"
  ttl   = 120
  value = "127.0.0.1"
}
resource "vercel_dns_record" "aaaa" {
  domain = "%[1]s"
  name  = "test-acc-%s-aaaa-record"
  type  = "AAAA"
  ttl   = 120
  value = "::1"
}
resource "vercel_dns_record" "alias" {
  domain = "%[1]s"
  name  = "test-acc-%s-alias"
  type  = "ALIAS"
  ttl   = 120
  value = "example.com"
}
resource "vercel_dns_record" "caa" {
  domain = "%[1]s"
  name   = "test-acc-%s-caa"
  type   = "CAA"
  ttl    = 120
  value  = "0 issue \"letsencrypt.org\""
}
resource "vercel_dns_record" "cname" {
  domain = "%[1]s"
  name  = "test-acc-%s-cname"
  type  = "CNAME"
  ttl   = 120
  value = "example.com"
}
resource "vercel_dns_record" "mx" {
  domain = "%[1]s"
  name        = "test-acc-%s-mx"
  type        = "MX"
  ttl         = 120
  mx_priority = 123
  value       = "example.com"
}
resource "vercel_dns_record" "srv" {
  domain = "%[1]s"
  name = "test-acc-%[2]s-srv"
  type = "SRV"
  ttl  = 120
  srv = {
      port     = 5000
      weight   = 120
      priority = 27
      target   = "example.com"
  }
}
resource "vercel_dns_record" "srv_no_target" {
  domain = "%[1]s"
  name = "test-acc-%[2]s-srv-no-target"
  type = "SRV"
  ttl  = 120
  srv = {
      port     = 5000
      weight   = 120
      priority = 27
      target = ""
  }
}
resource "vercel_dns_record" "txt" {
  domain = "%[1]s"
  name = "test-acc-%[2]s-txt"
  type = "TXT"
  ttl  = 120
  value = "terraform testing"
}
`, testDomain, nameSuffix)
}

func testAccDNSRecordConfigUpdated(testDomain, nameSuffix string) string {
	return fmt.Sprintf(`
resource "vercel_dns_record" "a" {
  domain = "%[1]s"
  name  = "test-acc-%[2]s-a-record-updated"
  type  = "A"
  ttl   = 60
  value = "192.168.0.1"
}
resource "vercel_dns_record" "aaaa" {
  domain = "%[1]s"
  name  = "test-acc-%s-aaaa-record-updated"
  type  = "AAAA"
  ttl   = 60
  value = "::0"
}
resource "vercel_dns_record" "alias" {
  domain = "%[1]s"
  name  = "test-acc-%s-alias-updated"
  type  = "ALIAS"
  ttl   = 60
  value = "example2.com"
}
resource "vercel_dns_record" "caa" {
  domain = "%[1]s"
  name   = "test-acc-%s-caa-updated"
  type   = "CAA"
  ttl    = 60
  value  = "1 issue \"letsencrypt.org\""
}
resource "vercel_dns_record" "cname" {
  domain = "%[1]s"
  name  = "test-acc-%s-cname-updated"
  type  = "CNAME"
  ttl   = 60
  value = "example2.com"
}
resource "vercel_dns_record" "mx" {
  domain = "%[1]s"
  name        = "test-acc-%s-mx-updated"
  type        = "MX"
  ttl         = 60
  mx_priority = 333
  value       = "example2.com"
}
resource "vercel_dns_record" "srv" {
  domain = "%[1]s"
  name = "test-acc-%[2]s-srv-updated"
  type = "SRV"
  ttl  = 60
  srv = {
      port     = 6000
      weight   = 60
      priority = 127
      target   = "example2.com"
  }
}
resource "vercel_dns_record" "txt" {
  domain = "%[1]s"
  name = "test-acc-%[2]s-txt-updated"
  type = "TXT"
  ttl  = 60
  value = "terraform testing two"
}
`, testDomain, nameSuffix)
}
