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
  resolveAssetUrl,
  normalizeSettlementRow,
  isSlotDetail,
};
`;
fs.writeFileSync(path.join(__dirname, "helpers.cjs"), helpersSrc);

const h = require("./helpers.cjs");
const file =
  "D:/project/fg-game-server/fgServer/static/global/game/game_FG/2534/40x40/card_01.png";
console.log(
  JSON.stringify(
    {
      fileExists: fs.existsSync(file),
      relative: h.resolveAssetUrl(
        "/global/game/game_FG/2534/40x40/card_01.png?v=1.23",
      ),
      absoluteKept: h.resolveAssetUrl(
        "https://pic.vewjn.com/global/game/game_FG/2534/40x40/card_01.png",
      ),
      slotGrid: h.normalizeSettlementRow({
        gameId: 2534,
        detail: {
          info: {
            game_id: 2534,
            type_id: 1,
            grids: [["/global/game/game_FG/2534/40x40/card_01.png"]],
          },
        },
      }).info.grids[0][0],
    },
    null,
    2,
  ),
);
