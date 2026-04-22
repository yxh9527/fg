<template>
  <div>
    <Card>
      <h2 style="margin-bottom: 10px">玩家信息</h2>
      <Table :columns="userColumns" :data="userData"></Table>
    </Card>
    <br />
    <Card>
      <h2 style="margin-bottom: 10px">盈亏明细</h2>
      <div class="inlineForm" style="margin-bottom: 10px">
        <div>时间选择：</div>
        <div style="display: flex; align-items: center">
          <DatePicker
            v-model="startDate"
            :options="startDateRestrict"
            type="month"
            placeholder="起始月份"
            style="width: 110px"
            :clearable="false"
          ></DatePicker>
          <div style="padding: 0 10px">—</div>
          <DatePicker
            v-model="endDate"
            :options="endDateRestrict"
            type="month"
            placeholder="结束月份"
            style="width: 110px"
            :clearable="false"
          ></DatePicker>
        </div>
        <Button @click="agentSearch" style="margin-left: 20px" type="primary"
          >搜索</Button
        >
      </div>
      <Table :columns="agentColumns" :data="agentData"></Table>
    </Card>
  </div>
</template>

<script>
import axios from '@/libs/api.request'
import { getToken } from '@/libs/util'
import * as dayjs from 'dayjs'
import expandRow from './userStatementExpand'

export default {
  components: {},
  data () {
    let _this = this
    return {
      /**
       * 查询参数
       */
      startDate: new Date(),
      startDateRestrict: {
        disabledDate (date) {
          if (date.getTime() > _this.endDate.getTime()) {
            return true
          }
          if (date.getTime() < Date.now() - 1000 * 60 * 60 * 24 * 30 * 6) {
            return true
          }
        }
      },
      endDate: new Date(),
      endDateRestrict: {
        disabledDate (date) {
          if (date.getTime() < _this.startDate.getTime()) {
            return true
          }
          if (date.getTime() > Date.now()) {
            return true
          }
        }
      },
      site: null,
      siteOption: [],
      agent: null,
      agentOption: [],

      /**
       * 用户信息
       */
      userColumns: [
        {
          title: '玩家ID',
          key: 'web_name',
          align: 'center'
        },
        {
          title: '玩家昵称',
          key: 'web_name',
          align: 'center'
        },
        {
          title: '最近登录时间',
          key: 'web_name',
          align: 'center'
        },
        {
          title: '最近登录IP',
          key: 'web_name',
          align: 'center'
        },
        {
          title: '注单数量',
          key: 'web_name',
          align: 'center'
        },
        {
          title: '有效下注',
          key: 'web_name',
          align: 'center'
        },
        {
          title: '玩家盈亏',
          key: 'web_name',
          align: 'center'
        },
        {
          title: '状态',
          key: 'web_name',
          align: 'center'
        }
      ],
      /**
       * 用户信息表格数据
       */
      userData: [],

      /**
       * 表格配置
       */
      agentColumns: [
        {
          type: 'expand',
          width: 50,
          render: (h, params) => {
            return h(expandRow, {
              props: {
                row: params.row
              }
            })
          }
        },
        {
          title: '时间',
          key: 'web_name',
          align: 'center'
        },
        {
          title: '星期',
          align: 'center',
          key: 'nickName'
        },
        {
          title: '有效下注',
          align: 'center',
          key: 'agentId'
        },
        {
          title: '主单号',
          align: 'center',
          key: 'userTotal'
        },
        {
          title: '盈亏',
          align: 'center',
          key: 'eNumber'
        }
      ],

      /**
       * 表格数据
       */
      agentData: []
    }
  },
  methods: {
    /**
     * 获取玩家信息
     */

    getUserInfo () {
      let data = {
        id: this.id,
        agentId: sessionStorage.getItem('agentVal')
      }
      getPlayerInfoData(data).then((res) => {
        this.userData.push(res.data.data)
      })
    },

    /**
     * 代理查询
     */
    async agentSearch () {
      // 转换到月初月末时间
      let startTime, endTime
      startTime = dayjs(this.startDate.getTime()).startOf('month').unix()
      endTime = dayjs(this.endDate.getTime()).endOf('month').unix()

      let params = {
        token: getToken(),
        webId: this.site,
        agentId: this.agent,
        startTime: startTime,
        endTime: endTime
      }

      if (this.agent == 9999999) {
        params.agentId = null
      }

      let data = await axios.request({
        url: 'v1/stat/agent/info',
        method: 'get',
        params
      })

      if (data && data.data && data.data.code == 200) {
        data.data.data.map((x) => {
          x.pTotal = -x.pTotal
        })
        this.agentData = data.data.data
      }
    }
  },
  mounted () {}
}
</script>

<style scoped lang="less">
.inlineForm {
  display: flex;
  align-items: center;
}
</style>
