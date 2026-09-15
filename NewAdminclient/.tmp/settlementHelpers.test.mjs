/** 把后台注单行 / fgServer 注单列表项统一成详情弹窗可用结构。 */

export const safeNumber = (value) => {
  if (value === null || value === undefined || value === "") return 0;
  return Number(value) || 0;
};

export const formatAmount = (value) => {
  if (value === null || value === undefined || value === "") return "-";
  if (typeof value === "string" && Number.isNaN(Number(value))) return value;
  return safeNumber(value).toFixed(2);
};

export const formatPlayedTime = (value) => {
  if (!value && value !== 0) return "";
  if (typeof value === "string" && value.includes("-")) return value;
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric <= 0) return String(value || "");
  const ms = numeric < 1e12 ? numeric * 1000 : numeric;
  const date = new Date(ms);
  if (Number.isNaN(date.getTime())) return String(value);
  const pad = (n) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(
    date.getHours(),
  )}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
};

export const parseMaybeJson = (value) => {
  if (typeof value !== "string") return value;
  try {
    return JSON.parse(value);
  } catch (error) {
    return value;
  }
};

const looksLikeDetailInfo = (value) =>
  Boolean(
    value &&
      typeof value === "object" &&
      (value.grids ||
        value.specific_bet_info ||
        value.result_show_url ||
        value.result_grid_url ||
        value.type_id != null ||
        value.game_id != null ||
        value.total_bet),
  );

const coerceDetailObject = (value) => {
  const detail = parseMaybeJson(value);
  if (!detail || typeof detail !== "object") return null;
  if (detail.info && typeof detail.info === "object") {
    return { detail, info: detail.info };
  }
  if (looksLikeDetailInfo(detail)) {
    return { detail: { info: detail }, info: detail };
  }
  return null;
};

/** 兼容 detail / log / detail.info / 直接 info 几种落库形态。 */
export const resolveDetailPayload = (row) => {
  if (!row || typeof row !== "object") {
    return { detail: null, info: null };
  }
  const fromDetail = coerceDetailObject(row.detail);
  if (fromDetail) return fromDetail;
  const fromLog = coerceDetailObject(row.log);
  if (fromLog) return fromLog;
  if (row.info && typeof row.info === "object") {
    return { detail: { info: row.info }, info: row.info };
  }
  return { detail: null, info: null };
};

export const normalizeSettlementRow = (row = {}) => {
  const { detail, info } = resolveDetailPayload(row);
  const gameId = Number(
    info?.game_id ?? row.gameId ?? row.game_id ?? 0,
  );
  return {
    raw: row,
    detail,
    info,
    gameId,
    gameName: row.gameName || row.game_name || info?.game_name || "",
    recordId: String(row.roundID || row.id || info?.ju_id || row.officeNumber || ""),
    time: info?.time || formatPlayedTime(row.playedDate) || row.time || "",
    allBets: info?.all_bets ?? info?.total_bet?.bet_num ?? row.bet ?? row.all_bets ?? "",
    allBonus:
      info?.sum_bonus ??
      info?.total_bet?.reward_money ??
      row.win ??
      row.all_bonus ??
      "",
    jpBonus: info?.jp_bonus ?? row.jp_bonus ?? "",
    lineBets: info?.line_bets ?? "",
    typeText: info?.type || row.game_type || "",
    currency: row.currency || row.symbol || "",
    playerName: info?.player_name || row.nickName || row.account || "",
  };
};

export const isFruitGameId = (gameId) => {
  const id = Number(gameId);
  return id >= 7000 && id < 8000;
};

export const isSlotDetail = (info) => {
  if (!info || typeof info !== "object") return false;
  return Array.isArray(info.grids) && info.grids.length > 0;
};

export const isFruitDetail = (info, gameId) => {
  if (!info || typeof info !== "object") return false;
  if (isSlotDetail(info)) return false;
  if (isFruitGameId(gameId || info.game_id)) return true;
  return Boolean(
    (Array.isArray(info.specific_bet_info) && info.specific_bet_info.length) ||
      (Array.isArray(info.result_show_url) && info.result_show_url.length) ||
      (Array.isArray(info.result_grid_url) && info.result_grid_url.length) ||
      info.total_bet,
  );
};

export const asImageList = (value) => {
  if (!value) return [];
  if (typeof value === "string") return value ? [value] : [];
  if (!Array.isArray(value)) return [];
  return value.filter((item) => typeof item === "string" && item);
};

/** result_grid_url 可能是一维图片数组，也可能是二维盘面。 */
export const asImageMatrix = (value) => {
  if (!Array.isArray(value) || !value.length) return [];
  if (Array.isArray(value[0])) {
    return value
      .map((row) => (Array.isArray(row) ? row.filter((item) => typeof item === "string" && item) : []))
      .filter((row) => row.length);
  }
  const flat = value.filter((item) => typeof item === "string" && item);
  return flat.length ? [flat] : [];
};

export const pickLineInfos = (info) => {
  if (!info || typeof info !== "object") return [];
  const isAllLines = Number(info.is_all_lines || 0);
  const source =
    isAllLines === 2 || isAllLines === 3
      ? info.m_list_info
      : isAllLines === 1
        ? info.m_list_info?.length
          ? info.m_list_info
          : info.lines_info
        : info.lines_info;
  return Array.isArray(source) ? source : [];
};
