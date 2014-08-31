var config = require('./config');

var newrelic = {
	getBrowserTimingHeader: function () {}
};
if (config.production && config.credentials.use_analytics) {
	newrelic = require('newrelic');
}

var express = require('express')
	, routes = require('./routes')
	, http = require('http')
	, path = require('path')
	, mongoose = require('mongoose')
	, util = require('./util')
	, session = require('express-session')
	, RedisStore = require('connect-redis')(session)
	, passport = require('passport')
	, auth = require('./auth')
	, bugsnag = require('bugsnag')
	, socketPassport = require('passport.socketio')
	, bodyParser = require('body-parser')
	, logtrail = require('logtrail');

var app = exports.app = express();

if (config.production) {
	require('newrelic');
}

logtrail.configure({
	timestamps: {
		enabled: false
	},
	stacktrace: true,
	basedir: __dirname
});
console.log = logtrail.log.bind(logtrail);

var releaseStage = config.production ? "production" : "development";

bugsnag.register(config.credentials.bugsnag_key, {
	notifyReleaseStages: ["production"],
	releaseStage: releaseStage
});

mongoose.connect(config.credentials.db, config.credentials.db_options);

var sessionStore = new RedisStore({
	host: config.credentials.redis_host,
	port: config.credentials.redis_port,
	ttl: 604800000,
	pass: config.credentials.redis_key
});

var db = mongoose.connection
db.on('error', console.error.bind(console, 'Mongodb Connection Error:'));
db.once('open', function callback () {
	if (!config.is_testing) console.log("Mongodb connection established")
});

// all environments
app.enable('trust proxy');
app.set('port', process.env.PORT || 3000); // Port
app.set('views', __dirname + '/views');
app.set('view engine', 'jade'); // Templating engine
app.set('app version', config.version); // App version
app.set('x-powered-by', false);

app.set('view cache', config.production);

app.locals.newrelic = newrelic;
config.configure(app);

app.use(function(req, res, next) {
	res.set('x-frame-options', 'SAMEORIGIN');
	res.set('x-xss-protection', '1; mode=block');
	next();
});

app.use(require('serve-static')(path.join(__dirname, '..', 'dist')));
app.use(require('morgan')(config.production ? 'combined' : 'dev'));

app.use(bugsnag.requestHandler);
app.use(bodyParser.urlencoded({
	extended: true
}));
app.use(bodyParser.json());
app.use(require('cookie-parser')());
app.use(session({
	secret: "#semaphore",
	name: 'semaphore',
	store: sessionStore,
	proxy: true,
	saveUninitialized: false,
	resave: false,
	cookie: {
		secure: config.credentials.is_ssl,
		maxAge: 604800000
	}
}));

app.use(passport.initialize());
app.use(passport.session());

// Custom middleware
app.use(function(req, res, next) {
	res.locals.user = req.user;
	res.locals.loggedIn = res.locals.user != null;
	
	next();
});

// routes
routes.router(app);

app.use(bugsnag.errorHandler);

var server = http.createServer(app)
server.listen(app.get('port'), function(){
	console.log('Semaphore listening on port ' + app.get('port'));
});
exports.io = io = require('socket.io').listen(server)

config.init();

io.use(socketPassport.authorize({
	cookieParser: require('cookie-parser'),
	secret: "#semaphore",
	key: 'semaphore',
	store: sessionStore,
	passport: passport,
	fail: function(data, message, error, accept) {
		accept(false);
	}
}))
