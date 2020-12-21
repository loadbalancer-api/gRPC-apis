package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"
	"lbservice"
	"os"
	"time"
)

const (
	vdirectBaseUrl       = "https://0.0.0.0:2189/api/"
	vdirectUserName      = "root"
	vdirectPassword      = "C!sc0123"
	alteonSshPort        = 22
	alteonConfigProtocol = "HTTPS"
	MAX_COUNT            = 10
)

var grpcLog glog.LoggerV2

type queryApiShortRsp struct {
	URI       string `json:"uri"`
	TargetURI string `json:"targetUri"`
	Complete  bool   `json:"complete"`
}
type queryApiRsp struct {
	URI        string `json:"uri"`
	TargetURI  string `json:"targetUri"`
	Complete   bool   `json:"complete"`
	Status     int    `json:"status"`
	Success    bool   `json:"success"`
	Action     string `json:"action"`
	Timestamp  string `json:"timestamp"`
	Duration   int    `json:"duration"`
	Parameters struct {
		Output struct {
			LbGroup []struct {
				Asacname string `json:"asacname"`
				Asacip   string `json:"asacip"`
			} `json:"LbGroup"`
		} `json:"output"`
	} `json:"parameters"`
	Info struct {
		GeneratedScript string `json:"generatedScript"`
		CliOutput       string `json:"cliOutput"`
	} `json:"info"`
	GeneratedScript string `json:"generatedScript"`
	CliOutput       string `json:"cliOutput"`
}

type server struct{}

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

func serverSend(r *http.Request) ([]byte, error) {

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: customTransport}

	resp, err := client.Do(r)
	if err != nil {
		log.Fatal("Error:  ", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body. ", err)
	}
	//log.Println("serverSend ", string(body))
	return body, err
}

func serverPrepareHttphdr(req []byte, url string, op string, content string) (r *http.Request, err error) {
	r, err = http.NewRequest(op, url, bytes.NewBuffer(req))
	r.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(vdirectUserName+":"+vdirectPassword)))
	if content != "" {
		r.Header.Set("Content-Type", content)
	}
	if err != nil {
		log.Fatal("Error create new Http request", err)
	}
	return r, err

}

func readResponce(req []byte, url string, op string, content string) (complete bool, uri string) {
	var s queryApiShortRsp
	r, err := serverPrepareHttphdr(req, url, op, content)

	res, err := serverSend(r)
	if err != nil {
		log.Fatal("Error sending to vdirect server", err)
	}

	err = json.Unmarshal(res, &s)
	if err != nil {
		log.Fatal("Json Unmarshall failed :", err)
	}
	//log.Printf("short queryapi response %+v", s)

	return s.Complete, s.URI
}

func readFullResponce(req []byte, url string) (s queryApiRsp, err error) {
	r, err := serverPrepareHttphdr(req, url, "GET", "")

	res, err := serverSend(r)
	if err != nil {
		log.Fatal("Error in Destroying service", err)
	}

	err = json.Unmarshal(res, &s)
	if err != nil {
		log.Println("Json Unmarshall failed :", err)
	}
	//log.Printf("full queryapi response %+v", s)

	return s, err
}

