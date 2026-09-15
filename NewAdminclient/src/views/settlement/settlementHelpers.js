import config from "@/config";

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
  return (
    date.getFullYear() +
    "-" +
    pad(date.getMonth() + 1) +
    "-" +
    pad(date.getDate()) +
    " " +
    pad(date.getHours()) +
    ":" +
    pad(date.getMinutes()) +
    ":" +
    pad(date.getSeconds())
  );
};

export const parseMaybeJson = (value) => {
  if (typeof value !== "string") return value;
  try {
    return JSON.parse(value);
  } catch (error) {
    return value;
  }
};

const normalizeOssBaseUrl = (url) => {
  if (!url) return "";
  return url.endsWith("/") ? url : url + "/";
};

/** 与 fgServer recordService.buildRecordAssetUrl 对齐：相对 /global 路径补 OSS。 */
export const resolveAssetUrl = (value, ossUrl) => {
  if (value === null || value === undefined || value === "") return "";
  if (typeof value === "number") return "";
  if (typeof value !== "string") return "";
  const text = value.trim();
  if (!text) return "";
  if (/^(?:https?:)?\/\//i.test(text) || text.indexOf("data:") === 0) {
    return text;
  }

  let assetPath = "";
  if (text.indexOf("/global/") === 0) assetPath = text.slice(1);
  else if (text.indexOf("global/") === 0) assetPath = text;

  if (!assetPath) {
    // 已是站点相对路径时，也尽量挂到 OSS，避免打到后台域名 404。
    if (text.charAt(0) === "/") {
      assetPath = text.slice(1);
    } else {
      return text;
    }
  }

  const base = normalizeOssBaseUrl(ossUrl || config.ossUrl || "");
  return base ? base + assetPath : "/" + assetPath;
};

