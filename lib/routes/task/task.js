var models = require('../../models')
var mongoose = require('mongoose')
var express = require('express')

var jobRunner = require('../../runner');

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
	if (req.task.status == 'Running') {
		return res.send(400, 'Job is Running.');
	}

	jobRunner.queue.pause();
	for (var i = 0; i < jobRunner.queue.tasks.length; i++) {
		if (jobRunner.queue.tasks[i].data._id.toString() == req.task._id.toString()) {
			// This is our task
			jobRunner.queue.tasks.splice(i, 1);
			break;
		}
	}
	jobRunner.queue.resume();

	req.task.remove(function (err) {
		res.send(201);
	})
}