var manifest = [
	'Host',
	'HostGroup',
	'Job',
	'Identity',
	'Playbook',
	'Task',
	'User'
];

manifest.forEach(function (model) {
	module.exports[model] = require('./'+model);
});