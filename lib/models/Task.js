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
	status: {
		type: String,
		enum: ['Failed', 'Running', 'Queued']
	}
});

schema.index({
	status: 1
});

module.exports = mongoose.model('Task', schema);