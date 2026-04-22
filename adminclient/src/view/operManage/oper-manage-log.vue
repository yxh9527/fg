<template>
  <div>
    <Card>
      <ul>
        <template v-for="item in labelList">
          <li
            class="label-style"
            v-if="!(item.type == 'index' && labelList[0].value == 3)"
            :key="item.label"
          >
            <template v-if="item.type == 'index'">
              <label v-if="labelList[0].value == 1">昵称/ID</label>
              <label v-if="labelList[0].value == 2">游戏名称</label>
            </template>
            <template v-else>
              <label>{{ item.label }}</label>
            </template>

            <Select
              v-if="item.type == 'select'"
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
              v-if="item.type == 'index'"
              v-model="item.value"
              placeholder="请输入"
              :maxlength="50"
              style="width: 180px"
            />
          </li>
        </template>
        <div class="search">
          <Button type="primary" @click="searchAction">搜索</Button>
          <Button
            v-if="labelList[0].value != 3"
            @click="resetParams"
            >重置</Button
          >
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
import { getControlLogData } from '@/api/data'
import { setting } from '@/config'
import { getDate } from '@/libs/tools'
export default {
  name: 'gameManage',
  components: {
    Tables
  },
  props: ['agentId'],
  data () {
    return {
      labelList: [
        {
          label: '控制类型',
          value: 1,
          type: 'select',
          option: [
            { label: '玩家单控', value: 1 },
            { label: '游戏单控', value: 2 },
            { label: '代理总控', value: 3 }
          ]
        },
        { label: '玩家昵称', type: 'index', value: '' }
      ],
      originColums: null,

      columns: [
        { title: '序号', type: 'index', width: 80 },
        { title: '控制类型', key: 'controlTypeName', minWidth: 95 },
        { title: '玩家ID/游戏', key: 'gameId', minWidth: 95 },
        { title: '控制内容', key: 'text', minWidth: 95 },
        // { title: "对战类胜率", key: "battleWinPorb", minWidth: 95 },
        // { title: "下注类胜率", key: "betsWinPorb", minWidth: 95 },
        // { title: "输赢", key: "winLosing", minWidth: 95 },
        // { title: "剩余输赢", key: "surWinLosing", minWidth: 95 },
        { title: '操作人', key: 'adminName', minWidth: 95 },
        {
          title: '最新操作时间',
          key: 'createTime',
          minWidth: 125,
          render (h, params) {
            return <span>{getDate(params.row.createTime * 1000)}</span>
          }
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
    resetParams () {
      this.labelList[1].value = ''
    },
    handleSearch (isReset) {
      this.spinShow = true

      if (isReset) {
        this.labelList[1].value = ''
        this.pageData.page = setting.page
        this.pageData.pageSize = setting.pageSize
      }

      let Data = {
        contType: this.labelList[0].value,
        name: this.labelList[1].value,
        agentId: this.agentId,
        page: this.pageData.page,
        pageSize: this.pageData.pageSize
      }
      getControlLogData(Data).then(res => {
        this.spinShow = false
        this.tableData = []
        this.tableData.push(...res.data.data.data)
        this.pageData.current = res.data.data.total

        if (!this.originColums) {
          this.originColums = this.columns.concat()
        }
        this.columns = this.originColums.concat()

        // 根据当前控制类型改变显示内容
        switch (this.labelList[0].value) {
          case 1:
            this.columns.splice(
              2,
              1,
              { title: '玩家昵称', key: 'userName', minWidth: 95 },
              { title: '玩家ID', key: 'userId', minWidth: 95 }
            )
            break
          case 2:
            this.columns.splice(
              2,
              1,
              { title: '游戏名称', key: 'gameName', minWidth: 95 },
              { title: '游戏ID', key: 'gameId', minWidth: 95 }
            )
            break
          case 3:
            this.columns.splice(2, 1)
            break
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
