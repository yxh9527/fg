 <template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in labelList" :key="item.label">
          <label>{{item.label}}</label>
          <Input v-model="item.value" :maxlength='50' placeholder="请输入" style="width: 180px" />
        </li>
        <li class="label-style">
          <label style="margin-right:5px">创建时间</label>
          <DatePicker type="datetimerange" placeholder="选择开始至结束日期"></DatePicker>
        </li>

        <div class="search">
          <Button type="primary">全部</Button>
          <Button type="primary">搜索</Button>
          <Button>重置</Button>
        </div>
      </ul>
    </Card>
    <br />
    <Card>
      <Button type="primary" style="margin-bottom: 10px" to="/agency-add">创建代理商</Button>
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
      labelList: [
        { label: '代理标识', value: '' },
        { label: '代理昵称', value: '' },
        { label: '代理邮箱', value: '' }
      ],
      columns: [
        { title: '序号', type: 'index', width: 80 },
        { title: '站点标识', key: 'id', sortable: true },
        { title: '站点昵称', key: 'sitename', sortable: true },
        { title: '域名', key: 'url' },
        { title: '代理游戏', key: 'game' },
        { title: '生效时间', key: 'date' },
        {
          title: '状态',
          key: 'state',
          render (h, params) {
            return (
              {
                正常: <span style="color:green">{params.row.state}</span>,
                维护: <span style="color:red">{params.row.state}</span>,
                下线: <span style="color:brown">{params.row.state}</span>,
                即将上线: (
                  <span style="color:steelblue">{params.row.state}</span>
                )
              }[params.row.state] || <span>{params.row.state}</span>
            )
          }
        },
        { title: '备注', key: 'name' },
        {
          title: '操作',
          key: 'handle',
          width: 250,
          align: 'center',
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
                '编辑'
              )
            },
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'warning',
                    size: 'small'
                  },
                  style: {
                    marginRight: '5px'
                  }
                },
                '维护'
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
                '冻结'
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

<style lang="less">
</style>