func addAdcToVdirect(req *lbservice.CreateInstanceRequest) (id string, err error) {
	instance := req.GetInstance()
	url := vdirectBaseUrl + "container/"

	type Config struct {
		Name          string        `json:"name"`
		Type          string        `json:"type"`
		Tenants       []interface{} `json:"tenants"`
		Configuration struct {
			ConfigProtocol string `json:"configProtocol"`
			Host           string `json:"host"`
			CliUser        string `json:"cli.user"`
			CliPassword    string `json:"cli.password"`
			CliSSH         bool   `json:"cli.ssh"`
			CliPort        int    `json:"cli.port"`
			HTTPSPort      int32  `json:"https.port"`
			HTTPSUser      string `json:"https.user"`
			HTTPSPassword  string `json:"https.password"`
		} `json:"configuration"`
		ExtensionProperties struct {
		} `json:"extensionProperties"`
	}

	type Response struct {
		URI  string `json:"uri"`
		Name string `json:"name"`
		Moi  struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"moi"`
		ID struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"id"`
		Tenants []interface{} `json:"tenants"`
		Links   struct {
			Capacity     string `json:"capacity"`
			AdcVersion   string `json:"adcVersion"`
			Adc          string `json:"adc"`
			AlteonDevice string `json:"alteonDevice"`
		} `json:"links"`
		Type          string `json:"type"`
		AlteonDevice  bool   `json:"alteonDevice"`
		InstanceType  string `json:"instanceType"`
		Configuration struct {
			CliSSH            string `json:"cli.ssh"`
			ConfigProtocol    string `json:"configProtocol"`
			SnmpPort          string `json:"snmp.port"`
			CliPort           string `json:"cli.port"`
			HTTPSPort         string `json:"https.port"`
			Host              string `json:"host"`
			SnmpV3PrivacyType string `json:"snmp.v3.privacy.type"`
			SnmpVersion       string `json:"snmp.version"`
			SnmpV3AuthType    string `json:"snmp.v3.auth.type"`
		} `json:"configuration"`
	}
	config := Config{}
	config.Name = instance.GetLabel()
	config.Type = "AlteonDedicated"
	config.Configuration.ConfigProtocol = alteonConfigProtocol
	config.Configuration.Host = instance.GetMgmtIpAddr()
	config.Configuration.CliUser = instance.GetLbUserName()
	config.Configuration.CliPassword = instance.GetLbPassword()
	config.Configuration.CliSSH = true
	config.Configuration.CliPort = alteonSshPort
	config.Configuration.HTTPSPort = instance.GetLbHttpsPort()
	config.Configuration.HTTPSUser = instance.GetLbUserName()
	config.Configuration.HTTPSPassword = instance.GetLbPassword()

	requestBody, err := json.Marshal(config)
	if err != nil {
		log.Fatalln(err)
	}

	if req.GetTestOnly() {
		return "inside", nil
	}

	r, err := serverPrepareHttphdr(requestBody, url, "POST", "application/vnd.com.radware.vdirect.container+json")
	body, err := serverSend(r)
	var data Response
	json.Unmarshal(body, &data)
	return data.ID.ID, err
}

func uploadConfigTemplate(template string, templateName string) error {
	url := vdirectBaseUrl + "template?failIfInvalid=true&name=" + templateName

	vmFile, err := os.Open(template)
	if err != nil {
		log.Fatal(err)
	}
	fileInfo, _ := vmFile.Stat()
	var size int64 = fileInfo.Size()
	bytes := make([]byte, size)

	// read file into bytes
	buffer := bufio.NewReader(vmFile)
	_, err = buffer.Read(bytes)
	r, err := serverPrepareHttphdr(bytes, url, "POST", "text/x-velocity")
	defer vmFile.Close()
	_, err = serverSend(r)
	return err
}

