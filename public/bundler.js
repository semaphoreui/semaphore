var fs = require('fs'),
	async = require('async'),
	bundle = require('./bundle.json'),
	out = fs.createWriteStream('./js/bundle.js'),
	path = require('path'),
	child_process = require('child_process'),
	uglify = null;

bundle.forEach(file => {
	var o = {}
	if (typeof file == 'object') {
		o = file;
		file = o.src;
	}

	if (file.substr(0, 1) != '/') {
		file = '/js/' + file;
	}

	file = file + '.js';

	console.log(file);
	var contents = fs.readFileSync('.' + file);

	out.write('\n/* BUNDLED FILE: ' + file + ' */\n')
	if (o.pre) {
		out.write(o.pre + '\n');
	}

	out.write(contents + '\n');

	if (o.post) {
		out.write(o.post + '\n');
	}
});