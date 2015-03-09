var passport = require('passport')
	, models = require('../models')
	, validator = require('validator')
	, util = require('../util')
	, config = require('../config')
	, async = require('async')
	, express = require('express')
	, mongoose = require('mongoose')

exports.unauthorized = function (app, template) {
	// Unrestricted -- non-authorized people can access!
	template([
		'login'
	], {
		prefix: 'auth'
	});

	var auth = express.Router();

	auth.post('/password', doLogin)
		.get('/loggedin', isLoggedIn)
		.get('/logout', doLogout)

	app.use('/auth', auth);
}

exports.router = function (app) {
	// Restricted -- only authorized people can access!
	app.post('/auth/register', doRegister)
}

function isLoggedIn (req, res) {
	res.send({
		hasSession: req.user != null,
		isLoggedIn: res.locals.loggedIn
	});
}

function doLogin (req, res) {
	var auth = req.body.auth;
	var isValid = true;

	if (!validator.isLength(auth, 4)) {
		isValid = false;
	}

	// validate password
	if (!validator.isLength(req.body.password, 6)) {
		isValid = false;
	}

	if (!isValid) {
		return authCallback(false, null, req, res);
	}

	var query = {
		email: auth.toLowerCase()
	};

	models.User.findOne(query, function(err, user) {
		if (err) {
			throw err;
		}

		if (user == null) {
			return authCallback(false, null, req, res);
		}

		user.comparePassword(req.body.password, function (matches) {
			authCallback(matches, user, req, res);
		});
	})
}

function authCallback (isValid, user, req, res) {
	if (!isValid) {
		res.send(400, {
			message: "Nope. Incorrect Credentials!"
		});

		return;
	}

	req.login(user, function(err) {
		if (err) throw err;

		res.send(201)
	})
}

function doRegister (req, res) {
	var errs = {
		name: false,
		email: false,
		password: false,
		username: false
	};

	var userObject = req.body.user;
	if (!(userObject && typeof userObject === 'object')) {
		return res.send(400, {
			message: 'Invalid Request'
		});
	}

	var email = userObject.email;
	if (email) {
		email = email.toLowerCase();
	}
	var password = userObject.password;
	var username = userObject.username;
	var name = userObject.name;

	errs.email = !validator.isEmail(email);
	errs.username = !validator.isLength(username, 3, 15);
	errs.name = !validator.isLength(name, 4, 50);

	if (!(username && username.match(/^[a-zA-Z0-9_-]{3,15}$/) && validator.isAscii(username))) {
		// Errornous username
		errs.username = true;
	}

	// validate password
	errs.password = !validator.isLength(password, 8, 100);

	if (!(errs.username == false && errs.password == false && errs.name == false && errs.email == false)) {
		res.send(400, {
			fields: errs,
			message: ''
		});

		return;
	}

	// Register
	var user = new models.User({
		email: email,
		username: username,
		name: name
	});

	models.User.hashPassword(password, function (hash) {
		user.password = hash;

		user.save();

		// log in now
		req.login(user, function(err) {
			if (err) throw err;

			res.send({
				message: "Registration Successful",
				user_id: user._id
			});
		});
	});
}

function doLogout (req, res) {
	req.logout();
	req.session.destroy();

	res.format({
		json: function() {
			res.send(201)
		},
		html: function() {
			res.redirect('/')
		}
	})
}