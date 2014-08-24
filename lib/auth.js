var passport = require('passport')
	, models = require('./models')
	, bugsnag = require('bugsnag')

passport.serializeUser(function(user, done) {
	done(null, user._id);
});

passport.deserializeUser(function(id, done) {
	models.User.findOne({
		_id: id
	}, done);
})