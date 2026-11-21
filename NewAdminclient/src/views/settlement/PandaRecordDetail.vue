<template>
  <div class="panda-detail">
    <section class="detail-section">
      <div class="section-title summary-title">游戏详情</div>
      <table class="summary-table">
        <tbody>
          <tr v-for="item in summary" :key="item.label">
            <th scope="row">{{ item.label }}</th>
            <td>{{ item.value }}</td>
          </tr>
        </tbody>
      </table>
    </section>

    <section v-if="columns.length || lines.length" class="detail-section">
      <div class="section-title">
        <i class="el-icon-caret-bottom" aria-hidden="true" />
        游戏结果：{{ record.typeText }}　奖金：{{ formatAmount(lineBonus) }}
      </div>
      <div v-if="columns.length" class="panda-board" aria-label="开奖盘面">
        <div v-for="(column, colIndex) in columns" :key="colIndex" class="panda-column">
          <img v-for="(src, rowIndex) in column" :key="rowIndex" :src="cardUrl(src)" alt="盘面符号" />
        </div>
      </div>
      <div class="line-table-wrap">
        <table class="line-table">
          <thead><tr><th>线</th><th>方向</th><th>符号</th><th>倍数</th><th>奖金</th></tr></thead>
          <tbody>
            <tr v-for="(line, index) in lines" :key="index">
              <td>
                <template v-if="!line.special && Number(line.ts) === 1">
                  {{ line.line_id }}
                  <svg v-if="linePoints(line)" class="line-shape" viewBox="0 0 50 30" role="img" :aria-label="'中奖线 ' + line.line_id">
                    <path d="M0 10H50 M0 20H50 M10 0V30 M20 0V30 M30 0V30 M40 0V30" stroke="#d9dfe7" stroke-width="0.6" />
                    <polyline :points="linePoints(line)" fill="none" stroke="#e55757" stroke-width="2" />
                  </svg>
                </template>
                <template v-else>特殊玩法</template>
              </td>
              <td>{{ line.direction }}</td>
              <td><div class="line-symbols"><img v-for="(src, symbolIndex) in lineSymbols(line)" :key="symbolIndex" :src="cardUrl(src)" alt="中奖符号" :class="{ 'scatter-symbol': line.special }" /></div></td>
              <td><span v-if="line.times">X {{ line.times }}</span></td>
              <td>{{ formatAmount(line.bonus) }}</td>
            </tr>
            <tr v-if="!lines.length"><td colspan="5" class="empty-lines">暂无数据</td></tr>
          </tbody>
        </table>
      </div>
    </section>

    <section v-if="bonusItems.length" class="detail-section">
      <div class="section-title">
        <i class="el-icon-caret-bottom" aria-hidden="true" />
        游戏结果：奖励游戏　奖金：{{ formatAmount(info.bonus_game_info.xiao_all_bonus) }}
      </div>
      <div v-for="(item, index) in bonusItems" :key="index" class="bonus-item">
        {{ item.bonus_type }}：{{ formatAmount(item.bonus_num) }}
      </div>
    </section>
  </div>
</template>

<script>
import { asImageList, formatAmount, pickLineInfos } from "./settlementHelpers";

// 参考 Panda 的 25 条赔付线；静态目录没有 l_img，用实际行位置绘制线形。
const payLines = [
  [1, 1, 1, 1, 1], [0, 0, 0, 0, 0], [2, 2, 2, 2, 2],
  [0, 1, 2, 1, 0], [2, 1, 0, 1, 2], [1, 0, 0, 0, 1],
  [1, 2, 2, 2, 1], [0, 0, 1, 2, 2], [2, 2, 1, 0, 0],
  [1, 2, 1, 0, 1], [1, 0, 1, 2, 1], [0, 1, 1, 1, 0],
  [2, 1, 1, 1, 2], [0, 1, 0, 1, 0], [2, 1, 2, 1, 2],
  [1, 1, 0, 1, 1], [1, 1, 2, 1, 1], [0, 0, 2, 0, 0],
  [2, 2, 0, 2, 2], [0, 2, 2, 2, 0], [2, 0, 0, 0, 2],
  [1, 2, 0, 2, 1], [1, 0, 2, 0, 1], [0, 2, 0, 2, 0],
  [2, 0, 2, 0, 2],
];

