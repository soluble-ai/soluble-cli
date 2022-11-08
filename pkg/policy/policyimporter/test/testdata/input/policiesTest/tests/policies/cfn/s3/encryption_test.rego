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
package policies.cfn_s3_encryption

import data.tests.policies.cfn.s3.inputs

test_valid_encryption {
    resources = inputs.valid_encryption_infra_yaml.mock_resources
    allow with input as resources["Bucket"]
}

test_invalid_encryption_missing {
    resources = inputs.invalid_encryption_missing_infra_yaml.mock_resources
    not allow with input as resources["Bucket"]
}

test_invalid_encryption_with_valid {
    resources = inputs.invalid_encryption_with_valid_infra_yaml.mock_resources
    allow with input as resources["Bucket"]
    not allow with input as resources["InvalidBucket"]
}
