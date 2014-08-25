var models = require('../../models')
var mongoose = require('mongoose')
var express = require('express')

exports.unauthorized = function (app, template) {
	template([
		'view'
	], {
		prefix: 'credential'
	});
}

exports.httpRouter = function (app) {
}

exports.router = function (app) {
	var credential = express.Router();

	credential.get('/', view)
		.put('/', save)
		.delete('/', remove)

	app.param('credential_id', get)
	app.use('/credential/:credential_id', credential);
}

function get (req, res, next, id) {
	models.Credential.findOne({
		_id: id
	}).select('-private_key -public_key -password').exec(function (err, credential) {
		if (err || !credential) {
			return res.send(404);
		}

		req.credential = credential;
		next();
	});
}

function view (req, res) {
	res.send(req.credential);
}

function save (req, res) {
	req.credential.name = req.body.name;
	req.credential.location = req.body.location;

	req.credential.save();
	res.send(201);
}

function remove (req, res) {
	req.credential.remove(function (err) {
		res.send(201);
	})
}