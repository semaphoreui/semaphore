var manifest = [
	'User'
];

manifest.forEach(function (model) {
	module.exports[model] = require('./'+model);
});