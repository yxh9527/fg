import axios from '@/libs/api.request'
import qs from 'qs'
import {
  getToken
} from '@/libs/util'

function parmaBasic() {
  let parmaBasic = {}
  if (sessionStorage.getItem("agentVal")) {
    parmaBasic = {
      webId: sessionStorage.getItem("siteVal"),
      agentId: sessionStorage.getItem("agentVal")
    }
  } else if (window.windowData_getLinkageList) {
    parmaBasic = {
      webId: window.windowData_getLinkageList[0].id,
      agentId: window.windowData_getLinkageList[0].agentList[0].id
    }
  }
  return parmaBasic
}

export const getHomeData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v1/report-form/index",
    method: 'get',
    params: Data
  })
}

export const getHomeGameData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v1/report-form/game",
    method: 'get',
    params: Data
  })
}

export const getHomeGameAvgData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data, parmaBasic())
  }
  return axios.request({
    url: "v1/report-form/gameAvg",
    method: 'get',
    params: Data
  })
}

export const getReportData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v2/report-form/listWithAgent",
    method: 'get',
    params: Data
  })
}


export const exportAgentData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v2/export/agent/data",
    method: 'get',
    params: Data
  })
}

//获取代理游戏统计数据
export const getAgentGameDataAggs = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v2/report-form/listAgentGameAggs",
    method: 'get',
    params: Data
  })
}

//拆分请求
export const getReportDatacount = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data, parmaBasic())
  }
  return axios.request({
    url: "v1/report-form/count",
    method: 'get',
    params: Data
  })
}

export const getSiteData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v1/web/list",
    method: 'get',
    params: Data
  })
}

export const createSiteData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v1/web/add",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const editSiteData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/web/edit",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const getAgentData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v1/agent/list",
    method: 'get',
    params: Data
  })
}

export const createAgentData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }

  let account_info = {
    account: Data.account,
    password: Data.password,
    uName: Data.uName
  };


  Data.account_info = account_info

  delete Data.password
  delete Data.account
  delete Data.uName



  return axios.request({
    url: "v2/agent/add",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const getAgentInfo = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/agent/getInfo",
    method: 'get',
    params: Data
  })
}

export const editAgentData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "/v1/agent/edit",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const addAgentPoint = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/agent/point/add",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const agentPointLog = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/agent/point/log",
    method: 'get',
    params: Data
  })
}

export const agentScoreLog = Data => {
  Data = {
    token: getToken(),
    // agentId: sessionStorage.getItem("agentVal"),
    ...Data
  }
  return axios.request({
    url: "v1/agent/user/score/log",
    method: 'get',
    params: Data
  })
}

export const getUserRecord = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v2/govern/user-record",
    method: 'get',
    params: Data
  })
}

export const agentHistoryLog = Data => {
  Data = {
    token: getToken(),
    // agentId: sessionStorage.getItem("agentVal"),
    ...Data
  }
  return axios.request({
    url: "v2/report-form/history",
    method: 'get',
    params: Data
  })
}

export const getPlayerData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v2/user/list",
    method: 'get',
    params: Data
  })
}

export const editPlayerState = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/user/upState",
    method: 'get',
    params: Data
  })
}

export const getPlayerFwData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v2/fw/list",
    method: 'get',
    params: Data
  })
}

export const getPlayerFwDetailData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  
  if (Data.gameId === "") {
    delete Data.gameId
  }

  if (!!Data.startTime) {
    Data.startTime = new Date(Data.startTime).getTime() / 1000
  }

  if (!!Data.endTime) {
    Data.endTime = new Date(Data.endTime).getTime() / 1000
  }
  return axios.request({
    url: "v2/settlement/list",
    method: 'get',
    params: Data
  })
}

export const getPlayerInfoData = data => {
  return axios.request({
    url: "v1/user/info",
    method: 'get',
    params: {
      token: getToken(),
      ...data
    }
  })
}

export const getGameData = Data => {
  Data = {
    token: getToken(),
  }
  return axios.request({
    url: "v1/game/list",
    method: 'get',
    params: Data
  })
}

export const getGameData2 = Data => {
  Data = {
    token: getToken(),
  }
  return axios.request({
    url: "v2/game/list",
    method: 'get',
    params: Data
  })
}

export const getGD = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: 'v2/game/agent',
    method: 'get',
    params: Data
  })
}