func createAlteonLb(req *lbservice.CreateInstanceRequest) error {
	instance := req.GetInstance()
	url := vdirectBaseUrl + "runnable/ConfigurationTemplate/create_lb.vm/run"
	type CreateLbReq struct {
		DryRun bool `json:"__dryRun"`
		Alteon struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"alteon"`
		VIRTNAME string `json:"VIRT_NAME"`
		VIP      string `json:"VIP"`
		L4PORT   int32  `json:"L4PORT"`
		METRIC   string `json:"METRIC"`
		HEALTH   string `json:"HEALTH"`
	}

	config := CreateLbReq{}
	config.DryRun = false
	config.Alteon.Type = "Container"
	config.Alteon.Name = instance.GetLabel()
	config.VIP = instance.GetVip()
	config.L4PORT = instance.GetLbL4Port()
	config.METRIC = instance.GetLbMetric()
	config.HEALTH = instance.GetLbHealth()
	config.VIRTNAME = instance.GetLbGroupName()

	requestBody, err := json.Marshal(config)
	if err != nil {
		log.Fatalln(err)
	}

	c, uri := readResponce(requestBody, url, "POST", "application/json")
	count := 0
	for c != true && count < MAX_COUNT {
		time.Sleep(2 * time.Second)
		c, uri = readResponce(requestBody, uri, "GET", "")
		count++
	}
	if count == MAX_COUNT {
		err = fmt.Errorf("Timed out creating Loadbalancer instance %s", instance.GetLabel())
	}

	return err
}

func licAlteonLb(req *lbservice.CreateInstanceRequest) error {
	instance := req.GetInstance()
	url := vdirectBaseUrl + "runnable/Plugin/license/allocateAlteonLicense"
	type LicLbReq struct {
		MgmtIp      string `json:"alteon"`
		Entitlement string `json:"entitlement"`
		Throughput  int32  `json:"throughput"`
		AddOn       bool   `json:"add-on"`
	}

	config := LicLbReq{}
	config.MgmtIp = instance.GetMgmtIpAddr()
	config.Entitlement = instance.GetLicToken()
	config.Throughput = instance.GetLic()
	config.AddOn = false
	requestBody, err := json.Marshal(config)
	if err != nil {
		log.Fatalln(err)
	}

	c, uri := readResponce(requestBody, url, "POST", "application/json")
	count := 0
	for c != true && count < MAX_COUNT {
		time.Sleep(2 * time.Second)
		c, uri = readResponce(requestBody, uri, "GET", "")
		count++
	}
	if count == MAX_COUNT {
		err = fmt.Errorf("Timed out configuring lb license")
	}

	return err
}

func configL3Network(req *lbservice.CfgL3InterfacesRequest) error {
	var err error = nil
	url := vdirectBaseUrl + "runnable/ConfigurationTemplate/setup_l3.vm/run"

	type ConfigL3NetworkReq struct {
		DryRun bool `json:"__dryRun"`
		Alteon struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"alteon"`
		L3Interface struct {
			Name                 string `json:"name"`
			Vlan                 int32  `json:"vlan"`
			L3PrimaryIPAddress   string `json:"l3_primary_ip_address"`
			L3SecondaryIPAddress string `json:"l3_secondary_ip_address"`
			FloatingIPAddress    string `json:"floating_ip_address"`
			IPNetmask            string `json:"ip_netmask"`
			IPPrefix             int    `json:"ip_prefix"`
			Gw                   string `json:"gw"`
			IPVersion            string `json:"ip_version"`
			Port                 int32  `json:"port"`
		} `json:"l3_interface"`
		Dgw          string `json:"dgw"`
		DgwIPVersion string `json:"dgw_ip_version"`
		HaEnabled    bool   `json:"ha_enabled"`
	}
	for _, instance := range req.GetInterfaces() {
		// log.Printf("Configure L3interface %+v", instance)
		config := ConfigL3NetworkReq{}
		config.DryRun = false
		config.Alteon.Type = "Container"
		config.Alteon.Name = instance.GetLabel()
		config.L3Interface.Name = instance.GetLbInterfaceName()
		config.L3Interface.Vlan = instance.GetLbVlan()
		config.L3Interface.L3PrimaryIPAddress = instance.GetLbPrimaryIp()
		config.L3Interface.L3SecondaryIPAddress = instance.GetLbSecondaryIp()
		config.L3Interface.FloatingIPAddress = "0.0.0.0"
		config.L3Interface.IPNetmask = instance.GetLbIpMask()
		config.L3Interface.Gw = "0.0.0.0"
		if instance.GetLbIsV4() {
			config.L3Interface.IPVersion = "v4"
			config.DgwIPVersion = "v4"
		} else {
			config.L3Interface.IPVersion = "v6"
			config.DgwIPVersion = "v6"
		}
		config.L3Interface.Port = instance.GetLbPort()
		config.Dgw = "0.0.0.0"
		config.HaEnabled = instance.GetEnableHa()

		requestBody, err := json.Marshal(config)
		if err != nil {
			log.Fatalln(err)
		}

		if req.GetTestOnly() {
			continue
		}
		c, uri := readResponce(requestBody, url, "POST", "application/json")
		count := 0
		for c != true && count < MAX_COUNT {
			time.Sleep(2 * time.Second)
			c, uri = readResponce(requestBody, uri, "GET", "")
			count++
		}
		if count == MAX_COUNT {
			err = fmt.Errorf("Timed out Configuring L3 interfaces")
		}
	}

	return err
}

