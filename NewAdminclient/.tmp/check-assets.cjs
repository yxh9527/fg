const fs = require("fs");
const path = require("path");

const configSrc = fs
  .readFileSync(path.resolve(__dirname, "../src/config/index.js"), "utf8")
  .replace(/export const setting[\s\S]*?;\s*/m, "")
  .replace("export default config;", "module.exports = config;");
fs.writeFileSync(path.join(__dirname, "config.cjs"), configSrc);

let helpersSrc = fs.readFileSync(
  path.resolve(__dirname, "../src/views/settlement/settlementHelpers.js"),
  "utf8",
);
helpersSrc = helpersSrc.replace(
  /import config from ["']@\/config["'];\s*/,
  'const config = require("./config.cjs");\n',
);
helpersSrc = helpersSrc.replace(/^export const /gm, "const ");
helpersSrc += `
module.exports = {
  normalizeSettlementRow,
  isSlotDetail,
  isFruitDetail,
  resolveAssetUrl,
  buildSlotCardImageUrl,
};
`;
fs.writeFileSync(path.join(__dirname, "helpers.cjs"), helpersSrc);

const h = require("./helpers.cjs");
const n = h.normalizeSettlementRow({
  gameId: 2534,
  detail: {
    info: {
      game_id: 2534,
      type_id: 1,
      grids: [
        [
          "/global/game/game_FG/2534/40x40/card_01.png?v=1.23",
          "/global/game/game_FG/2534/40x40/card_02.png?v=1.23",
        ],
      ],
      lines_info: [
        {
          line_shape_url: "/global/game/game_FG/2534/l_img/line_01.png?v=1.23",
          symbol: ["/global/game/game_FG/2534/20x20/card_02.png?v=1.23"],
        },
      ],
    },
  },
});
console.log(
  JSON.stringify(
    {
      isSlot: h.isSlotDetail(n.info),
      grid: n.info.grids[0][0],
      shape: n.info.lines_info[0].line_shape_url,
      symbol: n.info.lines_info[0].symbol[0],
      numeric: h.normalizeSettlementRow({
        gameId: 2534,
        detail: { info: { game_id: 2534, type_id: 1, grids: [[1, 2], [3, 4]] } },
      }).info.grids,
    },
    null,
    2,
  ),
);
