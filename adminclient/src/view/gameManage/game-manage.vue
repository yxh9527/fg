<template>
  <div>
    <Card>
      <div>
        <label>游戏名称:</label>
        <Input
          style="width: 400px"
          placeholder="输入游戏名称"
          v-model="searchName"
        />
        <div class="search" style="margin-left: 20px">
          <Button type="primary" @click="handleSearch">搜索</Button>
        </div>
        <div style="float: right; margin-right: 20px">
          <Button type="success" @click="startAllGame">解冻所有游戏</Button>
        </div>
        <div style="float: right; margin-right: 20px">
          <Button type="error" @click="stopAllGame">冻结所有游戏</Button>
        </div>
      </div>
    </Card>
    <Card style="margin-top: 5px">
      <tables ref="tables" v-model="tableData" :columns="columns" />
      <Spin fix v-if="spinShow">
        <Icon type="ios-loading" size="48" class="demo-spin-icon-load"></Icon>
        <div>数据加载中</div>
      </Spin>
      <div class="pageStyle">
        <Page
          :total="pageData.total"
          :current="pageData.page"
          :page-size="pageData.pageSize"
          show-total
          @on-change="changePage"
        />
      </div>
    </Card>
    <Modal
      v-model="modalEdit.state"
      title="修改游戏房间配置"
      width="1200"
      :styles="{ top: '20px' }"
      class="modalEditBox"
      @on-ok="submitConfig"
    >
      <Row
        v-for="(item, index) in configInfo"
        :key="item.id"
        :gutter="16"
        style="padding-right: 50px"
      >
        <template v-for="items in Object.keys(item)">
          <i-col v-if="mapConfigTitle(items)" :key="items" class="config-item">
            <i-input
              :maxlength="50"
              :type="'text'"
              v-model="item[items]"
              :placeholder="mapConfigTitle(items)"
            >
              <span slot="prepend">{{ mapConfigTitle(items) }}</span>
            </i-input>
          </i-col>
        </template>
        <Button
          v-if="index"
          icon="md-remove"
          style="height: 26px; width: 26px"
          title="移除配置项"
          shape="circle"
          @click="removeConfigItem(index)"
        ></Button>
      </Row>
      <Button
        type="info"
        icon="md-add"
        style="margin-top: 10px"
        @click="addConfigItem"
        >添加配置项</Button
      >
    </Modal>
  </div>
</template>

