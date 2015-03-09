var models = require('../../models')
var mongoose = require('mongoose')

var user = require('./user')

var validator = require('validator')

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
	models.User.find({
	}).sort('-created').select('-password').exec(function (err, identities) {
		res.send(users)
	})
}

function add (req, res) {
	if (!validator.isLength(req.body.name, 1)) {
		return res.send(400);
	}

	var user = new models.User({
		name: req.body.name,
		password: req.body.password,
		email: req.body.email,
		username: req.body.username
	});

	user.save(function () {
		res.send(user);
	});
}
