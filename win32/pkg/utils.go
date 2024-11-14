package pkg

import (
	"errors"
	"fmt"
	"github.com/asaka1234/go-mtmanapi/win32/mtmanapi"
	"github.com/shopspring/decimal"
	"math"
)

// 获取32/64位
func GetSysVersion() int {
	const intSize = 32 << (^uint(0) >> 63)
	return intSize
}

// 获取api version
func GetApiVersion() int {
	//(a << 16) | b
	return (mtmanapi.ManAPIProgramVersion << 16) | mtmanapi.ManAPIProgramBuild
}

// 获取组点
func GetGroupSpreadDiffByRequest(managerPump mtmanapi.CManagerInterface, req mtmanapi.RequestInfo) (*GroupSpreadValue, error) {
	symbol := req.GetTrade().GetSymbol()
	group := req.GetGroup()
	return GetGroupSpreadDiffBySymbol(managerPump, group, symbol)
}

// 获取组点
func GetGroupSpreadDiffByTrade(managerPump mtmanapi.CManagerInterface, trade mtmanapi.TradeRecord) (*GroupSpreadValue, error) {
	symbol := trade.GetSymbol()

	userInfo := mtmanapi.NewUserRecord()
	defer mtmanapi.DeleteUserRecord(userInfo)
	managerPump.UserRecordGet(trade.GetLogin(), userInfo)
	group := userInfo.GetGroup()

	return GetGroupSpreadDiffBySymbol(managerPump, group, symbol)
}

//-----------------------------------------------------------

type GroupSpreadValue struct {
	Bid decimal.Decimal //bid侧组点值
	Ask decimal.Decimal //ask侧组点值
}

// 获取组点(only can be used in pumping mode)
func GetGroupSpreadDiffBySymbol(managerPump mtmanapi.CManagerInterface, group string, symbol string) (*GroupSpreadValue, error) {
	//增加组点
	symbolInfo := mtmanapi.NewSymbolInfo()
	defer mtmanapi.DeleteSymbolInfo(symbolInfo)
	code := managerPump.SymbolInfoGet(symbol, symbolInfo)
	if code != mtmanapi.RET_OK {
		managerPump.SymbolAdd(symbol)
		return nil, errors.New(fmt.Sprintf("SymbolGet err, symbol:%s, errCode:%d", symbol, code))
	}

	groupInfo := mtmanapi.NewConGroup()
	defer mtmanapi.DeleteConGroup(groupInfo)
	code = managerPump.GroupRecordGet(group, groupInfo)
	if code != mtmanapi.RET_OK {
		return nil, errors.New(fmt.Sprintf("GroupRecordGet err, symbol:%s, errCode:%d", symbol, code))
	}

	return GetGroupSpreadDiff(groupInfo, symbolInfo)
}

// 获取组点(only can be used in pumping mode)
func GetGroupSpreadDiff(groupInfo mtmanapi.ConGroup, symbolInfo mtmanapi.SymbolInfo) (*GroupSpreadValue, error) {
	//增加组点
	xtype := symbolInfo.GetXtype()
	digit := symbolInfo.GetDigits()

	secGroups := groupInfo.GetSecgroups()
	singleGroup := mtmanapi.ConGroupSecArray_getitem(secGroups, int64(xtype))
	spreadDiff := singleGroup.GetSpread_diff() //获取组点

	//数量
	spreadBid := spreadDiff / 2
	spreadAsk := spreadDiff - spreadBid

	//基本单位
	denominator := math.Pow(0.1, float64(digit))

	//两个方向各自的组点值
	bidVal := decimal.NewFromInt(int64(spreadBid)).Mul(decimal.NewFromFloat(denominator))
	askVal := decimal.NewFromInt(int64(spreadAsk)).Mul(decimal.NewFromFloat(denominator))

	return &GroupSpreadValue{
		bidVal,
		askVal,
	}, nil
}

func GetGroupSpreadDiff2(spreadDiff int, symbolInfo mtmanapi.SymbolInfo) (*GroupSpreadValue, error) {
	//增加组点
	digit := symbolInfo.GetDigits()

	//数量
	spreadBid := spreadDiff / 2
	spreadAsk := spreadDiff - spreadBid

	//基本单位
	denominator := math.Pow(0.1, float64(digit))

	//两个方向各自的组点值
	bidVal := decimal.NewFromInt(int64(spreadBid)).Mul(decimal.NewFromFloat(denominator))
	askVal := decimal.NewFromInt(int64(spreadAsk)).Mul(decimal.NewFromFloat(denominator))

	return &GroupSpreadValue{
		bidVal,
		askVal,
	}, nil
}

type SymbolBase struct {
	XType  int //type索引
	Digit  int //精度
	Symbol string
}

func GetGroupSpreadDiff3(groupInfo mtmanapi.ConGroup, symbolInfo SymbolBase) (*GroupSpreadValue, error) {
	//增加组点
	xtype := symbolInfo.XType
	digit := symbolInfo.Digit

	secGroups := groupInfo.GetSecgroups()
	singleGroup := mtmanapi.ConGroupSecArray_getitem(secGroups, int64(xtype))
	spreadDiff := singleGroup.GetSpread_diff() //获取组点

	//数量
	spreadBid := spreadDiff / 2
	spreadAsk := spreadDiff - spreadBid

	//基本单位
	denominator := math.Pow(0.1, float64(digit))

	//两个方向各自的组点值
	bidVal := decimal.NewFromInt(int64(spreadBid)).Mul(decimal.NewFromFloat(denominator))
	askVal := decimal.NewFromInt(int64(spreadAsk)).Mul(decimal.NewFromFloat(denominator))

	return &GroupSpreadValue{
		bidVal,
		askVal,
	}, nil
}

