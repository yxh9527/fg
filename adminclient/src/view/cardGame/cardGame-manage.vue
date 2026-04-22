 <template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in req" :key="item.label">
          <label>{{item.label}}</label>
          <Select
            v-model="item.value"
            @on-change="setLinkage($event,item.key)"
            style="width: 180px"
          >
            <Option v-for="items in item.option" :value="items.id" :key="items.id">{{ items.name }}</Option>
          </Select>
        </li>
        <div class="search">
          <Button type="primary" @click="searchAction">搜索</Button>
          <Button @click="handleAllSearch">重置</Button>
        </div>
      </ul>
    </Card>
    <br />
    <Card>
      <tables ref="tables" v-model="tableData" :columns="columns" />
      <div class="pageStyle">
        <Page
          :total="pageData.current"
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
      :styles="{top: '20px'}"
      class="modalEditBox"
      @on-ok="submitConfig"
    >
      <Row
        v-for="(item,index) in configInfo"
        :key="item.id"
        :gutter="16"
        style="padding-right: 50px;"
      >
        <i-col v-for="items in Object.keys(item)" :key="items" class="config-item">

      <i-input v-if="items!='id'" :maxlength='50' type='number'  v-model="item[items]" :placeholder="items">
            <span slot="prepend">{{changeConfigName(items,modalEdit.gameServerId)}}</span>
          </i-input>
        </i-col>
        <Button
          v-if="index"
          icon="md-remove"
          style="height: 26px;width: 26px;"
          title="移除配置项"
          shape="circle"
          @click="removeConfigItem(index)"
        ></Button>
      </Row>
      <Button
        type="info"
        icon="md-add"
        style="margin-top:10px;margin-left:15px"
        @click="addConfigItem"
      >添加配置项</Button>
    </Modal>
  </div>
</template>

