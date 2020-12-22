package main

import (
	"context"
	"flag"
	grpc "google.golang.org/grpc"
	"lbservice"
	"log"
	"sync"
	"time"
)

var ip string
var port string

func init() {
	flag.StringVar(&ip, "ip", "127.0.0.1", "ip address")
	flag.StringVar(&port, "port", "8080", "L4 port")
}

func sandwichLb(conn *grpc.ClientConn) {
	var wg sync.WaitGroup
	wg.Add(2)

	//configure outside LB
	go sandwichLbOutside(&wg, conn)
	//configure inside LB
	go sandwichLbInside(&wg, conn)

	//wait for configuration to finish
	wg.Wait()
}

//configure Outside Lb
func sandwichLbOutside(wg *sync.WaitGroup, conn *grpc.ClientConn) {
	defer wg.Done()

	c := lbservice.NewLoadBalancerServiceClient(conn)
	// Create Outside Loadbalance instance
	lb := createInstance(c,
		"111.222.333",
		"192.168.5.18",
		"Outside",
		lbservice.InstanceLicense_value["LB_10GIG_LIC"],
		"va-1hZQ2rdk",
		"2.2.2.174",
		"admin",
		"Cisco@123",
		443,
		lbservice.InstanceLbhealthchk_name[int32(lbservice.Instance_icmp)],
		lbservice.InstanceLbmetric_name[int32(lbservice.Instance_roundrobin)],
		false,
		"ASAc-egress",
		"ASAc-lb-svc",
		8888)

	time.Sleep(30 * time.Second)

	interfaces := []*lbservice.L3Interface{
		{
			Id:              lb,
			Label:           "Outside",
			LbInterfaceName: "tooutsideclient",
			LbVlan:          1,
			LbPrimaryIp:     "192.168.3.180",
			LbSecondaryIp:   "0.0.0.0",
			LbIpMask:        "255.255.255.0",
			LbPort:          1,
			LbIsV4:          true,
			EnableHa:        false,
		},
		{
			Id:              lb,
			Label:           "Outside",
			LbInterfaceName: "tooutsideasac",
			LbVlan:          2,
			LbPrimaryIp:     "5.5.5.180",
			LbSecondaryIp:   "0.0.0.0",
			LbIpMask:        "255.255.255.0",
			LbPort:          2,
			LbIsV4:          true,
			EnableHa:        false,
		},
	}

	if configL3Interfaces(c, interfaces) != true {
		log.Fatalf("error configuring l3 interfaces for %v", lb)
	}

	time.Sleep(10 * time.Second)
	rules := []*lbservice.L4Filter{
		{
			Id:              lb,
			Label:           "Outside",
			Name:            "Incoming traffic",
			RuleId:          100,
			Act:             lbservice.L4FilterAction_name[int32(lbservice.L4Filter_redirect)],
			LbIsV4:          true,
			SrcIp:           "any",
			SrcMask:         "0.0.0.0",
			DstIp:           "any",
			DstMask:         "0.0.0.0",
			Group:           "ASAc-egress",
			Port:            1,
			Vlan:            "any",
			Protocol:        "any",
			ReverseSession:  false,
			ReturnToLastHop: false,
			Op:              lbservice.L4Filter_Operation_name[int32(lbservice.L4Filter_ADD)],
		},
		{
			Id:              lb,
			Label:           "Outside",
			Name:            "Outbound traffic",
			RuleId:          200,
			Act:             lbservice.L4FilterAction_name[int32(lbservice.L4Filter_allow)],
			LbIsV4:          true,
			SrcIp:           "any",
			SrcMask:         "0.0.0.0",
			DstIp:           "any",
			DstMask:         "0.0.0.0",
			Group:           "ASAc-egress",
			Port:            2,
			Vlan:            "any",
			Protocol:        "any",
			ReverseSession:  true,
			ReturnToLastHop: true,
			Op:              lbservice.L4Filter_Operation_name[int32(lbservice.L4Filter_ADD)],
		},
	}

	if configL4Filters(c, rules) != true {
		log.Fatalf("error configuring l4 filter rule for %v", lb)
	}

	time.Sleep(10 * time.Second)
	eps := []*lbservice.EndPointCfg{
		{
			Id:               lb,
			Label:            "Outside",
			IpAddress:        "5.5.5.15",
			Op:               lbservice.EndPointCfg_Operation_value["ADD"],
			AsacInstanceName: "ASAc1",
			LbGroupName:      "ASAc-egress",
			LbServiceName:    "ASAc-lb-svc",
		},
		/*		{
					Id:               lb,
					Label:            "Inside",
					IpAddress:        "1.1.1.11",
					Op:               lbservice.EndPointCfg_Operation_value["ADD"],
					AsacInstanceName: "ASAc2",
					LbGroupName:      "Vdirect-ASAc-Group",
					LbServiceName:    "ASAc-lb-svc",
				},
				{
					Id:               lb,
					Label:            "Inside",
					IpAddress:        "1.1.1.12",
					Op:               lbservice.EndPointCfg_Operation_value["ADD"],
					AsacInstanceName: "ASAc3",
					LbGroupName:      "Vdirect-ASAc-Group",
					LbServiceName:    "ASAc-lb-svc",
				},*/
	}

	if provisionEndPoints(c, eps) != true {
		log.Fatalf("error adding provisioning Endpoints for  %+v", eps)
	}

}

