package pkg

import (
	"errors"
	"fmt"
	"git.alpexglobal.vip/cy/go-mtmanapi/win64/mtmanapi"
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
	managerPump.UserRecordGet(trade.GetLogin(), userInfo)
	group := userInfo.GetGroup()

	return GetGroupSpreadDiffBySymbol(managerPump, group, symbol)
}

//-----------------------------------------------------------

type GroupSpreadValue struct {
	Bid    decimal.Decimal //bid侧组点值
	Ask    decimal.Decimal //ask侧组点值
	Symbol mtmanapi.SymbolInfo
	Group  mtmanapi.ConGroup
}

// 获取组点(only can be used in pumping mode)
func GetGroupSpreadDiffBySymbol(managerPump mtmanapi.CManagerInterface, group string, symbol string) (*GroupSpreadValue, error) {
	//增加组点
	symbolInfo := mtmanapi.NewSymbolInfo()
	code := managerPump.SymbolInfoGet(symbol, symbolInfo)
	if code != mtmanapi.RET_OK {
		managerPump.SymbolAdd(symbol)
		return nil, errors.New(fmt.Sprintf("SymbolGet err, symbol:%s, errCode:%d", symbol, code))
	}

	groupInfo := mtmanapi.NewConGroup()
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
		symbolInfo,
		groupInfo,
	}, nil
}

// ----------------------------------
type ManagerMode int

const (
	ManagerDirect  ManagerMode = 1
	ManagerPumping ManagerMode = 2
	ManagerDealing ManagerMode = 3
)

func GetAllGroups(mode ManagerMode, manager mtmanapi.CManagerInterface) map[string]mtmanapi.ConGroup {
	result := make(map[string]mtmanapi.ConGroup)
	totalNum := 0
	var groups mtmanapi.ConGroup
	if mode == ManagerDirect {
		groups = manager.GroupsRequest(&totalNum)
	} else if mode == ManagerPumping {
		groups = manager.GroupsGet(&totalNum)
	}
	for i := 0; i < totalNum; i++ {
		singleGroup := mtmanapi.ConGroupArray_getitem(groups, int64(i))
		result[singleGroup.GetGroup()] = singleGroup
	}
	return result
}