<script>
import Tables from '_c/tables'
import { setting } from '@/config'
import {
  getGD,
  editGameData,
  editGameState,
  getSelectGames,
  generateHallCache
} from '@/api/data'
import axios from '@/libs/api.request'
import { getToken } from '@/libs/util'
export default {
  name: 'gameManage',
  components: {
    Tables
  },
  data () {
    return {
      robotConfig: {
        dealer_bet: {
          1: {
            rand_num: 80
          },
          2: {
            rand_num: 87
          },
          3: {
            rand_num: 90
          },
          4: {
            rand_num: 10
          }
        },
        player_bet: {
          1: {
            rand_num: 8
          },
          2: {
            rand_num: 16
          },
          3: {
            rand_num: 37
          },
          4: {
            rand_num: 58
          },
          5: {
            rand_num: 79
          },
          6: {
            rand_num: 21
          }
        }
      },
      searchName: '',
      robotConfigView: '',
      miniGames: [],
      originalGames: null,
      site: null,
      siteOption: [],
      agent: null,
      agentOption: [],
      games: JSON.parse(sessionStorage.getItem('games') || '[]'),
      req: [
        {
          label: '游戏名称',
          key: 'gameId',
          value: '',
          option: []
        }
      ],
      configInfo: [],
      modalEdit: {
        id: '',
        state: false,
        agentId: '',
        gameId: '',
        gameServerId: ''
      },
      columns: [
        { title: '游戏ID', key: 'number', align: 'center', minWidth: 95 },
        {
          title: '游戏名称',
          key: 'gameName',
          align: 'center',
          minWidth: 95,
          render (h, params) {
            if (params.row.nameZH.length > 0) {
              return (
                <span>{params.row.name + ' [' + params.row.nameZH + ']'}</span>
              )
            } else {
              return <span>{params.row.name}</span>
            }
          }
        },
        {
          title: '游戏状态',
          key: 'state',
          align: 'center',
          minWidth: 95,
          render (h, params) {
            if (params.row.state == 1) {
              return (
                <span style="color:green">正常</span>
              )
            } else if (params.row.state == 0) {
              return (
                <span style="color:red">未上架</span>
              )
            } else {
              return (
                <span style="color:red">维护</span>
              )
            }
          }
        },
        {
          title: '操作',
          key: 'handle',
          width: 250,
          align: 'center',
          button: [
            (h, params) => {
              return [
                h(
                  'Poptip',
                  {
                    props: {
                      transfer: true,
                      confirm: true,
                      placement: 'left',
                      title:
                        '您确定要' +
                        (params.row.state == 0 ? '上架' : '下架') +
                        '游戏吗?'
                    },
                    on: {
                      'on-ok': () => {
                        let data = {
                          agentId: params.row.agentId,
                          id: params.row.number,
                          isFrozen: params.row.state == 0 ? 1 : 0
                        }
                        editGameState(data).then((res) => {
                          if (res.data.code == 200) {
                            this.$nextTick(() => {
                              this.handleSearch()
                              this.$Message.success(res.data.msg)
                            })
                          } else if (res.data.code == 400) {
                            this.$Message.error(res.data.msg)
                          }
                        })
                      }
                    }
                  },
                  [
                    h(
                      'Button',
                      {
                        props: {
                          type: params.row.state == 2 ? 'error' : 'success',
                          size: 'small'
                        }
                      },
                      params.row.state == 0 ? '上架' : '下架'
                    )
                  ]
                )
              ]
            },
            (h, params) => {
              return [
                h(
                  'Poptip',
                  {
                    props: {
                      transfer: true,
                      confirm: true,
                      placement: 'left',
                      title:
                        '您确定要' +
                        (params.row.state == 1 ? '冻结' : '启用') +
                        '游戏吗?'
                    },
                    style: { margin: '0 10px' },
                    on: {
                      'on-ok': () => {
                        let data = {
                          agentId: params.row.agentId,
                          id: params.row.number,
                          isFrozen: params.row.state == 2 ? 1 : 2
                        }
                        editGameState(data).then((res) => {
                          if (res.data.code == 200) {
                            this.$nextTick(() => {
                              this.handleSearch()
                              this.$Message.success(res.data.msg)
                            })
                          } else if (res.data.code == 400) {
                            this.$Message.error(res.data.msg)
                          }
                        })
                      }
                    }
                  },
                  [
                    h(
                      'Button',
                      {
                        props: {
                          type: params.row.state == 2 ? 'error' : 'success',
                          size: 'small'
                        }
                      },
                      params.row.state == 1 ? '冻结' : '启用'
                    )
                  ]
                )
              ]
            }
          ]
        }
      ],
      tableData: [],
      editData: [],
      mainStartTime: '',
      mainEndTime: '',
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        total: 0
      },
      spinShow: false
    }
  },
  methods: {
    async getMiniGames () {
      if (iplatform == '--kaiyuan') {
        let iparams = {
          token: getToken()
        }
        let data = await axios.request({
          url: 'v2/game/config',
          method: 'post',
          params: iparams
        })
        if (data && data.data && data.data.code == 200) {
          this.originalGames = data.data.data
          this.miniGames = Object.keys(data.data.data.gameSwitch).filter(
            (x) => !!data.data.data.gameSwitch[x]
          )
        }
      }
    },

    async miniGamesChange (games) {
      let value = this.originalGames.gameSwitch
      Object.keys(value).map((key) => {
        value[key] = games.includes(key) ? 1 : 0
      })
      let iparams = {
        token: getToken(),
        value: this.originalGames
      }

      let data = await axios.request({
        url: 'v2/game/setConfig',
        method: 'post',
        params: iparams
      })
      if (data && data.data && data.data.code == 200) {
        this.getMiniGames()
        this.$Message.success('修改成功')
      }
    },
    /**
     * 批量冻结解冻
     */
    stopAllGame () {
      this.$Modal.confirm({
        title: '提示',
        content: '确定冻结所有游戏吗？',
        onOk: async () => {
          let iparams = {
            token: getToken()
          }

          await axios.request({
            url: 'v1/game/stopAll',
            method: 'post',
            params: iparams
          })
          await this.handleSearch()
        }
      })
    },

    startAllGame () {
      this.$Modal.confirm({
        title: '提示',
        content: '确定解冻所有游戏吗？',
        onOk: async () => {
          let iparams = {
            token: getToken()
          }

          await axios.request({
            url: 'v1/game/startAll',
            method: 'post',
            params: iparams
          })
          await this.handleSearch()
        }
      })
    },

    /**
     * 切换站点
     */
    siteChanged (siteId) {
      this.agent = 9999999
      this.agentOption = this.siteOption.find(
        (site) => site.id == siteId
      ).agentList
      this.agentOption.map((item) => {
        item.label = item.name
      })
      sessionStorage.setItem('agentVal', '')
    },

    // 最后选择了代理 才更改session里面的站点和代理
    setSession (id) {
      if (id) {
        sessionStorage.setItem('siteVal', this.site)
        sessionStorage.setItem('agentVal', id)
      }
    },

    mapConfigTitle (items) {
      return this.currentConfigTitle[items]
    },
    showGameConfig (params) {
      /**
       * 按照配置转换游戏配置
       */
      let gameConfig =
        setting.configName.find((config) => {
          return config.id.includes(params.row.gameServerId)
        }) || setting.defaultConfigName

      this.currentConfigTitle = gameConfig.list

      this.configInfo = JSON.parse(JSON.stringify(params.row.config))
      this.modalEdit.id = params.row.id
      this.modalEdit.state = true
      this.modalEdit.agentId = params.row.agentId
      this.modalEdit.gameId = params.row.gameId
      this.modalEdit.gameServerId = params.row.gameServerId
    },
    generateHallCachefun () {
      let Data = { agentId: sessionStorage.getItem('agentVal') }
      generateHallCache(Data).then((res) => {
        this.$Message.success(res.data.msg)
        setTimeout(() => {
          this.handleSearch()
        }, 500)
      })
    },
    setLinkage (i, key) {
      if (key == 'webId') {
        let [k, m] = [null, null]
        for (const j in this.req) {
          if (this.req[j].key == 'webId') k = j
          if (this.req[j].key == 'agentId') m = j
        }
        this.req[m].option = []
        this.req[m].value =
          sessionStorage.getItem('agentVal') != null
            ? Number(sessionStorage.getItem('agentVal'))
            : ''
        this.req[k].option.forEach((item) => {
          if (item.agentList.length > 0 && item.id == i) {
            this.req[m].option.push(...item.agentList)
          }
        })
      }
    },
    async handleSearch () {
      if (this.agent === 9999999 || (this.agent !== 0 && !this.agent)) {
        this.$Message.error('必须选择一个站点和代理')
        return
      }
      this.spinShow = true
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize }
      ]
      Data = [
        { agentId: this.agent === 9999999 ? undefined : this.agent },
        { webId: this.site },
        { name: this.searchName },
        ...Data
      ]
      // }
      this.req.map((item) => {
        if (item.value) {
          Data.push({
            [item.key]: item.value
          })
        }
      })
      getGD(Data).then(({ data }) => {
        this.spinShow = false
        if (data.code == 200) {
          this.tableData = []
          this.tableData = data.data.list
          this.pageData.total = data.data.total
        } else {
          this.$Message.error(data.code + ' ：&nbsp;' + data.msg)
        }
      })
    },
    handleAllSearch () {
      this.spinShow = true
      for (const i in this.req) {
        this.req[i].value = ''
      }
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize },
        { agentId: this.agent === 9999999 ? undefined : this.agent },
        { webId: this.site }
      ]
      if (this.agent === 9999999 || (this.agent !== 0 && !this.agent)) {
        this.$Message.error('必须选择一个站点和代理')
        return
      }
      getGD(Data).then(({ data }) => {
        this.spinShow = false
        if (data.code == 200) {
          this.tableData = []
          this.tableData = data.data
        } else {
          this.$Message.error(data.code + ' ：&nbsp;' + data.msg)
        }
      })
    },
    changePage (index) {
      this.pageData.page = index
      this.handleSearch()
    },
    submitConfig () {
      let Data = []
      if (
        Object.values(this.configInfo[this.configInfo.length - 1]).indexOf('') >
        -1
      ) {
        this.$Message.warning('配置项不能为空')
      } else {
        const configData = this.configInfo.map((item, index) => {
          item.id = index + 1
          return { [index + 1]: item }
        })
        if (this.configInfo.length > 0) {
          Data = {
            id: this.modalEdit.id,
            agentId: this.modalEdit.agentId,
            gameId: this.modalEdit.gameId,
            config: JSON.stringify(Object.assign(...configData))
          }
        }
        editGameData(Data).then((res) => {
          if (res.data.code == 200) {
            this.$Message.success('修改配置成功')
            this.handleSearch()
          } else {
            this.$Message.error('修改配置失败：' + res.data.msg)
          }
          this.configInfo = []
        })
      }
    },
    addConfigItem () {
      let newConfig = {
        id: this.configInfo.length + 1
      }
      Object.keys(this.currentConfigTitle).map((key) => {
        newConfig[key] = ''
      })
      this.configInfo.push(newConfig)
    },
    removeConfigItem (index) {
      if (index > 0) {
        this.configInfo.splice(index, 1)
      }
    },
    changeConfigName (name, gid = 1) {
      let names = null
      setting.configName.forEach((item, index) => {
        if (item.id.indexOf(gid) != -1) {
          names = { ...item.list }[name] || name
        }
      })
      if (names) {
        return names
      } else return { ...setting.configNameDefault }[name] || name
    },
    /**
     * 根据代理id查询游戏
     */
    selectGameByAgent (val) {
      this.req = [
        {
          key: 'gameId',
          label: '游戏名称',
          value: '',
          option: []
        }
      ]
      if (val === 9999999) {
        //  选择的全部则不执行下面的代码
        return
      }
      this.req.map((item) => {
        /**
         * 更新游戏列表
         */
        if (item.key == 'gameId') {
          getSelectGames(val).then((res) => {
            return res.data.data
          })
        }
      })
    },

    async getGameCommon () {
      let params = {
        token: getToken()
        // gameId:
      }
      let data = await axios.request({
        url: 'v2/matchAward/list',
        method: 'get',
        params
      })
      if (data && data.data && data.data.code == 200) {
        this.matchAwardData = data.data.data
      }
    },

    async loadGames () {
      let newArray = JSON.parse(sessionStorage.getItem('games'))
      this.games.option = []
      newArray.forEach((item) => {
        this.games.option.push({
          id: item.number,
          name: item.name,
          nameZH: item.nameZH
        })
      })
    }
  },
  mounted () {
    let sid = sessionStorage.getItem('siteVal') // 获取当前session存储的已选择的站点id
    let siteOption = JSON.parse(sessionStorage.getItem('siteOption') || '[]') // 获取当前session存储的站点列表数据
    let agent = sessionStorage.getItem('agentVal')
    this.siteOption = siteOption
    this.site = sid * 1
    this.agent = agent * 1
    // 把选择的站点赋值到页面选中开始
    siteOption &&
      siteOption.map((item, index) => {
        if (item.id == sid) {
          this.agentOption = item.agentList
        }
      })
    this.agentOption.map((item) => {
      item.label = item.name
    })
    this.getMiniGames()
    this.handleSearch()
    this.loadGames()
    this.req.map((item) => {
      /**
       * 更新游戏列表
       */
      if (item.key == 'gameId') {
        let newArray = this.games
        item.option.push({ id: 0, name: '全部', nameZH: '' }, ...newArray)
        item.option.map((m) => {
          if (m.nameZH != '') {
            m.label = m.name + ' [' + m.nameZH + ']'
          } else {
            m.label = m.name
          }
          return m
        })
      }
    })
  }
}
</script>

<style lang="less">
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
.maintain {
  font-size: 15px;
  padding: 0 10px;
  > div {
    text-align: left;
    margin-bottom: 20px;
    label {
      margin-right: 10px;
    }
    span.ivu-input-suffix {
      margin: 5px;
    }
  }
  > button {
    margin-right: 20px;
  }
}
.config-item {
  display: inline;
  line-height: 40px;
  > .ivu-input-wrapper {
    width: 200px;
    display: inline-table;
  }
  > label {
    display: inline-block;
    margin-left: 10px;
    margin-right: 5px;
  }
  .modalEditBox div.ivu-modal-content div.ivu-modal-body {
    min-height: 75px !important;
  }
}
</style>