//configure Inside Lb
func sandwichLbInside(wg *sync.WaitGroup, conn *grpc.ClientConn) {
	defer wg.Done()

	c := lbservice.NewLoadBalancerServiceClient(conn)
	// Create Inside Loadbalance instance
	lb := createInstance(c,
		"111.222.44",
		"192.168.5.20",
		"Inside",
		lbservice.InstanceLicense_value["LB_10GIG_LIC"],
		"va-1hZQ2rdx",
		"2.2.2.174",
		"admin",
		"Cisco@123",
		443,
		lbservice.InstanceLbhealthchk_name[int32(lbservice.Instance_icmp)],
		lbservice.InstanceLbmetric_name[int32(lbservice.Instance_roundrobin)],
		false,
		"ASAc-inside",
		"ASAc-lb-svc",
		8888)

	time.Sleep(30 * time.Second)

	interfaces := []*lbservice.L3Interface{
		{
			Id:              lb,
			Label:           "Inside",
			LbInterfaceName: "insideofasac",
			LbVlan:          1,
			LbPrimaryIp:     "6.6.6.172",
			LbSecondaryIp:   "0.0.0.0",
			LbIpMask:        "255.255.255.0",
			LbPort:          1,
			LbIsV4:          true,
			EnableHa:        false,
		},
		{
			Id:              lb,
			Label:           "Inside",
			LbInterfaceName: "toinsideserver",
			LbVlan:          2,
			LbPrimaryIp:     "192.168.4.174",
			LbSecondaryIp:   "0.0.0.0",
			LbIpMask:        "255.255.255.0",
			LbPort:          2,
			LbIsV4:          true,
			EnableHa:        false,
		},
	}

	if configL3Interfaces(c, interfaces) != true {
		log.Fatalf("error configuring l3 interfaces for %v", lb)
	}

	time.Sleep(10 * time.Second)
	rules := []*lbservice.L4Filter{
		{
			Id:              lb,
			Label:           "Inside",
			Name:            "Outbound traffic",
			RuleId:          100,
			Act:             lbservice.L4FilterAction_name[int32(lbservice.L4Filter_redirect)],
			LbIsV4:          true,
			SrcIp:           "any",
			SrcMask:         "0.0.0.0",
			DstIp:           "any",
			DstMask:         "0.0.0.0",
			Group:           "ASAc-inside",
			Port:            2,
			Vlan:            "any",
			Protocol:        "any",
			ReverseSession:  false,
			ReturnToLastHop: false,
			Op:              lbservice.L4Filter_Operation_name[int32(lbservice.L4Filter_ADD)],
		},
		{
			Id:              lb,
			Label:           "Inside",
			Name:            "Incoming traffic",
			RuleId:          200,
			Act:             lbservice.L4FilterAction_name[int32(lbservice.L4Filter_allow)],
			LbIsV4:          true,
			SrcIp:           "any",
			SrcMask:         "0.0.0.0",
			DstIp:           "any",
			DstMask:         "0.0.0.0",
			Group:           "ASAc-inside",
			Port:            1,
			Vlan:            "any",
			Protocol:        "any",
			ReverseSession:  true,
			ReturnToLastHop: true,
			Op:              lbservice.L4Filter_Operation_name[int32(lbservice.L4Filter_ADD)],
		},
	}

	if configL4Filters(c, rules) != true {
		log.Fatalf("error configuring l4 filter rule for %v", lb)
	}

	time.Sleep(10 * time.Second)
	eps := []*lbservice.EndPointCfg{
		{
			Id:               lb,
			Label:            "Inside",
			IpAddress:        "6.6.6.14",
			Op:               lbservice.EndPointCfg_Operation_value["ADD"],
			AsacInstanceName: "ASAc1",
			LbGroupName:      "ASAc-inside",
			LbServiceName:    "ASAc-lb-svc",
		},
		/*		{
					Id:               lb,
					Label:            "Inside",
					IpAddress:        "1.1.1.11",
					Op:               lbservice.EndPointCfg_Operation_value["ADD"],
					AsacInstanceName: "ASAc2",
					LbGroupName:      "Vdirect-ASAc-Group",
					LbServiceName:    "ASAc-lb-svc",
				},
				{
					Id:               lb,
					Label:            "Inside",
					IpAddress:        "1.1.1.12",
					Op:               lbservice.EndPointCfg_Operation_value["ADD"],
					AsacInstanceName: "ASAc3",
					LbGroupName:      "Vdirect-ASAc-Group",
					LbServiceName:    "ASAc-lb-svc",
				},*/
	}

	if provisionEndPoints(c, eps) != true {
		log.Fatalf("error adding provisioning Endpoints for  %+v", eps)
	}

}

