var fs = require('fs'),
	env = process.env;

try {
	var credentials = require('./credentials.json');
} catch (e) {
	if (!(process.env.MONGODB_URL && process.env.REDIS_HOST)) {
		console.log("\nNo credentials.json File or env variables!\n");
		process.exit(1);
	} else {
		credentials = require('./credentials.default.json');
	}
}

exports.credentials = credentials;

['redis_port', 'redis_host', 'redis_key', 'bugsnag_key', 'port'].forEach(function (key) {
	if (env[key.toUpperCase()]) {
		exports.credentials[key] = env[key.toUpperCase()];
	}
});

if (env.SMTP_USER) {
	exports.credentials.smtp.user = env.SMTP_USER;
}
if (env.SMTP_PASS) {
	exports.credentials.smtp.pass = env.SMTP_PASS;
}
if (env.MONGODB_URL) {
	exports.credentials.db = env.MONGODB_URL;
}

console.log(exports.credentials)

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