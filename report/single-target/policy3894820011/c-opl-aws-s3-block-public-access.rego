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
package policies.c_opl_aws_s3_block_public_access

__rego__metadoc__ := {
  "description": "TEST S3 buckets should have all `block public access` options enabled. AWS's S3 Block Public Access feature has four settings: BlockPublicAcls, IgnorePublicAcls, BlockPublicPolicy, and RestrictPublicBuckets. All four settings should be enabled to help prevent the risk of a data breach",
  "custom": { "severity": "High" },
  "id": "c-opl-aws-s3-block-public-access",
  "title": "TEST S3 buckets should have all `block public access` options enabled"
}
# Or do we add the target type the policy applies to?
# package policies.opal.aws.s3.block_public_access_terraform

import data.lacework
import data.aws.s3.s3_library as lib

resource_type := "MULTIPLE"

policy[j] {
  b = buckets[bucket_id]
  bucket_is_blocked(b)
  j = lacework.allow_resource(b)
} {
  b = buckets[bucket_id]
  not bucket_is_blocked(b)
  j = lacework.deny_resource(b)
}

buckets = lacework.resources("aws_s3_bucket")
bucket_access_blocks = lacework.resources("aws_s3_bucket_public_access_block")
account_access_blocks = lacework.resources("aws_s3_account_public_access_block")

# Using the `bucket_access_blocks`, we construct a set of bucket IDs that have
# the public access blocked.
blocked_buckets[bucket_name] {
    block = bucket_access_blocks[_]
    bucket_name = block.bucket
    block.block_public_acls == true
    block.ignore_public_acls == true
    block.block_public_policy == true
    block.restrict_public_buckets == true
}

blocked_account {
    block := account_access_blocks[_]
    block.block_public_acls == true
    block.ignore_public_acls == true
    block.block_public_policy == true
    block.restrict_public_buckets == true
}

bucket_is_blocked(bucket) {
  blocked_account
} {
  lacework.input_type != "tf_runtime"
  blocked_buckets[bucket.id]
} {
  blocked_buckets[bucket.bucket]
}
