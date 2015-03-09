var util = require('../util')
	auth = require('./auth'),
	playbooks = require('./playbook/playbooks'),
	profile = require('./profile'),
	identities = require('./identity/identities'),
	jobs = require('./job/jobs'),
	hosts = require('./host/hosts'),
	tasks = require('./task/tasks'),
	users = require('./user/users')

exports.router = function(app) {
	var templates = require('../templates')(app);
	templates.route([
		auth,
		playbooks,
		identities,
		jobs,
		hosts,
		tasks,
		users
	]);

	templates.add('homepage')
	templates.add('abstract')
	templates.setup();

	app.get('/', layout);
	app.all('*', util.authorized);

	// Handle HTTP reqs
	playbooks.httpRouter(app);
	identities.httpRouter(app);
	jobs.httpRouter(app);
	hosts.httpRouter(app);
	tasks.httpRouter(app);
	users.httpRouter(app);

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
	identities.router(app);
	jobs.router(app);
	hosts.router(app);
	tasks.router(app);
	users.router(app);
}

function layout (req, res) {
	if (res.locals.loggedIn) {
		res.render('layout')
	} else {
		res.render('auth');
	}
}