export const buildSlotCardImageUrl = (gameId, symbolId, size) => {
  const id = Math.max(Number(symbolId) || 0, 0);
  const padded = String(id).padStart(2, "0");
  return (
    "/global/game/game_FG/" +
    Number(gameId) +
    "/" +
    (size || "40x40") +
    "/card_" +
    padded +
    ".png?v=1.23"
  );
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
    return { detail: detail, info: detail.info };
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

const resolveGridCell = (cell, gameId, size, ossUrl) => {
  if (typeof cell === "number" || (/^\d+$/.test(String(cell || "")))) {
    return resolveAssetUrl(buildSlotCardImageUrl(gameId, cell, size), ossUrl);
  }
  if (typeof cell !== "string") return "";
  const text = cell.trim();
  if (!text) return "";
  // card_01.png / 01.png 这类残缺路径，按 gameId 补全。
  const cardMatch = text.match(/(?:card_)?(\d{1,2})(?:\.png)?(?:\?.*)?$/i);
  if (
    cardMatch &&
    text.indexOf("/global/") < 0 &&
    text.indexOf("global/") < 0 &&
    text.indexOf("http") < 0
  ) {
    return resolveAssetUrl(
      buildSlotCardImageUrl(gameId, cardMatch[1], size),
      ossUrl,
    );
  }
  return resolveAssetUrl(text, ossUrl);
};

const mapImageMatrix = (value, gameId, size, ossUrl) => {
  if (!Array.isArray(value) || !value.length) return [];
  if (Array.isArray(value[0])) {
    return value
      .map((row) => {
        if (!Array.isArray(row)) return [];
        return row
          .map((cell) => resolveGridCell(cell, gameId, size, ossUrl))
          .filter(Boolean);
      })
      .filter((row) => row.length);
  }
  return [
    value
      .map((cell) => resolveGridCell(cell, gameId, size, ossUrl))
      .filter(Boolean),
  ].filter((row) => row.length);
};

const mapImageList = (value, gameId, size, ossUrl) => {
  if (!value && value !== 0) return [];
  if (!Array.isArray(value)) {
    const url = resolveGridCell(value, gameId, size, ossUrl);
    return url ? [url] : [];
  }
  return value
    .map((item) => resolveGridCell(item, gameId, size, ossUrl))
    .filter(Boolean);
};

const enrichDetailInfo = (rawInfo, gameId, ossUrl) => {
  if (!rawInfo || typeof rawInfo !== "object") return null;
  const info = Object.assign({}, rawInfo);
  const id = Number((info.game_id != null ? info.game_id : gameId) || 0);

  if (Array.isArray(info.grids)) {
    // slots grids 是“列优先”二维数组：[[col0 rows...],[col1 rows...]]
    info.grids = mapImageMatrix(info.grids, id, "40x40", ossUrl);
  }
  if (Array.isArray(info.result_show_url)) {
    info.result_show_url = mapImageList(info.result_show_url, id, "40x40", ossUrl);
  } else if (typeof info.result_show_url === "string") {
    info.result_show_url = mapImageList([info.result_show_url], id, "40x40", ossUrl);
  }
  if (Array.isArray(info.result_grid_url)) {
    info.result_grid_url = mapImageMatrix(info.result_grid_url, id, "40x40", ossUrl);
  }
  if (Array.isArray(info.lines_info)) {
    info.lines_info = info.lines_info.map((line) => {
      if (!line || typeof line !== "object") return line;
      const next = Object.assign({}, line);
      if (next.line_shape_url) {
        next.line_shape_url = resolveAssetUrl(next.line_shape_url, ossUrl);
      }
      if (next.symbol) {
        next.symbol = mapImageList(next.symbol, id, "20x20", ossUrl);
      }
      return next;
    });
  }
  if (Array.isArray(info.m_list_info)) {
    info.m_list_info = info.m_list_info.map((line) => {
      if (!line || typeof line !== "object") return line;
      const next = Object.assign({}, line);
      if (next.line_shape_url) {
        next.line_shape_url = resolveAssetUrl(next.line_shape_url, ossUrl);
      }
      if (next.symbol) {
        next.symbol = mapImageList(next.symbol, id, "20x20", ossUrl);
      }
      if (Array.isArray(next.grid_data)) {
        next.grid_data = mapImageMatrix(next.grid_data, id, "40x40", ossUrl);
      }
      return next;
    });
  }
  if (Array.isArray(info.specific_bet_info)) {
    info.specific_bet_info = info.specific_bet_info.map((bet) => {
      if (!bet || typeof bet !== "object") return bet;
      const next = Object.assign({}, bet);
      if (typeof next.bet_area_url === "string") {
        next.bet_area_url = resolveAssetUrl(next.bet_area_url, ossUrl);
      }
      if (Array.isArray(next.grid)) {
        next.grid = mapImageMatrix(next.grid, id, "40x40", ossUrl);
      }
      return next;
    });
  }
  return info;
};

const pickFirst = function () {
  for (let i = 0; i < arguments.length; i += 1) {
    const value = arguments[i];
    if (value !== null && value !== undefined && value !== "") return value;
  }
  return "";
};

export const normalizeSettlementRow = (row, ossUrl) => {
  const source = row || {};
  const resolved = resolveDetailPayload(source);
  const rawInfo = resolved.info;
  const gameId = Number(
    pickFirst(
      rawInfo && rawInfo.game_id,
      source.gameId,
      source.game_id,
      0,
    ),
  );
  const info = enrichDetailInfo(rawInfo, gameId, ossUrl || config.ossUrl);
  const totalBet = info && info.total_bet ? info.total_bet : null;

  return {
    raw: source,
    detail: resolved.detail,
    info: info,
    gameId: gameId,
    gameName: pickFirst(
      source.gameName,
      source.game_name,
      info && info.game_name,
    ),
    recordId: String(
      pickFirst(
        source.roundID,
        source.id,
        info && info.ju_id,
        source.officeNumber,
      ),
    ),
    time: pickFirst(info && info.time, formatPlayedTime(source.playedDate), source.time),
    allBets: pickFirst(
      info && info.all_bets,
      totalBet && totalBet.bet_num,
      source.bet,
      source.all_bets,
    ),
    allBonus: pickFirst(
      info && info.sum_bonus,
      totalBet && totalBet.reward_money,
      source.win,
      source.all_bonus,
    ),
    jpBonus: pickFirst(info && info.jp_bonus, source.jp_bonus),
    lineBets: pickFirst(info && info.line_bets),
    typeText: pickFirst(info && info.type, source.game_type),
    currency: pickFirst(source.currency, source.symbol),
    playerName: pickFirst(info && info.player_name, source.nickName, source.account),
  };
};

export const isFruitGameId = (gameId) => {
  const id = Number(gameId);
  return id >= 7000 && id < 8000;
};

export const isSlotDetail = (info) => {
  if (!info || typeof info !== "object") return false;
  if (Array.isArray(info.grids) && info.grids.length > 0) return true;
  // 有 type_id / lines_info 也按 slots 详情处理，避免缺 grids 时误判。
  if (info.type_id != null && !isFruitGameId(info.game_id)) return true;
  return false;
};

export const isFruitDetail = (info, gameId) => {
  if (!info || typeof info !== "object") return false;
  if (Array.isArray(info.grids) && info.grids.length > 0) return false;
  if (isFruitGameId(gameId || info.game_id)) return true;
  return Boolean(
    (Array.isArray(info.specific_bet_info) && info.specific_bet_info.length) ||
      (Array.isArray(info.result_show_url) && info.result_show_url.length) ||
      (Array.isArray(info.result_grid_url) && info.result_grid_url.length) ||
      info.total_bet,
  );
};

export const asImageList = (value) => {
  if (!value && value !== 0) return [];
  if (typeof value === "string") return value ? [value] : [];
  if (!Array.isArray(value)) return [];
  return value.filter((item) => typeof item === "string" && item);
};

/** result_grid_url 可能是一维图片数组，也可能是二维盘面。 */
export const asImageMatrix = (value) => {
  if (!Array.isArray(value) || !value.length) return [];
  if (Array.isArray(value[0])) {
    return value
      .map((row) =>
        Array.isArray(row) ? row.filter((item) => typeof item === "string" && item) : [],
      )
      .filter((row) => row.length);
  }
  const flat = value.filter((item) => typeof item === "string" && item);
  return flat.length ? [flat] : [];
};

export const pickLineInfos = (info) => {
  if (!info || typeof info !== "object") return [];
  const isAllLines = Number(info.is_all_lines || 0);
  let source = info.lines_info;
  if (isAllLines === 2 || isAllLines === 3) {
    source = info.m_list_info;
  } else if (isAllLines === 1) {
    source =
      info.m_list_info && info.m_list_info.length
        ? info.m_list_info
        : info.lines_info;
  }
  return Array.isArray(source) ? source : [];
};
