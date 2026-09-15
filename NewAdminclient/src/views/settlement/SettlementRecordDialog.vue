<template>
  <div class="settlement-record-dialog" :class="{ embedded: embedded }">
    <template v-if="normalized.info">
      <div class="summary-grid">
        <div
          v-for="item in summaryItems"
          :key="'summary-' + item.label"
          class="summary-item"
        >
          <div class="summary-label">{{ item.label }}</div>
          <div class="summary-value">{{ item.value }}</div>
        </div>
      </div>

      <!-- Slots：盘面 + 中奖线 -->
      <template v-if="mode === 'slot'">
        <div class="section-title">开奖盘面</div>
        <div v-if="slotColumns.length" class="grid-board" :style="slotBoardStyle">
          <div
            v-for="(col, colIndex) in slotColumns"
            :key="'col-' + colIndex"
            class="grid-column"
          >
            <img
              v-for="(src, rowIndex) in col"
              :key="'cell-' + colIndex + '-' + rowIndex"
              class="grid-icon"
              :src="src"
              alt=""
            />
          </div>
        </div>
        <el-alert
          v-else
          title="该 slots 注单没有可用的 grids 盘面图片路径。"
          type="warning"
          :closable="false"
          show-icon
        />

        <div v-if="lineInfos.length" class="section-title">中奖线</div>
        <div v-if="lineInfos.length" class="line-list">
          <div
            v-for="(line, index) in lineInfos"
            :key="'line-' + index"
            class="line-card"
          >
            <div class="line-head">
              <span>线 {{ line.line_id || index + 1 }}</span>
              <span>{{ line.direction || "" }}</span>
              <span>x{{ line.times || line.ts || 1 }}</span>
              <span class="positive">+{{ formatAmount(line.bonus) }}</span>
            </div>
            <div class="line-body">
              <img
                v-if="line.line_shape_url"
                class="line-shape"
                :src="line.line_shape_url"
                alt=""
              />
              <div class="line-symbols">
                <img
                  v-for="(symbol, symbolIndex) in asImageList(line.symbol)"
                  :key="'symbol-' + index + '-' + symbolIndex"
                  class="line-symbol"
                  :src="symbol"
                  alt=""
                />
              </div>
              <div class="line-meta">投注 {{ formatAmount(line.bets) }}</div>
            </div>
          </div>
        </div>
      </template>

      <!-- Fruit：开奖图 / 盘面 / 下注区 -->
      <template v-else-if="mode === 'fruit'">
        <div v-if="rewardTexts.length" class="section-title">开奖结果</div>
        <div v-if="rewardTexts.length" class="reward-tags">
          <el-tag
            v-for="(text, index) in rewardTexts"
            :key="'reward-' + index"
            type="info"
          >
            {{ text }}
          </el-tag>
        </div>

        <div v-if="resultShowImages.length" class="section-title">开奖图标</div>
        <div v-if="resultShowImages.length" class="image-row">
          <img
            v-for="(src, index) in resultShowImages"
            :key="'result-' + index"
            class="result-icon"
            :src="src"
            alt=""
          />
        </div>

        <div v-if="resultGridMatrix.length" class="section-title">开奖盘面</div>
        <div v-if="resultGridMatrix.length" class="matrix-board">
          <div
            v-for="(row, rowIndex) in resultGridMatrix"
            :key="'result-row-' + rowIndex"
            class="matrix-row"
          >
            <img
              v-for="(src, colIndex) in row"
              :key="'result-cell-' + rowIndex + '-' + colIndex"
              class="matrix-icon"
              :src="src"
              alt=""
            />
          </div>
        </div>

        <div v-if="specificBets.length" class="section-title">下注详情</div>
        <div v-if="specificBets.length" class="bet-list">
          <div
            v-for="(bet, index) in specificBets"
            :key="'bet-' + index"
            class="bet-card"
          >
            <div class="bet-head">
              <img
                v-if="isImageUrl(bet.bet_area_url)"
                class="bet-area"
                :src="bet.bet_area_url"
                alt=""
              />
              <div class="bet-head-text">
                <div>下注 {{ formatAmount(bet.bet_num) }}</div>
                <div>
                  赔率 {{ formatRate(bet.reward_rate) }} /
                  派彩 {{ formatAmount(bet.reward_money) }}
                </div>
              </div>
            </div>
            <div
              v-if="betGridMatrix(bet).length"
              class="matrix-board compact"
            >
              <div
                v-for="(row, rowIndex) in betGridMatrix(bet)"
                :key="'bet-grid-' + index + '-' + rowIndex"
                class="matrix-row"
              >
                <img
                  v-for="(src, colIndex) in row"
                  :key="'bet-cell-' + index + '-' + rowIndex + '-' + colIndex"
                  class="matrix-icon"
                  :src="src"
                  alt=""
                />
              </div>
            </div>
          </div>
        </div>

        <div v-if="fruitExtraItems.length" class="section-title">附加信息</div>
        <div v-if="fruitExtraItems.length" class="summary-grid">
          <div
            v-for="item in fruitExtraItems"
            :key="'extra-' + item.label"
            class="summary-item"
          >
            <div class="summary-label">{{ item.label }}</div>
            <div class="summary-value">{{ item.value }}</div>
          </div>
        </div>
      </template>

      <template v-else>
        <el-alert
          title="当前详情结构未识别为 slots/fruits 标准注单，已展示原始摘要。"
          type="warning"
          :closable="false"
          show-icon
        />
        <pre class="raw-json">{{ rawInfoJson }}</pre>
      </template>
    </template>

    <el-empty v-else description="没有可展示的注单详情数据" />
  </div>
