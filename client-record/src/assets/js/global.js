const commonFn = {
  j2s(obj) {
    return JSON.stringify(obj)
  },
  safeReplace(location) {
    const currentRoute = router.currentRoute
    const resolved = router.resolve(location, currentRoute)
    if (resolved.route.fullPath === currentRoute.fullPath) {
      return Promise.resolve(currentRoute)
    }
    return router.replace(location).catch((error) => {
      if (error && error.name === 'NavigationDuplicated') {
        return currentRoute
      }
      throw error
    })
  },
  shallowRefresh(name) {
    return this.safeReplace({ path: '/refresh', query: { name: name }})
  },
  closeGlobalLoading() {
    setTimeout(() => {
      store.dispatch('showLoading', false)
    }, 0)
  },
  openGlobalLoading() {
    setTimeout(() => {
      store.dispatch('showLoading', true)
    }, 0)
  },
  cloneJson(obj) {
    return JSON.parse(JSON.stringify(obj))
  },
  toastMsg(type, msg) {
    switch (type) {
      case 'normal':
        bus.$message(msg)
        break
      case 'success':
        bus.$message({
          message: msg,
          type: 'success'
        })
        break
      case 'warning':
        bus.$message({
          message: msg,
          type: 'warning'
        })
        break
      case 'error':
        bus.$message.error(msg)
        break
    }
  },
  clearVuex(cate) {
    store.dispatch(cate, [])
  },
  apiUrl() {
    return HOST
  },
  dyUrl() {
    return DY_HOST
  },
  getHasRule(val) {
    const moduleRule = 'admin'
    let userInfo = Lockr.get('userInfo')
    if (userInfo.id == 1) {
      return true
    } else {
      let authList = moduleRule + Lockr.get('authList')
      return _.includes(authList, val)
    }
  }
}

export default commonFn



// WEBPACK FOOTER //
// ./src/assets/js/global.js