func (*server) CreateService(ctx context.Context, req *lbservice.CreateInstanceRequest) (*lbservice.CreateInstanceResponse, error) {
	log.Printf("CreateService function was invoked with %v\n", req)

	lbID, err := addAdcToVdirect(req)
	if err != nil {
		log.Fatal("Error adding LB to vdirect", err)
	}

	res := &lbservice.CreateInstanceResponse{
		Id: &lbservice.InstanceId{
			InstanceId: lbID,
		},
	}

	if req.GetTestOnly() {
		return res, nil
	}
	err = uploadConfigTemplate("/workspace/radware/workflow_templates/create_lb.vm", "create_lb.vm")
	if err != nil {
		log.Fatal("Error upload create lb configuration", err)
	}
	err = uploadConfigTemplate("/workspace/radware/workflow_templates/add_reals.vm", "add_reals.vm")
	if err != nil {
		log.Fatal("Error upload add real server configuration", err)
	}
	err = uploadConfigTemplate("/workspace/radware/workflow_templates/read_reals.vm", "read_reals.vm")
	if err != nil {
		log.Fatal("Error upload read real server configuration", err)
	}
	err = uploadConfigTemplate("/workspace/radware/workflow_templates/delete_reals.vm", "delete_reals.vm")
	if err != nil {
		log.Fatal("Error upload delete real server configuration", err)
	}

	err = uploadConfigTemplate("/workspace/radware/workflow_templates/setup_l3.vm", "setup_l3.vm")
	if err != nil {
		log.Fatal("Error upload l3 setup configuration", err)
	}
	err = uploadConfigTemplate("/workspace/radware/workflow_templates/destroy_service.vm", "destroy_service.vm")
	if err != nil {
		log.Fatal("Error upload destroy service ", err)
	}
	err = uploadConfigTemplate("/workspace/radware/workflow_templates/setup_l4_filter.vm", "setup_l4_filter.vm")
	if err != nil {
		log.Fatal("Error upload L4 filter service ", err)
	}

	err = createAlteonLb(req)
	if err != nil {
		log.Fatal("Error create Alteon Lb", err)
	}
	// err = licAlteonLb(req)
	// if err != nil {
	// 	log.Fatal("Error to license Alteon Lb", err)
	// }

	return res, err
}

