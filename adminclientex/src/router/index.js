import Vue from 'vue'
import Router from 'vue-router'
import routes from './routers'
import store from '@/store'
import iView from 'view-design'
import Cookies from 'js-cookie'
import {
  setToken,
  getToken,
  canTurnTo,
  setTitle
} from '@/libs/util'
import config from '@/config'
const {
  homeName
} = config


Vue.use(Router)
const router = new Router({
  routes,
  mode: 'history'
})
const LOGIN_PAGE_NAME = 'login'

const turnTo = (to, access, next) => {

  if (canTurnTo(to.name, access, routes)) {
    next()
  } else next({
    replace: true,
    name: 'error_401'
  })
}

router.beforeEach((to, from, next) => {
  iView.LoadingBar.start()
  const token = getToken()
  if (to.name=="queryOrder") {
    next() // 跳转
  }else{
    if (!token && to.name !== LOGIN_PAGE_NAME) {
      // 未登录且要跳转的页面不是登录页
      next({
        name: LOGIN_PAGE_NAME // 跳转到登录页
      })
    } else if (!token && to.name === LOGIN_PAGE_NAME) {
      // 未登陆且要跳转的页面是登录页
      next() // 跳转
    } else if (token && to.name === LOGIN_PAGE_NAME) {
      // 已登录且要跳转的页面是登录页
      next({
        name: homeName // 跳转到homeName页
      })
    } else {
      let userData = JSON.parse(Cookies.get("userInfo"))
      turnTo(to, userData.access, next)
      store.dispatch('getUserInfo').then(user => {}).catch(() => {
        setToken('')
        next({
          name: 'login'
        })
      })
    }
  }
})

router.afterEach(to => {
  iView.LoadingBar.finish()
  setTitle(to, router.app)
  window.scrollTo(0, 0)
})

export default router