export default {
  name: "PandaRecordDetail",
  props: { record: { type: Object, required: true } },
  computed: {
    info() { return this.record.info; },
    summary() {
      return [
        { label: "游戏时间", value: this.record.time || "-" },
        { label: "游戏局号", value: this.record.recordId || "-" },
        { label: "游戏名称", value: "功夫熊猫" },
        { label: "线注", value: formatAmount(this.record.lineBets) },
        { label: "总投注", value: formatAmount(this.record.allBets) },
        { label: "总奖金", value: formatAmount(this.record.allBonus) },
      ];
    },
    columns() {
      const grids = this.info.grids || [];
      const visible = this.info.grid_show || [];
      return grids.map((column, index) => {
        if (!visible[index] || !visible[index].length) return column;
        // grid_show 使用从 1 开始的循环位置，整除时取最后一行。
        return visible[index].map(position => column[(Number(position) - 1) % column.length]);
      });
    },
    lines() { return pickLineInfos(this.info).filter(line => Number(line.bonus) > 0); },
    lineBonus() { return this.lines.reduce((sum, line) => sum + Number(line.bonus), 0); },
    bonusItems() { return this.info.bonus_game_info.bonus_info || []; },
  },
  methods: {
    formatAmount,
    lineSymbols(line) { return asImageList(line.symbol || line.img_url); },
    cardUrl(src) {
      // Panda 仅提供 40x40 资源，中奖表按 20px 缩放同一张图片。
      return typeof src === "string" ? src.replace("/2018/20x20/", "/2018/40x40/") : "";
    },
    linePoints(line) {
      const rows = payLines[Number(line.line_id) - 1];
      return rows ? rows.map((row, col) => `${col * 10 + 5},${row * 10 + 5}`).join(" ") : "";
    },
  },
};
</script>

<style scoped>
.panda-detail { padding: 24px 4%; background: #f0f0f0; color: #606266; font-size: 14px; }
.detail-section + .detail-section { margin-top: 10px; }
.section-title { padding: 12px 10px; background: #e2e6eb; color: #303133; font-weight: bold; line-height: 22px; }
.section-title i { margin-right: 10px; }
.summary-title { text-align: center; }
table { width: 100%; border-collapse: collapse; table-layout: fixed; background: #fff; }
th, td { padding: 12px 10px; border-bottom: 1px solid #ebeef5; font-weight: normal; overflow-wrap: anywhere; }
.summary-table th, .summary-table td { width: 50%; text-align: center; border: 1px solid #ebeef5; }
.summary-table tr:nth-child(even), .bonus-item:nth-child(odd) { background: #fafafa; }
.panda-board { display: flex; justify-content: center; padding: 10px 0; background: #cf9e9e; overflow-x: auto; }
.panda-column img { display: block; width: 40px; height: 40px; object-fit: contain; }
.line-table-wrap { overflow-x: auto; }
.line-table { min-width: 480px; }
.line-table th { color: #909399; font-weight: bold; }
.line-table th, .line-table td { text-align: left; }
.line-table th:first-child, .line-table td:first-child { width: 25%; text-align: center; }
.line-table th:nth-child(3) { width: 25%; }
.line-shape { width: 50px; height: 30px; vertical-align: middle; }
.line-symbols { display: flex; flex-wrap: wrap; }
.line-symbols img { width: 20px; height: 20px; object-fit: contain; }
.line-symbols img.scatter-symbol { width: 40px; height: 40px; }
.line-table .empty-lines { height: 36px; text-align: center; color: #909399; }
.bonus-item { padding: 12px 10px; background: #fff; }
</style>
