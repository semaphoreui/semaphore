var util = require('../util')
	, auth = require('./auth')
	, playbooks = require('./playbook/playbooks')
	, profile = require('./profile')
	, credentials = require('./credential/credentials')

exports.router = function(app) {
	var templates = require('../templates')(app);
	templates.route([
		auth,
		playbooks,
		credentials
	]);

	templates.add('homepage')
	templates.add('abstract')
	templates.setup();

	app.get('/', layout);
	app.all('*', util.authorized);

	// Handle HTTP reqs
	playbooks.httpRouter(app);
	credentials.httpRouter(app);
	
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
	playbooks.router(app);
	profile.router(app);
	credentials.router(app);
}

function layout (req, res) {
	if (res.locals.loggedIn) {
		res.render('layout')
	} else {
		res.render('auth');
	}
}