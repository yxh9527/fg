<template>
  <div>
    <Card>
      <Button type="primary" style="margin-bottom: 10px" to="/site-message-add"
        >添加消息</Button
      >
      <tables ref="tables" v-model="tableData" :columns="columns" />
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
import { getMsgData } from '@/api/data'
import { setting } from '@/config'
import { getDate } from '@/libs/tools'
export default {
  name: 'gameMessage',
  components: {
    Tables
  },
  inject: ['handleLogOut'],
  data () {
    return {
      labelList: [
        {
          label: '邮件类型',
          value: '',
          option: []
        }
      ],
      columns: [
        // { title: "ID", key: "id" },
        { title: '消息标题', key: 'title' },
        { title: '消息内容', key: 'info', width: 280, align: 'center' },
        {
          title: '发布时间',
          key: 'createTime',
          width: 180,
          align: 'center',
          render (h, params) {
            return <span>{getDate(params.row.createTime * 1000)}</span>
          }
        },
        {
          title: '消息类型',
          key: 'msgType',
          render (h, params) {
            return params.row.msgType == 1 ? (
              <span>管理消息</span>
            ) : (
              <span>{params.row.msgType}</span>
            )
          }
        },
        { title: '备注', key: 'remarks' },
        { title: '接收人', key: 'receiveName' }
      ],
      tableData: [],
      editData: [],
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts
      }
    }
  },
  methods: {
    handleSearch () {
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize }
      ]
      getMsgData(Data).then(res => {
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
