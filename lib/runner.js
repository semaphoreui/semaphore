var async = require('async'),
	fs = require('fs'),
	spawn = require('child_process').spawn;

var config = require('./config'),
	models = require('./models'),
	app = require('./app');

var home = process.env.HOME + '/';
var user = process.env.USER;

exports.queue = async.queue(worker, 1);

function worker (task, callback) {
	// Task is to be model Task

	// Download the git project
	// Set up hosts file
	// Set up vault pwd file
	// Set up private key
	// Execute ansible-playbook -i hosts --ask-vault-pass --private-key=~/.ssh/ansible_key task.yml

	async.waterfall([
		function (done) {
			task.populate('job', function (err) {
				done(err, task);
			})
		},
		function (task, done) {
			models.Playbook.findOne({ _id: task.playbook }, function (err, playbook) {
				done(err, task, playbook)
			});
		},
		function (task, playbook, done) {
			// mark task as running and send an update via socketio
			task.status = 'Running';

			app.io.emit('playbook.update', {
				task_id: task._id,
				playbook_id: playbook._id,
				task: task
			});

			models.Task.update({
				_id: task._id
			}, {
				$set: {
					status: 'Running'
				}
			}, function (err) {
				done(err, task, playbook);
			});
		},
		function (task, playbook, done) {
			playbook.populate('identity', function (err) {
				done(err, task, playbook)
			});
		},
		installHostKeys,
		pullGit,
		setupHosts,
		setupVault,
		playTheBook
	], function (err) {
		if (err) {
			task.status = 'Failed';
		} else {
			task.status = 'Completed';
		}

		var rmrf = spawn('rm', ['-rf', home + 'playbook_'+task.playbook])
		rmrf.on('close', function () {
			app.io.emit('playbook.update', {
				task_id: task._id,
				playbook_id: task.playbook,
				task: task
			});
			task.save();

			callback(err);
		});
	});
}

function installHostKeys (task, playbook, done) {
	// Install the private key
	playbookOutputHandler.call(task, "Updating SSH Keys\n");

	var location = home + '.ssh/id_rsa';
	fs.mkdir( home + '.ssh', 448, function() {
		async.parallel([
			function (done) {
				fs.writeFile(location, playbook.identity.private_key, {
					mode: 384 // base 8 = 0600
				}, done);
			},
			function (done) {
				fs.writeFile(location+'.pub', playbook.identity.public_key, {
					mode: 420 // base 8 = 0644
				}, done);
			},
			function (done) {
				var config = "Host *\n\
  StrictHostKeyChecking no\n\
  CheckHostIp no\n\
  PasswordAuthentication no\n\
  PreferredAuthentications publickey\n";

				fs.writeFile(home + '.ssh/config', config, {
					mode: 420 // 0644
				}, done);
			}
		], function (err) {
			playbookOutputHandler.call(task, "SSH Keys Updated.\n");
			done(err, task, playbook)
		});
	});
}

function pullGit (task, playbook, done) {
	// Pull from git
	playbookOutputHandler.call(task, "\nDownloading Playbook.\n");

	var install = spawn(config.path+"/scripts/pullGit.sh", [playbook.location, 'playbook_'+playbook._id], {
		cwd: home,
		env: {
			HOME: home,
			OLDPWD: home,
			PWD: home,
			LOGNAME: user,
			USER: user,
			TERM: 'xterm',
			SHELL: '/bin/bash',
			PATH: process.env.PATH+':/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin:/root/bin',
			LANG: 'en_GB.UTF-8'
		}
	});
	install.stdout.on('data', playbookOutputHandler.bind(task));
	install.stderr.on('data', playbookOutputHandler.bind(task));

	install.on('close', function(code) {
		playbookOutputHandler.call(task, "\n\nPlaybook Downloaded.\n");

		done(null, task, playbook);
	});
}

function setupHosts (task, playbook, done) {
	var hostfile = '';

	models.HostGroup.find({
		playbook: playbook._id
	}, function (err, hostgroups) {
		async.each(hostgroups, function (group, cb) {
			models.Host.find({
				group: group._id
			}, function (err, hosts) {
				hostfile += "["+group.name+"]\n";

				for (var i = 0; i < hosts.length; i++) {
					hostfile += hosts[i].hostname;
					if ( hosts[i].vars && hosts[i].vars.length > 0 ) {
						hostfile +=" " + hosts[i].vars;
					}

					hostfile += "\n";
				}

				cb();
			});
		}, function () {
			playbookOutputHandler.call(task, "\nSet up Ansible Hosts file with contents:\n"+hostfile+"\n");

			fs.writeFile(home + 'playbook_'+playbook._id+'/semaphore_hosts', hostfile, function (err) {
				done(err, task, playbook);
			});
		});
	});
}

function setupVault (task, playbook, done) {
	fs.writeFile(home + 'playbook_'+playbook._id+'/semaphore_vault_pwd', playbook.vault_password, function (err) {
		done(err, task, playbook);
	})
}

function playTheBook (task, playbook, done) {
	playbookOutputHandler.call(task, "\nStarting play "+task.job.play_file+".\n");

	var args = ['-i', 'semaphore_hosts'];
	if (task.job.use_vault && playbook.vault_password && playbook.vault_password.length > 0) {
		args.push('--vault-password-file='+'semaphore_vault_pwd');
	}

	// private key to login to server[s]
	args.push('--private-key=' + home + '.ssh/id_rsa');

	// the playbook file
	args.push(task.job.play_file);

	var playbook = spawn("ansible-playbook", args, {
		cwd: home + 'playbook_'+playbook._id,
		env: {
			HOME: home,
			OLDPWD: home,
			PWD: home + 'playbook_'+playbook._id,
			LOGNAME: user,
			USER: user,
			TERM: 'xterm',
			SHELL: '/bin/bash',
			PATH: process.env.PATH+':/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin:/root/bin',
			LANG: 'en_GB.UTF-8',
			PYTHONPATH: process.env.PYTHONPATH,
			PYTHONUNBUFFERED: 1
		}
	});
	playbook.stdout.on('data', playbookOutputHandler.bind(task));
	playbook.stderr.on('data', playbookOutputHandler.bind(task));

	playbook.on('close', function(code) {
		console.log('done.', code);

		if (code !== 0) {
			// Task failed
			return done('Failed with code '+code);
		}

		done();
	});
}

function playbookOutputHandler (chunk) {
	chunk = chunk.toString('utf8');

	if (!this.output) {
		this.output = "";
	}

	this.output += chunk;
	app.io.emit('playbook.output', {
		task_id: this._id,
		playbook_id: this.playbook,
		output: chunk
	});

	console.log(chunk);
}
