package sakura

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/XrayR-project/XrayR/api"
	"github.com/bitly/go-simplejson"
	"github.com/go-resty/resty/v2"
	"log"
	"os"
	"strconv"
	"time"
)

// APIClient create a api client to the panel.
type APIClient struct {
	client        *resty.Client
	APIHost       string
	NodeID        int
	Key           string
	NodeType      string
	EnableVless   bool
	EnableXTLS    bool
	SpeedLimit    float64
	DeviceLimit   int
	LocalRuleList []api.DetectRule
}

// New creat a api instance
func New(apiConfig *api.Config) *APIClient {

	client := resty.New()
	client.SetRetryCount(3)
	if apiConfig.Timeout > 0 {
		client.SetTimeout(time.Duration(apiConfig.Timeout) * time.Second)
	} else {
		client.SetTimeout(5 * time.Second)
	}
	client.OnError(func(req *resty.Request, err error) {
		if v, ok := err.(*resty.ResponseError); ok {
			// v.Response contains the last response from the server
			// v.Err contains the original error
			log.Print(v.Err)
		}
	})
	client.SetHostURL(apiConfig.APIHost)
	// Create Key for each requests
	client.SetHeaders(map[string]string{
		"key": apiConfig.Key,
	})
	// Read local rule list
	localRuleList := readLocalRuleList(apiConfig.RuleListPath)
	apiClient := &APIClient{
		client:        client,
		NodeID:        apiConfig.NodeID,
		Key:           apiConfig.Key,
		APIHost:       apiConfig.APIHost,
		NodeType:      apiConfig.NodeType,
		EnableVless:   apiConfig.EnableVless,
		EnableXTLS:    apiConfig.EnableXTLS,
		SpeedLimit:    apiConfig.SpeedLimit,
		DeviceLimit:   apiConfig.DeviceLimit,
		LocalRuleList: localRuleList,
	}
	return apiClient
}

// readLocalRuleList reads the local rule list file
func readLocalRuleList(path string) (LocalRuleList []api.DetectRule) {

	LocalRuleList = make([]api.DetectRule, 0)
	if path != "" {
		// open the file
		file, err := os.Open(path)

		//handle errors while opening
		if err != nil {
			log.Printf("Error when opening file: %s", err)
			return LocalRuleList
		}

		fileScanner := bufio.NewScanner(file)

		// read line by line
		for fileScanner.Scan() {
			LocalRuleList = append(LocalRuleList, api.DetectRule{
				ID:      -1,
				Pattern: fileScanner.Text(),
			})
		}
		// handle first encountered error while reading
		if err := fileScanner.Err(); err != nil {
			log.Fatalf("Error while reading file: %s", err)
			return make([]api.DetectRule, 0)
		}

		_ = file.Close()
	}

	return LocalRuleList
}

func (c *APIClient) GetNodeInfo() (nodeInfo *api.NodeInfo, err error) {
	var header json.RawMessage
	path := "/api/xray_r/node_info"
	res, err := c.client.R().
		ForceContentType("application/json").
		Get(path)
	response, err := c.parseResponse(res, path, err)
	if err != nil {
		return nil, err
	}

	Port, _ := response.Get("datas").Get("port").Int()
	AlterId, _ := response.Get("datas").Get("alter_id").Int()
	TransportProtocol, _ := response.Get("datas").Get("transport_protocol").String()
	EnableTLS, _ := response.Get("datas").Get("enable_tls").Bool()
	TlsType, _ := response.Get("datas").Get("tls_type").String()
	Path, _ := response.Get("datas").Get("path").String()
	Host, _ := response.Get("datas").Get("host").String()
	SpeedLimit, _ := response.Get("datas").Get("speed_limit").Uint64()
	ServiceName, _ := response.Get("datas").Get("service_name").String()
	if data, ok := response.Get("datas").CheckGet("header"); ok {
		if httpHeader, err := data.MarshalJSON(); err != nil {
			return nil, err
		} else {
			header = httpHeader
		}
	}

	nodeInfo = &api.NodeInfo{
		NodeType:          c.NodeType,
		NodeID:            c.NodeID,
		Port:              Port,
		SpeedLimit:        SpeedLimit,
		AlterID:           AlterId,
		TransportProtocol: TransportProtocol,
		Host:              Host,
		Path:              Path,
		EnableTLS:         EnableTLS,
		TLSType:           TlsType,
		EnableVless:       c.EnableVless,
		ServiceName:       ServiceName,
		Header:            header,
	}
	return nodeInfo, nil
}

