var mongoose = require('mongoose')
var ObjectId = mongoose.Schema.ObjectId;

var schema = mongoose.Schema({
	created: {
		type: Date,
		default: Date.now
	},
	credential_type: {
		type: String,
		enum: ['ssh', 'vault', 'git']
	},
	name: String,
	password: String,
	private_key: String,
	public_key: String
});

schema.index({
	name: 1
});

module.exports = mongoose.model('Credential', schema);