package main

import (
	"context"
	"errors"
	"testing"

	"remove-default-vpc/pkg/aws/ec2/mocks"

	"go.uber.org/mock/gomock"
)

func Test_cleanupVPCResources(t *testing.T) {
	tests := []struct {
		name      string
		vpcID     string
		setupMock func(m *mocks.MockAPI)
		wantErr   bool
	}{
		{
			name:  "Successful cleanup",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().DeleteInternetGatewaysInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteSubnetsInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteRouteTablesInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteNetworkACLsInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteSecurityGroupsInVPC(gomock.Any(), "vpc-12345").Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "Error deleting internet gateways",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().DeleteInternetGatewaysInVPC(gomock.Any(), "vpc-12345").Return(errors.New("failed to delete internet gateway"))
			},
			wantErr: true,
		},
		{
			name:  "Error deleting subnets",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().DeleteInternetGatewaysInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteSubnetsInVPC(gomock.Any(), "vpc-12345").Return(errors.New("failed to delete subnet"))
			},
			wantErr: true,
		},
		{
			name:  "Error deleting route tables",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().DeleteInternetGatewaysInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteSubnetsInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteRouteTablesInVPC(gomock.Any(), "vpc-12345").Return(errors.New("failed to delete route table"))
			},
			wantErr: true,
		},
		{
			name:  "Error deleting network ACLs",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().DeleteInternetGatewaysInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteSubnetsInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteRouteTablesInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteNetworkACLsInVPC(gomock.Any(), "vpc-12345").Return(errors.New("failed to delete network ACL"))
			},
			wantErr: true,
		},
		{
			name:  "Error deleting security groups",
			vpcID: "vpc-12345",
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().DeleteInternetGatewaysInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteSubnetsInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteRouteTablesInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteNetworkACLsInVPC(gomock.Any(), "vpc-12345").Return(nil)
				m.EXPECT().DeleteSecurityGroupsInVPC(gomock.Any(), "vpc-12345").Return(errors.New("failed to delete security group"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockAPI(ctrl)
			tt.setupMock(mockClient)

			if err := cleanupVPCResources(context.Background(), mockClient, tt.vpcID); (err != nil) != tt.wantErr {
				t.Errorf("cleanupVPCResources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
