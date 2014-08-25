var models = require('../../models')
var mongoose = require('mongoose')
var express = require('express')

exports.unauthorized = function (app, template) {
	template([
		'view'
	], {
		prefix: 'job'
	});
}

exports.httpRouter = function (app) {
}

exports.router = function (app) {
	var job = express.Router();

	job.get('/', view)
		.delete('/', remove)

	app.param('job_id', get)
	app.use('/playbook/:playbook_id/job/:job_id', job);
}

function get (req, res, next, id) {
	models.Job.findOne({
		_id: id
	}).exec(function (err, job) {
		if (err || !job) {
			return res.send(404);
		}

		req.job = job;
		next();
	});
}

function view (req, res) {
	res.send(req.job);
}

function remove (req, res) {
	req.job.remove(function (err) {
		res.send(201);
	})
}