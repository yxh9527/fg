<template>
  <div>
    <Card title="对局详情" style="margin-top: 10px">
      <tables ref="tables" v-model="tableData" :columns="columns" />
      <Spin fix v-if="spinShow">
        <Icon type="ios-loading" size="48" class="demo-spin-icon-load"></Icon>
        <div>数据加载中</div>
      </Spin>
    </Card>
    <Modal v-model="viewSettlementDetail" width="65%" title="游戏详情">
        <div>
          <iframe
            :src="this.viewGameDetailUrl"
            width="100%"
            height="500"
            frameborder="0"
          ></iframe>
        </div>
        <div slot="footer"></div>
      </Modal>
  </div>
</template>

<script>
import Tables from "_c/tables";
import { getDate } from "@/libs/tools";
import * as dayjs from "dayjs";
import {getGameServers,getQueryOrder } from "@/api/data";
export default {
  name: "detailManage",
  components: {
    Tables,
  },
  data() {
    let _this = this;
    return {
      spinShow: false,
      startDate: null,
      endDate: null,
      gameUrls: [],
      replays:[],
      viewSettlementDetail: false,
      viewGameDetailUrl: "",
      startDateRestrict: {
        disabledDate(date) {
          if (date.getTime() > dayjs(_this.endDate).unix() * 1000) {
            return true;
          }
          if (date.getTime() < Date.now() - 1000 * 60 * 60 * 24 * 30 * 6) {
            return true;
          }
        },
      },
      endDateRestrict: {
        disabledDate(date) {
          if (date.getTime() < dayjs(_this.startDate).unix() * 1000) {
            return true;
          }
          if (date.getTime() > Date.now()) {
            return true;
          }
        },
      },
      // 表列
      columns: [
        {
          title: "代理",
          key: "agentId",
          width: 80,
          render(h, params) {
            return (
              _this.siteOption &&
              _this.siteOption.map((item) => {
                if (item.agentList) {
                  return item.agentList.map((agent) => {
                    if (agent.id == params.row.agentId) {
                      return h("span", {}, agent.name);
                    }
                  });
                }
              })
            );
          },
        },
        {
          title: "游戏名称",
          key: "gameName",
          minWidth:250,
          render(h, params) {
            return h("span", {}, params.row.gameName);
          },
        },
        {
          title: "局号",
          key: "roundID",
          width: 180,
          render(h, params) {
            return <div>{params.row.roundID}</div>;
          },
        },
        { title: "用户ID", key: "userId", width: 80 },
        { title: "账号", key: "account" ,minWidth:80,},
        { title: "昵称", key: "nickName", width: 150 },
        { title: "Symbol", key: "symbol", width: 150 },
        {
          title: "状态",
          key: "complete",
          width: 120,
          render: (h, params) => {
            return params.row.complete ? (
              <span style="color:green">
                完成
              </span>
            ) : (
              <span style="color:red">未完成</span>
            );
          },
        },
        {
          title: "详情",
          align: "center",
          width: 80,
          render: (h, params) => {
            return h(
              "a",
              {
                on: {
                  click: () => {
                    _this.imgClick(params.row);
                  },
                },
              },
              "查看"
            );
          },
        },
        { title: "有效下注", key: "bet", width: 100 },
        {
          title: "返奖",
          key: "win",
          width: 100,
          render: (h, params) => {
            return params.row.win > 0 ? (
              <span style="color:green">
                {Number(params.row.win).toFixed(2)}
              </span>
            ) : (
              <span style="color:red">{Number(params.row.win).toFixed(2)}</span>
            );
          },
        },
        {
          title: "索引",
          key: "rowVersion",
          width: 220,
        },
        {
          title: "对局时间",
          key: "playedDate",
          width: 180,
          render(h, params) {
            return <span>{getDate(params.row.playedDate)}</span>;
          },
        },
      ],
      tableData: [],
    };
  },
  methods: {
    imgClick(row) {
      this.viewGameDetailUrl = this.replays[0] + "/share/" + row.hash;
      this.viewSettlementDetail = true;
    },
  },

  mounted() {
    /**
     * 填充页面的站点参数
     */
    getGameServers().then((data) => {
      if (data.data.data) this.replays = data.data.data.data.replays;
    });
    let params={
        token:this.$route.params.token,
        account:this.$route.params.account,
        roundId:this.$route.params.order,
    }
    getQueryOrder(params).then(({ data }) => {
      let result = data.data;
      if (data.code === 200) {
        this.tableData = [];
        this.tableData = (result && result.data) || [];
        this.tableData.map((item) => {
          item.detail = JSON.parse(item.detail || "{}");
        });
      } else {
        this.$Message.error("接口报错");
      }
      this.spinShow = false;
    })
  },
};
</script>

<style lang="less">
.label-style {
  font-size: 16px;
}
.pageStyle {
  margin-top: 20px;
  text-align: right;
}
.inlineForm {
  display: flex;
  align-items: center;
  padding: 10px 0;
}
</style>
