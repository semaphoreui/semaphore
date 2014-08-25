var models = require('../../models')
var mongoose = require('mongoose')

var credential = require('./credential')

exports.unauthorized = function (app, template) {
	template([
		'add',
		'list'
	], {
		prefix: 'credential'
	});

	credential.unauthorized(app, template);
}

exports.httpRouter = function (app) {
	credential.httpRouter(app);
}

exports.router = function (app) {
	app.get('/credentials', get)
		.post('/credentials', add)

	credential.router(app);
}

function get (req, res) {
	models.Credential.find({
	}).sort('-created').select('-public_key -private_key -password').exec(function (err, credentials) {
		res.send(credentials)
	})
}

function add (req, res) {
	var credential = new models.Credential({
		name: req.body.name,
		password: req.body.password,
		private_key: req.body.private_key,
		public_key: req.body.public_key
	})

	credential.save(function () {
		res.send(credential);
	});
}