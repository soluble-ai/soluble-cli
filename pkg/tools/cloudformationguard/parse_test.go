package cloudformationguard

import "testing"

func TestParseResults(t *testing.T) {
	failures := parseFailures(`[AmazonMQBroker] failed because [AutoMinorVersionUpgrade] is [false] and Version upgrades should be enabled to receive security updates
[AmazonMQBroker] failed because [EncryptionOptions.UseAwsOwnedKey] is [true] and CMKs should be used instead of AWS-provided KMS keys
[AmazonMQBroker] failed because [EngineVersion] is [5.15.9] and Broker engine version should be at least 5.15.10`)
	if len(failures) != 3 {
		t.Error(failures)
	}
	f := failures[2]
	if f.Resource != "AmazonMQBroker" || f.Attribute != "EngineVersion" || f.AttributeValue != "5.15.9" ||
		f.Message != "Broker engine version should be at least 5.15.10" {
		t.Error(f)
	}
}
