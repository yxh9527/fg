import Main from '@/components/main'

/**
 * iview-admin中meta除了原生参数外可配置的参数:
 * meta: {
 *  title: { String|Number|Function }
 *         显示在侧边栏、面包屑和标签栏的文字
 *         使用'{{ 多语言字段 }}'形式结合多语言使用，例子看多语言的路由配置;
 *         可以传入一个回调函数，参数是当前路由对象，例子看动态路由和带参路由
 *  hideInBread: (false) 设为true后此级路由将不会出现在面包屑中，示例看QQ群路由配置
 *  hideInMenu: (false) 设为true后在左侧菜单不会显示该页面选项
 *  notCache: (false) 设为true后页面在切换标签后不会缓存，如果需要缓存，无需设置这个字段，而且需要设置页面组件name属性和路由配置的name一致
 *  access: (null) 可访问该页面的权限数组，当前路由设置的权限会影响子路由
 *  icon: (-) 该页面在左侧菜单、面包屑和标签导航处显示的图标，如果是自定义图标，需要在图标名称前加下划线'_'
 *  beforeCloseName: (-) 设置该字段，则在关闭当前tab页时会去'@/router/before-close.js'里寻找该字段名对应的方法，作为关闭前的钩子函数
 * }
 */

export default [{
  path: '/login',
  name: 'login',
  meta: {
    title: 'Login - 登录',
    hideInMenu: true
  },
  component: () => import('@/view/login/login.vue')
},
{
  path: '/agent-aggs-detail',
  name: 'agent-aggs-detail',
  meta: {
    title: '统计详情',
    hideInMenu: true
  },
  component: () => import('@/view/agent/agent-aggs-detail.vue')
},
{
  path: '/queryOrder/:order/:account/:token',
  name: 'queryOrder',
  meta: {
    title: '注单详情',
    hideInMenu: true
  },
  component: () => import('@/view/orderCheck/detail.vue')
},
{
  path: '/',
  name: 'newHome-menu',
  component: Main,
  meta: {
    access: ['administrator'],
    icon: 'md-apps',
    title: '首页'
  },
  children: [{
    path: '/newHome',
    name: 'newHome',
    meta: {
      title: '首页',
      icon: 'md-home'
    },
    component: () => import('@/view/single-page/newHome')
  }, {
    path: '/newHome-agentGamesStatistic/:id/:st/:et',
    props: true,
    name: 'newHome-agentGamesStatistic/:id/:st/:et',
    meta: {
      hideInMenu: true,
      title: '代理各游戏数据统计'
    },
    component: () => import('@/view/single-page/agentGames.vue')
  },
  {
    path: '/newHome-agentPerformanceStatistic/:id',
    props: true,
    name: 'newHome-agentPerformanceStatistic/:id',
    meta: {
      hideInMenu: true,
      title: '代理各游戏数据统计'
    },
    component: () => import('@/view/single-page/agentPerformance.vue')
  }, {
    path: '/newHome-agentOrderStatistic/:id/:st/:et',
    props: true,
    name: 'newHome-agentOrderStatistic/:id/:st/:et',
    meta: {
      hideInMenu: true,
      icon: 'md-apps',
      notCache: false,
      title: '用户流水'
    },
    component: () => import('@/view/single-page/agentOrders.vue')
  },
  {
    path: '/newHome-weekIndicator',
    props: true,
    name: 'newHome-weekIndicator',
    meta: {
      hideInMenu: true,
      icon: 'md-apps',
      notCache: false,
      title: '单周指标'
    },
    component: () => import('@/view/single-page/weekIndicator.vue')
  }
  ]
},
{
  path: '/report',
  name: 'report',
  component: Main,
  meta: {
    access: ['administrator']
  },
  children: [{
    path: '/report-page',
    name: 'report-page',
    meta: {
      icon: 'md-apps',
      title: '报表'
    },
    component: () => import('@/view/single-page/report.vue')
  }]
},
{
  path: '/website-menu',
  name: 'website-menu',
  meta: {
    access: ['administrator'],
    icon: 'md-apps',
    title: '站点'
  },
  component: Main,
  children: [{
    path: '/website',
    name: 'website',
    meta: {
      icon: 'md-apps',
      notCache: false,
      title: '站点中心'
    },
    component: () => import('@/view/website/website.vue')
  },
  {
    path: '/website-add',
    name: 'website-add',
    meta: {
      hideInMenu: true,
      icon: 'md-apps',
      notCache: false,
      title: '创建站点',
      state: 0
    },
    component: () => import('@/view/website/website-add.vue')
  },
  {
    path: '/website-edit',
    name: 'website-edit',
    meta: {
      icon: 'md-apps',
      hideInMenu: true,
      title: '编辑站点',
      state: 1
    },
    component: () => import('@/view/website/website-add.vue')
  }
  ]
},
{
  path: '/agent-menu',
  name: 'agent-menu',
  meta: {
    access: ['operation'],
    icon: 'md-apps',
    title: '代理'
  },
  component: Main,
  children: [{
    path: '/agent',
    name: 'agent',
    meta: {
      notCache: false,
      icon: 'md-apps',
      title: '代理中心'
    },
    component: () => import('@/view/agent/agent.vue')
  },
  {
    path: '/agent-add',
    name: 'agent-add',
    meta: {
      notKeep: true,
      hideInMenu: true,
      icon: 'md-apps',
      notCache: false,
      title: '创建代理商',
      state: 0
    },
    component: () => import('@/view/agent/agent-add.vue')
  },
  {
    path: '/agent-domain',
    name: 'agent-domain',
    meta: {
      notKeep: true,
      hideInMenu: true,
      icon: 'md-apps',
      notCache: false,
      title: '代理商域名设置',
      state: 0
    },
    component: () => import('@/view/agent/agent-domain.vue')
  },
  {
    path: '/agent-edit',
    name: 'agent-edit',
    meta: {
      icon: 'md-apps',
      hideInMenu: true,
      title: '编辑代理商',
      state: 2
    },
    component: () => import('@/view/agent/agent-add.vue')
  }
  ]
},
{
  path: '/players-menu',
  name: 'players-menu',
  meta: {
    access: ['userCenter'],
    icon: 'md-apps',
    title: '用户'
  },
  component: Main,
  children: [{
    path: '/players',
    name: 'players',
    meta: {
      icon: 'md-apps',
      notCache: false,
      title: '用户中心'
    },
    component: () => import('@/view/players/players.vue')
  }, {
    path: '/players-control',
    name: 'players-control',
    meta: {
      hideInMenu: true,
      icon: 'md-apps',
      notCache: false,
      title: '用户中心'
    },
    component: () => import('@/view/players/players-control.vue')
  },
  {
    path: '/players-record-:id',
    props: true,
    name: 'players-record-:id',
    meta: {
      hideInMenu: true,
      icon: 'md-apps',
      notCache: false,
      title: '用户流水'
    },
    component: () => import('@/view/players/players-record.vue')
  }, {
    path: '/players-panel-:id',
    props: true,
    name: 'players-panel-:id',
    meta: {
      hideInMenu: true,
      icon: 'md-apps',
      notCache: false,
      title: '用户看板'
    },
    component: () => import('@/view/players/players-panel.vue')
  }, {
    path: '/userStatement/:id',
    props: true,
    name: '/userStatement/:id',
    meta: {
      hideInMenu: true,
      icon: 'md-apps',
      notCache: false,
      title: '用户明细'
    },
    component: () => import('@/view/single-page/userStatement.vue')
  },
  {
    path: '/players-game-:id',
    props: true,
    name: 'players-game-:id',
    meta: {
      hideInMenu: true,
      icon: 'md-apps',
      notCache: false,
      title: '用户对局'
    },
    component: () => import('@/view/players/players-game.vue')
  }
  ]
},
{
  path: '/',
  name: 'control',
  component: Main,
  meta: {
    access: ['userCenter'],
    icon: 'md-apps',
    title: '注单'
  },
  children: [{
    path: '/note-detail',
    name: 'note-detail',
    meta: {
      title: '注单详情',
      icon: 'md-home'
    },
    component: () => import('@/view/notes/detail.vue')
  }]
},
{
  path: '/userControl-menu',
  name: 'userControl-menu',
  meta: {
    access: ['administrator'],
    icon: 'md-apps',
    title: '控制管理'
  },
  component: Main,
  children: [{
    path: '/userControl',
    name: 'userControl',
    meta: {
      icon: 'md-apps',
      notCache: false,
      title: '控制管理'
    },
    component: () => import('@/view/single-page/userControl.vue')
  }]
},
{
  path: '/game-manage-menu',
  name: 'game-manage-menu',
  meta: {
    access: ['operation'],
    icon: 'md-apps',
    title: '游戏管理'
  },
  component: Main,
  children: [{
    path: '/game-manage',
    name: 'game-manage',
    meta: {
      icon: 'md-apps',
      title: '游戏管理'
    },
    component: () => import('@/view/gameManage/game-manage.vue')
  },
  {
    path: '/game-allAgent',
    name: 'game-allAgent',
    meta: {
      icon: 'md-apps',
      hideInMenu: true,
      title: '全部代理游戏管理'
    },
    component: () => import('@/view/gameManage/game-manage-allAgent.vue')
  },
  {
    path: '/game-add',
    name: 'game-add',
    meta: {
      icon: 'md-apps',
      title: '创建游戏',
      hideInMenu: true,
      state: 0
    },
    component: () => import('@/view/gameManage/game-add.vue')
  },
  {
    path: '/game-edit',
    name: '/game-edit',
    meta: {
      icon: 'md-apps',
      title: '编辑游戏',
      hideInMenu: true,
      state: 1
    },
    component: () => import('@/view/gameManage/game-add.vue')
  },
  /**
   * 房群游戏列表
   */
  {
    path: '/game-groups-list',
    name: '/game-groups-list',
    meta: {
      icon: 'md-apps',
      title: '房群游戏',
      hideInMenu: true,
      state: 1
    },
    component: () => import('@/view/gameManage/game-groups-list.vue')
  },
  /**
   * 创建官群
   */
  {
    path: '/game-groups-add',
    name: '/game-groups-add',
    meta: {
      icon: 'md-apps',
      title: '创建官群',
      hideInMenu: true,
      state: 1
    },
    component: () => import('@/view/gameManage/game-groups-add.vue')
  }
  ]
},

{
  path: '/manage-log-menu',
  name: 'manage-log-menu',
  component: Main,
  meta: {
    access: ['administrator']
  },
  children: [{
    path: '/manage-log',
    name: 'manage-log',
    meta: {
      access: ['administrator'],
      icon: 'md-apps',
      title: '日志管理'
    },
    component: () => import('@/view/single-page/manage-log.vue')
  }]
},
{
  path: '/account-menu',
  name: 'account-menu',
  component: Main,
  meta: {
    access: ['administrator']
  },
  children: [{
    path: '/account',
    name: 'account',
    meta: {
      access: ['administrator'],
      icon: 'md-apps',
      title: '账号管理'
    },
    component: () => import('@/view/single-page/account-manage.vue')
  }]
},
{
  path: '/exception-menu',
  name: 'exception-menu',
  meta: {
    hideInMenu: true,
    access: ['administrator'],
    icon: 'md-apps',
    title: '异常提示'
  },
  component: Main,
  children: [{
    path: '/exception',
    name: 'exception',
    meta: {
      access: ['administrator'],
      icon: 'md-apps',
      title: '异常提示列表'
    },
    component: () => import('@/view/exception/exception.vue')
  },
  {
    path: '/exception-rule',
    name: 'exception-rule',
    meta: {
      access: ['administrator'],
      icon: 'md-apps',
      title: '异常规则管理'
    },
    component: () => import('@/view/exception/exception-rule.vue')
  },
  {
    path: 'key_page',
    name: 'key_page',
    meta: {
      access: ['administrator'],
      title: 'Key',
      hideInMenu: true
    },
    component: () => import('@/view/single-page/key.vue')
  }
  ]
},
{
  path: '/systemSetting',
  name: 'systemSetting',
  component: Main,
  meta: {
    access: ['administrator'],
    icon: 'md-apps',
    title: '系统配置',
    hideInMenu: true
  },
  children: [{
    path: '/config',
    name: 'config',
    meta: {
      title: '系统配置',
      icon: 'md-home'
    },
    component: () => import('@/view/single-page/config.vue')
  }]
},
{
  path: '/401',
  name: 'error_401',
  meta: {
    hideInMenu: true
  },
  component: () => import('@/view/error-page/401.vue')
},
{
  path: '/500',
  name: 'error_500',
  meta: {
    hideInMenu: true
  },
  component: () => import('@/view/error-page/500.vue')
},
{
  path: '*',
  name: 'error_404',
  meta: {
    hideInMenu: true
  },
  component: () => import('@/view/error-page/404.vue')
}
]
