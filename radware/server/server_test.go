package main

import (
	"context"
	"log"
	"net"
	"testing"

	"lbservice"
	"mocks"

	"github.com/golang/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	lbservice.RegisterLoadBalancerServiceServer(s, &server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestCreateService(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	defer conn.Close()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	radwareSender := mocks.NewMockSender(mockCtrl)
	SetSender(radwareSender)
	reply := []byte(`{"uri" : "test", "targetUri" : "test", "complete":true}`)

	radwareSender.EXPECT().ServerSend(gomock.Any()).Return(reply, nil)
	client := lbservice.NewLoadBalancerServiceClient(conn)
	req := &lbservice.CreateInstanceRequest{
		Instance: &lbservice.Instance{
			MgmtMacAddr:   "aaa.bbb.ccc",
			MgmtIpAddr:    "1.2.3.4",
			Label:         "inside",
			Lic:           lbservice.InstanceLicense_value["LB_10GIG_LIC"],
			LicToken:      "abc",
			Vip:           "10.1.1.1",
			LbUserName:    "admin",
			LbPassword:    "Cisco@123",
			LbHttpsPort:   443,
			LbHealth:      lbservice.InstanceLbhealthchk_name[int32(lbservice.Instance_icmp)],
			LbMetric:      lbservice.InstanceLbmetric_name[int32(lbservice.Instance_roundrobin)],
			LbDsr:         true,
			LbGroupName:   "group1",
			LbServiceName: "svc1",
			LbL4Port:      5001,
		},
	}

	client.CreateService(ctx, req)
}

func TestConfigL3InterfacesService(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())

	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	radwareSender := mocks.NewMockSender(mockCtrl)
	SetSender(radwareSender)
	reply := []byte(`{"uri" : "test", "targetUri" : "test", "complete":true}`)

	radwareSender.EXPECT().ServerSend(gomock.Any()).Return(reply, nil)

	client := lbservice.NewLoadBalancerServiceClient(conn)
	interfaces := []*lbservice.L3Interface{
		{
			Label:           "Inside",
			LbInterfaceName: "int1",
			LbVlan:          2,
			LbPrimaryIp:     "1.1.1.100",
			LbSecondaryIp:   "0.0.0.0",
			LbIpMask:        "255.255.255.0",
			LbPort:          1,
			LbIsV4:          true,
			EnableHa:        false,
		},
	}
	req := &lbservice.CfgL3InterfacesRequest{
		Interfaces: interfaces,
	}
	res, err := client.ConfigL3InterfacesService(ctx, req)
	if err != nil {
		t.Fatalf("error while calling ConfigL3Interfaces RPC: %v", err)
	}
	if res.GetCfgL3InterfacesResp() != true {
		t.Fatalf("Failed to run Config L3 interfaces service")
	}
}

func TestConfigL4FilterService(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())

	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	defer conn.Close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	radwareSender := mocks.NewMockSender(mockCtrl)
	SetSender(radwareSender)
	reply := []byte(`{"uri" : "test", "targetUri" : "test", "complete":true}`)
	radwareSender.EXPECT().ServerSend(gomock.Any()).Return(reply, nil)

	client := lbservice.NewLoadBalancerServiceClient(conn)
	rules := []*lbservice.L4Filter{
		{
			Label:           "Inside",
			Name:            "rule1",
			RuleId:          100,
			Act:             lbservice.L4FilterAction_name[int32(lbservice.L4Filter_allow)],
			LbIsV4:          true,
			SrcIp:           "any",
			SrcMask:         "0.0.0.0",
			DstIp:           "any",
			DstMask:         "0.0.0.0",
			Group:           "Vdirect-ASAc-Group",
			Port:            1,
			Vlan:            "any",
			Protocol:        "any",
			ReverseSession:  false,
			ReturnToLastHop: false,
			Op:              lbservice.L4Filter_Operation_name[int32(lbservice.L4Filter_ADD)],
		},
	}
	req := &lbservice.CfgL4FilterRequest{
		Filt:     rules,
	}

	res, err := client.ConfigL4FilterService(ctx, req)
	if err != nil {
		t.Fatalf("error while calling ConfigL4Filter RPC: %v", err)
	}

	if res.GetCfgL4FilterResp() != true {
		t.Fatalf("Failed to configure L4 filter rules")
	}
}

func TestProvisionEndPointService(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())

	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	defer conn.Close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	radwareSender := mocks.NewMockSender(mockCtrl)
	SetSender(radwareSender)
	reply := []byte(`{"uri" : "test", "targetUri" : "test", "complete":true}`)

	radwareSender.EXPECT().ServerSend(gomock.Any()).Return(reply, nil)
	client := lbservice.NewLoadBalancerServiceClient(conn)
	eps := []*lbservice.EndPointCfg{
		{
			Label:            "Inside",
			IpAddress:        "1.1.1.10",
			Op:               lbservice.EndPointCfg_Operation_value["ADD"],
			AsacInstanceName: "ASAc1",
			LbGroupName:      "Vdirect-ASAc-Group",
			LbServiceName:    "ASAc-lb-svc",
		},
	}
	req := &lbservice.ProvisionEndPointRequest{
		Ep:       eps,
	}

	res, err := client.ProvisionEndPointService(ctx, req)
	if err != nil {
		t.Fatalf("error while calling ProvisionEndPoint RPC: %v", err)
	}

	if res.GetProvisionEndPointResp() != true {
		t.Fatalf("Configure Endpoints service test failed")
	}
}

func TestDestroyService(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	defer conn.Close()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	radwareSender := mocks.NewMockSender(mockCtrl)
	SetSender(radwareSender)
	reply := []byte(`{"uri" : "test", "targetUri" : "test", "complete":true}`)
	radwareSender.EXPECT().ServerSend(gomock.Any()).Return(reply, nil)
	client := lbservice.NewLoadBalancerServiceClient(conn)
	req := &lbservice.DestroyInstanceRequest{
		Label:         "Inside",
		LbServiceName: "svc1",
	}

	res, err := client.DestroyService(ctx, req)
	if err != nil {
		t.Fatalf("error while calling destroyInstance RPC: %v", err)
	}

	if res.GetDestroyInstanceResp() != true {
		t.Fatalf("Destory service test failed")
	}
}
