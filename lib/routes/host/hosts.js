var models = require('../../models')
var mongoose = require('mongoose')

var host = require('./host')

exports.unauthorized = function (app, template) {
	template([
		'hosts'
	], {
		prefix: 'host'
	});

	host.unauthorized(app, template);
}

exports.httpRouter = function (app) {
	host.httpRouter(app);
}

exports.router = function (app) {
	app.get('/playbook/:playbook_id/hostgroups', get)
		.post('/playbook/:playbook_id/hostgroups', add)

	host.router(app);
}

function get (req, res) {
	models.HostGroup.find({
		playbook: req.playbook._id
	}).sort('-created').exec(function (err, hosts) {
		res.send(hosts)
	})
}

function add (req, res) {
	var hostgroup = new models.HostGroup({
		playbook: req.playbook._id,
		name: req.body.name
	});
	
	hostgroup.save(function () {
		res.send(hostgroup);
	});
}