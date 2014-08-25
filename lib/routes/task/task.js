var models = require('../../models')
var mongoose = require('mongoose')
var express = require('express')

exports.unauthorized = function (app, template) {
	template([
		'view'
	], {
		prefix: 'task'
	});
}

exports.httpRouter = function (app) {
}

exports.router = function (app) {
	var task = express.Router();

	task.get('/', view)
		.delete('/', remove)

	app.param('task_id', get)
	app.use('/playbook/:playbook_id/job/:job_id/task/:task_id', task);
}

function get (req, res, next, id) {
	models.Task.findOne({
		_id: id
	}).exec(function (err, task) {
		if (err || !task) {
			return res.send(404);
		}

		req.task = task;
		next();
	});
}

function view (req, res) {
	res.send(req.task);
}

function remove (req, res) {
	req.task.remove(function (err) {
		res.send(201);
	})
}