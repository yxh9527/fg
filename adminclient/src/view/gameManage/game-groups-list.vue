<template>
  <div>
    <Card>
      <Button type="primary">创建官群</Button>
    </Card>
    <br />
    <Card>
      <Table :columns="groupColumns" :data="groupListData">
        <!-- 操作按钮 -->
        <template slot="actions" slot-scope="{ row }">
          <Button type="info" size="small" style="margin-right:10px"
            >对局战绩</Button
          >
          <Button type="info" size="small" style="margin-right:10px"
            >配置</Button
          >
          <Button type="error" size="small" style="margin-right:10px"
            >冻结</Button
          >
          <Button type="success" size="small">启用</Button>
        </template>
        <!-- 操作按钮end -->
      </Table>
      <br />
      <center>
        <Page :current="currentPage" :page-size="30" :total="100" show-total />
      </center>
    </Card>
  </div>
</template>

<script>
export default {
  data () {
    return {
      groupColumns: [
        { title: '位置', key: '', align: 'center' },
        { title: '房间ID', key: '', align: 'center' },
        { title: '创建人', key: '', align: 'center' },
        { title: '群名称', key: '', align: 'center' },
        { title: '游戏类型', key: '', align: 'center' },
        { title: '准入金额', key: '', align: 'center' },
        { title: '公告', key: '', align: 'center' },
        { title: '玩家人数', key: '', align: 'center' },
        { title: '游戏状态', key: '', align: 'center' },
        { title: '操作', key: '', align: 'center', slot: 'actions', width: 290 }
      ],
      groupListData: [{}],
      currentPage: 1
    }
  },
  methods: {}
}
</script>
