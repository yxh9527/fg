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
      <tables v-model="tableData" :columns="columns" @on-delete="handleDelete" />
    </Card>
  </div>
</template>

<script>
import Tables from '_c/tables'
import { getTableData } from '@/api/data'
export default {
  name: 'gameManage',
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
        { title: '序号', width: 80, type: 'index' },
        { title: '代理标识', key: 'id' },
        { title: '代理昵称', key: 'sitename' },
        {
          title: '当日盈亏',
          key: 'pay',
          render (h, params) {
            return params.row.pay >= 0 ? (
              <span style="color:green">{params.row.pay}</span>
            ) : (
              <span style="color:red">{params.row.pay} </span>
            )
          }
        },
        { title: '总库存', key: 'number' },
        {
          title: '总盈亏',
          key: 'pay',
          render (h, params) {
            return params.row.pay >= 0 ? (
              <span style="color:green">{params.row.pay * 15}</span>
            ) : (
              <span style="color:red">{params.row.pay * 2} </span>
            )
          }
        },
        { title: '代理游戏', key: 'game' },
        {
          title: '操作',
          key: 'handle',
          width: 520,
          align: 'center',
          button: [
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small',
                    to: '/operManage-single'
                  },
                  style: {
                    marginRight: '8px',
                    padding: '0 20px'
                  }
                },
                '玩家单控'
              )
            },
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small',
                    to: '/operManage-game'
                  },
                  style: {
                    marginRight: '8px',
                    padding: '0 20px'
                  }
                },
                '游戏控制'
              )
            },
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small',
                    to: '/operManage-agency'
                  },
                  style: {
                    marginRight: '8px',
                    padding: '0 20px'
                  }
                },
                '代理总控'
              )
            },
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small'
                  },
                  style: {
                    marginRight: '8px',
                    padding: '0 20px'
                  }
                },
                '库存'
              )
            },
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small',
                    to: '/operManage-log'
                  },
                  style: {
                    marginRight: '8px',
                    padding: '0 20px'
                  }
                },
                '日志'
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
