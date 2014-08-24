var mongoose = require('mongoose')
var ObjectId = mongoose.Schema.ObjectId;

var schema = mongoose.Schema({
	created: {
		type: Date,
		default: Date.now
	},
	name: String,
	location: String, // Git URL
	vault_password: String,
	credential: {
		type: ObjectId,
		ref: 'Credential'
	}
});

schema.index({
	name: 1
});

module.exports = mongoose.model('Playbook', schema);