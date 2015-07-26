var models = require('../../models'),
	mongoose = require('mongoose'),
	validator = require('validator'),
	user = require('./user');

exports.unauthorized = function (app, template) {
	template([
		'add',
		'list'
	], {
		prefix: 'user'
	});

	user.unauthorized(app, template);
}

exports.httpRouter = function (app) {
	user.httpRouter(app);
}

exports.router = function (app) {
	app.get('/users', get)
		.post('/users', add)

	user.router(app);
}

function get (req, res) {
	models.User.find({})
	.sort('-created')
	.select('-password')
	.exec(function (err, users) {
		res.send(users);
	});
}

function add (req, res) {
	var user = new models.User({
		name: req.body.name,
		email: req.body.email,
		username: req.body.username
	});

	if (user.name.length == 0 || user.name.email == 0) {
		return res.send(400);
	}

	models.User.findOne({
		email: user.email
	}, function (_, existingUser) {
		if (existingUser) {
			return res.send(400);
		}

		models.User.hashPassword(req.body.password, function (hash) {
			user.password = hash;

			user.save(function () {
				res.send(user);
			});
		});
	});
}