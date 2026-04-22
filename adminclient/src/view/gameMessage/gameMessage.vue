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
        <li class="label-style">
          <label>时间选择</label>&nbsp;
          <DatePicker type="datetime" placeholder="选择开始时间"></DatePicker>&emsp;—&emsp;
          <DatePicker type="datetime" placeholder="选择结束时间"></DatePicker>
        </li>
        <div class="search">
          <Button type="primary">全部</Button>
          <Button type="primary">搜索</Button>
        </div>
      </ul>
    </Card>
    <br />
    <Card>
      <Button type="primary" style="margin-bottom: 10px" to="/gameMessage-add">添加消息</Button>
      <tables ref="tables" v-model="tableData" :columns="columns" @on-delete="handleDelete" />
    </Card>
  </div>
</template>

<script>
import Tables from '_c/tables'
import { getTableData } from '@/api/data'
export default {
  name: 'gameMessage',
  components: {
    Tables
  },
  data () {
    return {
      labelList: [
        {
          label: '邮件类型',
          value: '',
          option: [
            { label: '全部', value: '' },
            { label: '系统消息', value: '' },
            { label: '维护公告', value: '' }
          ]
        }
      ],
      columns: [
        { title: '序号', type: 'index', width: 80 },
        { title: '消息标题', key: 'ctitle' },
        { title: '消息内容', key: 'ccontent', width: 280 },
        { title: '发布时间', key: 'date' },
        { title: '收件代理', key: 'sitename' },
        { title: '游戏类型', key: 'mode' },
        { title: '游戏名称', key: 'game' },
        { title: '状态', key: 'state' },
        { title: '类型', key: 'mode' },
        { title: '备注', key: 'plat' },
        {
          title: '操作',
          key: 'handle',
          align: 'center',
          width: 250,
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
