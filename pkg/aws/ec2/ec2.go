// Package ec2 provides a wrapper around the AWS EC2 client with an interface
// for easy mocking and testing.
package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

//go:generate go tool mockgen -source=ec2.go -destination=mocks/mock_ec2.go -package=mocks

// API defines the EC2 operations used by this application.
// This interface allows for easy mocking in tests.
type API interface {
	GetRegions(ctx context.Context) ([]string, error)
	GetDefaultVPCs(ctx context.Context) ([]string, error)
	DeleteSubnetsInVPC(ctx context.Context, vpcID string) error
	DeleteRouteTablesInVPC(ctx context.Context, vpcID string) error
	DeleteInternetGatewaysInVPC(ctx context.Context, vpcID string) error
	DeleteSecurityGroupsInVPC(ctx context.Context, vpcID string) error
	DeleteNetworkACLsInVPC(ctx context.Context, vpcID string) error
	DeleteVPC(ctx context.Context, vpcID string) error
}

// sdkClient defines the subset of EC2 SDK operations used by this package.
// This interface allows for mocking the underlying SDK client in tests.
type sdkClient interface {
	DescribeRegions(ctx context.Context, params *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error)
	DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
	DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	DeleteSubnet(ctx context.Context, params *ec2.DeleteSubnetInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSubnetOutput, error)
	DescribeRouteTables(ctx context.Context, params *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error)
	DeleteRouteTable(ctx context.Context, params *ec2.DeleteRouteTableInput, optFns ...func(*ec2.Options)) (*ec2.DeleteRouteTableOutput, error)
	DescribeInternetGateways(ctx context.Context, params *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error)
	DetachInternetGateway(ctx context.Context, params *ec2.DetachInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DetachInternetGatewayOutput, error)
	DeleteInternetGateway(ctx context.Context, params *ec2.DeleteInternetGatewayInput, optFns ...func(*ec2.Options)) (*ec2.DeleteInternetGatewayOutput, error)
	DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	DeleteSecurityGroup(ctx context.Context, params *ec2.DeleteSecurityGroupInput, optFns ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error)
	DescribeNetworkAcls(ctx context.Context, params *ec2.DescribeNetworkAclsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkAclsOutput, error)
	DeleteNetworkAcl(ctx context.Context, params *ec2.DeleteNetworkAclInput, optFns ...func(*ec2.Options)) (*ec2.DeleteNetworkAclOutput, error)
	DeleteVpc(ctx context.Context, params *ec2.DeleteVpcInput, optFns ...func(*ec2.Options)) (*ec2.DeleteVpcOutput, error)
}

// Client implements API and wraps the real EC2 client.
type Client struct {
	sdk sdkClient
}

// NewClient creates a new EC2 Client from an AWS config.
func NewClient(cfg aws.Config) *Client {
	return &Client{sdk: ec2.NewFromConfig(cfg)}
}

// newClientWithSDK creates a new EC2 Client with a custom SDK client (for testing).
func newClientWithSDK(sdk sdkClient) *Client {
	return &Client{sdk: sdk}
}

// GetRegions returns a list of all available AWS region names.
func (c *Client) GetRegions(ctx context.Context) ([]string, error) {
	resp, err := c.sdk.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}

	regions := make([]string, 0, len(resp.Regions))
	for _, region := range resp.Regions {
		regions = append(regions, aws.ToString(region.RegionName))
	}
	return regions, nil
}

// GetDefaultVPCs returns a list of default VPC IDs.
func (c *Client) GetDefaultVPCs(ctx context.Context) ([]string, error) {
	resp, err := c.sdk.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, err
	}

	vpcs := []string{}
	for _, vpc := range resp.Vpcs {
		if aws.ToBool(vpc.IsDefault) {
			vpcs = append(vpcs, aws.ToString(vpc.VpcId))
		}
	}
	return vpcs, nil
}

