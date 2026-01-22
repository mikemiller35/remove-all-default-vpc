package ec2

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/mock/gomock"

	"remove-default-vpc/pkg/aws/ec2/mocks"
)

// TestClientImplementsAPI ensures that Client implements the API interface.
func TestClientImplementsAPI(t *testing.T) {
	var _ API = (*Client)(nil)
}

// TestMockAPIImplementsAPI ensures that MockAPI implements the API interface.
func TestMockAPIImplementsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var _ API = mocks.NewMockAPI(ctrl)
}

func TestNewClient(t *testing.T) {
	cfg := aws.Config{Region: "us-east-1"}
	client := NewClient(cfg)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.sdk == nil {
		t.Fatal("NewClient did not set internal sdk client")
	}
}

// Test that MockAPI works correctly for common operations
func TestMockAPI_GetRegions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAPI(ctrl)
	mockClient.EXPECT().GetRegions(gomock.Any()).Return([]string{"us-east-1", "us-west-2"}, nil)

	regions, err := mockClient.GetRegions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(regions) != 2 {
		t.Fatalf("expected 2 regions, got %d", len(regions))
	}
}

func TestMockAPI_GetDefaultVPCs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAPI(ctrl)
	mockClient.EXPECT().GetDefaultVPCs(gomock.Any()).Return([]string{"vpc-123"}, nil)

	vpcs, err := mockClient.GetDefaultVPCs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vpcs) != 1 {
		t.Fatalf("expected 1 VPC, got %d", len(vpcs))
	}
}

func TestMockAPI_DeleteVPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAPI(ctrl)
	mockClient.EXPECT().DeleteVPC(gomock.Any(), "vpc-123").Return(nil)

	err := mockClient.DeleteVPC(context.Background(), "vpc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockAPI_DeleteSubnetsInVPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAPI(ctrl)
	mockClient.EXPECT().DeleteSubnetsInVPC(gomock.Any(), "vpc-123").Return(nil)

	err := mockClient.DeleteSubnetsInVPC(context.Background(), "vpc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockAPI_DeleteInternetGatewaysInVPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAPI(ctrl)
	mockClient.EXPECT().DeleteInternetGatewaysInVPC(gomock.Any(), "vpc-123").Return(nil)

	err := mockClient.DeleteInternetGatewaysInVPC(context.Background(), "vpc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockAPI_DeleteRouteTablesInVPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAPI(ctrl)
	mockClient.EXPECT().DeleteRouteTablesInVPC(gomock.Any(), "vpc-123").Return(nil)

	err := mockClient.DeleteRouteTablesInVPC(context.Background(), "vpc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockAPI_DeleteSecurityGroupsInVPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAPI(ctrl)
	mockClient.EXPECT().DeleteSecurityGroupsInVPC(gomock.Any(), "vpc-123").Return(nil)

	err := mockClient.DeleteSecurityGroupsInVPC(context.Background(), "vpc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockAPI_DeleteNetworkACLsInVPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAPI(ctrl)
	mockClient.EXPECT().DeleteNetworkACLsInVPC(gomock.Any(), "vpc-123").Return(nil)

	err := mockClient.DeleteNetworkACLsInVPC(context.Background(), "vpc-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Client implementation tests using MocksdkClient
// =============================================================================

func TestClient_GetRegions(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mocks.MocksdkClient)
		wantRegions []string
		wantErr     bool
	}{
		{
			name: "success with multiple regions",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeRegions(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRegionsOutput{
					Regions: []types.Region{
						{RegionName: aws.String("us-east-1")},
						{RegionName: aws.String("us-west-2")},
						{RegionName: aws.String("eu-west-1")},
					},
				}, nil)
			},
			wantRegions: []string{"us-east-1", "us-west-2", "eu-west-1"},
			wantErr:     false,
		},
		{
			name: "success with empty regions",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeRegions(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRegionsOutput{
					Regions: []types.Region{},
				}, nil)
			},
			wantRegions: []string{},
			wantErr:     false,
		},
		{
			name: "error from API",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeRegions(gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))
			},
			wantRegions: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSDK := mocks.NewMocksdkClient(ctrl)
			tt.setupMock(mockSDK)

			client := newClientWithSDK(mockSDK)
			regions, err := client.GetRegions(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRegions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(regions) != len(tt.wantRegions) {
				t.Errorf("GetRegions() got %d regions, want %d", len(regions), len(tt.wantRegions))
			}
		})
	}
}

