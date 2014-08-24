var mongoose = require('mongoose')
var ObjectId = mongoose.Schema.ObjectId;

var schema = mongoose.Schema({
	created: {
		type: Date,
		default: Date.now
	},
	name: String,
	playbook: {
		type: ObjectId,
		ref: 'Playbook'
	}
});

schema.index({
	name: 1
});

module.exports = mongoose.model('HostGroup', schema);