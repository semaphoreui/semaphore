var mongoose = require('mongoose')
var ObjectId = mongoose.Schema.ObjectId;

var schema = mongoose.Schema({
	created: {
		type: Date,
		default: Date.now
	},
	playbook: {
		type: ObjectId,
		ref: 'Playbook'
	},
	name: String,
	play_file: String, //x.yml
	use_vault: Boolean
});

schema.index({
	name: 1
});

module.exports = mongoose.model('Job', schema);