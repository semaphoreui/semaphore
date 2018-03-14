var fs = require('fs'),
	bundle = require('./bundle.json'),
	out = fs.createWriteStream('./public/js/bundle.js');

bundle.forEach(function(file) {
	var o = {};
	if (typeof file === 'object') {
		o = file;
		file = o.src;
	}

	file = file + '.js';

	var contents = fs.readFileSync(file);

	out.write('\n/* BUNDLED FILE: ' + file + ' */\n');

	if (o.pre) {
		out.write(o.pre + '\n');
	}

	out.write(contents + '\n');

	if (o.post) {
		out.write(o.post + '\n');
	}
});