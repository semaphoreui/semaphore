var models = require('../../models')
var mongoose = require('mongoose')

var identity = require('./identity')

var validator = require('validator')

exports.unauthorized = function (app, template) {
	template([
		'add',
		'list'
	], {
		prefix: 'identity'
	});

	identity.unauthorized(app, template);
}

exports.httpRouter = function (app) {
	identity.httpRouter(app);
}

exports.router = function (app) {
	app.get('/identities', get)
		.post('/identities', add)

	identity.router(app);
}

function get (req, res) {
	models.Identity.find({
	}).sort('-created').select('-public_key -private_key -password').exec(function (err, identities) {
		res.send(identities)
	})
}

function add (req, res) {
	if (!validator.isLength(req.body.name, 1)) {
		return res.send(400);
	}
	
	var identity = new models.Identity({
		name: req.body.name,
		password: req.body.password,
		private_key: req.body.private_key,
		public_key: req.body.public_key
	});

	identity.save(function () {
		res.send(identity);
	});
}