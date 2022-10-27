# Copyright 2022 Lacework, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

package policies.opal.aws.s3.bucket_access_logging
# Or do we add the target type the policy applies to?
# package policies.opal.aws.s3.bucket_access_logging_terraform

import data.aws.s3.s3_library as lib
import data.lacework

__rego__metadoc__ := {
  "custom": {
    "severity": "Medium"
  },
  "description": "S3 bucket access logging should be enabled. Enabling server access logging provides detailed records for the requests that are made to a S3 bucket. This information is useful for security and compliance auditing purposes.",
  "id": "FG_R00274",
  "title": "S3 bucket access logging should be enabled"
}
input_type := "tf"
resource_type := "MULTIPLE"

buckets := lacework.resources("aws_s3_bucket")

bucket_has_logging(bucket) {
  _ = bucket.logging[_]
}

bucket_has_logging(bucket) {
  _ = lib.bucket_logging_by_bucket[lib.bucket_name_or_id(bucket)]
}

policy[p] {
  bucket := buckets[_]
  bucket_has_logging(bucket)
  p := lacework.allow_resource(bucket)
} {
  bucket := buckets[_]
  not bucket_has_logging(bucket)
  p := lacework.deny_resource(bucket)
}
