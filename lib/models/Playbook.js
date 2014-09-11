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
	identity: {
		type: ObjectId,
		ref: 'Identity'
	}
});

schema.index({
	name: 1
});

module.exports = mongoose.model('Playbook', schema);