func (*server) DestroyService(ctx context.Context, req *lbservice.DestroyInstanceRequest) (*lbservice.DestroyInstanceResponse, error) {
	log.Printf("DestroyService function was invoked with %v\n", req)
	url := vdirectBaseUrl + "runnable/ConfigurationTemplate/destroy_service.vm/run"
	type DestroyConfigReq struct {
		DryRun bool `json:"__dryRun"`
		Alteon struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"alteon"`
		ServiceName string `json:"service_name"`
	}

	config := DestroyConfigReq{}
	config.DryRun = false
	config.Alteon.Type = "Container"
	config.Alteon.Name = req.GetLabel()
	config.ServiceName = req.GetLbServiceName()

	requestBody, err := json.Marshal(config)
	if err != nil {
		log.Fatalln(err)
	}
	res := &lbservice.DestroyInstanceResponse{
		DestroyInstanceResp: true,
	}

	if req.GetTestOnly() {
		return res, nil
	}

	c, uri := readResponce(requestBody, url, "POST", "application/json")
	count := 0
	for c != true && count < MAX_COUNT {
		time.Sleep(2 * time.Second)
		c, uri = readResponce(requestBody, uri, "GET", "")
		count++
	}
	if count == MAX_COUNT {
		err = fmt.Errorf("Timed out Destroying service")
		res = &lbservice.DestroyInstanceResponse{
			DestroyInstanceResp: false,
		}
	}

	if err != nil {
		log.Fatal("Error in Destroying service", err)
	}

	return res, err
}

func (*server) ConfigL3InterfacesService(ctx context.Context, req *lbservice.CfgL3InterfacesRequest) (*lbservice.CfgL3InterfacesResponse, error) {
	log.Printf("ConfigL3InterfacesService function was invoked with %v\n", req)
	err := configL3Network(req)
	res := &lbservice.CfgL3InterfacesResponse{
		CfgL3InterfacesResp: true,
	}
	if err != nil {
		log.Fatal("Error configure L3 network", err)
		res := &lbservice.CfgL3InterfacesResponse{
			CfgL3InterfacesResp: false,
		}

		return res, err
	}

	return res, err
}

func (*server) ConfigL4FilterService(ctx context.Context, req *lbservice.CfgL4FilterRequest) (*lbservice.CfgL4FilterResponse, error) {
	log.Printf("ConfigL4FilterService function was invoked with %v\n", req)
	url := vdirectBaseUrl + "runnable/ConfigurationTemplate/setup_l4_filter.vm/run"
	rules := req.GetFilt()
	var err error = nil
	res := &lbservice.CfgL4FilterResponse{
		CfgL4FilterResp: false,
	}

	type FiltConfigReq struct {
		DryRun bool `json:"__dryRun"`
		Alteon struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"alteon"`
		L4Filter struct {
			FilterName      string `json:"name"`
			FilterId        int32  `json:"id"`
			Action          string `json:"action"`
			IPVersion       string `json:"ip_version"`
			SrcIP           string `json:"src_ip_address"`
			SrcMask         string `json:"src_ip_mask"`
			DstIP           string `json:"dst_ip_address"`
			DstMask         string `json:"dst_ip_mask"`
			GroupName       string `json:"group"`
			Port            int32  `json:"port"`
			Reverse         string `json:"reverse"`
			ReturnToLastHop string `json:"returntolasthop"`
			Operation       string `json:"op"`
			Vlan            string `json:"vlan"`
			Protocol        string `json:"proto"`
		} `json:"filter"`
	}

	for _, cfg := range rules {
		config := FiltConfigReq{}
		config.DryRun = false
		config.Alteon.Type = "Container"
		config.Alteon.Name = cfg.GetLabel()
		config.L4Filter.FilterName = cfg.GetName()
		config.L4Filter.FilterId = cfg.GetRuleId()
		config.L4Filter.Action = cfg.GetAct()
		if cfg.GetLbIsV4() {
			config.L4Filter.IPVersion = "v4"
		} else {
			config.L4Filter.IPVersion = "v6"
		}
		config.L4Filter.SrcIP = cfg.GetSrcIp()
		config.L4Filter.SrcMask = cfg.GetSrcMask()
		config.L4Filter.DstIP = cfg.GetDstIp()
		config.L4Filter.DstMask = cfg.GetDstMask()
		config.L4Filter.GroupName = cfg.GetGroup()
		config.L4Filter.Port = cfg.GetPort()
		config.L4Filter.Vlan = cfg.GetVlan()
		config.L4Filter.Protocol = cfg.GetProtocol()
		if cfg.GetReverseSession() {
			config.L4Filter.Reverse = "enable"
		} else {
			config.L4Filter.Reverse = "disable"
		}
		if cfg.GetReturnToLastHop() {
			config.L4Filter.ReturnToLastHop = "enable"
		} else {
			config.L4Filter.ReturnToLastHop = "disable"
		}
		config.L4Filter.Operation = cfg.GetOp()
		requestBody, err := json.Marshal(config)
		if err != nil {
			log.Fatalln(err)
		}

		if req.GetTestOnly() {
			continue
		}

		c, uri := readResponce(requestBody, url, "POST", "application/json")
		count := 0
		for c != true && count < MAX_COUNT {
			time.Sleep(2 * time.Second)
			c, uri = readResponce(requestBody, uri, "GET", "")
			count++
		}
		if count == MAX_COUNT {
			err = fmt.Errorf("Timed out Configuring L4 filters")
			return res, err
		}
	}

	res = &lbservice.CfgL4FilterResponse{
		CfgL4FilterResp: true,
	}

	return res, err
}