export const createGameData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v1/game/add",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const editGameData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/game/edit",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const configGameData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/game/extraConfig",
    method: 'post',
    data: qs.stringify(Data),
  })
}

/**
 * 生成大厅缓存
 */
export const generateHallCache = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/game/hall/cache",
    method: 'get',
    params: Data
  })
}

export const editGameState = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/game/upState",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const getGameHallList = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/game/hall/list",
    method: 'get',
    params: Data
  })
}

export const setGameHallUpSort = Data => {
  Data = {
    data: [],
    agentId: sessionStorage.getItem("agentVal"),
    token: getToken(),
    ...Data
  }

  for (let x in Data) {
    if (!isNaN(x)) {
      Data.data.push(Data[x])
      delete Data[x]
    }
  }
  Data.data = JSON.stringify(Data.data)
  return axios.request({
    url: "v1/game/hall/upSort",
    method: 'post',
    data: qs.stringify(Data)
  })
}

export const changeGameHallUpState = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/game/hall/upState",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const setGameHallUpData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/game/hall/upData",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const changeGameHallUpMaintain = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/game/hall/upMaintain",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const getGameHallTypeList = () => {
  return axios.request({
    url: "v1/game/hall/getTypeList",
    method: 'get',
    params: {
      token: getToken(),
      agentId: sessionStorage.getItem("agentVal")
    }
  })
}

export const getGameMsgData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data, parmaBasic())
  }
  return axios.request({
    url: "v1/game-msg/list",
    method: 'get',
    params: Data
  })
}

export const createGameMsgData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v1/game-msg/add",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const editGameMsgData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/game-msg/edit",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const deleteGameMsgData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/game-msg/del",
    method: 'get',
    params: Data
  })
}


export const getMsgData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data, parmaBasic())
  }
  return axios.request({
    url: "v1/msg/list",
    method: 'get',
    params: Data
  })
}

export const createMsgData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v1/msg/add",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const editSiteMsgData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/msg/edit",
    method: 'post',
    data: qs.stringify(Data),
  })
}

/**
 * 玩家胜率控制
 *
 * @param [Data]
 * @memberof ControlMange
 */

export const getControlUserData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v1/control/user/getList",
    method: 'get',
    params: Data
  })
}

export const addControlUserProb = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/control/user/add",
    method: 'POST',
    data: qs.stringify(Data),
  })
}

export const removeControlUserProb = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/control/user/upState",
    method: 'POST',
    data: qs.stringify(Data),
  })
}

export const getHistoryControlUserData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/control/history/getList",
    method: 'get',
    params: Data
  })
}

/**
 * 代理胜率控制
 *
 * @param [Data]
 * @memberof ControlMange
 */

export const addControlAgentPomp = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/agent/pomp",
    method: 'POST',
    data: qs.stringify(Data),
  })
}

export const getControlAgentData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/control/agent/getList",
    method: 'get',
    params: Data
  })
}

export const addControlAgentProb = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/control/agent/add",
    method: 'POST',
    data: qs.stringify(Data),
  })
}

export const editControlAgentProb = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/control/agent/edit",
    method: 'POST',
    data: qs.stringify(Data),
  })
}

export const delControlAgentProb = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/control/agent/del",
    method: 'POST',
    data: qs.stringify(Data),
  })
}

/**
 * 游戏胜率控制
 *
 * @param [Data]
 * @memberof ControlMange
 */

export const getControlGameData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }



  return axios.request({
    url: "v1/control/game-list",
    method: 'get',
    params: Data
  })
}

export const setControlGameProb = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  let Url = Data.gameId == undefined ? "v1/control/game-edit" : "v1/control/game-add"
  return axios.request({
    url: Url,
    method: 'POST',
    data: qs.stringify(Data),
  })
}

// todo
export const addControlGame = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/control/game-add",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const getLogListData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/log/list",
    method: 'get',
    params: Data
  })
}

export const getControlLogData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/control/log",
    method: 'get',
    params: Data
  })
}

export const getFeedbackData = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v1/feedback/list",
    method: 'get',
    params: Data
  })
}

export const editFeedbackState = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/feedback/state/up",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const getControlList = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v1/ip/control/list",
    method: 'get',
    params: Data
  })
}

