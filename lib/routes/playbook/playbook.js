var models = require('../../models')
var mongoose = require('mongoose')
var express = require('express')

exports.unauthorized = function (app, template) {
	template([
		'view'
	], {
		prefix: 'playbook'
	});
}

exports.httpRouter = function (app) {
}

exports.router = function (app) {
	var playbook = express.Router();

	playbook.get('/', view)
		.put('/', save)
		.delete('/', remove)

	app.param('playbook_id', getPlaybook)
	app.use('/playbook/:playbook_id', playbook);
}

function getPlaybook (req, res, next, id) {
	models.Playbook.findOne({
		_id: id
	}).select('-vault_password').exec(function (err, playbook) {
		if (err || !playbook) {
			return res.send(404);
		}

		req.playbook = playbook;
		next();
	});
}

function view (req, res) {
	res.send(req.playbook);
}

function save (req, res) {
	req.playbook.name = req.body.name;
	req.playbook.location = req.body.location;

	if (req.body.vault_password.length > 0) {
		req.playbook.vault_password = req.body.vault_password;
	}

	try {
		req.playbook.credential = mongoose.Types.ObjectId(req.body.credential);
	} catch (e) {}

	req.playbook.save();
	res.send(201);
}

function remove (req, res) {
	req.playbook.remove(function (err) {
		res.send(201);
	})
}