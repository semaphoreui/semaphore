var fs = require('fs');

try {
	var credentials = require('./credentials.json')
} catch (e) {
	console.log("\nNo credentials.json File!\n")
	process.exit(1);
}

exports.credentials = credentials;

exports.version = require('../package.json').version;
exports.hash = 'dirty';
exports.production = process.env.NODE_ENV == "production";
exports.port = process.env.PORT || credentials.port;
exports.path = __dirname;

if (process.platform.match(/^win/) == null) {
	try {
		var spawn_process = require('child_process').spawn
		var readHash = spawn_process('git', ['rev-parse', '--short', 'HEAD']);
		readHash.stdout.on('data', function (data) {
			exports.hash = data.toString().trim();
			require('./app').app.locals.versionHash = exports.hash;
		})
	} catch (e) {
		console.log("\n~= Unable to obtain git commit hash =~\n")
	}
}

exports.configure = function (app) {
	app.locals.pretty = exports.production // Pretty HTML outside production mode
	app.locals.version = exports.version;
	app.locals.versionHash = exports.hash;
	app.locals.production = exports.production;
	app.locals.use_analytics = credentials.use_analytics;
}

exports.init = function () {
	var models = require('./models');

	models.User.findOne({
		email: 'admin@semaphore.local'
	}).exec(function (err, admin) {
		if (!admin) {
			console.log("Creating Admin user admin@semaphore.local!");

			admin = new models.User({
				email: 'admin@semaphore.local',
				username: 'semaphore',
				name: 'Administrator'
			});
			models.User.hashPassword('CastawayLabs', function (hash) {
				admin.password = hash;

				admin.save();
			});
		}
	})
}