func GetGroupSpreadDiff4(gSecList []int, symbolInfo mtmanapi.SymbolInfo) (*GroupSpreadValue, error) {
	//增加组点
	xtype := symbolInfo.GetXtype()
	digit := symbolInfo.GetDigits()

	spreadDiff := gSecList[xtype]

	//数量
	spreadBid := spreadDiff / 2
	spreadAsk := spreadDiff - spreadBid

	//基本单位
	denominator := math.Pow(0.1, float64(digit))

	//两个方向各自的组点值
	bidVal := decimal.NewFromInt(int64(spreadBid)).Mul(decimal.NewFromFloat(denominator))
	askVal := decimal.NewFromInt(int64(spreadAsk)).Mul(decimal.NewFromFloat(denominator))

	return &GroupSpreadValue{
		bidVal,
		askVal,
	}, nil
}

func GetGroupSpreadDiff5(gSecList []int, symbolInfo SymbolBase) (*GroupSpreadValue, error) {
	//增加组点
	xtype := symbolInfo.XType
	digit := symbolInfo.Digit

	spreadDiff := gSecList[xtype]

	//数量
	spreadBid := spreadDiff / 2
	spreadAsk := spreadDiff - spreadBid

	//基本单位
	denominator := math.Pow(0.1, float64(digit))

	//两个方向各自的组点值
	bidVal := decimal.NewFromInt(int64(spreadBid)).Mul(decimal.NewFromFloat(denominator))
	askVal := decimal.NewFromInt(int64(spreadAsk)).Mul(decimal.NewFromFloat(denominator))

	return &GroupSpreadValue{
		bidVal,
		askVal,
	}, nil
}

// ----------------------------------
type ManagerMode int

const (
	ManagerDirect  ManagerMode = 1
	ManagerPumping ManagerMode = 2
	ManagerDealing ManagerMode = 3
)

// 获取所有group的spread_diff
func GetAllGroupSpreadDiff(mode ManagerMode, manager mtmanapi.CManagerInterface) map[string][]int {
	result := make(map[string][]int)
	totalNum := 0
	var groups mtmanapi.ConGroup
	if mode == ManagerDirect {
		groups = manager.GroupsRequest(&totalNum)
		//https://support.metaquotes.net/en/docs/mt4/api/manager_api/manager_api_config/manager_api_config_group/cmanagerinterface_groupsrequest
		defer manager.MemFree(groups.Swigcptr())
	} else if mode == ManagerPumping {
		groups = manager.GroupsGet(&totalNum)
	}
	for i := 0; i < totalNum; i++ {
		groupInfo := mtmanapi.ConGroupArray_getitem(groups, int64(i))
		defer mtmanapi.DeleteConGroup(groupInfo)

		secList := make([]int, mtmanapi.MAX_SEC_GROUPS)
		for j := 0; j < mtmanapi.MAX_SEC_GROUPS; j++ {
			secItem := mtmanapi.ConGroupSecArray_getitem(groupInfo.GetSecgroups(), int64(j))
			defer mtmanapi.DeleteConGroupSec(secItem)
			secList[j] = secItem.GetSpread_diff()
		}
		result[groupInfo.GetGroup()] = secList
	}
	return result
}

// 获取指定group的spread_diff
func GetGroupSpreadDiffRecord(group mtmanapi.ConGroup) []int {
	secList := make([]int, mtmanapi.MAX_SEC_GROUPS)
	for j := 0; j < mtmanapi.MAX_SEC_GROUPS; j++ {
		secItem := mtmanapi.ConGroupSecArray_getitem(group.GetSecgroups(), int64(j))
		defer mtmanapi.DeleteConGroupSec(secItem)
		secList[j] = secItem.GetSpread_diff()
	}
	return secList
}

func GetAllGroups(mode ManagerMode, manager mtmanapi.CManagerInterface) map[string]mtmanapi.ConGroup {
	result := make(map[string]mtmanapi.ConGroup)
	totalNum := 0
	var groups mtmanapi.ConGroup
	if mode == ManagerDirect {
		groups = manager.GroupsRequest(&totalNum)
		//https://support.metaquotes.net/en/docs/mt4/api/manager_api/manager_api_config/manager_api_config_group/cmanagerinterface_groupsrequest
		defer manager.MemFree(groups.Swigcptr())
	} else if mode == ManagerPumping {
		groups = manager.GroupsGet(&totalNum)
	}
	for i := 0; i < totalNum; i++ {
		singleGroup := mtmanapi.ConGroupArray_getitem(groups, int64(i))
		result[singleGroup.GetGroup()] = singleGroup
	}
	return result
}

func GetAllSymbols(mode ManagerMode, manager mtmanapi.CManagerInterface) map[string]mtmanapi.ConSymbol {
	result := make(map[string]mtmanapi.ConSymbol)
	totalNum := 0
	var symbols mtmanapi.ConSymbol
	if mode == ManagerDirect {
		symbols = manager.CfgRequestSymbol(&totalNum)
		//https://support.metaquotes.net/en/docs/mt4/api/manager_api/manager_api_config/manager_api_config_symbol/cmanagerinterface_cfgrequestsymbol
		defer manager.MemFree(symbols.Swigcptr())
	} else if mode == ManagerPumping {
		symbols = manager.SymbolsGetAll(&totalNum)
	}
	for i := 0; i < totalNum; i++ {
		singleSymbol := mtmanapi.ConSymbolArray_getitem(symbols, int64(i))
		result[singleSymbol.GetSymbol()] = singleSymbol
	}
	return result
}
