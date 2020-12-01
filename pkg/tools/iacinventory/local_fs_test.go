package iacinventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeDupe(t *testing.T) {
	assert := assert.New(t)

	raw := map[string][]string{
		"terraform_dirs": {
			"terraform/aws",
			"terraform/azure",
			"terraform/gcp",
			"terraform/gcp/wordpress-mysql-gce-pv",
		},
		"cloudformation_dirs": {
			"cloudformation/AutoScaling",
			"cloudformation/Cloud9",
			"cloudformation/CloudFormation",
			"cloudformation/CloudFormation/MacrosExamples/Count",
			"cloudformation/CloudFormation/MacrosExamples/ExecutionRoleBuilder",
			"cloudformation/CloudFormation/MacrosExamples/Explode",
			"cloudformation/CloudFormation/MacrosExamples/Public-and-Private-Subnet-per-AZ",
			"cloudformation/CloudFormation/MacrosExamples/PyPlate",
			"cloudformation/CloudFormation/MacrosExamples/StringFunctions",
			"cloudformation/Config",
			"cloudformation/DMS",
			"cloudformation/DirectoryService",
			"cloudformation/DynamoDB",
			"cloudformation/EC2",
			"cloudformation/ECS/EC2LaunchType/clusters",
			"cloudformation/ECS/EC2LaunchType/services",
			"cloudformation/ECS",
			"cloudformation/ECS/FargateLaunchType/clusters",
			"cloudformation/ECS/FargateLaunchType/services",
			"cloudformation/EMR",
			"cloudformation/ElasticLoadBalancing",
			"cloudformation/IAM",
			"cloudformation/IoT",
			"cloudformation/NeptuneDB",
			"cloudformation/RDS",
			"cloudformation/S3",
			"cloudformation/SNS",
			"cloudformation/SQS",
			"cloudformation/ServiceCatalog",
			"cloudformation/VPC",
		},
		"k8s_dirs": {
			"kubernetes/deployments",
			"kubernetes/deployments/hello-deployment",
			"kubernetes/deployments/vsphere-deployment",
			"kubernetes/job",
			"kubernetes/pod",
			"kubernetes/statefulset",
		},
	}

	desired := map[string][]string{
		"terraform_dirs": {
			"terraform/aws",
			"terraform/azure",
			"terraform/gcp",
		},
		"cloudformation_dirs": {
			"cloudformation/AutoScaling",
			"cloudformation/Cloud9",
			"cloudformation/CloudFormation",
			"cloudformation/Config",
			"cloudformation/DMS",
			"cloudformation/DirectoryService",
			"cloudformation/DynamoDB",
			"cloudformation/EC2",
			"cloudformation/ECS",
			"cloudformation/EMR",
			"cloudformation/ElasticLoadBalancing",
			"cloudformation/IAM",
			"cloudformation/IoT",
			"cloudformation/NeptuneDB",
			"cloudformation/RDS",
			"cloudformation/S3",
			"cloudformation/SNS",
			"cloudformation/SQS",
			"cloudformation/ServiceCatalog",
			"cloudformation/VPC",
		},
		"k8s_dirs": {
			"kubernetes/deployments",
			"kubernetes/job",
			"kubernetes/pod",
			"kubernetes/statefulset",
		},
	}

	assert.EqualValues(dedupe(raw), desired)
}
