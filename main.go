package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	ec2client "remove-default-vpc/pkg/aws/ec2"
)

// cleanupVPCResources cleans up all resources in a VPC before deleting it.
func cleanupVPCResources(ctx context.Context, client ec2client.API, vpcID string) error {
	if err := client.DeleteInternetGatewaysInVPC(ctx, vpcID); err != nil {
		return err
	}

	if err := client.DeleteSubnetsInVPC(ctx, vpcID); err != nil {
		return err
	}

	if err := client.DeleteRouteTablesInVPC(ctx, vpcID); err != nil {
		return err
	}

	if err := client.DeleteNetworkACLsInVPC(ctx, vpcID); err != nil {
		return err
	}

	if err := client.DeleteSecurityGroupsInVPC(ctx, vpcID); err != nil {
		return err
	}

	return nil
}

// DeleteAllDefaultVPCs deletes all default VPCs in all regions.
func DeleteAllDefaultVPCs(ctx context.Context, regions []string, cfg aws.Config) {
	var wg sync.WaitGroup
	for _, region := range regions {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()
			fmt.Printf("Processing region: %s\n", region)
			regionCfg := cfg.Copy()
			regionCfg.Region = region
			client := ec2client.NewClient(regionCfg)

			vpcs, err := client.GetDefaultVPCs(ctx)
			if err != nil {
				fmt.Printf("Error fetching default VPCs in region %s: %v\n", region, err)
				return
			}

			for _, vpcID := range vpcs {
				err := cleanupVPCResources(ctx, client, vpcID)
				if err != nil {
					fmt.Printf("Error cleaning up resources for VPC %s: %v", vpcID, err)
					os.Exit(1)
				}

				fmt.Printf("Deleting default VPC %s in region %s\n", vpcID, region)
				err = client.DeleteVPC(ctx, vpcID)
				if err != nil {
					fmt.Printf("Error deleting VPC %s in region %s: %v", vpcID, region, err)
					os.Exit(1)
				}
			}
		}(region)
	}

	wg.Wait()
	fmt.Println("All default VPCs deleted.")
}

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Printf("Unable to load AWS SDK config: %v", err)
		os.Exit(1)
	}

	client := ec2client.NewClient(cfg)

	regions, err := client.GetRegions(ctx)
	if err != nil {
		fmt.Printf("Unable to describe regions: %v", err)
		os.Exit(1)
	}

	DeleteAllDefaultVPCs(ctx, regions, cfg)
}
