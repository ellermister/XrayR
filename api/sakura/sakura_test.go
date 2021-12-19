package sakura_test

import (
	"fmt"
	"github.com/XrayR-project/XrayR/api"
	"github.com/XrayR-project/XrayR/api/sakura"
	"testing"
)

func CreateClient() api.API {
	apiConfig := &api.Config{
		APIHost:  "http://127.0.0.1:8000",
		Key:      "qwertyuiopasdfghjkl",
		NodeID:   1,
		NodeType: "V2ray",
	}
	client := sakura.New(apiConfig)
	return client
}

func TestGetV2rayNodeInfo(t *testing.T) {
	client := CreateClient()
	nodeInfo, err := client.GetNodeInfo()
	if err != nil {
		t.Error(err)
	}
	t.Log(nodeInfo)
}

func TestGetUserList(t *testing.T) {
	client := CreateClient()

	userList, err := client.GetUserList()
	if err != nil {
		t.Error(err)
	}

	t.Log(userList)
}

func TestReportNodeStatus(t *testing.T) {
	client := CreateClient()
	nodeStatus := &api.NodeStatus{
		1, 1, 1, 256,
	}
	err := client.ReportNodeStatus(nodeStatus)
	if err != nil {
		t.Error(err)
	}
}

func TestReportNodeOnlineUsers(t *testing.T) {
	client := CreateClient()
	userList, err := client.GetUserList()
	if err != nil {
		t.Error(err)
	}

	onlineUserList := make([]api.OnlineUser, len(*userList))
	for i, userInfo := range *userList {
		onlineUserList[i] = api.OnlineUser{
			UID: userInfo.UID,
			IP:  fmt.Sprintf("1.1.1.%d", i),
		}
	}
	//client.Debug()
	err = client.ReportNodeOnlineUsers(&onlineUserList)
	if err != nil {
		t.Error(err)
	}
}

func TestReportUserTraffic(t *testing.T) {
	client := CreateClient()
	userList, err := client.GetUserList()
	if err != nil {
		t.Error(err)
	}
	generalUserTraffic := make([]api.UserTraffic, len(*userList))
	for i, userInfo := range *userList {
		generalUserTraffic[i] = api.UserTraffic{
			UID:      userInfo.UID,
			Upload:   114514,
			Download: 114514,
		}
	}
	//client.Debug()
	err = client.ReportUserTraffic(&generalUserTraffic)
	if err != nil {
		t.Error(err)
	}
}

func TestGetNodeRule(t *testing.T) {
	client := CreateClient()

	ruleList, err := client.GetNodeRule()
	if err != nil {
		t.Error(err)
	}

	t.Log(ruleList)
}

func TestReportIllegal(t *testing.T) {
	client := CreateClient()

	detectResult := []api.DetectResult{
		{UID: 1, RuleID: 2},
		{UID: 1, RuleID: 3},
	}
	err := client.ReportIllegal(&detectResult)
	if err != nil {
		t.Error(err)
	}
}