<script>
import Tables from '_c/tables'
import { setting } from '@/config'
import {
  getGameData,
  editGameData,
  editGameState
} from '@/api/data'
export default {
  name: 'gameManage',
  components: {
    Tables
  },
  inject: ['handleLogOut'],
  data () {
    return {
      req: [
        {
          label: '站点',
          key: 'webId',
          value: '',
          option: []
        },
        {
          label: '代理',
          key: 'agentId',
          value: '',
          option: []
        },
        {
          label: '游戏分类',
          key: 'gameType',
          value: '',
          option: []
        },
        {
          label: '游戏名称',
          key: 'gameId',
          value: '',
          option: []
        },
        {
          label: '游戏平台',
          key: 'platformId',
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
        { title: '游戏ID', key: 'gameServerId', align: 'center' },
        { title: '站点', key: 'webName', align: 'center' },
        { title: '代理', key: 'agentName', align: 'center' },
        { title: '游戏分类', key: 'gameTypeName', align: 'center' },
        { title: '游戏名称', key: 'gameName', align: 'center' },
        { title: '游戏平台', key: 'gamePlatForm', align: 'center' },
        {
          title: '游戏状态',
          key: 'state',
          align: 'center',
          render (h, params) {
            return params.row.state == 1 ? (
              <span style="color:green">正常</span>
            ) : (
              <span style="color:red">维护</span>
            )
          }
        },
        {
          title: '查询',
          key: 'gameList',
          align: 'center',
          render (h, params) {
            return h(
              'Button',
              {
                props: {
                  type: 'info',
                  size: 'small',
                  to: '/cardGame-game-info'
                },
                style: {
                  marginRight: '5px'
                }
              },
              '战绩'
            )
          }
        },
        {
          title: '维护开始',
          key: 'mainStartTime',
          align: 'center',
          width: 180
        },
        { title: '维护结束', key: 'mainEndTime', align: 'center', width: 180 },
        {
          title: '操作',
          key: 'handle',
          width: 300,
          align: 'center',
          button: [
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'primary',
                    size: 'small'
                  },
                  style: {
                    marginRight: '5px'
                  },
                  on: {
                    click: () => {
                      // this.configInfo = params.row.config;
                      this.modalEdit.id = params.row.id
                      this.modalEdit.state = true
                      this.modalEdit.agentId = params.row.agentId
                      this.modalEdit.gameId = params.row.gameId
                      this.modalEdit.gameServerId = params.row.gameServerId
                    }
                  }
                },
                '房间配置'
              )
            },
            (h, params) => {
              return [
                h(
                  'Poptip',
                  {
                    props: {
                      transfer: true,
                      confirm: true,
                      placement: 'left-start',
                      // width: "800",
                      title:
                        '您确定要' +
                        (params.row.state == 1 ? '维护' : '恢复') +
                        '游戏吗?'
                    },
                    style: { textAlign: 'left' },
                    on: {
                      'on-ok': val => {
                        let Data = {
                          id: params.row.id,
                          agentId: params.row.agentId,
                          gameId: params.row.gameId,
                          state: params.row.state == 1 ? 2 : 1,
                          mainStartTime: this.mainStartTime
                            ? getDate(this.mainStartTime * 1000)
                            : '',
                          mainEndTime: this.mainEndTime
                            ? getDate(this.mainEndTime * 1000)
                            : ''
                        }
                        editGameData(Data).then(res => {
                          if (res.data.code == 200) {
                            this.$nextTick(() => {
                              this.handleSearch()
                            })
                          } else if (res.data.code == 400) {
                            this.$Message.error(res.data.msg)
                          }
                        })
                      }
                    }
                  },
                  [
                    params.row.state == 1
                      ? h(
                        'div',
                        {
                          props: {
                            className: 'api'
                          },
                          slot: 'title',
                          class: 'maintain'
                        },
                        [
                          h('div', [
                            h('label', '开始维护时间'),
                            h('DatePicker', {
                              props: {
                                type: 'datetime',
                                format: 'YYYY-MM-dd HH:mm',
                                value: this.mainStartTime
                              },
                              on: {
                                'on-change': val => {
                                  this.mainStartTime = val
                                }
                              }
                            })
                          ]),
                          h('div', [
                            h('label', '结束维护时间'),
                            h('DatePicker', {
                              props: {
                                type: 'datetime',
                                format: 'YYYY-MM-dd HH:mm',
                                value: this.mainEndTime
                              },
                              on: {
                                'on-change': val => {
                                  this.mainEndTime = val
                                }
                              }
                            })
                          ])
                        ]
                      )
                      : '',
                    h(
                      'Button',
                      {
                        props: {
                          type: params.row.state == 1 ? 'warning' : 'success',
                          size: 'small'
                        },
                        style: {
                          marginRight: '5px'
                        }
                      },
                      params.row.state == 1 ? '维护' : '恢复'
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
                        (params.row.isFrozen == 0 ? '冻结' : '启用') +
                        '游戏吗?'
                    },
                    style: { textAlign: 'left', zIndex: '99' },
                    on: {
                      'on-ok': () => {
                        let data = {
                          id: params.row.id,
                          isFrozen: params.row.isFrozen == 0 ? 1 : 0
                        }
                        editGameState(data).then(res => {
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
                          type: params.row.isFrozen == 0 ? 'error' : 'success',
                          size: 'small'
                        }
                      },
                      params.row.isFrozen == 0 ? '冻结' : '启用'
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
        pageSize: setting.pageSize
      }
    }
  },
  methods: {
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
        this.req[k].option.forEach(item => {
          if (item.agentList.length > 0 && item.id == i) {
            this.req[m].option.push(...item.agentList)
          }
        })
      }
    },
    handleSearch () {
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize }
      ]
      if (sessionStorage.getItem('siteVal')) {
        Data = [{ webId: Number(sessionStorage.getItem('siteVal')) }, ...Data]
      }
      if (sessionStorage.getItem('agentVal')) {
        Data = [
          { agentId: Number(sessionStorage.getItem('agentVal')) },
          ...Data
        ]
      }
      this.req.map(item => {
        if (item.value) {
          Data.push({
            [item.key]: item.value
          })
        }
      })
      getGameData(Data)
        .then(res => {
          if (res.data.code == 200) {
            this.tableData = res.data.data.data
            this.pageData.current = res.data.data.total
          } else {
            // 判断响应状态是否为Token失效，如果失效则执行退出函数并刷新页面。
            this.$nextTick(() => {
              if (setting.arrStatus.indexOf(res.data.code) != -1) {
                this.$Message.error(
                  res.data.code + ' ：&nbsp;' + res.data.msg + '请重新登录'
                )
                this.handleLogOut()
              } else {
                this.$Message.error(res.data.code + ' ：&nbsp;' + res.data.msg)
              }
            })
          }
        })
    },
    handleAllSearch () {
      for (const i in this.req) {
        this.req[i].value = ''
      }
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize }
      ]
      getGameData(Data).then(res => {
        if (res.data.code == 200) {
          this.tableData = []
          this.pageData.page = 1
          this.tableData = res.data.data.data
          this.pageData.current = res.data.data.total
        } else {
          this.$Message.error(res.data.code + ' ：&nbsp;' + res.data.msg)
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
        editGameData(Data).then(res => {
          if (res.data.code == 200) {
            this.$Message.success('修改配置成功')
          } else {
            this.$Message.error('修改配置失败：' + res.data.msg)
          }
          this.configInfo = []
        })
      }
    },
    addConfigItem () {
      let item = this.configInfo.length
      if (
        !this.configInfo.length ||
        Object.values(this.configInfo[item - 1]).indexOf('') == -1
      ) {
        if (setting.maxGameConfig.indexOf(this.modalEdit.gameServerId) != -1) {
          this.configInfo.push({
            id: this.configInfo.length + 1,
            name: '',
            stakes: '',
            min_game_currency: '',
            max_game_currency: '',
            commission_rate: '',
            avg_control_game_currency: ''
          })
        } else {
          this.configInfo.push({
            id: this.configInfo.length + 1,
            name: '',
            stakes: '',
            min_game_currency: '',
            commission_rate: '',
            avg_control_game_currency: ''
          })
        }
      } else {
        this.$Message.warning('配置项不能为空')
      }
    },
    removeConfigItem (index) {
      if (index > 0) {
        this.configInfo.splice(index, 1)
      }
    },
    changeConfigName (name, gid = 1) {
      for (const item of setting.configName) {
        if (item.id.indexOf(gid) != -1) return { ...item.list }[name] || name
        else return { ...setting.configNameDefault }[name] || name
      }
    },
    searchAction () {
      this.pageData.page = 1
      this.handleSearch()
    }
  },
  created () {
    window.self = this
  },
  mounted () {
    this.req.map(item => {
      if (item.key == 'webId') {
        item.option.push(
          ...Object.assign(JSON.parse(sessionStorage.getItem('siteOption')))
        )
        if (sessionStorage.getItem('siteVal')) {
          item.value = Number(sessionStorage.getItem('siteVal'))
          this.$nextTick(() => {
            this.setLinkage(item.value, 'webId')
          })
        }
      }
      if (item.key == 'agentId' && sessionStorage.getItem('agentVal')) {
        item.value = Number(sessionStorage.getItem('agentVal'))
      }
      if (item.key == 'platformId') {
        item.option.push(
          ...Object.assign(JSON.parse(sessionStorage.getItem('classOption')))
        )
      }
      if (item.key == 'gameType') {
        item.option.push(
          ...Object.assign(JSON.parse(sessionStorage.getItem('typeOption')))
        )
      }
      if (item.key == 'gameId') {
        item.option.push(
          ...Object.assign(JSON.parse(sessionStorage.getItem('gameOption')))
        )
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
