var manifest = [
	'Credential',
	'Host',
	'HostGroup',
	'Job',
	'Playbook',
	'Task',
	'User'
];

manifest.forEach(function (model) {
	module.exports[model] = require('./'+model);
});