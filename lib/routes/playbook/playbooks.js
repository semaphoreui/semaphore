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
	
	if (req.body.credential && req.body.credential.length > 0) {
		try {
			playbook.credential = mongoose.Types.ObjectId(req.body.credential)
		} catch (e) {}
	}

	playbook.save(function () {
		res.send(playbook);
	});
}