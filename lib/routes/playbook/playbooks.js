var models = require('../../models')

exports.unauthorized = function (app, template) {
	template([
		'add'
	], {
		prefix: 'playbook'
	});
}

exports.httpRouter = function () {
	
}

exports.router = function (app) {
	app.get('/playbooks', getPlaybooks)
}

function getPlaybooks (req, res) {
	models.Playbook.find({
	}, function (err, playbooks) {
		res.send(playbooks)
	})
}