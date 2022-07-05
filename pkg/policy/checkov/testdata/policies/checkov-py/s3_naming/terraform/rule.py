from checkov.common.models.enums import CheckResult, CheckCategories
from checkov.terraform.checks.resource.base_resource_check import BaseResourceCheck
from typing import Dict, List, Any

class S3Naming(BaseResourceCheck):
    def __init__(self):
        supported_resources = ['aws_s3_bucket']
        super().__init__(name="", id="", categories=[], supported_resources=supported_resources)

    def scan_resource_conf(self, conf):
        if 'bucket' in conf.keys():
            name = conf['bucket'][0]
            if name and name.startswith("soluble-"):
                return CheckResult.PASSED
        if 'bucket_prefix' in conf.keys():
            prefix = conf['bucket_prefix'][0]
            if prefix and prefix.startswith("soluble-"):
                return CheckResult.PASSED
        return CheckResult.FAILED

check = S3Naming()