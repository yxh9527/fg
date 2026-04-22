<template>
  <div>
    <Card style="margin:10px 0;">
      <Table :columns="agentColumns" :data="agentData"></Table>
      <div style="margin-top:20px; text-align:center;">
        <Page :total="total" :page-size="pagesize" @on-change="currentChanged" />
      </div>
    </Card>
  </div>
</template>

<script>
import axios from '@/libs/api.request'
import { getToken } from '@/libs/util'

export default {
  components: {},
  props: ['id', 'st', 'et'],
  data () {
    return {
      /**
       * 表格配置
       */
      agentColumns: [
        {
          title: '日期',
          key: 'day',
          align: 'center'
        },
        {
          title: '代理',
          align: 'center',
          key: 'agent_name'
        },
        {
          align: 'center',
          title: '期间有效投注',
          key: 'eTotal',
          render (h, params) {
            let jsx = <span>{params.row.eTotal.toFixed(2)}</span>
            return jsx
          }
        },
        {
          align: 'center',
          title: '期间注单',
          key: 'eNumber'
        },
        {
          align: 'center',
          title: '盈亏',
          key: 'pTotal',
          render (h, params) {
            let jsx = (
              <div style="color:red">{params.row.pTotal.toFixed(2)}</div>
            )
            if (Number(params.row.pTotal) > 0) {
              jsx = (
                <div style="color:green">{params.row.pTotal.toFixed(2)}</div>
              )
            }
            return jsx
          }
        },
        {
          align: 'center',
          title: '杀数',
          key: '',
          render (h, params) {
            let jsx = (
              <div>{(params.row.pTotal / params.row.eTotal).toFixed(3)}</div>
            )
            return jsx
          }
        }
      ],

      /**
       * 表格数据
       */
      agentData: [],
      page: 1,
      total: 1,
      pagesize: 1
    }
  },
  methods: {
    /**
     * 切换分页
     */
    currentChanged (page) {
      this.page = page
      this.fetchGameList()
    },

    /**
     * 查询游戏数据
     */
    async fetchGameList () {
      let data = await axios.request({
        url: 'v2/stat/agent/detail',
        method: 'get',
        params: {
          token: getToken(),
          page: this.page,
          pageSize: 50,
          agentId: this.id,
          startTime: this.st,
          endTime: this.et
        }
      })

      if (data && data.data && data.data.code == 200) {
        data.data.data.data.map((x) => {
          x.pTotal = -x.pTotal
        })
        this.agentData = data.data.data.data
        this.total = data.data.data.total
        this.pagesize = data.data.data.page_size
      }
    }
  },
  mounted () {
    this.fetchGameList()
  }
}
</script>

<style scoped lang="less">
</style>
