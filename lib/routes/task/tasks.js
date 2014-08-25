var models = require('../../models')
var mongoose = require('mongoose')

var task = require('./task')

var app = require('../../app')

exports.unauthorized = function (app, template) {
	template([
		'tasks'
	], {
		prefix: 'task'
	});

	task.unauthorized(app, template);
}

exports.httpRouter = function (app) {
	task.httpRouter(app);
}

exports.router = function (app) {
	app.get('/playbook/:playbook_id/tasks', getTasks)
		.get('/playbook/:playbook_id/job/:job_id/tasks', get)
		.post('/playbook/:playbook_id/job/:job_id/tasks', add)
		.post('/playbook/:playbook_id/job/:job_id/run', runJob)

	task.router(app);
}

function get (req, res) {
	models.Task.find({
		job: req.job._id
	}).populate('job').sort('-created').exec(function (err, tasks) {
		res.send(tasks)
	})
}

function getTasks (req, res) {
	models.Task.find({
		playbook: req.playbook._id
	}).populate('job').sort('-created').exec(function (err, tasks) {
		res.send(tasks)
	})
}

function add (req, res) {
	var task = new models.Task({
		job: req.job._id,
		status: 'Queued'
	})
	
	task.save(function () {
		res.send(task);
	});
}

function runJob (req, res) {
	var task = new models.Task({
		job: req.job._id,
		playbook: req.playbook._id,
		status: 'Queued'
	});

	task.save(function (err) {
		task.populate('job', function () {
			app.io.emit('playbook.update', {
				task_id: task._id,
				playbook_id: req.playbook._id,
				task: task
			});
		})
	});

	res.send(201)
}