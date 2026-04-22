 <template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in req" :key="item.label">
          <label>{{item.label}}</label>
          <Input
            v-model="item.value"
            v-if="item.key!='fbState'"
            placeholder="请输入"
            style="width: 180px"
            :maxlength='50'
          />
          <Select v-if="item.key=='fbState'" v-model="item.value" style="width: 150px;">
            <Option value="">全部</Option>
            <Option value="0">未处理</Option>
            <Option value="1">已处理</Option>
          </Select>
        </li>

        <div class="search">
          <Button type="primary" @click="searchAction">搜索</Button>
        </div>
      </ul>
    </Card>
    <br />
    <Card>
      <tables ref="tables" v-model="tableData" :columns="columns" />
      <Spin fix v-if="spinShow">
        <Icon type="ios-loading" size="48" class="demo-spin-icon-load"></Icon>
        <div>数据加载中</div>
      </Spin>
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
import { getFeedbackData, editFeedbackState } from '@/api/data'
import { setting } from '@/config'
import { getDate } from '@/libs/tools'
export default {
  name: 'gameManage',
  components: {
    Tables
  },
  inject: ['handleLogOut'],
  data () {
    return {
      req: [
        { label: '游戏关键词', key: 'gameKey', value: '' },
        // { label: '代理关键词', key: 'agentKey', value: '' },
        { label: '状态', key: 'fbState', value: '' }
      ],
      columns: [
        { title: '序号', width: 80, key: 'id', align: 'center' },
        {
          title: '接收时间',
          key: 'createTime',
          align: 'center',
          sortable: true,
          sortType: 'desc',
          width: 280,
          render (h, params) {
            return <span>{getDate(params.row.createTime * 1000)}</span>
          }
        },
        { title: '站点', key: 'webName', width: 120, align: 'center' },
        { title: '代理', key: 'agentName', width: 120, align: 'center' },
        { title: '游戏', key: 'gameName', width: 120, align: 'center' },
        // { title: '平台', key: 'platFormName', width: 120, align: 'center' },
        { title: '玩家ID', key: 'userId', width: 120, align: 'center' },
        { title: '消息内容', key: 'msg', width: 400, align: 'center' },
        {
          title: '状态',
          key: 'fbState',
          width: 120,
          align: 'center',
          render (h, params) {
            return params.row.state == 1 ? (
              <span style="color:green">已处理</span>

            ) : (
              <span style="color:red">未处理</span>
            )
          }
        },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          width: 250,
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
                        '您确定修改状态为' +
                        (params.row.state == 0 ? '已处理' : '未处理') +
                        '吗?'
                    },
                    style: { textAlign: 'left', zIndex: '99' },
                    on: {
                      'on-ok': () => {
                        let Data = {
                          agentId: params.row.agentId,
                          id: params.row.id,
                          state: params.row.state == 0 ? 1 : 0
                        }
                        editFeedbackState(Data).then(res => {
                          this.$nextTick(() => {
                            this.handleSearch()
                          })
                        })
                      }
                    }
                  },
                  [
                    h(
                      'Button',
                      {
                        props: {
                          type: params.row.state === 1 ? 'warning' : 'success',
                          size: 'small'
                        }
                      },
                      params.row.state === 1 ? '未处理' : '处理'
                    )
                  ]
                )
              ]
            }
          ]
        }
      ],
      tableData: [],
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      },
      spinShow: false
    }
  },
  methods: {
    handleSearch () {
      this.spinShow = true
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize },
        { agentId: sessionStorage.getItem('agentVal') }
      ]

      this.req.map(item => {
        if (item.value) {
          Data.push({ [item.key]: item.value })
        }
      })
      getFeedbackData(Data)
        .then(res => {
          this.spinShow = false
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
        .catch(err => {
          this.spinShow = false
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
    this.handleSearch()
  }
}
</script>

<style>
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
</style>
