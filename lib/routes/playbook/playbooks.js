var models = require('../../models')
var mongoose = require('mongoose')

var playbook = require('./playbook')

exports.unauthorized = function (app, template) {
	template([
		'add',
		'list'
	], {
		prefix: 'playbook'
	});

	playbook.unauthorized(app, template);
}

exports.httpRouter = function (app) {
	playbook.httpRouter(app);
}

exports.router = function (app) {
	app.get('/playbooks', getPlaybooks)
		.post('/playbooks', addPlaybook)

	playbook.router(app);
}

function getPlaybooks (req, res) {
	models.Playbook.find({
	}).sort('-created').select('-vault_password').exec(function (err, playbooks) {
		res.send(playbooks)
	})
}

function addPlaybook (req, res) {
	var playbook = new models.Playbook({
		name: req.body.name,
		location: req.body.location,
		vault_password: req.body.vault_password
	})
	
	if (typeof req.body.identity == 'string' && req.body.identity.length > 0) {
		try {
			playbook.identity = mongoose.Types.ObjectId(req.body.identity);
		} catch (e) {}
	}

	playbook.save(function () {
		res.send(playbook);
	});
}