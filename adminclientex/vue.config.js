const path = require("path");
const webpack = require("webpack");
const CopyWebpackPlugin = require("copy-webpack-plugin");
const WebpackShellPlugin = require("webpack-shell-plugin");
const os = require("os");
const resolve = dir => {
  return path.join(__dirname, dir);
};

// 项目部署配置
let iplatform = "--leyou";
const BASE_URL = process.env.NODE_ENV === "production" ? "/" : "/";
// let proxyUrl = 'http://172.21.211.215:9529'
// let proxyUrl='http://127.0.0.1:8000';
let proxyUrl='';

module.exports = {
  devServer: {
    disableHostCheck: true
  },
  publicPath: BASE_URL,
  lintOnSave: false,
  chainWebpack: config => {
    // 目录别名配置
    config.resolve.alias
      .set("@", resolve("src"))
      .set("_c", resolve("src/components"));
  },
  // 设为false打包时不生成.map文件
  productionSourceMap: false,
  // 调用接口的基础路径，来解决跨域，如果设置了代理，那你本地开发环境的axios的baseUrl要写为 '' ，即空字符串
  devServer: {
    proxy: {
      v2: {
        target: proxyUrl
      },
      v1: {
        target: proxyUrl
      }
    }
  },
  outputDir: "web-client-manager-ex",
  configureWebpack: {
    plugins: [
      new webpack.DefinePlugin({
        iplatform: JSON.stringify(iplatform) // 将proxy当做全局变量PROXY注入到项目的上下文环境中，注意这个时候，这个全局变量并不是附加到了window对象上，而是直接的一个全局变量。访问的时候就是直接PROXY就可以访问到该变量
      }),
      new CopyWebpackPlugin({
        patterns: [
          {
            from: path.join(__dirname, "./public/404.html"),
            to: path.join(__dirname, "./web-client-manager-ex/")
          }
        ],
        options: {
          concurrency: 100
        }
      })
    ]
  }
};