//func (c APIClient) GetUserList() (userList *[]api.UserInfo, err error) {
func (c *APIClient) GetUserList() (UserList *[]api.UserInfo, err error) {
	path := "/api/xray_r/user_list"
	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.NodeID)).
		ForceContentType("application/json").
		Get(path)

	response, err := c.parseResponse(res, path, err)
	if err != nil {
		return nil, err
	}
	numOfUsers := len(response.Get("datas").Get("user_list").MustArray())
	userList := make([]api.UserInfo, numOfUsers)
	for i := 0; i < numOfUsers; i++ {
		user := api.UserInfo{}
		user.UID = response.Get("datas").Get("user_list").GetIndex(i).Get("port").MustInt()
		user.SpeedLimit = uint64(c.SpeedLimit * 1000000 / 8)
		user.DeviceLimit = c.DeviceLimit
		// v2ray
		user.UUID = response.Get("datas").Get("user_list").GetIndex(i).Get("pass").MustString()
		user.Email = response.Get("datas").Get("user_list").GetIndex(i).Get("port").MustString()
		user.AlterID = response.Get("datas").Get("alter_id").MustInt()

		userList[i] = user
	}
	return &userList, nil
}

func (c *APIClient) ReportNodeStatus(nodeStatus *api.NodeStatus) (err error) {
	path := "/api/xray_r/report_node_status"
	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.NodeID)).
		ForceContentType("application/json").
		SetBody(nodeStatus).
		Post(path)

	_, err = c.parseResponse(res, path, err)
	if err != nil {
		return err
	}
	return nil
}

func (c *APIClient) ReportNodeOnlineUsers(onlineUser *[]api.OnlineUser) (err error) {
	path := "/api/xray_r/report_online_user"
	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.NodeID)).
		ForceContentType("application/json").
		SetBody(onlineUser).
		Post(path)

	_, err = c.parseResponse(res, path, err)
	if err != nil {
		return err
	}
	return nil
}

func (c *APIClient) ReportUserTraffic(userTraffic *[]api.UserTraffic) (err error) {
	path := "/api/xray_r/report_user_traffic"
	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.NodeID)).
		ForceContentType("application/json").
		SetBody(userTraffic).
		Post(path)

	_, err = c.parseResponse(res, path, err)
	if err != nil {
		return err
	}
	return nil
}

// Describe return a description of the client
func (c *APIClient) Describe() api.ClientInfo {
	return api.ClientInfo{APIHost: c.APIHost, NodeID: c.NodeID, Key: c.Key, NodeType: c.NodeType}
}

// Debug set the client debug for client
func (c *APIClient) Debug() {
	c.client.SetDebug(true)
}

func (c *APIClient) GetNodeRule() (*[]api.DetectRule, error) {
	ruleList := c.LocalRuleList

	// V2board only support the rule for v2ray
	path := "/api/xray_r/node_rule"
	res, err := c.client.R().
		ForceContentType("application/json").
		Get(path)

	response, err := c.parseResponse(res, path, err)
	if err != nil {
		return nil, err
	}
	ruleListResponse := response.Get("datas").Get("rules").MustStringArray()
	for i, rule := range ruleListResponse {
		ruleListItem := api.DetectRule{
			ID:      i,
			Pattern: rule,
		}
		ruleList = append(ruleList, ruleListItem)
	}
	return &ruleList, nil
}

func (c *APIClient) ReportIllegal(detectResultList *[]api.DetectResult) (err error) {

	data := make([]IllegalItem, len(*detectResultList))
	for i, r := range *detectResultList {
		data[i] = IllegalItem{
			ID:  r.RuleID,
			UID: r.UID,
		}
	}
	path := "/api/xray_r/report_illegal"
	res, err := c.client.R().
		SetQueryParam("node_id", strconv.Itoa(c.NodeID)).
		SetBody(&data).
		ForceContentType("application/json").
		Post(path)
	_, err = c.parseResponse(res, path, err)
	if err != nil {
		return err
	}
	return nil
}

func (c *APIClient) assembleURL(path string) string {
	return c.APIHost + path
}

func (c *APIClient) parseResponse(res *resty.Response, path string, err error) (*simplejson.Json, error) {
	if err != nil {
		return nil, fmt.Errorf("request %s failed: %s", c.assembleURL(path), err)
	}

	if res.StatusCode() > 400 {
		body := res.Body()
		return nil, fmt.Errorf("request %s failed: %s, %s", c.assembleURL(path), string(body), err)
	}
	rtn, err := simplejson.NewJson(res.Body())

	if err != nil {
		return nil, fmt.Errorf("Ret %s invalid", res.String())
	}

	responseCode, _ := rtn.Get("response").Get("code").Int()
	if responseCode != 200 {
		errorMessage := rtn.Get("response").Get("message")
		body := res.Body()
		return nil, fmt.Errorf("request %s failed: %s, error:%s %s", c.assembleURL(path), errorMessage, string(body), err)
	}
	return rtn, nil
}
