import dayjs from "dayjs";

export const safeNumber = (value) => {
  if (value === null || value === undefined || value === "") return 0;
  return Number(value) || 0;
};

export const toFixedValue = (value, digits = 2) => safeNumber(value).toFixed(digits);

export const formatDateTime = (value) => {
  if (!value) return "";
  return dayjs(value * 1000).format("YYYY-MM-DD HH:mm:ss");
};

export const formatPickerDayStart = (value) =>
  dayjs(Number(value)).startOf("day").valueOf();

export const formatPickerDayEnd = (value) =>
  dayjs(Number(value)).endOf("day").valueOf();

/**
 * 新流水：bet=下注(负)，award=到账(>=0)
 * 旧流水：无有效 award，bet 即账变 delta（下注负、返奖正）
 *
 * 本条对余额的实际影响：
 * - 下注：bet
 * - 返奖/结算/回退：award（本金已在「下注」流水扣过；回退只有 award）
 * - 旧返奖：正 bet
 */
export const resolveBillDelta = (row) => {
  if (!row) return 0;
  const award = safeNumber(row.award);
  const bet = safeNumber(row.bet);
  const desc = String(row.desc || "").trim();
  if (desc === "下注") return bet;
  if (award > 0) return award;
  return bet;
};

export const resolveBillBeforeScore = (row) => {
  const after = safeNumber(row && row.currentScore);
  return after - resolveBillDelta(row);
};

export const formatBillBet = (row) => {
  if (!row) return "-";
  const bet = safeNumber(row.bet);
  const award = safeNumber(row.award);
  const desc = String(row.desc || "").trim();
  // 旧返奖/结算：bet>0 表示到账，不应显示在「下注」列
  if (award === 0 && bet > 0 && (desc === "返奖" || desc === "结算")) {
    return "-";
  }
  return toFixedValue(bet);
};

export const formatBillAward = (row) => {
  if (!row) return "0.00";
  const bet = safeNumber(row.bet);
  const award = safeNumber(row.award);
  const desc = String(row.desc || "").trim();
  if (award > 0) return toFixedValue(award);
  // 旧数据：正 bet 当作到账
  if (bet > 0 && desc !== "下注") return toFixedValue(bet);
  return "0.00";
};
