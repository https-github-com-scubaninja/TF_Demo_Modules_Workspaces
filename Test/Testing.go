package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

// This function tests the Azure DevOps agent pools Terraform module
func TestAgentPoolHasBeenDeployed(t *testing.T) {
	t.Parallel()

	// load the Terraform template from the directory and the testing variables
	terraformOptions := &terraform.Options{
		TerraformDir: "./Test",
	}

	// call terraform destroy at the end of the test
	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	// in order to validate that what has been deployed is valid, we need to get an Azure Go SDK authorizer (from the environment variable)
	// see: https://docs.microsoft.com/en-us/azure/go/azure-sdk-go-authorization#use-environment-based-authentication
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		t.Fatalf("Cannot get an Azure SDK Authorizer: %v", err)
	}

	// load variables from environment
	// AzureSubscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")

	// load variables from Terraform deployment outputs
	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name")
	vmssName := terraform.Output(t, terraformOptions, "vmss_name")
	vmssComputerNamePrefix := terraform.Output(t, terraformOptions, "vmss_computer_name_prefix")
	expectedCapacityAsString := terraform.Output(t, terraformOptions, "vmss_capacity")

	// Create a the virtual machine scale set client used to request the Azure subscription to validate the deployment
	vmssClient := compute.NewVirtualMachineScaleSetsClient(AzureSubscriptionID)
	vmssClient.Authorizer = authorizer

	ctx := context.Background()

	// First check of the test: the Virtual Machine Scale Set must have been deployed
	t.Logf("Checking that VM Scale '%s' exists in resource group '%s'.", vmssName, resourceGroupName)
	vmScaleSet, err := vmssClient.Get(ctx, resourceGroupName, vmssName)
	if err != nil {
		t.Fatalf("Cannot retrieve Virtual Machine Scale Set: %v", err)
	}

	// Second check of the test: the number of machines inside the scale set is the one that is expected
	vmScaleSetCapacity := *vmScaleSet.Sku.Capacity
	expectedCapacity, err := strconv.ParseInt(expectedCapacityAsString, 10, 64)

	if err != nil {
		t.Fatalf("Cannot convert expected Virtual Machine Scale Set capacity from String to Int64: %v", err)
	}

	t.Logf("Checking VM Scale Set Capacity. Actual = %d, Expected = %d.", vmScaleSetCapacity, expectedCapacity)
	if vmScaleSetCapacity != expectedCapacity {
		t.Error("The actual capacity of the Virtual Machine Scale Set does not match the expected capacity")
	}