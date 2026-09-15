<template>
  <div class="page-shell">
    <el-card shadow="never" class="content-card">
      <div class="panel-head">
        <div>
          <div class="panel-kicker">Settlement</div>
          <div class="panel-title">注单详情</div>
          <div class="panel-note">按 fgServer 注单详情结构展示 slots / fruits 游戏详情。</div>
        </div>
      </div>
      <div class="toolbar-row">
        <div class="field-inline">
          <label>开始时间</label>
          <el-date-picker v-model="startTime" type="datetime" value-format="timestamp" />
        </div>
        <div class="field-inline">
          <label>结束时间</label>
          <el-date-picker v-model="endTime" type="datetime" value-format="timestamp" />
        </div>
        <div class="field-inline">
          <label>游戏</label>
          <el-select v-model="gameId" filterable clearable class="wide-select">
            <el-option
              v-for="item in gameOptions"
              :key="item.number"
              :label="item.label"
              :value="item.number"
            />
          </el-select>
        </div>
        <div class="field-inline">
          <label>玩家ID</label>
          <el-input v-model.trim="userId" clearable />
        </div>
        <div class="field-inline">
          <label>注单号</label>
          <el-input v-model.trim="officeNumber" clearable />
        </div>
        <div class="field-inline">
          <el-button type="primary" @click="searchFirstPage">搜索</el-button>
        </div>
      </div>
    </el-card>

    <el-card shadow="never" class="content-card">
      <div class="table-toolbar">
        <div>
          <div class="panel-kicker">Orders</div>
          <div class="panel-title">注单列表</div>
        </div>
        <div class="table-meta">共 {{ pageData.current }} 条</div>
      </div>
      <app-table :data="tableData" :columns="columns" :loading="loading" />
      <div class="pager-wrap">
        <el-pagination
          background
          layout="total, sizes, prev, pager, next, jumper"
          :current-page="pageData.page"
          :page-size="pageData.pageSize"
          :page-sizes="pageData.pageOpts"
          :total="pageData.current"
          @current-change="changePage"
          @size-change="changePageSize"
        />
      </div>
    </el-card>

    <el-dialog
      title="游戏详情"
      :visible.sync="detailVisible"
      width="70%"
      custom-class="settlement-detail-dialog"
      append-to-body
    >
      <settlement-record-dialog :row="detailRow" embedded />
      <span slot="footer"></span>
    </el-dialog>
  </div>
</template>

<script>
import AppTable from "@/components/AppTable.vue";
import { getGameData2, getSettlement } from "@/api/data";
import { setting } from "@/config";
import SettlementRecordDialog from "./SettlementRecordDialog.vue";
import { formatAmount, formatPlayedTime, parseMaybeJson } from "./settlementHelpers";

export default {
  name: "SettlementDetailPage",
  components: {
    AppTable,
    SettlementRecordDialog,
  },
  data() {
    return {
      loading: false,
      startTime: "",
      endTime: "",
      gameId: "",
      userId: "",
      officeNumber: "",
      gameOptions: [],
      tableData: [],
      detailVisible: false,
      detailRow: null,
      pageData: {
        current: 0,
        page: setting.page,
        pageSize: setting.pageSize,
        pageOpts: setting.pageOpts,
      },
    };
  },
  computed: {
    columns() {
      return [
        { title: "游戏ID", key: "gameId", width: 100, align: "center" },
        { title: "游戏名称", key: "gameName", minWidth: 160, align: "center" },
        { title: "玩家ID", key: "userId", minWidth: 120, align: "center" },
        { title: "局号", key: "roundID", minWidth: 180, align: "center" },
        {
          title: "对局时间",
          key: "playedDate",
          minWidth: 170,
          align: "center",
          render: (h, { row }) => h("span", formatPlayedTime(row.playedDate)),
        },
        { title: "货币", key: "currency", width: 100, align: "center" },
        {
          title: "有效下注",
          key: "bet",
          minWidth: 120,
          align: "center",
          render: (h, { row }) => h("span", formatAmount(row.bet)),
        },
        {
          title: "总输赢",
          key: "win",
          minWidth: 120,
          align: "center",
          render: (h, { row }) => {
            const value = Number(row.win || 0);
            return h(
              "span",
              { class: value > 0 ? "positive" : "negative" },
              formatAmount(value),
            );
          },
        },
        {
          title: "操作",
          type: "action",
          width: 100,
          buttons: [
            {
              label: "查看",
              onClick: (row) => this.openSettlementDetail(row),
            },
          ],
        },
      ];
    },
  },
  methods: {
    async initGames() {
      const response = await getGameData2();
      this.gameOptions = [{ number: 0, label: "全部" }].concat(
        (response.data.data || []).map((item) => ({
          ...item,
          label: item.nameZH ? `${item.name} [${item.nameZH}]` : item.name,
        })),
      );
    },
    buildQuery() {
      const params = {
        page: this.pageData.page,
        pageSize: this.pageData.pageSize,
      };
      if (this.userId) params.userId = this.userId;
      if (this.officeNumber) params.officeNumber = this.officeNumber;
      if (this.gameId) params.gameId = this.gameId;
      if (this.startTime) params.startTime = Number(this.startTime) / 1000;
      if (this.endTime) params.endTime = Number(this.endTime) / 1000;
      return params;
    },
    normalizeRow(item) {
      const row = { ...item };
      row.detail = parseMaybeJson(row.detail);
      return row;
    },
    async fetchList() {
      this.loading = true;
      try {
        const response = await getSettlement(this.buildQuery());
        const payload = response.data.data || {};
        this.tableData = (payload.data || []).map((item) => this.normalizeRow(item));
        this.pageData.current = payload.total || 0;
      } finally {
        this.loading = false;
      }
    },
    searchFirstPage() {
      this.pageData.page = 1;
      this.fetchList();
    },
    changePage(page) {
      this.pageData.page = page;
      this.fetchList();
    },
    changePageSize(size) {
      this.pageData.pageSize = size;
      this.pageData.page = 1;
      this.fetchList();
    },
    openSettlementDetail(row) {
      this.detailRow = row;
      this.detailVisible = true;
    },
  },
  async mounted() {
    if (this.$route.query.userId) this.userId = String(this.$route.query.userId);
    if (this.$route.query.on) this.officeNumber = String(this.$route.query.on);
    if (this.$route.query.gameId) this.gameId = Number(this.$route.query.gameId) || "";
    await this.initGames();
    await this.fetchList();
  },
};
</script>

<style scoped>
.wide-select {
  min-width: 280px;
}

:global(.settlement-detail-dialog) {
  width: min(1360px, calc(100vw - 48px)) !important;
  max-width: calc(100vw - 48px);
  margin: 0 auto !important;
  top: 50%;
  transform: translateY(-50%);
}

:global(.settlement-detail-dialog .el-dialog__body) {
  padding: 12px 16px 16px;
}
</style>
