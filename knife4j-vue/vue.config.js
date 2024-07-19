const TerserPlugin = require("terser-webpack-plugin");
// const WebpackBundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin
const CompressionWebpackPlugin = require('compression-webpack-plugin');
const productionGzipExtensions = ["js", "css"];
module.exports = {
  publicPath: ".",
  assetsDir: "webjars",
  outputDir: "dist",
  lintOnSave: false,
  productionSourceMap: false,
  indexPath: "index.tmpl",
  css: {
    loaderOptions: {
      less: {
        javascriptEnabled: true
      }
    }
  },
  devServer: {
    watchOptions:{
      ignored: /node_modules/
    },
    proxy: {
      "/": {
      target: 'http://127.0.0.1:8001/swagger/',
      // target: 'http://127.0.0.1:28087/v1/doc/api/frontend/swagger/',
      /*    target: 'http://localhost:18568/',   */
        /* target: 'http://knife4j.xiaominfo.com/',*/
        ws: true,
        changeOrigin: true
      }
    }
  },
  configureWebpack: {
    optimization: {
      minimizer: [
        new TerserPlugin({
          terserOptions: {
            ecma: undefined,
            warnings: false,
            parse: {},
            compress: {
              drop_console: true,
              drop_debugger: true,
              pure_funcs: ['console.log', 'console.debug', 'window.console.log', 'window.console.debug'] // 移除console
            }
          },
        }),

      ]
    },
    plugins: [
      new CompressionWebpackPlugin({
        algorithm: "gzip",
        test: new RegExp("\\.(" + productionGzipExtensions.join("|") + ")$"),
        threshold: 10240,
        minRatio: 0.8
      })
    ]
  }
};
