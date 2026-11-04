<template>
  <div class="page-shell">
    <el-card shadow="never" class="content-card">
      <div class="panel-head">
        <div>
          <div class="panel-kicker">Detail</div>
          <div class="panel-title">统计详情</div>
          <div class="panel-note">按代理拆分游戏统计，查看有效下注、税收和杀数。</div>
        </div>
        <div class="panel-actions">
          <span class="badge-inline">{{ rangeText }}</span>
        </div>
      </div>
      <div class="toolbar-row">
        <div class="field-inline">
          <label>代理选择</label>
          <el-select v-model="agentId" filterable placeholder="选择代理" @change="refreshPage">
            <el-option v-for="item in agentOption" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </div>
      </div>
    </el-card>

    <el-card shadow="never" class="content-card">
      <div class="table-toolbar">
        <div>
          <div class="panel-kicker">Breakdown</div>
          <div class="panel-title">游戏维度明细</div>
        </div>
        <div class="table-meta">共 {{ tableData.length }} 条记录</div>
      </div>
      <app-table :data="tableData" :columns="columns" :loading="loading" />
    </el-card>
  </div>
</template>

<script>
import dayjs from "dayjs";
import AppTable from "@/components/AppTable.vue";
import { getAgentGameDataAggs, getLinkageList, getPoolGroups } from "@/api/data";
import { getScheme } from "@/api/poolScheme";
import { calcKillRate, calcProfit, toFixedNumber, toNumber } from "./reportHelpers";

export default {
  name: "AgentAggsDetailPage",
  components: {
    AppTable,
  },
  data() {
    return {
      loading: false,
      tableData: [],
      poolGroupNames: {},
      agentOption: [],
      agentId: 0,
      requestSequence: 0,
    };
  },
  computed: {
    rangeText() {
      if (!this.$route.query.startTime || !this.$route.query.endTime) return "全部时间";
      return `${dayjs(Number(this.$route.query.startTime) * 1000).format("YYYY-MM-DD HH:mm")} - ${dayjs(
        Number(this.$route.query.endTime) * 1000
      ).format("YYYY-MM-DD HH:mm")}`;
    },
    columns() {
      const columns = [
        { title: "代理ID", key: "agentId", width: 90, align: "center" },
        { title: "代理名称", key: "agentName", minWidth: 160, align: "center" },
        { title: "游戏名称", key: "gameName", minWidth: 240, align: "center" },
        { title: "Symbol", key: "symbol", minWidth: 140, align: "center" },
        { title: "人次", key: "userNumber", minWidth: 100, align: "center" },
        { title: "局数", key: "gameNumber", minWidth: 100, align: "right" },
        {
          title: "有效下注",
          key: "effectiveBetsTotal",
          align: "right",
          minWidth: 120,
          render: (h, { row }) => h("span", toFixedNumber(row.effectiveBetsTotal)),
        },
        {
          title: "有效打码",
          key: "chipsTotal",
          align: "right",
          minWidth: 120,
          render: (h, { row }) => h("span", toFixedNumber(row.chipsTotal)),
        },
        {
          title: "返奖",
          key: "profitLossTotal",
          align: "right",
          minWidth: 120,
          render: (h, { row }) => h("span", toFixedNumber(row.profitLossTotal)),
        },
        {
          title: "盈亏",
          key: "profit",
          align: "right",
          minWidth: 120,
          render: (h, { row }) => {
            const value = calcProfit(row.effectiveBetsTotal, row.profitLossTotal);
            return h("span", { class: value >= 0 ? "positive" : "negative" }, value.toFixed(2));
          },
        },
        {
          title: "税收",
          key: "revenueTotal",
          align: "right",
          minWidth: 100,
          render: (h, { row }) => h("span", toFixedNumber(row.revenueTotal)),
        },
        {
          title: "杀数",
          key: "killRate",
          align: "right",
          minWidth: 100,
          render: (h, { row }) =>
            h("span", calcKillRate(calcProfit(row.effectiveBetsTotal, row.profitLossTotal), row.chipsTotal)),
        },
      ];
      const gameNameColumn = columns.find((column) => column.key === "gameName");
      gameNameColumn.render = (h, { row }) =>
        h("div", { style: { position: "relative", minHeight: "34px", paddingTop: row.poolGroupName ? "14px" : "0" } }, [
          row.poolGroupName
            ? h(
                "span",
                {
                  style: {
                    position: "absolute",
                    top: "0",
                    right: "0",
                    maxWidth: "100%",
                    overflow: "hidden",
                    color: "#2563eb",
                    fontSize: "11px",
                    lineHeight: "13px",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                  },
                },
                row.poolGroupName
              )
            : null,
          h("span", { style: { display: "block", overflowWrap: "anywhere" } }, row.gameName || "-"),
        ]);
      return columns;
    },
  },
  methods: {
    async getAgentGameAggs(agentId) {
      const sequence = ++this.requestSequence;
      this.loading = true;
      try {
        const [response] = await Promise.all([
          getAgentGameDataAggs([
            { agentId },
            { startTime: this.$route.query.startTime },
            { endTime: this.$route.query.endTime },
            { webId: this.$route.query.webId },
          ]),
          this.loadPoolGroupNames(agentId, sequence),
        ]);
        if (sequence !== this.requestSequence) return;
        this.tableData = (response.data.data.data || []).map((item) => ({
          ...item,
          poolGroupName: item.poolGroupName || this.poolGroupNames[item.symbol] || "",
          userNumber:
            item.userNumber !== undefined && item.userNumber !== null && item.userNumber !== ""
              ? item.userNumber
              : item.userTotal,
          gameNumber:
            item.gameNumber !== undefined && item.gameNumber !== null && item.gameNumber !== ""
              ? item.gameNumber
              : item.docCount,
          userTotal: toNumber(item.userTotal),
          docCount: toNumber(item.docCount),
        }));
      } finally {
        if (sequence === this.requestSequence) this.loading = false;
      }
    },
    async loadPoolGroupNames(agentId, sequence) {
      try {
        const requests = [getPoolGroups({ agentId, webId: this.$route.query.webId })];
        if (Number(agentId) === 0) {
          requests.push(getScheme({ webId: this.$route.query.webId }));
          if (Number(this.$route.query.webId) !== 0) requests.push(getScheme({ webId: 0 }));
        }
        const responses = await Promise.allSettled(requests);
        if (sequence !== this.requestSequence) return;
        const groups = responses.flatMap((response) => {
          if (response.status !== "fulfilled") return [];
          const data = response.value.data.data || {};
          return data.groups || [];
        });
        this.poolGroupNames = groups.reduce((names, group) => {
          (group.symbols || []).forEach((symbol) => {
            if (!names[symbol]) names[symbol] = group.name;
          });
          return names;
        }, {});
      } catch (error) {
        if (sequence === this.requestSequence) this.poolGroupNames = {};
      }
    },
    refreshPage() {
      this.getAgentGameAggs(this.agentId);
    },
    async initAgents() {
      let siteOption = JSON.parse(sessionStorage.getItem("siteOption") || "[]");
      if (!siteOption.length) {
        const response = await getLinkageList();
        siteOption = response.data.data || [];
        sessionStorage.setItem("siteOption", JSON.stringify(siteOption));
      }
      const site = siteOption.find((item) => item.id === Number(this.$route.query.webId));
      this.agentOption = site ? site.agentList || [] : [];
      this.agentId = Number(this.$route.query.agent) || (this.agentOption[0] && this.agentOption[0].id) || 0;
    },
  },
  async mounted() {
    await this.initAgents();
    await this.getAgentGameAggs(this.agentId);
  },
};
</script>
