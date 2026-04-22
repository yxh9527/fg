export default {
  /**
   *  配置显示在浏览器标签的title
   */
  title: '游戏管理后台',
  /**
   *  配置后台访问地址，api请求基础路径
   */
  baseUrl: {
    // 开发环境Api地址
    dev: '/api/auth/',
    // dev: 'http://127.0.0.1:8000/api/auth',
    // // 生产环境Api地址
    pro: '/api/auth/'
    // pro: 'http://127.0.0.1:8000/api/auth/'

    // dev: 'http://172.21.211.215:9529/api/auth',
    // pro: 'http://172.21.211.215:9529/api/auth/'

  },
  /**
   *  默认打开的首页的路由name，默认为home页面
   */
  homeName: 'home',

  /**
   *  需要加载的插件
   */
  plugin: {
    'error-store': {
      showInHeader: true, // 设为false后不会在顶部显示错误日志徽标
      developmentOff: true // 设为true后在开发环境不会收集错误信息，方便开发中排查错误
    }
  }
}

/**
 *  配置各类全局参数
 */
export const setting = {

  /**
   *  表格分页配置，page 默认第几页，pageSize 默认每页条数 pageOpts 分页数量选择列表
   */
  page: 1,
  pageSize: 15,
  pageOpts: [15, 30, 50, 100, 200, 300],

  /**
   *  游戏配置，对配置名称的替换，按参数Id列表和游戏key进行名称匹配
   */
  defaultConfigName: {
    id: [],
    list: {
      name: '房间名',
      stakes: '底注',
      min_game_currency: '限入',
      commission_rate: '抽水 %',
    }
  },
  configName: [{
    id: [1, 5, 18, 19, 33, 36, 41, 42],
    list: {
      name: '房间名',
      stakes: '底注',
      min_game_currency: '限入',
      max_game_currency: '限红',
      commission_rate: '抽水 %',
    }
  }, {
    id: [6, 43],
    list: {
      name: '房间名',
      stakes: '底注',
      min_game_currency: '限入',
      max_game_currency: '最大带入',
      commission_rate: '抽水 %',
    }
  },
  {
    id: [26],
    list: {
      name: '房间名',
      stakes: '最低红包',
      min_game_currency: '限入',
      max_game_currency: '红包个数',
      commission_rate: '抽水 %',
    }
  }
  ],

  /**
   *  配置token错误或过期的处理状态码
   */
  arrStatus: [401]
}