func TestClient_GetDefaultVPCs(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.MocksdkClient)
		wantVPCs  []string
		wantErr   bool
	}{
		{
			name: "success with default and non-default VPCs",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeVpcs(gomock.Any(), gomock.Any()).Return(&ec2.DescribeVpcsOutput{
					Vpcs: []types.Vpc{
						{VpcId: aws.String("vpc-default"), IsDefault: aws.Bool(true)},
						{VpcId: aws.String("vpc-custom"), IsDefault: aws.Bool(false)},
						{VpcId: aws.String("vpc-default2"), IsDefault: aws.Bool(true)},
					},
				}, nil)
			},
			wantVPCs: []string{"vpc-default", "vpc-default2"},
			wantErr:  false,
		},
		{
			name: "success with no default VPCs",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeVpcs(gomock.Any(), gomock.Any()).Return(&ec2.DescribeVpcsOutput{
					Vpcs: []types.Vpc{
						{VpcId: aws.String("vpc-custom"), IsDefault: aws.Bool(false)},
					},
				}, nil)
			},
			wantVPCs: []string{},
			wantErr:  false,
		},
		{
			name: "error from API",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeVpcs(gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))
			},
			wantVPCs: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSDK := mocks.NewMocksdkClient(ctrl)
			tt.setupMock(mockSDK)

			client := newClientWithSDK(mockSDK)
			vpcs, err := client.GetDefaultVPCs(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDefaultVPCs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(vpcs) != len(tt.wantVPCs) {
				t.Errorf("GetDefaultVPCs() got %d VPCs, want %d", len(vpcs), len(tt.wantVPCs))
			}
		})
	}
}

