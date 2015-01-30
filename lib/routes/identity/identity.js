var models = require('../../models')
var mongoose = require('mongoose')
var express = require('express')

exports.unauthorized = function (app, template) {
	template([
		'view'
	], {
		prefix: 'identity'
	});
}

exports.httpRouter = function (app) {
}

exports.router = function (app) {
	var identity = express.Router();

	identity.get('/', view)
		.put('/', save)
		.delete('/', remove)

	app.param('identity_id', get)
	app.use('/identity/:identity_id', identity);
}

function get (req, res, next, id) {
	models.Identity.findOne({
		_id: id
	}).select('-private_key -password').exec(function (err, identity) {
		if (err || !identity) {
			return res.send(404);
		}

		req.identity = identity;
		next();
	});
}

function view (req, res) {
	res.send(req.identity);
}

function save (req, res) {
	req.identity.name = req.body.name;
	req.identity.password = req.body.password;

	req.identity.save();
	res.send(201);
}

function remove (req, res) {
	req.identity.remove(function (err) {
		res.send(201);
	})
}