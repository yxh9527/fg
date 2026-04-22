 <template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in req" :key="item.label">
          <label>{{ item.label }}</label>
          <Input
            v-model="item.value"
            placeholder="请输入"
            v-if="item.type == 'text'"
            style="width: 180px"
            :maxlength="50"
          />
          <Select
            v-model="item.value"
            v-if="item.type == 'select'"
            @on-change="setLinkage($event, item.key)"
            style="width: 180px"
          >
            <Option
              v-for="items in item.option"
              :value="items.id"
              :key="items.id"
              >{{ items.name }}</Option
            >
          </Select>
        </li>

        <li class="label-style">
          <Button
            type="primary"
            @click="searchAction"
            style="margin-right: 15px"
            >搜索</Button
          >
          <Button @click="handleAllSearch">重置</Button>
        </li>
      </ul>
    </Card>
    <br />
    <div
      class
      style="text-align: right; margin-bottom: 10px; margin-right: 15px"
    >
      <Button type="primary" size="small" @click="editBatchState(2)"
        >批量冻结</Button
      >
      <Button
        type="primary"
        size="small"
        style="margin-left: 10px"
        @click="editBatchState(1)"
        >批量解冻</Button
      >
    </div>
    <Card>
      <tables
        ref="tables"
        v-model="tableData"
        :columns="columns"
        @on-selection-change="onSelect"
      />
      <div class="pageStyle">
        <Page
          :total="pageData.current"
          :current="pageData.page"
          :page-size="pageData.pageSize"
          :page-size-opts="pageData.pageOpts"
          show-elevator
          show-sizer
          show-total
          title="输入页数后，按Enter键跳转页面"
          @on-change="changePage"
          @on-page-size-change="changePageSize"
        />
      </div>
    </Card>
  </div>
</template>

<script>
import Tables from '_c/tables'
import { setting } from '@/config'
import { getPlayerData, editPlayerState } from '@/api/data'
export default {
  name: 'players',
  components: {
    Tables
  },
  inject: ['viewAccess', 'handleLogOut'],
  data () {
    return {
      req: [
        {
          label: '站点',
          type: 'select',
          key: 'webId',
          value: '',
          option: []
        },
        {
          label: '代理',
          key: 'agentId',
          type: 'select',
          value: '',
          option: []
        },
        {
          label: '状态',
          key: 'gameType',
          type: 'select',
          value: '',
          option: []
        },
        { label: '玩家ID', key: 'userId', type: 'text', value: '' }
      ],
      columns: [
        { title: '', type: 'selection', width: 60, align: 'center' },
        {
          title: '玩家ID',
          key: 'userId',
          width: 85,
          align: 'center',
          tooltip: true
        },
        {
          title: '玩家昵称',
          key: 'nickName',
          width: 140,
          align: 'center',
          tooltip: true
        },
        {
          title: '站点',
          key: 'webName',
          width: 120,
          align: 'center',
          tooltip: true
        },
        {
          title: '所属代理',
          key: 'agentName',
          width: 120,
          align: 'center',
          tooltip: true
        },
        {
          title: '最近登录时间',
          key: 'logTime',
          width: 180,
          align: 'center',
          render (h, params) {
            return (
              <span>
                {new Date(params.row.logTime * 1000).toLocaleString('chinese', {
                  hour12: false
                })}
              </span>
            )
          }
        },
        { title: '余额', key: 'times', width: 120, align: 'center' },
        { title: '局数', key: 'totalNumber', align: 'center', width: 80 },
        { title: '有效下注', key: 'times', width: 120, align: 'center' },
        {
          title: '总盈亏',
          key: 'totalProfLoss',
          width: 150,
          align: 'center'
        },
        { title: '控制次数', key: 'times', width: 120, align: 'center' },
        {
          title: '状态',
          key: 'state',
          width: 80,
          align: 'center',
          render (h, params) {
            return params.row.state == 1 ? (
              <span style="color:green">正常</span>
            ) : (
              <span style="color:red">冻结</span>
            )
          }
        },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          width: 180,
          button: [
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small',
                    // to: "/players-record-" + params.row.userId,
                    target: '_blank'
                  },
                  style: {
                    marginRight: '5px'
                  },
                  on: {
                    click: () => {
                      let routeData = this.$router.resolve({
                        path: '/players-record-' + params.row.userI,
                        query: { agent: params.row.agentId }
                      })
                      window.open(routeData.href, '_blank')
                    }
                  }
                },
                '流水'
              )
            },
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'warning',
                    size: 'small',
                    // to: "/players-game-" + params.row.userId,
                    target: '_blank'
                  },
                  style: {
                    marginRight: '5px'
                  },
                  on: {
                    click: () => {
                      const params2 = { agent: params.row.agentId }
                      let routeData = this.$router.resolve({
                        path: '/players-game-' + params.row.userId,
                        query: params2
                      })
                      window.open(routeData.href, '_blank')
                    }
                  }
                },
                '战绩'
              )
            },
            (h, params) => {
              if (this.viewAccess) {
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
                          (params.row.state == 2 ? '解冻' : '冻结') +
                          '此玩家吗?'
                      },
                      style: { textAlign: 'left', zIndex: '99' },
                      on: {
                        'on-ok': () => {
                          let data = {
                            id: JSON.stringify([params.row.id]),
                            state: params.row.state == 2 ? 1 : 2
                          }
                          editPlayerState(data).then((res) => {
                            if (res.data.code == 200) {
                              this.$nextTick(() => {
                                this.handleSearch()
                                this.$Message.success(res.data.msg)
                              })
                            } else {
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
                            type: params.row.state == 1 ? 'error' : 'success',
                            size: 'small'
                          }
                        },
                        params.row.state == 1 ? '冻结' : '解冻'
                      )
                    ]
                  )
                ]
              }
            }
          ]
        }
      ],
      tableData: [],
      userSelect: [],
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      }
    }
  },
  methods: {
    onSelect (sel) {
      this.userSelect = sel.map((item) => {
        return item.id
      })
    },
    editBatchState (status) {
      let Data = { id: JSON.stringify(this.userSelect), state: status }
      editPlayerState(Data).then((res) => {
        if (res.data.code == 200) this.$Message.success('批量修改状态成功')
        this.$nextTick(() => {
          this.handleSearch()
        })
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
      this.req.map((item) => {
        if (item.value !== '') {
          Data.push({
            [item.key]: item.value
          })
        }
      })
      getPlayerData(Data).then((res) => {
        if (res.data.code == 200) {
          this.tableData = []
          this.tableData.push(...res.data.data.data)
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
      getPlayerData(Data).then((res) => {
        if (res.data.code == 200) {
          this.tableData = []
          this.pageData.page = 1
          this.tableData.push(...res.data.data.data)
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
    changePageSize (index) {
      this.pageData.pageSize = index
      this.handleSearch()
    },
    searchAction () {
      this.pageData.page = 1
      this.handleSearch()
    }
  },
  mounted () {
    this.req.map((item) => {
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
      if (item.key == 'gameType') {
        item.option.push(
          ...Object.assign(JSON.parse(sessionStorage.getItem('typeOption')))
        )
      }
    })
  }
}
</script>

<style lang="less">
.search {
  display: inline;
  > button {
    margin-right: 10px;
    padding-left: 20px;
    padding-right: 20px;
  }
}
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
</style>
