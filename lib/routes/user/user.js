var models = require('../../models'),
	mongoose = require('mongoose'),
	express = require('express');

exports.unauthorized = function (app, template) {
	template([
		'view'
	], {
		prefix: 'user'
	});
}

exports.httpRouter = function (app) {
}

exports.router = function (app) {
	var user = express.Router();

	user.get('/', view)
		.put('/', save)
		.delete('/', remove)

	app.param('user_id', get)
	app.use('/user/:user_id', user);
}

function get (req, res, next, id) {
	models.User.findOne({
		_id: id
	})
	.select('-password')
	.exec(function (err, user) {
		if (err || !user) {
			return res.send(404);
		}

		req._user = user;
		next();
	});
}

function view (req, res) {
	res.send(req._user);
}

function save (req, res) {
	req._user.name = req.body.name;
	models.User.hashPassword(req.body.password, function (hash) {
		req._user.password = hash;
	});

	req._user.save();
	res.send(201);
}

function remove (req, res) {
	req._user.remove(function (err) {
		res.send(201);
	})
}
