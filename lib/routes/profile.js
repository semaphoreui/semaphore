
exports.router = function (app, template) {
	app.get('/profile', getProfile)
}

function getProfile (req, res) {
	res.send({
		_id: req.user._id,
		name: req.user.name,
		username: req.user.username,
		email: req.user.email
	})
}