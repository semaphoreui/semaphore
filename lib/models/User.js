var bcrypt = require('bcrypt')

var mongoose = require('mongoose')
var ObjectId = mongoose.Schema.ObjectId;

var schema = mongoose.Schema({
	created: {
		type: Date,
		default: Date.now
	},
	username: String,
	name: String,
	email: String,
	password: String
});

schema.index({
	email: 1
});

schema.statics.hashPassword = function(password, cb) {
	bcrypt.hash(password, 10, function(err, hash) {
		cb(hash);
	});
}

schema.methods.comparePassword = function (password, cb) {
	bcrypt.compare(password, this.password, function(err, res) {
		// res is boolean
		cb(res);
	})
}

module.exports = mongoose.model('User', schema);