export const addControlIP = Data => {
  Data = {
    token: getToken(),
    ...Object.assign(...Data)
  }
  return axios.request({
    url: "v1/ip/control/add",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const deletControlIP = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/ip/control/del",
    method: 'get',
    params: Data
  })
}

export const getAccountData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/account/list",
    method: 'get',
    params: Data
  })
}

export const addAccountData = Data => {

  Data = {
    token: getToken(),
    ...Data
  }

  return axios.request({
    url: "v1/account/add",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const editAccountData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/account/edit",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const editAccountState = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/account/upState",
    method: 'post',
    data: qs.stringify(Data),
  })
}

export const deleteAccountState = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: "v1/account/del",
    method: 'get',
    params: {
      token: getToken(),
      id: Data.id
    }
  })
}

// 获取站点列表
export const getSelectSite = () => {
  return axios.request({
    url: 'v1/web/select-list',
    method: 'get',
    params: {
      token: getToken()
    }
  })
}

// 获取代理列表
export const getSelectAgent = () => {
  return axios.request({
    url: 'v1/agent/select-list',
    // url: 'v1/agent/select/all-list',
    method: 'get',
    params: {
      token: getToken()
    }
  })
}

// 获取游戏列表
export const getSelectGames = (agent) => {
  return axios.request({
    url: 'v1/game/select-list',
    method: 'get',
    params: {
      token: getToken(),
      agentId: agent
    }
  })
}

// 获取带分类的游戏列表
export const getSelectClassGames = () => {
  return axios.request({
    url: 'v1/game/gamer-game-list',
    method: 'get',
    params: {
      token: getToken()
    }
  })
}

// 获取游戏服务端地址列表
export const getGameServers = () => {
  return axios.request({
    url: 'v2/game/gameUrl',
    method: 'get',
    params: {
      token: getToken()
    }
  })
}

// 获取平台列表
export const getClassList = (agent) => {
  return axios.request({
    url: 'v1/game/getClassList',
    method: 'get',
    params: {
      token: getToken(),
      agentId: agent
    }
  })
}

// 获取类型列表
export const getTypeList = (agent) => {
  return axios.request({
    url: 'v1/game/getTypeList',
    method: 'get',
    params: {
      token: getToken(),
      agentId: agent
    }
  })
}

// 获取站点、代理、游戏三级联动列表
export const getLinkageList = () => {
  return axios.request({
    url: 'v1/web/linkage',
    method: 'get',
    params: {
      token: getToken()
    }
  })
}

// 获取收件人列表
export const getMsgReceiveList = () => {
  return axios.request({
    url: "v1/receive/list",
    method: 'get',
    params: {
      token: getToken()
    },
  })
}

export const generateCustom = (agent) => {
  return axios.request({
    url: 'v1/generate/custom',
    method: 'get',
    params: {
      token: getToken(),
      agentId: agent
    }
  })
}

export const uploadImg = formData => {
  return axios.request({
    url: 'image/upload',
    data: formData
  })
}


export const getReward = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: 'v2/reward/big-list',
    method: 'get',
    params: Data
  })
}

// 获取用户控制数据
export const getControllData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: 'v2/govern/player',
    method: 'get',
    params: Data
  })
}

// 修改用户控制数据
export const updateControllerData = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: 'v2/govern/player',
    method: 'put',
    params: Data
  })
}

//查询注单详情
export const getSettlement = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: 'v2/settlement/list',
    method: 'get',
    params: Data
  })
}

//查询注单详情
export const getQueryOrder = Data => {
  Data = {
    ...Data
  }
  return axios.request({
    url: 'v2/queryOrder',
    method: 'get',
    params: Data
  })
}

//查询注单详情
export const getExportSettlementCount = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: 'v2/export/settlements/count',
    method: 'get',
    params: Data
  })
}

//清理玩家指定游戏状态
export const clearPlayerGameState = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: 'v2/clear/gameState',
    method: 'get',
    params: Data
  })
}

//查询注单详情
export const getExportSettlements = Data => {
  Data = {
    token: getToken(),
    ...Data
  }
  return axios.request({
    url: 'v2/export/settlements',
    method: 'get',
    params: Data
  })
}



export const getTableData = () => {
  return axios.request({
    url: 'get_table_data',
    method: 'get'
  })
}

//暂停/开启所有游戏
export const stopAllGame = () => { 
  
}