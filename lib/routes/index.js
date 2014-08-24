var util = require('../util')
	, auth = require('./auth')
	, apps = require('./apps')

exports.router = function(app) {
	var templates = require('../templates')(app);
	templates.route([
		auth,
	//	apps
	]);

	templates.add('homepage')
	templates.setup();

	app.get('/', layout);
	app.all('*', util.authorized);

	// Handle HTTP reqs
	apps.httpRouter(app);
	
	// only json beyond this point
	app.get('*', function(req, res, next) {
		res.format({
			json: function() {
				next()
			},
			html: function() {
				layout(req, res);
			}
		});
	});

	auth.router(app);
	apps.router(app);
}

function layout (req, res) {
	if (res.locals.loggedIn) {
		res.render('layout')
	} else {
		res.render('auth');
	}
}