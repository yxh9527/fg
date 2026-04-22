const path = require('node:path')
const webpack = require('webpack')
const HtmlWebpackPlugin = require('html-webpack-plugin')
const { VueLoaderPlugin } = require('vue-loader')

const rootDir = __dirname
const srcDir = path.resolve(rootDir, 'src')
const publicDir = path.resolve(rootDir, 'public')

function ensureTrailingSlash(value) {
  return value.endsWith('/') ? value : `${value}/`
}

const host = ensureTrailingSlash(process.env.HOST || 'https://h5.chichengwld.com/')
const dyHost = ensureTrailingSlash(process.env.DY_HOST || host)
const imgHost = ensureTrailingSlash(process.env.IMG_HOST || host)
const routerBase = process.env.ROUTER_BASE || '/'

module.exports = {
  entry: path.resolve(srcDir, 'main.js'),
  output: {
    path: path.resolve(rootDir, 'dist'),
    filename: 'static/js/[name].[contenthash:8].js',
    assetModuleFilename: 'static/assets/[name].[contenthash:8][ext][query]',
    publicPath: routerBase
  },
  resolve: {
    extensions: ['.js', '.vue', '.json'],
    alias: {
      vue$: 'vue/dist/vue.esm.js',
      '@': srcDir,
      assets: path.resolve(srcDir, 'assets')
    }
  },
  module: {
    rules: [
      {
        test: /\.vue$/,
        loader: 'vue-loader'
      },
      {
        test: /\.js$/,
        include: srcDir,
        use: {
          loader: 'babel-loader'
        }
      },
      {
        test: /\.css$/,
        use: ['vue-style-loader', 'css-loader']
      },
      {
        test: /\.scss$/,
        use: ['vue-style-loader', 'css-loader', 'sass-loader']
      },
      {
        test: /\.(png|jpe?g|gif|svg|ico|webp)$/i,
        type: 'asset/resource',
        generator: {
          filename: 'static/images/[name][ext][query]'
        }
      },
      {
        test: /\.(woff2?|eot|ttf|otf)$/i,
        type: 'asset/resource',
        generator: {
          filename: 'static/fonts/[name][ext][query]'
        }
      }
    ]
  },
  plugins: [
    new webpack.DefinePlugin({
      HOST: JSON.stringify(host),
      DY_HOST: JSON.stringify(dyHost),
      IMG_HOST: JSON.stringify(imgHost),
      ROUTER_BASE: JSON.stringify(routerBase)
    }),
    new VueLoaderPlugin(),
    new HtmlWebpackPlugin({
      template: path.resolve(publicDir, 'index.html')
    })
  ],
  devServer: {
    port: Number(process.env.PORT || 8080),
    host: '0.0.0.0',
    historyApiFallback: true,
    static: {
      directory: publicDir
    }
  }
}
