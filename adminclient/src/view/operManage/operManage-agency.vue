 <template>
  <div>
    <Card>
      <Table :columns="columns1" :data="data1"></Table>
      <br />
      <div class style="text-align: right;margin-bottom: 10px;margin-right: 15px;">
        <h2 style="left: 15px;position: absolute;">控制规则</h2>
        <Button type="primary">玩家类型</Button>
        <Button type="primary" style="margin-left:10px;">增加新控制</Button>
      </div>
      <tables
        ref="tables"
        editable
        v-model="tableData"
        :columns="columns"
        @on-delete="handleDelete"
      />
    </Card>
  </div>
</template>

<script>
import Tables from '_c/tables'
import { getTableData } from '@/api/data'
export default {
  name: 'website',
  components: {
    Tables
  },
  data () {
    return {
      columns1: [
        {
          title: '站点',
          key: 'site'
        },
        {
          title: '代理标识',
          key: 'id'
        },
        {
          title: '代理昵称',
          key: 'nikename'
        },
        {
          title: '初始点数',
          key: 'money1'
        },
        {
          title: '剩余点数',
          key: 'money2'
        },
        {
          title: '总消耗',
          key: 'money',
          render (h, params) {
            return params.row.money > 0 ? (
              <span style="color:green">{params.row.money * 5}</span>
            ) : (
              <span style="color:red">{params.row.money * 2} </span>
            )
          }
        },
        {
          title: '当日消耗',
          key: 'money3',
          render (h, params) {
            return params.row.money >= 0 ? (
              <span style="color:green">{params.row.money * 5}</span>
            ) : (
              <span style="color:red">{params.row.money * 2} </span>
            )
          }
        },
        {
          title: '剩余库存点数',
          key: 'money4'
        },
        {
          title: '总玩家',
          key: 'num1'
        },
        {
          title: '当日下注玩家',
          key: 'num2'
        },
        {
          title: '总有效下注',
          key: 'money5'
        },
        {
          title: '总抽水分数',
          key: 'money6'
        }
      ],
      data1: [
        {
          site: '游戏网',
          id: 'LL',
          nikename: '乐乐网',
          money1: '1000000',
          money2: '900000',
          money: '100000',
          money3: '1000',
          money4: '90000',
          num1: '1100',
          num2: '100',
          money5: '231100',
          money6: '10100'
        }
      ],
      columns: [
        { title: '序号', type: 'index', width: 80 },
        { title: '控制级别', key: 'num', width: 150 },
        { title: '被控类型', key: 'mode', width: 200 },
        { title: '触发条件', key: 'ccontent', width: 500 },
        { title: '控制方式', key: 'ctitle' },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          width: 300,
          button: [
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small'
                  },
                  style: {
                    marginRight: '5px'
                  }
                },
                '修改'
              )
            },
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'error',
                    size: 'small'
                  }
                },
                '删除'
              )
            }
          ]
        }
      ],
      tableData: []
    }
  },
  methods: {
    handleDelete (params) {

    },
    exportExcel () {
      this.$refs.tables.exportCsv({
        filename: `table-${new Date().valueOf()}.csv`
      })
    }
  },
  mounted () {
    getTableData().then(res => {
      this.tableData = res.data
    })
  }
}
</script>

<style>
</style>
