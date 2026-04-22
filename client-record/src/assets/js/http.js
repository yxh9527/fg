const apiMethods = {
  methods: {
    logRefreshRedirect(res, source) {
      console.warn('[refresh-redirect]', {
        source,
        code: res && res.code,
        error: res && (res.error || res.err),
        route: router && router.currentRoute ? router.currentRoute.fullPath : '',
        response: res
      })
    },
    apiGet(url, data) {
      return new Promise((resolve, reject) => {
        axios.get(url, data).then((response) => {
          resolve(response.data)
        }, (response) => {
          reject(response)
          _g.closeGlobalLoading()
          bus.$message({
            // message: '请求超时，请检查网络111111111',
            message: this.$t('message.request_timeout'),
            type: 'warning'
          })
        })
      })
    },
    apiPost(url, data) {
      return new Promise((resolve, reject) => {
        axios.post(url, data).then((response) => {
          resolve(response.data)
        }).catch((response) => {
          console.log('f', response)
          resolve(response)
          bus.$message({
            // message: '请求超时，请检查网络',
            message: this.$t('message.request_timeout'),
            type: 'warning'
          })
        })
      })
    },
    apiDelete(url, id) {
      return new Promise((resolve, reject) => {
        axios.delete(url + id).then((response) => {
          resolve(response.data)
        }, (response) => {
          reject(response)
          _g.closeGlobalLoading()
          bus.$message({
            // message: '请求超时，请检查网络',
            message: this.$t('message.request_timeout'),
            type: 'warning'
          })
        })
      })
    },
    apiPut(url, id, obj) {
      return new Promise((resolve, reject) => {
        axios.put(url + id, obj).then((response) => {
          resolve(response.data)
        }, (response) => {
          _g.closeGlobalLoading()
          bus.$message({
            // message: '请求超时，请检查网络',
            message: this.$t('message.request_timeout'),
            type: 'warning'
          })
          reject(response)
        })
      })
    },
    handelResponse(res, cb, errCb) {
      _g.closeGlobalLoading()
      if (res.code == 101) {
        this.logRefreshRedirect(res, 'handelResponse')
        // _g.toastMsg('error', '页面已过期')
        setTimeout(() => {
          _g.safeReplace('/refresh')
        }, 1500)
      } else if (res.code === 104) {
        // 提示不支持旧版详情
        _g.toastMsg('error', res.error)
        _g.closeGlobalLoading()
      } else {
        if (res) {
          console.log(res)
          cb(res)
        } else {
          if (typeof errCb == 'function') {
            errCb()
          }
          this.handleError()
        }
      }
    },
    handleError(res) {
      if (res.code) {
        switch (res.code) {
          case 101:
            _g.toastMsg('error', res.error)
            this.logRefreshRedirect(res, 'handleError')
            setTimeout(() => {
              _g.safeReplace('/refresh')
            }, 1500)
            break
          case 103:
            _g.toastMsg('error', res.error)
            this.logRefreshRedirect(res, 'handleError')
            setTimeout(() => {
              _g.safeReplace('/refresh')
            }, 1500)
            break
          default :
            _g.toastMsg('error', res.error)
        }
      } else {
        // _g.toastMsg('error', '请求超时，请检查网络')
        _g.toastMsg('error', this.$t('message.request_timeout'))
        console.log('default error!')
      }
    },
    resetCommonData(data) {
      _(data.menusList).forEach((res, key) => {
        if (key == 0) {
          res.selected = true
        } else {
          res.selected = false
        }
      })
      Lockr.set('menus', data.menusList)              // 菜单数据
      Lockr.set('authKey', data.authKey)              // 权限认证
      Lockr.set('rememberKey', data.rememberKey)      // 记住密码的加密字符串
      Lockr.set('authList', data.authList)            // 权限节点列表
      Lockr.set('userInfo', data.userInfo)            // 用户信息
      Lockr.set('sessionId', data.sessionId)          // 用户sessionid
      window.axios.defaults.headers.authKey = Lockr.get('authKey')
      let routerUrl = ''
      if (data.menusList[0].url) {
        routerUrl = data.menusList[0].url
      } else {
        routerUrl = data.menusList[0].child[0].child[0].url
      }
      setTimeout(() => {
        let path = this.$route.path
        if (routerUrl != path) {
          _g.safeReplace(routerUrl)
        } else {
          _g.shallowRefresh(this.$route.name)
        }
      }, 1000)
    },
    reAjax(url, data) {
      return new Promise((resolve, reject) => {
        axios.post(url, data).then((response) => {
          resolve(response.data)
        }, (response) => {
          reject(response)
          bus.$message({
            // message: '请求超时，请检查网络',
            message: this.$t('message.request_timeout'),
            type: 'warning'
          })
        })
      })
    }
  },
  computed: {
    showLoading() {
      return store.state.globalLoading
    }
  }
}

export default apiMethods



// WEBPACK FOOTER //
// ./src/assets/js/http.js
