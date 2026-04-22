 <template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in labelList" :key="item.label">
          <label>{{item.label}}</label>
          <Input v-model="item.value" :maxlength='50' placeholder="请输入" style="width: 180px" />
        </li>

        <div class="search">
          <Button type="primary">全部</Button>
          <Button type="primary">历史控制</Button>
          <Button type="primary">搜索</Button>
          <Button>重置</Button>
        </div>
      </ul>
    </Card>
    <br />
    <div style="text-align: right;;margin-bottom: 10px;">
      <Button type="primary" style="margin-right: 60px;" to="/website-add">批量新增</Button>
    </div>
    <Card>
      <tables ref="tables" v-model="tableData" :columns="columns" @on-delete="handleDelete" />
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
        { label: '玩家总控ID', value: '' },
        { label: '玩家昵称', value: '' },
        { label: '玩家代理ID', value: '' }
      ],
      columns: [
        { title: '', type: 'selection', width: 60, align: 'center' },
        { title: '玩家总控ID', key: 'id' },
        { title: '玩家代理ID', key: 'id' },
        {
          title: '对战类胜率',
          key: 'ratio',
          render (h, params) {
            return <sapn>{params.row.ratio + 3}%</sapn>
          }
        },
        {
          title: '下注类胜率',
          key: 'ratio',
          render (h, params) {
            return <sapn>{params.row.ratio}%</sapn>
          }
        },
        { title: '输赢', key: 'number' },
        { title: '剩余输赢', key: 'pay' },
        { title: '历史控制次数', key: 'num' },
        { title: '累计盈亏', key: 'number' },
        { title: '操作人', key: 'name' },
        { title: '最新操作时间', key: 'date' },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          button: [
            (h, params) => {
              return h(
                'Button',
                {
                  props: {
                    type: 'info',
                    size: 'small'
                  }
                },
                '修改'
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