func singleLb(conn *grpc.ClientConn) {
	c := lbservice.NewLoadBalancerServiceClient(conn)
	// Create Outside Loadbalance instance
	lb1 := createInstance(c,
		"111.222.333",
		"192.168.0.19",
		"Inside",
		lbservice.InstanceLicense_value["LB_10GIG_LIC"],
		"va-1hZQ2rdk",
		"3.3.3.50",
		"admin",
		"Cisco-123456",
		443,
		lbservice.InstanceLbhealthchk_name[int32(lbservice.Instance_icmp)],
		lbservice.InstanceLbmetric_name[int32(lbservice.Instance_roundrobin)],
		false,
		"Vdirect-ASAc-Group",
		"ASAc-lb-svc",
		5001)

	interfaces := []*lbservice.L3Interface{
		{
			Id:              lb1,
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

	if configL3Interfaces(c, interfaces) != true {
		log.Fatalf("error configuring l3 interfaces for %v", lb1)
	}

	rules := []*lbservice.L4Filter{
		{
			Id:              lb1,
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
		{
			Id:              lb1,
			Label:           "Inside",
			Name:            "rule2",
			RuleId:          200,
			Act:             lbservice.L4FilterAction_name[int32(lbservice.L4Filter_redirect)],
			LbIsV4:          true,
			SrcIp:           "any",
			SrcMask:         "0.0.0.0",
			DstIp:           "any",
			DstMask:         "0.0.0.0",
			Group:           "Vdirect-ASAc-Group",
			Port:            1,
			Vlan:            "any",
			Protocol:        "any",
			ReverseSession:  true,
			ReturnToLastHop: true,
			Op:              lbservice.L4Filter_Operation_name[int32(lbservice.L4Filter_ADD)],
		},
	}

	if configL4Filters(c, rules) != true {
		log.Fatalf("error configuring l4 filter rule for %v", lb1)
	}

	eps := []*lbservice.EndPointCfg{
		{
			Id:               lb1,
			Label:            "Inside",
			IpAddress:        "1.1.1.10",
			Op:               lbservice.EndPointCfg_Operation_value["ADD"],
			AsacInstanceName: "ASAc1",
			LbGroupName:      "Vdirect-ASAc-Group",
			LbServiceName:    "ASAc-lb-svc",
		},
		{
			Id:               lb1,
			Label:            "Inside",
			IpAddress:        "1.1.1.11",
			Op:               lbservice.EndPointCfg_Operation_value["ADD"],
			AsacInstanceName: "ASAc2",
			LbGroupName:      "Vdirect-ASAc-Group",
			LbServiceName:    "ASAc-lb-svc",
		},
		{
			Id:               lb1,
			Label:            "Inside",
			IpAddress:        "1.1.1.12",
			Op:               lbservice.EndPointCfg_Operation_value["ADD"],
			AsacInstanceName: "ASAc3",
			LbGroupName:      "Vdirect-ASAc-Group",
			LbServiceName:    "ASAc-lb-svc",
		},
	}

	if provisionEndPoints(c, eps) != true {
		log.Fatalf("error adding provisioning Endpoints for  %+v", eps)
	}

	queryInstance(c, "Inside")

	//destroy all rsources created in Alteon
	destroyInstance(c, "Inside", "Vdirect-ASAc-Group")
}

func main() {
	flag.Parse()
	socket := ip + ":" + port
	log.Println(socket)
	conn, err := grpc.Dial(socket, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect %v", err)
	}

	defer conn.Close()
	// Test single LB deployment.
	singleLb(conn)

	// Test sandwich LB deployment.
	// sandwichLb(conn)

}

func createInstance(c lbservice.LoadBalancerServiceClient,
	mgmtMacAddr string,
	mgmtIP string,
	label string,
	lic int32, lictoken string,
	vip string,
	lbusername string,
	lbpassword string,
	lbhttpsport int32,
	lbhealth string, lbmetric string, lbdsr bool,
	lbgroupname string, lbservicename string, lbl4port int32) (ID *lbservice.InstanceId) {
	log.Printf("Starting LibcreateInstance gRPC")

	req := &lbservice.CreateInstanceRequest{
		Instance: &lbservice.Instance{
			MgmtMacAddr:   mgmtMacAddr,
			MgmtIpAddr:    mgmtIP,
			Label:         label,
			Lic:           lic,
			LicToken:      lictoken,
			Vip:           vip,
			LbUserName:    lbusername,
			LbPassword:    lbpassword,
			LbHttpsPort:   lbhttpsport,
			LbHealth:      lbhealth,
			LbMetric:      lbmetric,
			LbDsr:         lbdsr,
			LbGroupName:   lbgroupname,
			LbServiceName: lbservicename,
			LbL4Port:      lbl4port,
		},
	}

	res, err := c.CreateService(context.Background(), req)

	if err != nil {
		log.Fatalf("error while calling CreateService RPC: %v", err)
	}

	log.Printf("Respone from CreateService: %v", res.GetId())
	return res.GetId()
}

func configL3Interfaces(c lbservice.LoadBalancerServiceClient,
	interfaces []*lbservice.L3Interface) (resp bool) {

	log.Printf("Starting Config L3 Interfaces for %+v", interfaces)

	req := &lbservice.CfgL3InterfacesRequest{
		Interfaces: interfaces,
	}
	res, err := c.ConfigL3InterfacesService(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling ConfigL3Interfaces RPC: %v", err)
	}

	log.Printf("Respone from ConfigL3Interfaces: %v", res)
	return res.GetCfgL3InterfacesResp()
}

func configL4Filters(c lbservice.LoadBalancerServiceClient,
	rules []*lbservice.L4Filter) (resp bool) {
	log.Printf("Starting Config L4filter rule for %+v", rules)

	req := &lbservice.CfgL4FilterRequest{
		Filt: rules,
	}

	res, err := c.ConfigL4FilterService(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling ConfigL4Filter RPC: %v", err)
	}

	log.Printf("Respone from ConfigL4Filter: %v", res)
	return res.GetCfgL4FilterResp()
}

func provisionEndPoints(c lbservice.LoadBalancerServiceClient,
	eps []*lbservice.EndPointCfg) (resp bool) {

	log.Printf("Starting ProvisionEndPoint gRPC for %+v", eps)

	req := &lbservice.ProvisionEndPointRequest{
		Ep: eps,
	}

	res, err := c.ProvisionEndPointService(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling ProvisionEndPoint RPC: %v", err)
	}

	log.Printf("Respone from ProvisionEndPoint: %v", res)
	return res.GetProvisionEndPointResp()
}

func destroyInstance(c lbservice.LoadBalancerServiceClient,
	label string, service string) (resp bool) {

	req := &lbservice.DestroyInstanceRequest{
		Label:         label,
		LbServiceName: service,
	}

	res, err := c.DestroyService(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling destroyInstance RPC: %v", err)
	}

	log.Printf("Respone from destroyInstance: %v", res)
	return res.GetDestroyInstanceResp()
}

func queryInstance(c lbservice.LoadBalancerServiceClient,
	label string) (resp []*lbservice.EndPointInstance) {
	log.Printf("Starting QueryInstance gRPC for %v", label)

	req := &lbservice.QueryInstanceRequest{Label: label}

	res, err := c.QueryInstanceService(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling dQueryInstance RPC: %v", err)
	}

	log.Printf("Respone from QueryInstance: %v", res)
	return res.GetQueryInstance()
}