func TestClient_DeleteSubnetsInVPC(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(*mocks.MocksdkClient)
		wantErr   bool
	}{
		{
			name:  "success with multiple subnets",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeSubnets(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []types.Subnet{
						{SubnetId: aws.String("subnet-1")},
						{SubnetId: aws.String("subnet-2")},
					},
				}, nil)
				m.EXPECT().DeleteSubnet(gomock.Any(), gomock.Any()).Return(&ec2.DeleteSubnetOutput{}, nil).Times(2)
			},
			wantErr: false,
		},
		{
			name:  "success with no subnets",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeSubnets(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []types.Subnet{},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:  "error describing subnets",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeSubnets(gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))
			},
			wantErr: true,
		},
		{
			name:  "error deleting subnet",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeSubnets(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []types.Subnet{
						{SubnetId: aws.String("subnet-1")},
					},
				}, nil)
				m.EXPECT().DeleteSubnet(gomock.Any(), gomock.Any()).Return(nil, errors.New("delete error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSDK := mocks.NewMocksdkClient(ctrl)
			tt.setupMock(mockSDK)

			client := newClientWithSDK(mockSDK)
			err := client.DeleteSubnetsInVPC(context.Background(), tt.vpcID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteSubnetsInVPC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_DeleteRouteTablesInVPC(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(*mocks.MocksdkClient)
		wantErr   bool
	}{
		{
			name:  "success with non-main route tables",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeRouteTables(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRouteTablesOutput{
					RouteTables: []types.RouteTable{
						{RouteTableId: aws.String("rtb-1"), Associations: []types.RouteTableAssociation{}},
						{RouteTableId: aws.String("rtb-2"), Associations: []types.RouteTableAssociation{}},
					},
				}, nil)
				m.EXPECT().DeleteRouteTable(gomock.Any(), gomock.Any()).Return(&ec2.DeleteRouteTableOutput{}, nil).Times(2)
			},
			wantErr: false,
		},
		{
			name:  "skips main route table",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeRouteTables(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRouteTablesOutput{
					RouteTables: []types.RouteTable{
						{RouteTableId: aws.String("rtb-main"), Associations: []types.RouteTableAssociation{
							{Main: aws.Bool(true)},
						}},
						{RouteTableId: aws.String("rtb-custom"), Associations: []types.RouteTableAssociation{}},
					},
				}, nil)
				m.EXPECT().DeleteRouteTable(gomock.Any(), gomock.Any()).Return(&ec2.DeleteRouteTableOutput{}, nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:  "error describing route tables",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeRouteTables(gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))
			},
			wantErr: true,
		},
		{
			name:  "error deleting route table",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeRouteTables(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRouteTablesOutput{
					RouteTables: []types.RouteTable{
						{RouteTableId: aws.String("rtb-1"), Associations: []types.RouteTableAssociation{}},
					},
				}, nil)
				m.EXPECT().DeleteRouteTable(gomock.Any(), gomock.Any()).Return(nil, errors.New("delete error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSDK := mocks.NewMocksdkClient(ctrl)
			tt.setupMock(mockSDK)

			client := newClientWithSDK(mockSDK)
			err := client.DeleteRouteTablesInVPC(context.Background(), tt.vpcID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteRouteTablesInVPC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_DeleteInternetGatewaysInVPC(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(*mocks.MocksdkClient)
		wantErr   bool
	}{
		{
			name:  "success with internet gateways",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{
						{InternetGatewayId: aws.String("igw-1")},
					},
				}, nil)
				m.EXPECT().DetachInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DetachInternetGatewayOutput{}, nil)
				m.EXPECT().DeleteInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DeleteInternetGatewayOutput{}, nil)
			},
			wantErr: false,
		},
		{
			name:  "success with no internet gateways",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:  "error describing internet gateways",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))
			},
			wantErr: true,
		},
		{
			name:  "error detaching internet gateway",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{
						{InternetGatewayId: aws.String("igw-1")},
					},
				}, nil)
				m.EXPECT().DetachInternetGateway(gomock.Any(), gomock.Any()).Return(nil, errors.New("detach error"))
			},
			wantErr: true,
		},
		{
			name:  "error deleting internet gateway",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{
						{InternetGatewayId: aws.String("igw-1")},
					},
				}, nil)
				m.EXPECT().DetachInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DetachInternetGatewayOutput{}, nil)
				m.EXPECT().DeleteInternetGateway(gomock.Any(), gomock.Any()).Return(nil, errors.New("delete error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSDK := mocks.NewMocksdkClient(ctrl)
			tt.setupMock(mockSDK)

			client := newClientWithSDK(mockSDK)
			err := client.DeleteInternetGatewaysInVPC(context.Background(), tt.vpcID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteInternetGatewaysInVPC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_DeleteSecurityGroupsInVPC(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(*mocks.MocksdkClient)
		wantErr   bool
	}{
		{
			name:  "success with non-default security groups",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeSecurityGroups(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{
						{GroupId: aws.String("sg-1"), GroupName: aws.String("custom-sg")},
						{GroupId: aws.String("sg-2"), GroupName: aws.String("another-sg")},
					},
				}, nil)
				m.EXPECT().DeleteSecurityGroup(gomock.Any(), gomock.Any()).Return(&ec2.DeleteSecurityGroupOutput{}, nil).Times(2)
			},
			wantErr: false,
		},
		{
			name:  "skips default security group",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeSecurityGroups(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{
						{GroupId: aws.String("sg-default"), GroupName: aws.String("default")},
						{GroupId: aws.String("sg-custom"), GroupName: aws.String("custom-sg")},
					},
				}, nil)
				m.EXPECT().DeleteSecurityGroup(gomock.Any(), gomock.Any()).Return(&ec2.DeleteSecurityGroupOutput{}, nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:  "error describing security groups",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeSecurityGroups(gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))
			},
			wantErr: true,
		},
		{
			name:  "error deleting security group",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeSecurityGroups(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{
						{GroupId: aws.String("sg-1"), GroupName: aws.String("custom-sg")},
					},
				}, nil)
				m.EXPECT().DeleteSecurityGroup(gomock.Any(), gomock.Any()).Return(nil, errors.New("delete error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSDK := mocks.NewMocksdkClient(ctrl)
			tt.setupMock(mockSDK)

			client := newClientWithSDK(mockSDK)
			err := client.DeleteSecurityGroupsInVPC(context.Background(), tt.vpcID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteSecurityGroupsInVPC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_DeleteNetworkACLsInVPC(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(*mocks.MocksdkClient)
		wantErr   bool
	}{
		{
			name:  "success with non-default ACLs",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeNetworkAcls(gomock.Any(), gomock.Any()).Return(&ec2.DescribeNetworkAclsOutput{
					NetworkAcls: []types.NetworkAcl{
						{NetworkAclId: aws.String("acl-1"), IsDefault: aws.Bool(false)},
						{NetworkAclId: aws.String("acl-2"), IsDefault: aws.Bool(false)},
					},
				}, nil)
				m.EXPECT().DeleteNetworkAcl(gomock.Any(), gomock.Any()).Return(&ec2.DeleteNetworkAclOutput{}, nil).Times(2)
			},
			wantErr: false,
		},
		{
			name:  "skips default network ACL",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeNetworkAcls(gomock.Any(), gomock.Any()).Return(&ec2.DescribeNetworkAclsOutput{
					NetworkAcls: []types.NetworkAcl{
						{NetworkAclId: aws.String("acl-default"), IsDefault: aws.Bool(true)},
						{NetworkAclId: aws.String("acl-custom"), IsDefault: aws.Bool(false)},
					},
				}, nil)
				m.EXPECT().DeleteNetworkAcl(gomock.Any(), gomock.Any()).Return(&ec2.DeleteNetworkAclOutput{}, nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:  "error describing network ACLs",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeNetworkAcls(gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))
			},
			wantErr: true,
		},
		{
			name:  "error deleting network ACL",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DescribeNetworkAcls(gomock.Any(), gomock.Any()).Return(&ec2.DescribeNetworkAclsOutput{
					NetworkAcls: []types.NetworkAcl{
						{NetworkAclId: aws.String("acl-1"), IsDefault: aws.Bool(false)},
					},
				}, nil)
				m.EXPECT().DeleteNetworkAcl(gomock.Any(), gomock.Any()).Return(nil, errors.New("delete error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSDK := mocks.NewMocksdkClient(ctrl)
			tt.setupMock(mockSDK)

			client := newClientWithSDK(mockSDK)
			err := client.DeleteNetworkACLsInVPC(context.Background(), tt.vpcID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteNetworkACLsInVPC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_DeleteVPC(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(*mocks.MocksdkClient)
		wantErr   bool
	}{
		{
			name:  "success",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DeleteVpc(gomock.Any(), gomock.Any()).Return(&ec2.DeleteVpcOutput{}, nil)
			},
			wantErr: false,
		},
		{
			name:  "error deleting VPC",
			vpcID: "vpc-123",
			setupMock: func(m *mocks.MocksdkClient) {
				m.EXPECT().DeleteVpc(gomock.Any(), gomock.Any()).Return(nil, errors.New("delete error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSDK := mocks.NewMocksdkClient(ctrl)
			tt.setupMock(mockSDK)

			client := newClientWithSDK(mockSDK)
			err := client.DeleteVPC(context.Background(), tt.vpcID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteVPC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsMainRouteTable(t *testing.T) {
	tests := []struct {
		name       string
		routeTable types.RouteTable
		want       bool
	}{
		{
			name: "main route table",
			routeTable: types.RouteTable{
				RouteTableId: aws.String("rtb-main"),
				Associations: []types.RouteTableAssociation{
					{Main: aws.Bool(true)},
				},
			},
			want: true,
		},
		{
			name: "non-main route table",
			routeTable: types.RouteTable{
				RouteTableId: aws.String("rtb-custom"),
				Associations: []types.RouteTableAssociation{
					{Main: aws.Bool(false)},
				},
			},
			want: false,
		},
		{
			name: "route table with no associations",
			routeTable: types.RouteTable{
				RouteTableId: aws.String("rtb-no-assoc"),
				Associations: []types.RouteTableAssociation{},
			},
			want: false,
		},
		{
			name: "route table with multiple associations including main",
			routeTable: types.RouteTable{
				RouteTableId: aws.String("rtb-multi"),
				Associations: []types.RouteTableAssociation{
					{Main: aws.Bool(false)},
					{Main: aws.Bool(true)},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMainRouteTable(tt.routeTable)
			if got != tt.want {
				t.Errorf("isMainRouteTable() = %v, want %v", got, tt.want)
			}
		})
	}
}
