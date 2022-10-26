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
package policies.opal.aws.s3.block_public_access

input_type := "cfn"
resource_type := "AWS::S3::Bucket"

default allow = false

allow {
  access_block := input.PublicAccessBlockConfiguration
  access_block.BlockPublicAcls == true
  access_block.BlockPublicPolicy == true
  access_block.IgnorePublicAcls == true
  access_block.RestrictPublicBuckets == true
}