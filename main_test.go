package main

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"remove-default-vpc/mocks"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/mock/gomock"
)

func Test_getRegions(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(m *mocks.MockEC2API)
		want      []string
		wantErr   bool
	}{
		{
			name: "success - multiple regions",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeRegions(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRegionsOutput{
					Regions: []types.Region{
						{RegionName: aws.String("us-east-1")},
						{RegionName: aws.String("us-west-2")},
					},
				}, nil)
			},
			want:    []string{"us-east-1", "us-west-2"},
			wantErr: false,
		},
		{
			name: "error fetching regions",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeRegions(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to fetch regions"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no regions found",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeRegions(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRegionsOutput{
					Regions: []types.Region{},
				}, nil)
			},
			want:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockEC2API(ctrl)
			tt.setupMock(mockClient)

			got, err := getRegions(context.Background(), mockClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRegions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRegions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDefaultVPCs(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(m *mocks.MockEC2API)
		want      []string
		wantErr   bool
	}{
		{
			name: "success - multiple default VPCs",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeVpcs(gomock.Any(), gomock.Any()).Return(&ec2.DescribeVpcsOutput{
					Vpcs: []types.Vpc{
						{VpcId: aws.String("vpc-12345"), IsDefault: aws.Bool(true)},
						{VpcId: aws.String("vpc-67890"), IsDefault: aws.Bool(true)},
					},
				}, nil)
			},
			want:    []string{"vpc-12345", "vpc-67890"},
			wantErr: false,
		},
		{
			name: "error fetching VPCs",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeVpcs(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to fetch VPCs"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no default VPCs",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeVpcs(gomock.Any(), gomock.Any()).Return(&ec2.DescribeVpcsOutput{
					Vpcs: []types.Vpc{},
				}, nil)
			},
			want:    []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockEC2API(ctrl)
			tt.setupMock(mockClient)

			got, err := getDefaultVPCs(context.Background(), mockClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDefaultVPCs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDefaultVPCs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_deleteSubnets(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(m *mocks.MockEC2API)
		wantErr   bool
	}{
		{
			name:  "success - delete multiple subnets",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
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
			name:  "error fetching subnets",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeSubnets(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to fetch subnets"))
			},
			wantErr: true,
		},
		{
			name:  "error deleting subnet",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeSubnets(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []types.Subnet{
						{SubnetId: aws.String("subnet-1")},
					},
				}, nil)
				m.EXPECT().DeleteSubnet(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to delete subnet subnet-1"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockEC2API(ctrl)
			tt.setupMock(mockClient)

			err := deleteSubnets(context.Background(), mockClient, tt.vpcID)
			if (err != nil) != tt.wantErr {
				t.Errorf("deleteSubnets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_isMainRouteTable(t *testing.T) {
	type args struct {
		rt types.RouteTable
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "main route table",
			args: args{
				rt: types.RouteTable{
					Associations: []types.RouteTableAssociation{
						{
							Main: aws.Bool(true),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "non-main route table",
			args: args{
				rt: types.RouteTable{
					Associations: []types.RouteTableAssociation{
						{
							Main: aws.Bool(false),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "empty associations",
			args: args{
				rt: types.RouteTable{
					Associations: []types.RouteTableAssociation{},
				},
			},
			want: false,
		},
		{
			name: "nil main association",
			args: args{
				rt: types.RouteTable{
					Associations: []types.RouteTableAssociation{
						{
							Main: nil,
						},
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMainRouteTable(tt.args.rt); got != tt.want {
				t.Errorf("isMainRouteTable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_deleteRouteTables(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(m *mocks.MockEC2API)
		wantErr   bool
	}{
		{
			name:  "success - delete multiple route tables",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeRouteTables(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRouteTablesOutput{
					RouteTables: []types.RouteTable{
						{RouteTableId: aws.String("rtb-1")},
						{RouteTableId: aws.String("rtb-2")},
					},
				}, nil)
				m.EXPECT().DeleteRouteTable(gomock.Any(), gomock.Any()).Return(&ec2.DeleteRouteTableOutput{}, nil).Times(2)
			},
			wantErr: false,
		},
		{
			name:  "error describing route tables",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeRouteTables(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to describe route tables"))
			},
			wantErr: true,
		},
		{
			name:  "error deleting route table",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeRouteTables(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRouteTablesOutput{
					RouteTables: []types.RouteTable{
						{RouteTableId: aws.String("rtb-1")},
					},
				}, nil)
				m.EXPECT().DeleteRouteTable(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to delete route table"))
			},
			wantErr: true,
		},
		{
			name:  "no route tables found",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeRouteTables(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRouteTablesOutput{
					RouteTables: []types.RouteTable{},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:  "success - skip main route table",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeRouteTables(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRouteTablesOutput{
					RouteTables: []types.RouteTable{
						{
							RouteTableId: aws.String("rtb-1"),
							Associations: []types.RouteTableAssociation{{Main: aws.Bool(true)}},
						},
						{
							RouteTableId: aws.String("rtb-2"),
							Associations: []types.RouteTableAssociation{{Main: aws.Bool(false)}},
						},
					},
				}, nil)
				m.EXPECT().DeleteRouteTable(gomock.Any(), gomock.Any()).Return(&ec2.DeleteRouteTableOutput{}, nil).Times(1)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockEC2API(ctrl)
			tt.setupMock(mockClient)

			if err := deleteRouteTables(context.Background(), mockClient, tt.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("deleteRouteTables() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deleteInternetGateways(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(m *mocks.MockEC2API)
		wantErr   bool
	}{
		{
			name:  "success - multiple internet gateways",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{
						{InternetGatewayId: aws.String("igw-12345")},
						{InternetGatewayId: aws.String("igw-67890")},
					},
				}, nil)
				m.EXPECT().DetachInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DetachInternetGatewayOutput{}, nil).Times(2)
				m.EXPECT().DeleteInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DeleteInternetGatewayOutput{}, nil).Times(2)
			},
			wantErr: false,
		},
		{
			name:  "error detaching internet gateway",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{
						{InternetGatewayId: aws.String("igw-12345")},
					},
				}, nil)
				m.EXPECT().DetachInternetGateway(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to detach internet gateway"))
			},
			wantErr: true,
		},
		{
			name:  "error deleting internet gateway",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{
						{InternetGatewayId: aws.String("igw-12345")},
					},
				}, nil)
				m.EXPECT().DetachInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DetachInternetGatewayOutput{}, nil)
				m.EXPECT().DeleteInternetGateway(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to delete internet gateway"))
			},
			wantErr: true,
		},
		{
			name:  "no internet gateways found",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:  "error describing internet gateways",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to describe internet gateways"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockEC2API(ctrl)
			tt.setupMock(mockClient)

			if err := deleteInternetGateways(context.Background(), mockClient, tt.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("deleteInternetGateways() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deleteSecurityGroups(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(m *mocks.MockEC2API)
		wantErr   bool
	}{
		{
			name:  "Successfully delete a security group",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeSecurityGroups(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{
						{GroupId: aws.String("sg-12345"), GroupName: aws.String("test-group")},
					},
				}, nil)
				m.EXPECT().DeleteSecurityGroup(gomock.Any(), gomock.Any()).Return(&ec2.DeleteSecurityGroupOutput{}, nil)
			},
			wantErr: false,
		},
		{
			name:  "Successfully delete multiple security groups",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeSecurityGroups(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{
						{GroupId: aws.String("sg-12345"), GroupName: aws.String("test-group-1")},
						{GroupId: aws.String("sg-67890"), GroupName: aws.String("test-group-2")},
					},
				}, nil)
				m.EXPECT().DeleteSecurityGroup(gomock.Any(), gomock.Any()).Return(&ec2.DeleteSecurityGroupOutput{}, nil).Times(2)
			},
			wantErr: false,
		},
		{
			name:  "Skip Default Security Group",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeSecurityGroups(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{
						{GroupId: aws.String("sg-default"), GroupName: aws.String("default")},
					},
				}, nil)
				// DeleteSecurityGroup should NOT be called for default group
			},
			wantErr: false,
		},
		{
			name:  "No Security Groups to Delete",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeSecurityGroups(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:  "Error in DescribeSecurityGroups",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeSecurityGroups(gomock.Any(), gomock.Any()).Return(nil, errors.New("describe security groups failed"))
			},
			wantErr: true,
		},
		{
			name:  "Error in DeleteSecurityGroup",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeSecurityGroups(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{
						{GroupId: aws.String("sg-12345"), GroupName: aws.String("test-group")},
					},
				}, nil)
				m.EXPECT().DeleteSecurityGroup(gomock.Any(), gomock.Any()).Return(nil, errors.New("delete security group failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockEC2API(ctrl)
			tt.setupMock(mockClient)

			if err := deleteSecurityGroups(context.Background(), mockClient, tt.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("deleteSecurityGroups() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deleteNetworkACLs(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(m *mocks.MockEC2API)
		wantErr   bool
	}{
		{
			name:  "Successfully delete non-default network ACLs",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeNetworkAcls(gomock.Any(), gomock.Any()).Return(&ec2.DescribeNetworkAclsOutput{
					NetworkAcls: []types.NetworkAcl{
						{NetworkAclId: aws.String("acl-12345"), IsDefault: aws.Bool(false)},
					},
				}, nil)
				m.EXPECT().DeleteNetworkAcl(gomock.Any(), gomock.Any()).Return(&ec2.DeleteNetworkAclOutput{}, nil)
			},
			wantErr: false,
		},
		{
			name:  "Successfully skip default network ACL",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeNetworkAcls(gomock.Any(), gomock.Any()).Return(&ec2.DescribeNetworkAclsOutput{
					NetworkAcls: []types.NetworkAcl{
						{NetworkAclId: aws.String("acl-default"), IsDefault: aws.Bool(true)},
					},
				}, nil)
			},
			wantErr: false,
		},
		{
			name:  "Error describing network ACLs",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeNetworkAcls(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to describe network ACLs"))
			},
			wantErr: true,
		},
		{
			name:  "Error deleting a network ACL",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeNetworkAcls(gomock.Any(), gomock.Any()).Return(&ec2.DescribeNetworkAclsOutput{
					NetworkAcls: []types.NetworkAcl{
						{NetworkAclId: aws.String("acl-12345"), IsDefault: aws.Bool(false)},
					},
				}, nil)
				m.EXPECT().DeleteNetworkAcl(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to delete network ACL"))
			},
			wantErr: true,
		},
		{
			name:  "No network ACLs found",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeNetworkAcls(gomock.Any(), gomock.Any()).Return(&ec2.DescribeNetworkAclsOutput{
					NetworkAcls: []types.NetworkAcl{},
				}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockEC2API(ctrl)
			tt.setupMock(mockClient)

			if err := deleteNetworkACLs(context.Background(), mockClient, tt.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("deleteNetworkACLs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_deleteVPC(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(m *mocks.MockEC2API)
		wantErr   bool
	}{
		{
			name:  "Successfully delete VPC",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DeleteVpc(gomock.Any(), gomock.Any()).Return(&ec2.DeleteVpcOutput{}, nil)
			},
			wantErr: false,
		},
		{
			name:  "Error deleting VPC",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DeleteVpc(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to delete VPC"))
			},
			wantErr: true,
		},
		{
			name:  "Delete non-existent VPC",
			vpcID: "vpc-non-existent",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DeleteVpc(gomock.Any(), gomock.Any()).Return(nil, errors.New("VPC not found"))
			},
			wantErr: true,
		},
		{
			name:  "Error deleting due to dependency",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DeleteVpc(gomock.Any(), gomock.Any()).Return(nil, errors.New("DependencyViolation: VPC has dependent resources"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockEC2API(ctrl)
			tt.setupMock(mockClient)

			if err := deleteVPC(context.Background(), mockClient, tt.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("deleteVPC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_cleanupVPCResources(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(m *mocks.MockEC2API)
		wantErr   bool
	}{
		{
			name:  "Successful cleanup",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				// Internet gateways
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{{InternetGatewayId: aws.String("igw-12345")}},
				}, nil)
				m.EXPECT().DetachInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DetachInternetGatewayOutput{}, nil)
				m.EXPECT().DeleteInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DeleteInternetGatewayOutput{}, nil)

				// Subnets
				m.EXPECT().DescribeSubnets(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []types.Subnet{
						{SubnetId: aws.String("subnet-1")},
						{SubnetId: aws.String("subnet-2")},
					},
				}, nil)
				m.EXPECT().DeleteSubnet(gomock.Any(), gomock.Any()).Return(&ec2.DeleteSubnetOutput{}, nil).Times(2)

				// Route tables
				m.EXPECT().DescribeRouteTables(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRouteTablesOutput{
					RouteTables: []types.RouteTable{
						{RouteTableId: aws.String("rtb-1")},
						{RouteTableId: aws.String("rtb-2")},
					},
				}, nil)
				m.EXPECT().DeleteRouteTable(gomock.Any(), gomock.Any()).Return(&ec2.DeleteRouteTableOutput{}, nil).Times(2)

				// Network ACLs
				m.EXPECT().DescribeNetworkAcls(gomock.Any(), gomock.Any()).Return(&ec2.DescribeNetworkAclsOutput{
					NetworkAcls: []types.NetworkAcl{
						{NetworkAclId: aws.String("acl-12345"), IsDefault: aws.Bool(false)},
						{NetworkAclId: aws.String("acl-67890"), IsDefault: aws.Bool(false)},
					},
				}, nil)
				m.EXPECT().DeleteNetworkAcl(gomock.Any(), gomock.Any()).Return(&ec2.DeleteNetworkAclOutput{}, nil).Times(2)

				// Security groups
				m.EXPECT().DescribeSecurityGroups(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{
						{GroupId: aws.String("sg-12345"), GroupName: aws.String("test-group-0")},
						{GroupId: aws.String("sg-67890"), GroupName: aws.String("test-group-1")},
					},
				}, nil)
				m.EXPECT().DeleteSecurityGroup(gomock.Any(), gomock.Any()).Return(&ec2.DeleteSecurityGroupOutput{}, nil).Times(2)
			},
			wantErr: false,
		},
		{
			name:  "Error deleting internet gateways",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{{InternetGatewayId: aws.String("igw-12345")}},
				}, nil)
				m.EXPECT().DetachInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DetachInternetGatewayOutput{}, nil)
				m.EXPECT().DeleteInternetGateway(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to delete internet gateway"))
			},
			wantErr: true,
		},
		{
			name:  "Error deleting subnets",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				// Internet gateways - succeed
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{{InternetGatewayId: aws.String("igw-12345")}},
				}, nil)
				m.EXPECT().DetachInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DetachInternetGatewayOutput{}, nil)
				m.EXPECT().DeleteInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DeleteInternetGatewayOutput{}, nil)

				// Subnets - fail
				m.EXPECT().DescribeSubnets(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []types.Subnet{
						{SubnetId: aws.String("subnet-1")},
						{SubnetId: aws.String("subnet-2")},
					},
				}, nil)
				m.EXPECT().DeleteSubnet(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to delete subnet"))
			},
			wantErr: true,
		},
		{
			name:  "Error deleting route tables",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				// Internet gateways - succeed
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{{InternetGatewayId: aws.String("igw-12345")}},
				}, nil)
				m.EXPECT().DetachInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DetachInternetGatewayOutput{}, nil)
				m.EXPECT().DeleteInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DeleteInternetGatewayOutput{}, nil)

				// Subnets - succeed
				m.EXPECT().DescribeSubnets(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []types.Subnet{
						{SubnetId: aws.String("subnet-1")},
						{SubnetId: aws.String("subnet-2")},
					},
				}, nil)
				m.EXPECT().DeleteSubnet(gomock.Any(), gomock.Any()).Return(&ec2.DeleteSubnetOutput{}, nil).Times(2)

				// Route tables - fail
				m.EXPECT().DescribeRouteTables(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRouteTablesOutput{
					RouteTables: []types.RouteTable{
						{RouteTableId: aws.String("rtb-1")},
						{RouteTableId: aws.String("rtb-2")},
					},
				}, nil)
				m.EXPECT().DeleteRouteTable(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to delete route table"))
			},
			wantErr: true,
		},
		{
			name:  "Error deleting network ACLs",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				// Internet gateways - succeed
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{{InternetGatewayId: aws.String("igw-12345")}},
				}, nil)
				m.EXPECT().DetachInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DetachInternetGatewayOutput{}, nil)
				m.EXPECT().DeleteInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DeleteInternetGatewayOutput{}, nil)

				// Subnets - succeed
				m.EXPECT().DescribeSubnets(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []types.Subnet{
						{SubnetId: aws.String("subnet-1")},
						{SubnetId: aws.String("subnet-2")},
					},
				}, nil)
				m.EXPECT().DeleteSubnet(gomock.Any(), gomock.Any()).Return(&ec2.DeleteSubnetOutput{}, nil).Times(2)

				// Route tables - succeed
				m.EXPECT().DescribeRouteTables(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRouteTablesOutput{
					RouteTables: []types.RouteTable{
						{RouteTableId: aws.String("rtb-1")},
						{RouteTableId: aws.String("rtb-2")},
					},
				}, nil)
				m.EXPECT().DeleteRouteTable(gomock.Any(), gomock.Any()).Return(&ec2.DeleteRouteTableOutput{}, nil).Times(2)

				// Network ACLs - fail
				m.EXPECT().DescribeNetworkAcls(gomock.Any(), gomock.Any()).Return(&ec2.DescribeNetworkAclsOutput{
					NetworkAcls: []types.NetworkAcl{
						{NetworkAclId: aws.String("acl-12345"), IsDefault: aws.Bool(false)},
						{NetworkAclId: aws.String("acl-67890"), IsDefault: aws.Bool(false)},
					},
				}, nil)
				m.EXPECT().DeleteNetworkAcl(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to delete network ACL"))
			},
			wantErr: true,
		},
		{
			name:  "Error deleting security groups",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockEC2API) {
				// Internet gateways - succeed
				m.EXPECT().DescribeInternetGateways(gomock.Any(), gomock.Any()).Return(&ec2.DescribeInternetGatewaysOutput{
					InternetGateways: []types.InternetGateway{{InternetGatewayId: aws.String("igw-12345")}},
				}, nil)
				m.EXPECT().DetachInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DetachInternetGatewayOutput{}, nil)
				m.EXPECT().DeleteInternetGateway(gomock.Any(), gomock.Any()).Return(&ec2.DeleteInternetGatewayOutput{}, nil)

				// Subnets - succeed
				m.EXPECT().DescribeSubnets(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []types.Subnet{
						{SubnetId: aws.String("subnet-1")},
						{SubnetId: aws.String("subnet-2")},
					},
				}, nil)
				m.EXPECT().DeleteSubnet(gomock.Any(), gomock.Any()).Return(&ec2.DeleteSubnetOutput{}, nil).Times(2)

				// Route tables - succeed
				m.EXPECT().DescribeRouteTables(gomock.Any(), gomock.Any()).Return(&ec2.DescribeRouteTablesOutput{
					RouteTables: []types.RouteTable{
						{RouteTableId: aws.String("rtb-1")},
						{RouteTableId: aws.String("rtb-2")},
					},
				}, nil)
				m.EXPECT().DeleteRouteTable(gomock.Any(), gomock.Any()).Return(&ec2.DeleteRouteTableOutput{}, nil).Times(2)

				// Network ACLs - succeed
				m.EXPECT().DescribeNetworkAcls(gomock.Any(), gomock.Any()).Return(&ec2.DescribeNetworkAclsOutput{
					NetworkAcls: []types.NetworkAcl{
						{NetworkAclId: aws.String("acl-12345"), IsDefault: aws.Bool(false)},
						{NetworkAclId: aws.String("acl-67890"), IsDefault: aws.Bool(false)},
					},
				}, nil)
				m.EXPECT().DeleteNetworkAcl(gomock.Any(), gomock.Any()).Return(&ec2.DeleteNetworkAclOutput{}, nil).Times(2)

				// Security groups - fail
				m.EXPECT().DescribeSecurityGroups(gomock.Any(), gomock.Any()).Return(&ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []types.SecurityGroup{
						{GroupId: aws.String("sg-12345"), GroupName: aws.String("test-group-0")},
						{GroupId: aws.String("sg-67890"), GroupName: aws.String("test-group-1")},
					},
				}, nil)
				m.EXPECT().DeleteSecurityGroup(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to delete security group"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockEC2API(ctrl)
			tt.setupMock(mockClient)

			if err := cleanupVPCResources(context.Background(), mockClient, tt.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("cleanupVPCResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
