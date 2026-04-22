import 'core-js/stable'
import 'regenerator-runtime/runtime'
import Vue from 'vue'
import App from './App'
import axios from 'axios'
import Lockr from 'lockr'
import Cookies from 'js-cookie'
import _ from 'lodash'
import moment from 'moment'
import ElementUI from 'element-ui'
import 'element-ui/lib/theme-chalk/index.css'
import routes from './routes'
import VueRouter from 'vue-router'
import store from './vuex/store'
import filter from './assets/js/filter'
import _g from './assets/js/global'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import 'assets/css/global.css'
import 'assets/css/base.css'
import VueI18n from 'vue-i18n'
import enLocale from 'element-ui/lib/locale/lang/en'
import zhtwLocale from 'element-ui/lib/locale/lang/zh-TW'
import zhcnLocale from 'element-ui/lib/locale/lang/zh-CN'
import jaLocale from 'element-ui/lib/locale/lang/ja'
import koLocale from 'element-ui/lib/locale/lang/ko'
import thLocale from 'element-ui/lib/locale/lang/th'
import viLocale from 'element-ui/lib/locale/lang/vi'
import langlocale from 'element-ui/lib/locale'

// localStorage.game_id = `{"data":"3304"}`
// localStorage.query_data = `{"data":{"query_type":0,"game_type":"老虎机","game_type_id":"10","game_name":"粉红女郎","game_id":"3304","game_size":20,"bill_type":"","bill_type_id":0,"game_time_id":1,"game_time_label":"今天","game_time":["2026-01-13 00:00:00","2026-01-13 22:48:18"]}}`
// localStorage.language = `{"data":"zh-cn"}`
// localStorage.slot_detail = `{"data":{"all_bpay":"-0.36","all_bonus":"0.04","all_bets":"0.40","start_chips":"981.52","end_chips":"981.16","jp_bonus":0,"game_id":3304,"game_type":"普通旋转","id":1634875,"time":"2026-01-13 22:48:08","agent_id":143,"game_name":"粉红女郎","dt_url":"https://h5.sit.fg.fgsit.me/slot_log_details?id=1634875&buckets=143&language=zh-cn"}}`
// localStorage.timezone = `{"data":"-4"}`
// localStorage.token = `{"data":"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiIsImp0aSI6IjRmMWcyM2ExMmFhYmxpenppIn0.eyJpc3MiOiJodHRwczpcL1wvaDUuc2l0LmZnLmZnc2l0Lm1lXC8iLCJhdWQiOiJodHRwczpcL1wvaDUuc2l0LmZnLmZnc2l0Lm1lXC8iLCJqdGkiOiI0ZjFnMjNhMTJhYWJsaXp6aSIsImlhdCI6MTc2ODM1ODg5OCwibmJmIjoxNzY4MzU4ODk4LCJleHAiOjE3Njg2MTgwOTgsInVzZXJfaW5mbyI6IntcInVzZXJfaWRcIjoyOTgyNDA4LFwidXNlcl9uYW1lXCI6XCJ0ZXN0MjAyNlwiLFwiZ2FtZV9pZFwiOlwiMzMwNFwiLFwiYWdlbnRfaWRcIjoxNDMsXCJ0b3BfYWdlbnRfdWlkXCI6MTQzLFwiY3Vycm5lY3lpZFwiOlwiLVwifSJ9.r28IOsP3IU4Jh7clZaa2jqoqUAeSffwKRFzfXWXpY-0"}`
// localStorage.user_id = `{"data":2982408}`

axios.defaults.baseURL = HOST
axios.defaults.timeout = 1000 * 15
axios.defaults.headers.webToken = Lockr.get('token')
axios.defaults.headers.authKey = Lockr.get('authKey')
axios.defaults.headers.sessionId = Lockr.get('sessionId')
axios.defaults.headers['Content-Type'] = 'application/json'

function syncRuntimeQuery(query) {
  if (!query || typeof query !== 'object') {
    return
  }

  if (query.token) {
    Lockr.set('token', query.token)
    axios.defaults.headers.webToken = query.token
  }

  if (query.game_id) {
    Lockr.set('game_id', query.game_id)
  } else if (Object.prototype.hasOwnProperty.call(query, 'game_id')) {
    Lockr.set('game_id', 0)
  }

  if (query.language) {
    Lockr.set('language', query.language)
  }

  if (query.timezone) {
    Lockr.set('timezone', query.timezone)
  }
}

const router = new VueRouter({
  mode: 'history',
  base: ROUTER_BASE,
  routes: routes
})

router.beforeEach((to, from, next) => {
  syncRuntimeQuery(to.query)
  const hideLeft = to.meta.hideLeft
  store.dispatch('showLeftMenu', hideLeft)
  store.dispatch('showLoading', true)
  NProgress.start()
  next()
})

router.afterEach(transition => {
  NProgress.done()
})

Vue.use(ElementUI)
Vue.use(VueRouter)
Vue.use(VueI18n)

window.router = router
window.store = store
window.HOST = HOST
window.axios = axios
window._ = _
window.moment = moment
window.Lockr = Lockr
window.Cookies = Cookies
window._g = _g
window.imgUrl = IMG_HOST
window.pageSize = 15

const bus = new Vue()
window.bus = bus

// 初始化多语言
const i18n = new VueI18n({
  locale: 'zh-cn',
  messages: {
    'zh-cn': Object.assign(require('./components/Common/lang/zh'), zhcnLocale),
    'zh-hk': Object.assign(require('./components/Common/lang/zhhk'), zhtwLocale),
    'en-us': Object.assign(require('./components/Common/lang/en'), enLocale),
    'ja-jp': Object.assign(require('./components/Common/lang/jajp'), jaLocale),
    'ko-kr': Object.assign(require('./components/Common/lang/kokr'), koLocale),
    'th-th': Object.assign(require('./components/Common/lang/thth'), thLocale),
    'vi-vn': Object.assign(require('./components/Common/lang/vivn'), viLocale)
  }
})
langlocale.i18n((key, value) => i18n.t(key, value)) // 为了实现element插件的多语言切换

new Vue({
  el: '#app',
  template: '<App/>',
  filters: filter,
  router,
  store,
  i18n,
  components: { App }
}).$mount('#app')



// WEBPACK FOOTER //
// ./src/main.js
