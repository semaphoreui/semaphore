var models = require('../../models')
var mongoose = require('mongoose')
var express = require('express')

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
	}).select('-password').exec(function (err, identity) {
		if (err || !user) {
			return res.send(404);
		}

		req.user = user;
		next();
	});
}

function view (req, res) {
	res.send(req.user);
}

function save (req, res) {
	req.user.name = req.body.name;
	models.User.hashPassword(req.body.password, function (hash) {
		req.user.password = hash;
	});

	req.user.save();
	res.send(201);
}

function remove (req, res) {
	req.user.remove(function (err) {
		res.send(201);
	})
}