// DeleteSubnetsInVPC deletes all subnets in the specified VPC.
func (c *Client) DeleteSubnetsInVPC(ctx context.Context, vpcID string) error {
	resp, err := c.sdk.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, subnet := range resp.Subnets {
		_, err := c.sdk.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
			SubnetId: subnet.SubnetId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete subnet %s: %w", aws.ToString(subnet.SubnetId), err)
		}
		fmt.Printf("Deleted subnet: %s\n", aws.ToString(subnet.SubnetId))
	}
	return nil
}

// DeleteRouteTablesInVPC deletes all non-main route tables in the specified VPC.
func (c *Client) DeleteRouteTablesInVPC(ctx context.Context, vpcID string) error {
	resp, err := c.sdk.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to describe route tables: %w", err)
	}

	for _, rt := range resp.RouteTables {
		if isMainRouteTable(rt) {
			continue
		}

		_, err := c.sdk.DeleteRouteTable(ctx, &ec2.DeleteRouteTableInput{
			RouteTableId: rt.RouteTableId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete route table %s: %w", aws.ToString(rt.RouteTableId), err)
		}
		fmt.Printf("Deleted route table: %s\n", aws.ToString(rt.RouteTableId))
	}
	return nil
}

func isMainRouteTable(rt types.RouteTable) bool {
	for _, association := range rt.Associations {
		if aws.ToBool(association.Main) {
			return true
		}
	}
	return false
}

// DeleteInternetGatewaysInVPC detaches and deletes all internet gateways in the specified VPC.
func (c *Client) DeleteInternetGatewaysInVPC(ctx context.Context, vpcID string) error {
	resp, err := c.sdk.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("attachment.vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, igw := range resp.InternetGateways {
		_, err := c.sdk.DetachInternetGateway(ctx, &ec2.DetachInternetGatewayInput{
			InternetGatewayId: igw.InternetGatewayId,
			VpcId:             aws.String(vpcID),
		})
		if err != nil {
			return fmt.Errorf("failed to detach internet gateway %s: %w", aws.ToString(igw.InternetGatewayId), err)
		}

		_, err = c.sdk.DeleteInternetGateway(ctx, &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: igw.InternetGatewayId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete internet gateway %s: %w", aws.ToString(igw.InternetGatewayId), err)
		}
		fmt.Printf("Deleted internet gateway: %s\n", aws.ToString(igw.InternetGatewayId))
	}
	return nil
}

// DeleteSecurityGroupsInVPC deletes all non-default security groups in the specified VPC.
func (c *Client) DeleteSecurityGroupsInVPC(ctx context.Context, vpcID string) error {
	resp, err := c.sdk.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, sg := range resp.SecurityGroups {
		if aws.ToString(sg.GroupName) == "default" {
			continue
		}
		_, err := c.sdk.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
			GroupId: sg.GroupId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete security group %s: %w", aws.ToString(sg.GroupId), err)
		}
		fmt.Printf("Deleted security group: %s\n", aws.ToString(sg.GroupId))
	}
	return nil
}

// DeleteNetworkACLsInVPC deletes all non-default network ACLs in the specified VPC.
func (c *Client) DeleteNetworkACLsInVPC(ctx context.Context, vpcID string) error {
	resp, err := c.sdk.DescribeNetworkAcls(ctx, &ec2.DescribeNetworkAclsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, acl := range resp.NetworkAcls {
		if aws.ToBool(acl.IsDefault) {
			continue
		}
		_, err := c.sdk.DeleteNetworkAcl(ctx, &ec2.DeleteNetworkAclInput{
			NetworkAclId: acl.NetworkAclId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete network ACL %s: %w", aws.ToString(acl.NetworkAclId), err)
		}
		fmt.Printf("Deleted network ACL: %s\n", aws.ToString(acl.NetworkAclId))
	}
	return nil
}

// DeleteVPC deletes the specified VPC.
func (c *Client) DeleteVPC(ctx context.Context, vpcID string) error {
	_, err := c.sdk.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(vpcID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete VPC %s: %w", vpcID, err)
	}

	fmt.Printf("Deleted VPC: %s\n", vpcID)
	return nil
}
