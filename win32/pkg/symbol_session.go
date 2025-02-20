package pkg

import (
	"git.alpexglobal.vip/cy/go-mtmanapi/win32/mtmanapi"
)

type OpenDuration struct {
	OpenHour    int16 //开始的小时(24小时制)
	OpenMinute  int16 //开始的分钟
	CloseHour   int16 //结束的小时(24小时制)
	CloseMinute int16 //结束的分钟
}

type SymbolSessionInfo struct {
	Symbol         string
	Quote          [][]OpenDuration
	Trade          [][]OpenDuration
	WeekOpenMinute int16 // 一周开盘的当天时间(距离当天0点的minute分钟数)
}

// 获取symbol的交易时间/报价时间
// https://support.metaquotes.net/en/docs/mt4/api/reference_structures/structure_config/consessions
func GetSymbolSessions(singleSymbol mtmanapi.ConSymbol) SymbolSessionInfo {
	quoteDuration := make([][]OpenDuration, 7) // 星期几->多个时间段 (注意:0是周日),如果这一天没有则[]OpenDuration是空.
	tradeDuration := make([][]OpenDuration, 7) // 星期几->多个时间段

	var weekOpenHour int16 = -1   //每周开盘时间(h)
	var weekOpenMinute int16 = -1 //每周开盘时间(m)

	sessions := singleSymbol.GetSessions()
	for i := 0; i < 7; i++ {
		//0是周日, 1是周一......
		session := mtmanapi.ConSessionsArray_getitem(sessions, int64(i))

		quoteSessions := session.GetQuote()
		for j := 0; j < 3; j++ {
			//每天最多配置3段
			quoteSession := mtmanapi.ConSessionArray_getitem(quoteSessions, int64(j))
			if quoteSession.GetOpen_hour() == 0 && quoteSession.GetOpen_min() == 0 && quoteSession.GetClose_hour() == 0 && quoteSession.GetClose_min() == 0 {
				//说明不存在
				continue
			}
			quoteDuration[i] = append(quoteDuration[i], OpenDuration{
				quoteSession.GetOpen_hour(),
				quoteSession.GetOpen_min(),
				quoteSession.GetClose_hour(),
				quoteSession.GetClose_min(),
			})

			//取最开始的一段的开始时间
			if weekOpenHour == -1 || weekOpenMinute == -1 {
				weekOpenHour = quoteSession.GetOpen_hour()
				weekOpenMinute = quoteSession.GetOpen_min()
			}
		}

		tradeSessions := session.GetTrade()
		for j := 0; j < 3; j++ {
			tradeSession := mtmanapi.ConSessionArray_getitem(tradeSessions, int64(j))
			if tradeSession.GetOpen_hour() == 0 && tradeSession.GetOpen_min() == 0 && tradeSession.GetClose_hour() == 0 && tradeSession.GetClose_min() == 0 {
				//说明不存在
				continue
			}
			tradeDuration[i] = append(tradeDuration[i], OpenDuration{
				tradeSession.GetOpen_hour(),
				tradeSession.GetOpen_min(),
				tradeSession.GetClose_hour(),
				tradeSession.GetClose_min(),
			})
		}
	}
	weekOpenTime := weekOpenHour*60 + weekOpenMinute //是周开盘时间(影响四小时kline的累计)

	return SymbolSessionInfo{
		singleSymbol.GetSymbol(),
		quoteDuration,
		tradeDuration,
		weekOpenTime,
	}
}

//--------------------------------------------------------

type HolidayInfo struct {
	Symbol string //symbol
	Year   int    //年份 (如果是每一年, 则该值为0)
	Month  int    //月
	Day    int    //日

	FromMinute int //开始的时间(0点0分到现在的minute数)
	ToMinute   int //结束的时间(0点0分到现在的minute数)

	Enable int //规则是否生效
}

// 获取symbol的节假日时间段安排
// https://support.metaquotes.net/en/docs/mt4/api/manager_api/manager_api_config/manager_api_config_holiday/cmanagerinterface_cfgrequestholiday
func GetConHoliday(singleHoliday mtmanapi.ConHoliday) HolidayInfo {
	return HolidayInfo{
		singleHoliday.GetSymbol(),
		singleHoliday.GetYear(),
		singleHoliday.GetMonth(),
		singleHoliday.GetDay(),
		singleHoliday.GetFrom(),
		singleHoliday.GetTo(),
		singleHoliday.GetEnable(),
	}
}

// 返回各个symbol的holiday列表(manager 必须是direct manager)
func GetAllConHolidays(manager mtmanapi.CManagerInterface) map[string][]HolidayInfo {
	//获取所有的Holiday
	holidayMap := make(map[string][]HolidayInfo) //一个symbol可能有多个holiday
	var hTotal int
	hs := manager.CfgRequestHoliday(&hTotal) //获取holiday列表
	for i := 0; i < hTotal; i++ {
		singleHoliday := mtmanapi.ConHolidayArray_getitem(hs, int64(i))
		symbol := singleHoliday.GetSymbol()
		//初始化一下
		if _, ok := holidayMap[symbol]; !ok {
			holidayMap[symbol] = make([]HolidayInfo, 0)
		}
		holidayMap[symbol] = append(holidayMap[symbol], GetConHoliday(singleHoliday))
		mtmanapi.Delete_ConHolidayArray(singleHoliday)
	}
	//https://support.metaquotes.net/en/docs/mt4/api/manager_api/manager_api_config/manager_api_config_holiday/cmanagerinterface_cfgrequestholiday
	manager.MemFree(hs.Swigcptr())

	return holidayMap
}
