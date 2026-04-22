<template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in req" :key="item.label">
          <label>{{ item.label }}</label>
          <Select
            v-if="item.key == 'type'"
            v-model="item.value"
            style="width: 180px"
          >
            <Option
              v-for="items in item.option"
              :value="items.value"
              :key="items.label"
              >{{ items.label }}</Option
            >
          </Select>
          <Input
            v-else
            v-model="item.value"
            :maxlength="50"
            placeholder="请输入"
            style="width: 180px"
          />
        </li>
        <li class="label-style">
          <label style="margin-right:5px">时间选择</label>
          <DatePicker
            type="datetime"
            style="width:175px"
            v-model="startTime"
            placeholder="选择开始时间"
          ></DatePicker
          >&emsp;—&emsp;
          <DatePicker
            type="datetime"
            style="width:175px"
            v-model="endTime"
            placeholder="选择结束时间"
          ></DatePicker>
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
import { getLogListData } from '@/api/data'
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
        {
          label: '日志类型',
          key: 'type',
          value: '',
          option: [{ label: '后台总控', value: 1 }]
        },
        { label: '操作人账号', key: 'name', value: '' }
      ],
      startTime: '',
      endTime: '',
      columns: [
        { title: '序号', width: 80, key: 'id', align: 'center' },
        { title: '来源', key: 'source', align: 'center' },
        { title: '操作', key: 'text', width: 400 },
        { title: 'IP地址', key: 'ip' },
        { title: 'url', key: 'url', width: 400 },
        {
          title: '操作时间',
          key: 'createTime',
          render (h, params) {
            return <span>{getDate(params.row.createTime * 1000)}</span>
          }
        },
        { title: '操作人', key: 'adminName' }
      ],
      tableData: [],
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
      let Data = {
        type: this.req[0].value,
        name: this.req[1].value,
        startTime: getDate(this.startTime),
        endTime: getDate(this.endTime),
        page: this.pageData.page,
        pageSize: this.pageData.pageSize
      }

      /**
       * 验证时间范围合法
       */
      if (Data.startTime && Data.endTime) {
        let startTime = new Date(Data.startTime).getTime()
        let endTime = new Date(Data.endTime).getTime()
        if (endTime - startTime <= 0) {
          this.$Message.error('开始时间不允许大于结束时间')
          return
        }
      }

      getLogListData(Data).then(res => {
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
      // 重置时间
      this.startTime = ''
      this.endTime = ''

      for (const i in this.req) {
        this.req[i].value = ''
      }
      let Data = [
        { page: this.pageData.page },
        { pageSize: this.pageData.pageSize }
      ]

      getLogListData(Data).then(res => {
        if (res.data.code == 200) {
          this.pageData.page = 1
          this.tableData = []
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
