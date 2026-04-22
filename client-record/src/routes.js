import Login from './components/Login.vue'
import Home from './components/Home.vue'
import refresh from './components/refresh.vue'
import slot_record from './components/Common/slot_record.vue'
import slot_account from './components/Common/slot_account.vue'
import slot_detail from './components/Common/slot_detail.vue'
import chess_detail from './components/Common/chess_detail.vue'
import chess_record from './components/Common/chess_record.vue'
import chess_account from './components/Common/chess_account.vue'
import fruit_record from './components/Common/fruit_record.vue'
import fruit_account from './components/Common/fruit_account.vue'
import fruit_detail from './components/Common/fruit_detail.vue'
import hunting_record from './components/Common/hunting_record.vue'
import hunting_account from './components/Common/hunting_account.vue'
import hunting_detail from './components/Common/hunting_detail.vue'

const routes = [
  { path: '/', component: Login, name: 'Login' },
  {
    path: '/',
    component: Home,
    children: [
      { path: '/refresh', component: refresh, name: 'refresh' },
      { path: '/slot_record', component: slot_record, name: 'slot_record' },
      { path: '/slot_account', component: slot_account, name: 'slot_account' },
      { path: '/fruit_record', component: fruit_record, name: 'fruit_record' },
      { path: '/fruit_account', component: fruit_account, name: 'fruit_account' },
      { path: '/chess_record', component: chess_record, name: 'chess_record' },
      { path: '/chess_account', component: chess_account, name: 'chess_account' },
      { path: '/hunting_record', component: hunting_record, name: 'hunting_record' },
      { path: '/hunting_account', component: hunting_account, name: 'hunting_account' },
      { path: '/', component: slot_record, name: 'slot_record' }
    ]
  },
  { path: '/slot_detail', component: slot_detail, name: 'slot_detail' },
  { path: '/chess_detail', component: chess_detail, name: 'chess_detail' },
  { path: '/fruit_detail', component: fruit_detail, name: 'fruit_detail' },
  { path: '/hunting_detail', component: hunting_detail, name: 'hunting_detail' }
]
export default routes




// WEBPACK FOOTER //
// ./src/routes.js