var mongoose = require('mongoose')
var ObjectId = mongoose.Schema.ObjectId;

var schema = mongoose.Schema({
	created: {
		type: Date,
		default: Date.now
	},
	job: {
		type: ObjectId,
		ref: 'Job'
	},
	playbook: {
		type: ObjectId,
		ref: 'Playbook'
	},
	output: String,
	status: {
		type: String,
		enum: ['Completed', 'Failed', 'Running', 'Queued']
	}
});

schema.index({
	status: 1
});

module.exports = mongoose.model('Task', schema);