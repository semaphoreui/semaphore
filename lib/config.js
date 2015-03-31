var fs = require('fs'),
	env = process.env;

exports.credentials = {
	"redis_port": env.REDIS_PORT,
	"redis_host": env.REDIS_HOST,
	"redis_key": env.REDIS_KEY,
	"use_analytics": env.USE_ANALYTICS,
	"bugsnag_key": env.BUGSNAG_KEY,
	"smtp": {
		"user": env.SMTP_USER,
		"pass": env.SMTP_PASS
	},
	"db": env.MONGODB_URL,
	"db_options": {
		"auto_reconnect": true,
		"native_parser": true,
		"server": {
			"auto_reconnect": true
		}
	}
};

exports.version = require('../package.json').version;
exports.hash = 'dirty';
exports.production = process.env.NODE_ENV == "production";
exports.port = env.SSL ? 443 : 80;
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
	app.locals.use_analytics = exports.credentials.use_analytics;
}

exports.init = function () {
	var models = require('./models');

	models.User.findOne({
		email: env.ADMIN_EMAIL
	}).exec(function (err, admin) {
		if (!admin) {
			console.log("Creating Admin user admin@semaphore.local!");

			admin = new models.User({
				email: env.ADMIN_EMAIL,
				username: env.ADMIN_USERNAME,
				name: env.ADMIN_REALNAME
			});

			models.User.hashPassword(env.PASSWORD_HASH, function (hash) {
				admin.password = hash;

				admin.save();
			});
		}
	})
}