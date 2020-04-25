const path = require('path');
const { CleanWebpackPlugin } = require('clean-webpack-plugin');

module.exports = {
	entry: './src/index.jsx',
	module: {
		rules: [
			{
				test: /\.jsx?$/,
				exclude: /node_modules/,
				use: {
					loader: 'babel-loader',
					options: {
						plugins: [
							["@babel/plugin-transform-react-jsx", {
								"pragma": "h", // default pragma is React.createElement
								"pragmaFrag": "Preact.Fragment", // default is React.Fragment
							}],
							"@babel/plugin-proposal-class-properties"
						]
					}
				}
			}
		]
	},
	output: {
		path: path.resolve(__dirname, 'dist'),
		filename: 'bundle.js'
	},
	resolve: {
		extensions: ['.js', '.jsx']
	},
	plugins: [
		new CleanWebpackPlugin()
	],
	devtool: 'source-map',
	externals: {
		'fetch': 'node-fetch',
		'WebSocket': 'ws'
	}
};