func (*server) ProvisionEndPointService(ctx context.Context, req *lbservice.ProvisionEndPointRequest) (*lbservice.ProvisionEndPointResponse, error) {
	log.Printf("ProvisionEndPointService function was invoked with %v\n", req)
	var err error = nil
	url := ""
	res := &lbservice.ProvisionEndPointResponse{
		ProvisionEndPointResp: false,
	}
	type EpConfigReq struct {
		DryRun bool `json:"__dryRun"`
		Alteon struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"alteon"`
		VIRTNAME string `json:"VIRT_NAME"`
		Real     string `json:"real"`
		RealName string `json:"realname"`
	}

	epscfg := req.GetEp()

	for _, epcfg := range epscfg {
		op := epcfg.GetOp()
		if op == int32(lbservice.EndPointCfg_ADD) {
			// log.Println("Create ASAc EP", epcfg.GetIpAddress())
			url = vdirectBaseUrl + "runnable/ConfigurationTemplate/add_reals.vm/run"
		} else {
			// log.Println("Delete ASAc EP", epcfg.GetIpAddress())
			url = vdirectBaseUrl + "runnable/ConfigurationTemplate/delete_reals.vm/run"
		}

		config := EpConfigReq{}
		config.DryRun = false
		config.Alteon.Type = "Container"
		config.Alteon.Name = epcfg.GetLabel()
		config.Real = epcfg.GetIpAddress()
		config.RealName = epcfg.GetAsacInstanceName()
		config.VIRTNAME = epcfg.GetLbGroupName()

		requestBody, err := json.Marshal(config)
		if err != nil {
			log.Fatalln(err)
		}

		if req.GetTestOnly() {
			continue
		}
		c, uri := readResponce(requestBody, url, "POST", "application/json")
		count := 0
		for c != true && count < MAX_COUNT {
			time.Sleep(2 * time.Second)
			c, uri = readResponce(requestBody, uri, "GET", "")
			count++
		}
		if count == MAX_COUNT {
			err = fmt.Errorf("Timed out programming endpoints")
			return res, err
		}
	}

	res = &lbservice.ProvisionEndPointResponse{
		ProvisionEndPointResp: true,
	}
	return res, err
}

func (*server) QueryInstanceService(ctx context.Context, req *lbservice.QueryInstanceRequest) (*lbservice.QueryInstanceResponse, error) {
	log.Printf("QueryInstanceService function was invoked with %v\n", req)
	var err error = nil

	url := vdirectBaseUrl + "runnable/ConfigurationTemplate/read_reals.vm/run"
	type QueryReq struct {
		DryRun bool `json:"__dryRun"`
		Alteon struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"alteon"`
	}

	config := QueryReq{}
	config.DryRun = false
	config.Alteon.Type = "Container"
	config.Alteon.Name = req.GetLabel()

	requestBody, err := json.Marshal(config)
	if err != nil {
		log.Fatalln(err)
	}

	if req.GetTestOnly() {
		res := &lbservice.QueryInstanceResponse{
			QueryInstance: []*lbservice.EndPointInstance{
				{
					IpAddress:        "1.1.1.1",
					AsacInstanceName: "ASAc1",
				},
				{
					IpAddress:        "2.2.2.2",
					AsacInstanceName: "ASAc2",
				},
			},
		}
		return res, nil
	}
	c, uri := readResponce(requestBody, url, "POST", "application/json")
	count := 0
	for c != true && count < MAX_COUNT {
		time.Sleep(2 * time.Second)
		c, uri = readResponce(requestBody, uri, "GET", "")
		count++
	}

	if count == MAX_COUNT {
		log.Fatal("Timed out while waiting for query response")
	}

	s, err := readFullResponce(requestBody, uri)

	var qresp *lbservice.QueryInstanceResponse = new(lbservice.QueryInstanceResponse)
	var rspList *[]*lbservice.EndPointInstance = &qresp.QueryInstance

	for _, e := range s.Parameters.Output.LbGroup {
		*rspList = append(*rspList, &lbservice.EndPointInstance{
			IpAddress:        e.Asacip,
			AsacInstanceName: e.Asacname,
		})
	}

	log.Printf("Query response is %+v", qresp)
	return qresp, err
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}
	s := grpc.NewServer()

	lbservice.RegisterLoadBalancerServiceServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to listen to server %v", err)
	}
}
