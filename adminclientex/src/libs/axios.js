import axios from 'axios'
import store from '@/store'
import {
  getToken
} from '@/libs/util'
// import { Spin } from 'iview'
const addErrorLog = errorInfo => {
  const {
    statusText,
    status,
    request: {
      responseURL
    }
  } = errorInfo
  let info = {
    type: 'ajax',
    code: status,
    mes: statusText,
    url: responseURL
  }
  if (!responseURL.includes('save_error_logger')) store.dispatch('addErrorLog', info)
}

class HttpRequest {
  constructor(baseUrl = baseURL) {
    this.baseUrl = baseUrl
    this.queue = {}
  }
  getInsideConfig() {
    const config = {
      baseURL: this.baseUrl,
      headers: {
        //
      }
    }
    return config
  }
  destroy(url) {
    delete this.queue[url]
    if (!Object.keys(this.queue).length) {
      // Spin.hide()
    }
  }
  interceptors(instance, url) {
    // 请求拦截
    instance.interceptors.request.use(config => {
      // 添加全局的loading...
      if (!Object.keys(this.queue).length) {
        // Spin.show() // 不建议开启，因为界面不友好
      }
      this.queue[url] = true

      if (config.url == 'v1/agent/point/add' && window.lastInterfaceUrlTime && config.url == window.lastInterfaceUrl && Date.now() - window.lastInterfaceUrlTime <= 10000) {
        return Promise.reject({
          error: 'unknown'
        })
      }

      window.lastInterfaceUrl = config.url
      window.lastInterfaceUrlTime = Date.now()

      if (window.urlVersion == 'v1') {
        config.url = config.url.replace('v2/', 'v1/')
      }
      // config.baseURL = `${sessionStorage.getItem('node_url') || ''}${config.baseURL}`
      if (config.params) {
        config.params.sign = sessionStorage.getItem('sign');
      }
      return config
    }, error => {
      return Promise.reject(error)
    })
    // 响应拦截
    if (window.lastMsgTime === undefined) {
      window.lastMsgTime = 1
      window.lastMsgContent = ""
    }

    instance.interceptors.response.use(res => {
      if (res.data.code != 200) {

        if (appInstance && appInstance.$Message) {
          if (Date.now() - lastMsgTime > 300 || lastMsgContent != res.data.msg) {
            //屏蔽多个重复报错
            res.data.msg && appInstance.$Message.error(res.data.msg);
          }
          lastMsgTime = Date.now()
          lastMsgContent = res.data.msg
          if (res.data.code == 401) {

            setTimeout(() => {
              axios.request({
                url: this.baseUrl + 'v1/logout',
                method: 'get',
                params: {
                  token: getToken()
                }
              }).then(() => {
                sessionStorage.setItem('token', '')
                location.pathname = '/login'
              })
            }, 2000);
          }
        } else {
          alert(res.data.msg)
        }
        return Promise.reject({
          response: {
            status: res.data.code,
            msg: res.data.msg
          }
        })
      }

      const {
        data,
        status
      } = res

      if (data.data && data.data.total) {
        data.data.total = Number(data.data.total)
      }
      try {
        /**
         * V2添加V1字段
         */
        if (res.config.url.includes('v2')) {
          if (data && data.data && data.data.length) {
            data.data.map(item => {
              //首页字段映射
              item.docCount && (item.eNumber = item.docCount)
              item.profitLossTotal && (item.pTotal = item.profitLossTotal)
              item.effectiveBetsTotal && (item.eTotal = item.effectiveBetsTotal)
              item.webName && (item.web_name = item.webName)
              item.gameName && (item.name = item.gameName)
              item.agentName && (item.agent_name = item.agentName)

            })
          }
          if (data && data.data && data.data.data && data.data.data.length) {
            data.data.data.map(item => {
              //首页字段映射
              item.docCount && (item.eNumber = item.docCount)
              item.profitLossTotal && (item.pTotal = item.profitLossTotal)
              item.effectiveBetsTotal && (item.eTotal = item.effectiveBetsTotal)
              item.webName && (item.web_name = item.webName)
              item.gameName && (item.name = item.gameName)
              item.agentName && (item.agent_name = item.agentName)
              item.userTotal && (item.userNumber = item.userTotal)
              item.docCount && (item.gameNumber = item.docCount)
              item.userBetsTotal && (item.betsScore = item.userBetsTotal)
              item.effectiveBetsTotal && (item.effeBetsScore = item.effectiveBetsTotal)
              item.profitLossTotal && (item.profitLoss = item.profitLossTotal)
              item.pumpTotal && (item.pump = item.pumpTotal)
              item.platformName && (item.gameClassName = item.platformName);
              (!item.gameName && item.name) && (item.gameName = item.name)
            })
          }
        }

      } catch (error) {

      }
      return {
        data,
        status
      }
    }, error => {
      this.destroy(url)
      let errorInfo = error.response
      if (!errorInfo) {
        const {
          request: {
            statusText,
            status
          },
          config
        } = JSON.parse(JSON.stringify(error))
        errorInfo = {
          statusText: statusText ? statusText : '',
          status,
          request: {
            responseURL: config.url
          }
        }
      }
      addErrorLog(errorInfo)
      return Promise.reject(error)
    })
  }
  request(options) {
    const instance = axios.create()
    options = Object.assign(this.getInsideConfig(), options)
    this.interceptors(instance, options.url)
    return instance(options)
  }
}
export default HttpRequest
