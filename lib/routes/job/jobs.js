var models = require('../../models')
var mongoose = require('mongoose')

var job = require('./job')

exports.unauthorized = function (app, template) {
	template([
		'jobs'
	], {
		prefix: 'job'
	});

	job.unauthorized(app, template);
}

exports.httpRouter = function (app) {
	job.httpRouter(app);
}

exports.router = function (app) {
	app.get('/playbook/:playbook_id/jobs', get)
		.post('/playbook/:playbook_id/jobs', add)

	job.router(app);
}

function get (req, res) {
	models.Job.find({
		playbook: req.playbook._id
	}).sort('-created').exec(function (err, jobs) {
		res.send(jobs)
	})
}

function add (req, res) {
	var job = new models.Job({
		playbook: req.playbook._id,
		name: req.body.name,
		play_file: req.body.play_file,
		use_vault: req.body.use_vault
	})
	
	job.save(function () {
		res.send(job);
	});
}