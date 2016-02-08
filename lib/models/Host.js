var mongoose = require('mongoose')
var ObjectId = mongoose.Schema.ObjectId;

var schema = mongoose.Schema({
	created: {
		type: Date,
		default: Date.now
	},
	name: String,
	hostname: String,
	vars: String,
	group: {
		type: ObjectId,
		ref: 'HostGroup'
	},
	playbook: {
		type: ObjectId,
		ref: 'Playbook'
	}
});

schema.index({
	name: 1,
	hostname: 1,
	vars: 1
});

module.exports = mongoose.model('Host', schema);