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
          <Button type="primary">搜索</Button>
          <Button>重置</Button>
        </div>
      </ul>
    </Card>
    <br />
    <Card>
      <tables ref="tables" v-model="tableData" :columns="columns" />
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
        {
          label: '异常类型',
          value: '',
          option: [
            { label: '全部', value: '' },
            { label: '玩家数据异常', value: '' },
            { label: '代理数据异常', value: '' }
          ]
        },
        {
          label: '异常级别',
          value: '',
          option: [
            { label: '全部', value: '' },
            { label: '1', value: '' },
            { label: '2', value: '' },
            { label: '3', value: '' },
            { label: '4', value: '' },
            { label: '5', value: '' }
          ]
        }
      ],
      columns: [
        { title: '序号', width: 80, type: 'index' },
        { title: '规则类型', key: 'game' },
        { title: '级别', key: 'num', width: 100 },
        { title: '规则说明', key: 'ccontent', width: 280 },
        { title: '值', key: 'ratio' },
        { title: '提示原因', key: 'ctitle' },
        { title: '自动处理', key: 'state' },
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
  methods: {},
  mounted () {
    getTableData().then(res => {
      if (res.data.code == 200) {
        this.tableData = res.data
      } else {
        this.$Message.warning('没有异常数据')
      }
    })
  }
}
</script>

<style>
</style>
