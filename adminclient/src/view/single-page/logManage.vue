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
        { label: '日志类型', value: '' },
        { label: '操作人', value: '' }
      ],
      columns: [
        { title: '序号', width: 80, type: 'index' },
        { title: '游戏名称', key: 'game' },
        { title: '来源', key: 'game' },
        { title: '操作', key: 'state' },
        { title: 'IP地址', key: 'id' },
        { title: 'url', key: 'url' },
        { title: '操作时间', key: 'date' },
        { title: '操作人', key: 'name' }
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
