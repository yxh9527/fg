import {
  login,
  logout,
  getContentByMsgId,
} from '@/api/user'
import {
  setToken,
  getToken
} from '@/libs/util'
import Cookies from 'js-cookie'

export default {
  state: {
    userName: '',
    userId: '',
    avatarImgPath: '',
    token: getToken(),
    access: '',
    hasGetInfo: false,
    unreadCount: 0,
    messageUnreadList: [],
    messageReadedList: [],
    messageTrashList: [],
    messageContentStore: {}
  },
  mutations: {
    setAvatar(state, avatarPath) {
      state.avatarImgPath = avatarPath
    },
    setUserId(state, id) {
      state.userId = id
    },
    setUserName(state, name) {
      state.userName = name
    },
    setAccess(state, access) {
      state.access = access
    },
    setToken(state, token) {
      state.token = token
      setToken(token)
    },
    setHasGetInfo(state, status) {
      state.hasGetInfo = status
    },
    setMessageCount(state, count) {
      state.unreadCount = count
    },
    setMessageUnreadList(state, list) {
      state.messageUnreadList = list
    },
    setMessageReadedList(state, list) {
      state.messageReadedList = list
    },
    setMessageTrashList(state, list) {
      state.messageTrashList = list
    },
    updateMessageContentStore(state, {
      msg_id,
      content
    }) {
      state.messageContentStore[msg_id] = content
    },
    moveMsg(state, {
      from,
      to,
      msg_id
    }) {
      const index = state[from].findIndex(_ => _.msg_id === msg_id)
      const msgItem = state[from].splice(index, 1)[0]
      msgItem.loading = false
      // state[to].unshift(msgItem)
    }
  },
  getters: {
    messageUnreadCount: state => state.messageUnreadList.length,
    messageReadedCount: state => state.messageReadedList.length,
    messageTrashCount: state => state.messageTrashList.length
  },
  actions: {
    // 登录
    handleLogin({
      commit
    }, {
      name,
      password
    }) {
      name = name.trim()
      return new Promise((resolve, reject) => {
        login({
          name,
          password
        }).then(res => {
          if (res.data.code == 200 && res.data.data.token) {
            const path = res.data.data
            commit('setToken', path.token)

            // { label: "总控账号", value: 1 },
            // { label: "客服账号", value: 2 },
            // { label: "代理账号", value: 3 },
            // { label: "操作账号", value: 4 },
            const data = {
              name: path.user.name,
              userType: path.user.userType,
              avatar: path.user.userType == 1 ? '/color/head1.png' : '/color/head2.png',
              userid: path.user.id,
              access: (() => {
                switch (path.user.userType) {
                  case 1:
                    return ['administrator', 'userCenter', 'operation'];
                  case 2:
                    return ['userCenter'];
                  case 3:
                    return [];
                  case 4:
                    return ['userCenter', 'operation']

                }
              })()
            }
            Cookies.set("userInfo", data, {
              expires: 1
            })
            resolve(res)

          } else {
            alert("登录失败：" + res.data.msg)
          }
          resolve()
        }).catch(err => {
          reject(err)
        })
      })
    },
    // 退出登录
    handleLogOut({
      state,
      commit
    }) {
      return new Promise((resolve, reject) => {
        logout(state.token).then(() => {
          commit('setToken', '')
          commit('setAccess', [])
          sessionStorage.removeItem('typeOption')
          sessionStorage.removeItem('siteVal')
          sessionStorage.removeItem('games')
          sessionStorage.removeItem('siteOption')
          sessionStorage.removeItem('agentVal')
          sessionStorage.removeItem('classOption')
          sessionStorage.removeItem('token')
          sessionStorage.removeItem('node_url')
          sessionStorage.removeItem('sign')
          resolve()
        }).catch(err => {
          reject(err)
        })
        // 如果你的退出登录无需请求接口，则可以直接使用下面三行代码而无需使用logout调用接口
        // commit('setToken', '')
        // commit('setAccess', [])
        // resolve()
      })
    },
    // 获取用户相关信息
    getUserInfo({
      state,
      commit
    }) {
      return new Promise((resolve, reject) => {
        try {
          const data = JSON.parse(Cookies.get("userInfo"))
          commit('setAvatar', data.avatar)
          commit('setUserName', data.name)
          commit('setUserId', data.userid)
          commit('setAccess', data.access)
          resolve(1)
        } catch (error) {
          reject(error)
        }
      })
    },
    // 根据当前点击的消息的id获取内容
    getContentByMsgId({
      state,
      commit
    }, {
      msg_id
    }) {
      return new Promise((resolve, reject) => {
        let contentItem = state.messageContentStore[msg_id]
        if (contentItem) {
          resolve(contentItem)
        } else {
          getContentByMsgId(msg_id).then(res => {
            const content = res.data.data[0].info
            commit('updateMessageContentStore', {
              msg_id,
              content
            })
            resolve(content)
          })
        }
      })
    },
  }
}