</template>

<script>
import config from "@/config";
import {
  asImageList,
  asImageMatrix,
  formatAmount,
  isFruitDetail,
  isSlotDetail,
  normalizeSettlementRow,
  pickLineInfos,
} from "./settlementHelpers";

export default {
  name: "SettlementRecordDialog",
  props: {
    row: {
      type: Object,
      default: null,
    },
    embedded: {
      type: Boolean,
      default: false,
    },
    ossUrl: {
      type: String,
      default: "",
    },
  },
  computed: {
    assetBase() {
      return this.ossUrl || config.ossUrl || "";
    },
    normalized() {
      return normalizeSettlementRow(this.row || {}, this.assetBase);
    },
    mode() {
      const info = this.normalized.info;
      if (isSlotDetail(info)) return "slot";
      if (isFruitDetail(info, this.normalized.gameId)) return "fruit";
      return "unknown";
    },
    summaryItems() {
      const item = this.normalized;
      const info = item.info || {};
      const rows = [
        { label: "游戏", value: item.gameName || item.gameId || "-" },
        { label: "游戏ID", value: item.gameId || "-" },
        { label: "局号", value: item.recordId || "-" },
        { label: "时间", value: item.time || "-" },
        { label: "类型", value: item.typeText || "-" },
        { label: "投注", value: formatAmount(item.allBets) },
        { label: "派彩", value: formatAmount(item.allBonus) },
      ];
      if (item.lineBets !== "" && item.lineBets != null) {
        rows.splice(5, 0, { label: "线注", value: formatAmount(item.lineBets) });
      }
      if (item.jpBonus !== "" && item.jpBonus != null && Number(item.jpBonus) !== 0) {
        rows.push({ label: "JP", value: formatAmount(item.jpBonus) });
      }
      if (item.currency) {
        rows.push({ label: "货币", value: item.currency });
      }
      if (item.playerName) {
        rows.push({ label: "玩家", value: item.playerName });
      }
      if (info.times != null && info.times !== "") {
        rows.push({ label: "倍数", value: String(info.times) });
      }
      return rows;
    },
    slotColumns() {
      const info = this.normalized.info || {};
      return Array.isArray(info.grids) ? info.grids : [];
    },
    slotBoardStyle() {
      const count = Math.max(this.slotColumns.length, 1);
      return {
        gridTemplateColumns: "repeat(" + count + ", minmax(48px, 1fr))",
      };
    },
    lineInfos() {
      return pickLineInfos(this.normalized.info);
    },
    rewardTexts() {
      const info = this.normalized.info || {};
      const rewards = info.reward_result;
      if (!Array.isArray(rewards)) return [];
      return rewards
        .map(function (item) {
          if (typeof item === "string" || typeof item === "number") return String(item);
          return "";
        })
        .filter(Boolean);
    },
    resultShowImages() {
      const info = this.normalized.info || {};
      return asImageList(info.result_show_url);
    },
    resultGridMatrix() {
      const info = this.normalized.info || {};
      return asImageMatrix(info.result_grid_url);
    },
    specificBets() {
      const info = this.normalized.info || {};
      return Array.isArray(info.specific_bet_info) ? info.specific_bet_info : [];
    },
    fruitExtraItems() {
      const info = this.normalized.info || {};
      const rows = [];
      if (info.bet_num != null && info.bet_num !== "") {
        rows.push({ label: "弹珠数/注数", value: String(info.bet_num) });
      }
      if (info.risk) {
        rows.push({ label: "风险", value: String(info.risk) });
      }
      if (info.num != null && info.num !== "") {
        rows.push({ label: "档位", value: String(info.num) });
      }
      if (info.level != null && Number(info.level) !== 0) {
        rows.push({ label: "等级", value: String(info.level) });
      }
      if (info.rate != null && info.rate !== "") {
        rows.push({ label: "倍率", value: String(info.rate) });
      }
      if (info.reward_money != null && info.reward_money !== "") {
        rows.push({ label: "奖励金额", value: formatAmount(info.reward_money) });
      }
      return rows;
    },
    rawInfoJson() {
      return JSON.stringify(this.normalized.info || {}, null, 2);
    },
  },
  methods: {
    formatAmount: formatAmount,
    asImageList: asImageList,
    betGridMatrix: function (bet) {
      if (!bet || !bet.grid) return [];
      return asImageMatrix(bet.grid);
    },
    isImageUrl: function (value) {
      return (
        typeof value === "string" &&
        (/^(https?:)?\/\//.test(value) || value.indexOf("/") === 0)
      );
    },
    formatRate: function (value) {
      const number = Number(value);
      if (!Number.isFinite(number) || number === 0) return formatAmount(0);
      if (Math.abs(number) >= 10) return (number / 100).toFixed(2);
      return formatAmount(number);
    },
  },
};
</script>

<style scoped>
.settlement-record-dialog {
  min-height: 120px;
}

.settlement-record-dialog.embedded {
  max-height: calc(100vh - 220px);
  overflow: auto;
  padding-right: 4px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 10px;
  margin-bottom: 16px;
}

.summary-item {
  padding: 10px 12px;
  border: 1px solid var(--line-color);
  border-radius: 10px;
  background: #f8fafc;
}

.summary-label {
  color: var(--text-sub);
  font-size: 12px;
}

.summary-value {
  margin-top: 4px;
  color: var(--text-main);
  font-size: 14px;
  font-weight: 600;
  word-break: break-all;
}

.section-title {
  margin: 18px 0 10px;
  color: var(--text-main);
  font-size: 15px;
  font-weight: 700;
}

.grid-board {
  display: grid;
  gap: 8px;
  width: fit-content;
  max-width: 100%;
  padding: 12px;
  border: 1px solid var(--line-color);
  border-radius: 12px;
  background: #0f172a;
}

.grid-column {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.grid-icon,
.matrix-icon,
.result-icon,
.bet-area,
.line-symbol,
.line-shape {
  display: block;
  object-fit: contain;
  background: rgba(255, 255, 255, 0.04);
  border-radius: 6px;
}

.grid-icon {
  width: 48px;
  height: 48px;
}

.matrix-board {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: fit-content;
  max-width: 100%;
  padding: 12px;
  border: 1px solid var(--line-color);
  border-radius: 12px;
  background: #0f172a;
  overflow: auto;
}

.matrix-board.compact {
  margin-top: 10px;
  padding: 8px;
}

.matrix-row {
  display: flex;
  gap: 6px;
}

.matrix-icon {
  width: 40px;
  height: 40px;
}

.image-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.result-icon {
  width: 56px;
  height: 56px;
}

.reward-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.line-list,
.bet-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.line-card,
.bet-card {
  padding: 12px;
  border: 1px solid var(--line-color);
  border-radius: 12px;
  background: #fff;
}

.line-head,
.bet-head {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.line-head {
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 13px;
  color: var(--text-sub);
}

.line-body {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.line-shape {
  width: 72px;
  height: 36px;
}

.line-symbols {
  display: flex;
  gap: 4px;
}

.line-symbol {
  width: 28px;
  height: 28px;
}

.line-meta {
  color: var(--text-sub);
  font-size: 12px;
}

.bet-area {
  width: 48px;
  height: 48px;
}

.bet-head-text {
  color: var(--text-main);
  font-size: 13px;
  line-height: 1.5;
}

.positive {
  color: #16a34a;
  font-weight: 700;
}

.raw-json {
  margin-top: 12px;
  padding: 12px;
  border-radius: 10px;
  background: #0f172a;
  color: #e2e8f0;
  font-size: 12px;
  overflow: auto;
  max-height: 360px;
}
</style>
