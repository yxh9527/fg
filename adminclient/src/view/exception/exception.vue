 <template>
  <div>
    <Card>
      <ul>
        <li class="label-style" v-for="item in labelList" :key="item.label">
          <label>{{item.label}}</label>
          <Input v-model="item.value" :maxlength='50' placeholder="请输入" style="width: 180px" />
        </li>
        <li class="label-style">
          <label style="margin-right:5px">时间选择</label>
          <DatePicker type="datetime" style="width:175px" placeholder="选择开始时间"></DatePicker>&emsp;—&emsp;
          <DatePicker type="datetime" style="width:175px" placeholder="选择结束时间"></DatePicker>
        </li>

        <div class="search">
          <Button type="primary">搜索</Button>
          <Button>重置</Button>
        </div>
      </ul>
    </Card>
    <br />
    <Card>
      <Button type="primary" style="margin-bottom: 10px" to="/exception-rule">规则管理</Button>
      <tables ref="tables" v-model="tableData" :columns="columns" />
    </Card>
  </div>
</template>

<script>
import Tables from '_c/tables'
import { setting } from '@/config'
import { getTableData } from '@/api/data'
export default {
  name: 'gameManage',
  components: {
    Tables
  },
  inject: ['handleLogOut'],
  data () {
    return {
      labelList: [
        { label: '异常类型', value: '' },
        { label: '异常级别', value: '' }
      ],
      columns: [
        { title: '序号', width: 80, type: 'index' },
        { title: '异常类型', key: 'game' },
        { title: '异常级别', key: 'num' },
        { title: '异常位置', key: 'plat' },
        { title: '异常原因', key: 'ctitle' },
        { title: '时间', key: 'date' },
        { title: '处理结果', key: 'state' },
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
                '确定'
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
    // $$ 需要修改
    getTableData()
      .then(res => {
        if (res.data.code == 200) {
          this.tableData = res.data
        } else {
          // 判断响应状态是否为Token失效，如果失效则执行退出函数并刷新页面。
          this.$nextTick(() => {
            if (setting.arrStatus.indexOf(res.data.code) != -1) {
              this.$Message.error(
                res.data.code + ' ：&nbsp;' + res.data.msg + '请重新登录'
              )
              this.handleLogOut()
            } else {
              this.$Message.warning('没有异常数据')
            }
          })
        }
      })
  }
}
</script>

<style>
</style>
