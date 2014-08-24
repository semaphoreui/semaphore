exports.authorized = function (req, res, next) {
	if (res.locals.loggedIn) {
		next()
	} else {
		res.format({
			html: function() {
				res.render('auth');
			},
			json: function() {
				res.send(403, {
					message: "Unauthorized"
				})
			}
		})
	}
}
