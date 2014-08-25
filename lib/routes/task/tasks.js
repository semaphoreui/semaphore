var models = require('../../models')
var mongoose = require('mongoose')

var task = require('./task')

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
	app.get('/playbook/:playbook_id/job/:job_id/tasks', get)
		.post('/playbook/:playbook_id/job/:job_id/tasks', add)

	task.router(app);
}

function get (req, res) {
	models.Task.find({
		job: req.job._id
	}).sort('-created').exec(function (err, tasks) {
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