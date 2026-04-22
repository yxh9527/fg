 <template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in labelList" :key="item.label">
          <label>{{item.label}}</label>
          <Select v-model="item.value" style="width: 180px">
            <Option
              v-for="items in item.option"
              :value="items.value"
              :key="items.label"
            >{{ items.label }}</Option>
          </Select>
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
        { label: '游戏名称', value: '' },
        {
          label: '游戏平台',
          value: '',
          option: [
            { label: '全部', value: '' },
            { label: 'H5', value: '' },
            { label: 'APP', value: '' },
            { label: 'Android', value: '' },
            { label: 'IOS', value: '' }
          ]
        },
        {
          label: '游戏类型',
          value: '',
          option: [
            { label: '全部', value: '' },
            { label: '下注型', value: '' },
            { label: '对战型', value: '' }
          ]
        },
        {
          label: '游戏状态',
          value: '',
          option: [
            { label: '全部', value: '' },
            { label: '正常', value: '' },
            { label: '冻结', value: '' },
            { label: '未激活', value: '' }
          ]
        }
      ],
      columns: [
        { title: '游戏名称', key: 'game' },
        { title: '房间名', key: 'roomname' },
        { title: '游戏平台', key: 'plat' },
        { title: '游戏分类', key: 'mode' },
        {
          title: '房间总盈亏',
          key: 'pay',
          render (h, params) {
            return params.row.pay >= 0 ? (
              <span style="color:green">{params.row.pay * 11}</span>
            ) : (
              <span style="color:red">{params.row.pay * 3} </span>
            )
          }
        },
        {
          title: '今日房间盈亏',
          key: 'pay',
          render (h, params) {
            return params.row.pay >= 0 ? (
              <span style="color:green">{params.row.pay}</span>
            ) : (
              <span style="color:red">{params.row.pay} </span>
            )
          }
        },
        { title: '房间胜率', key: 'ratio' },
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
