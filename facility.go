package facility

import (
	
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	xir "gitlab.com/mergetb/xir/v0.3/go"
	. "gitlab.com/mergetb/xir/v0.3/go/build"
)

type Chassis struct {
	NodeId string `json:"nodeId"`
	NodeIdType string `json:"nodeIdType"`
	SysName string `json:"sysName"`
	SysDescription string `json:"sysDescription"`
	MgmtIp string `json:"mgmtIp"`
	Capability struct {
		Bridge bool `json:"Bridge"`
		Router bool `json:"Router"`
		Wlan bool `json:"Wlan"`
		Station bool `json:"Station"`
	}
}

type Neighbor struct {
	Interface string `json:"interface"`
	NeighborId string `json:"neighborId"`
	NeighborIdType string `json:"neighborIdType"`
	Name string `json:"name"`
	MgmtIp string `json:"mgmtIp"`
	PortIdType string `json:"portIdType"`
	PortId string `json:"portId"`
	PortDescription string `json:"portDescription"`
	PortTtl string `json:"portTtl"`
	Capability struct {
		Bridge bool `json:"Bridge"`
		Router bool `json:"Router"`
		Wlan bool `json:"Wlan"`
		Station bool `json:"Station"`
	}
}

type LldpData struct {
	Chassis Chassis
	Neighbor []Neighbor 
}

type ConnectivityInfo struct {
	Data []LldpData
}

var (
	ifr    *xir.Resource
	nodes []*xir.Resource
	ixp    *xir.Resource
	ops    *xir.Resource
	mgmt   *xir.Resource
)
	

func CreateFacility() *xir.Facility {

	lldpresponse, err := GetFacilityData()

	if err != nil {
        log.Fatalf("Error getting API response: %v", err)
    }

	// fmt.Println(lldpresponse)
	for _, x := range lldpresponse.Data {
		fmt.Print(x.Chassis)
		// fmt.Println(x.Neighbor[0])
	}

	tb, err := NewBuilder("mergetestbed", "mergetestbed")
	if err != nil {
		log.Fatalf("new builder: %v", err)
	}


	// Ops server (Mergeops)
	ops = tb.OpsServer("ops",
		Eth(1, Gbps(1)),
		Eno(1, Gbps(1)),
		Enp(4, 1, Gbps(1), Mgmt(2), Product("Intel Corporation", "I350 Gigabit Network Connection", "X710")),
		Product("Unknown", "Test", "Mergeops"),
	)
	ops.Mgmt().Mac = "a0:36:9f:02:58:03"

	// switches
	mgmt = tb.MgmtLeaf("mgmt",
		Eth(1, Gbps(1), Mgmt()),
		Swp(8, Gbps(1), Mgmt()),
		Product("Unknown", "test bridge", "Bridge"),
	)
	
	// Infrastructure Server
	ifr = tb.Infraserver("ifr",
		Eno(2, Mgmt(1)),
		Enp(4,1),
		Product("Unknown", "Test 1", "IFR"),
	)

  	ifr.Mgmt().Mac = "00:1e:67:19:58:89"
	ifr.NICs[1].Ports[0].Name = "ens261f0"
	ifr.NICs[1].Ports[0].Mac = "3c:fd:fe:9e:f0:28"
	ifr.NICs[1].Ports[1].Name = "ens261f1"
	ifr.NICs[1].Ports[1].Mac = "3c:fd:fe:9e:f0:29"
	ifr.NICs[1].Ports[2].Name = "ens261f2"
	ifr.NICs[1].Ports[2].Mac = "3c:fd:fe:9e:f0:2a"
	ifr.NICs[1].Ports[3].Name = "ens261f3"
	ifr.NICs[1].Ports[3].Mac = "3c:fd:fe:9e:f0:2b"

  	// IXP switches
  	ixp = tb.InfraLeaf("ixp",
		Eth(1, Gbps(1), Mgmt()),
    	Swp(8, Gbps(100), Infranet()),
		Swp(8, Gbps(100), Xpnet()),
    	Product("Unknown", "Test 2", "IXP"),
	)
	ixp.Mgmt().Mac = "0c:42:a1:36:b3:e8"

	//nodes
	nodes = tb.Nodes(2, "x",
		Ipmi(1, Gbps(1)),
		Eno(2, Gbps(1), Mgmt(0)),
		Enp(1, 1, Gbps(50), Infranet(0)), // connect-x5
		Enp(2, 1, Gbps(100), Xpnet()), // connect-x6
	)

	nodes[0].NICs[0].Ports[0].Mac = "b4:2e:99:f7:cb:6c"
	nodes[0].Mgmt().Mac = "b4:2e:99:f7:cb:6a"
	nodes[0].Infranet().Mac = "0c:42:a1:42:b0:26"
	nodes[0].Infranet().Name = "enp66s0np0"
	nodes[0].NICs[3].Ports[0].Mac = "0c:42:a1:2c:e4:ba"
	nodes[0].NICs[3].Ports[1].Mac = "0c:42:a1:2c:e4:bb"

	nodes[1].NICs[0].Ports[0].Mac = "b4:2e:99:f7:cb:70"
	nodes[1].Mgmt().Mac = "b4:2e:99:f7:cb:6e"
	nodes[1].Infranet().Mac = "0c:42:a1:42:b0:1a"
	nodes[1].Infranet().Name = "enp66s0np0"
	nodes[1].NICs[3].Ports[0].Mac = "0c:42:a1:98:55:20"
	nodes[1].NICs[3].Ports[1].Mac = "0c:42:a1:98:55:21"

	// mgmt
	tb.Connect(ixp.Mgmt(), mgmt.NextSwpG(1))
	tb.Connect(ops.Mgmt(), mgmt.NextSwpG(1))
	tb.Connect(ifr.Mgmt(), mgmt.NextSwpG(1))
	tb.Connect(nodes[0].Mgmt(), mgmt.NextSwpG(1))
	tb.Connect(nodes[1].Mgmt(), mgmt.NextSwpG(1))

  	// infra
	tb.BreakoutTrunk(
		ixp.NICs[1],
		ixp.SwpIndex(1),
		ifr.NICs[1].Ports,
		"ifr",
		"ixp",
		DACBreakout(),
	)

	// DAC 100G
	tb.Connect(nodes[0].Infranet(), ixp.SwpIndex(7))
	tb.Connect(nodes[1].Infranet(), ixp.SwpIndex(8))

	// xp
	// DAC 40 Gbps
	tb.Connect(nodes[0].NICs[3].Ports[0], ixp.SwpIndex(9))
	tb.Connect(nodes[0].NICs[3].Ports[1], ixp.SwpIndex(10))
	tb.Connect(nodes[1].NICs[3].Ports[0], ixp.SwpIndex(11))
	tb.Connect(nodes[1].NICs[3].Ports[1], ixp.SwpIndex(12))

  return tb.Facility()
}


func GetFacilityData() (*ConnectivityInfo, error) {
	//define the API endpoint
	url := "http://localhost:9000/lldp-data"

	// HTTP request
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make the request: %v", err)
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Unmarshall the JSON response
	var apiresponse ConnectivityInfo
	err = json.Unmarshal(body, &apiresponse)
	// fmt.Print(body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall JSON response: %v", err)
	}

	return &apiresponse, nil
}

