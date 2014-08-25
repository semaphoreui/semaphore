var async = require('async');
var config = require('./config');
var models = require('./models');
var fs = require('fs');
var spawn = require('child_process').spawn
var app = require('./app');

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
			models.Playbook.findOne({ _id: task.job.playbook }, function (err, playbook) {
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
			playbook.populate('credential', function (err) {
				done(err, task, playbook)
			});
		},
		installHostKeys,
		pullGit,
		setupHosts,
		setupVault,
		playTheBook,
		function (task, playbook, done) {
			var rmrf = spawn('rm', ['-rf', '/root/playbook_'+playbook._id])
			rmrf.on('close', function () {
				done(null, playbook)
			})
		}
	], callback);
}

function installHostKeys (task, playbook, done) {
	// Install the private key
	var location = '/root/.ssh/id_rsa';
	fs.mkdir('/root/.ssh', 448, function() {
		async.parallel([
			function (done) {
				fs.writeFile(location, playbook.credential.private_key, {
					mode: 384 // base 8 = 0600
				}, done);
			},
			function (done) {
				fs.writeFile(location+'.pub', playbook.credential.public_key, {
					mode: 420 // base 8 = 0644
				}, done);
			},
			function (done) {
				var config = "Host *\n\
StrictHostKeyChecking no\n\
CheckHostIp no\n\
PasswordAuthentication no\n";

				fs.writeFile('/root/.ssh/config', config, {
					mode: 420 // 0644
				}, done);
			}
		], function (err) {
			done(err, task, playbook)
		});
	});
}

function pullGit (task, playbook, done) {
	// Pull from git
	var install = spawn(config.path+"/scripts/pullGit.sh", [playbook.location, 'playbook_'+playbook._id], {
		cwd: '/root/',
		env: {
			HOME: '/root/',
			OLDPWD: '/root/',
			PWD: '/root/',
			LOGNAME: 'root',
			USER: 'root',
			TERM: 'xterm',
			SHELL: '/bin/bash',
			PATH: '/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin:/root/bin',
			LANG: 'en_GB.UTF-8'
		}
	});
	install.stdout.on('data', function (chunk) {
		console.log('out', chunk.toString('utf8'))
	});
	install.stderr.on('data', function (chunk) {
		console.log('err', chunk.toString('utf8'))
	});

	install.on('close', function(code) {
		console.log('done.', code)
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
					hostfile += hosts[i].hostname+"\n";
				}

				cb();
			});
		}, function () {
			console.log(hostfile);

			fs.writeFile('/root/playbook_'+playbook._id+'/semaphore_hosts', hostfile, function (err) {
				done(err, task, playbook);
			});
		});
	});
}

function setupVault (task, playbook, done) {
	fs.writeFile('/root/playbook_'+playbook._id+'/semaphore_vault_pwd', playbook.vault_password, function (err) {
		done(err, task, playbook);
	})
}

function playTheBook (task, playbook, done) {
	var playbook = spawn("ansible-playbook", ['-i', 'semaphore_hosts', '--vault-password-file='+'semaphore_vault_pwd', '--private-key=/root/.ssh/id_rsa', task.job.play_file], {
		cwd: '/root/playbook_'+playbook._id,
		env: {
			HOME: '/root/',
			OLDPWD: '/root/',
			PWD: '/root/playbook_'+playbook._id,
			LOGNAME: 'root',
			USER: 'root',
			TERM: 'xterm',
			SHELL: '/bin/bash',
			PATH: '/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin:/root/bin',
			LANG: 'en_GB.UTF-8'
		}
	});
	playbook.stdout.on('data', function (chunk) {
		console.log('out', chunk.toString('utf8'))
	});
	playbook.stderr.on('data', function (chunk) {
		console.log('err', chunk.toString('utf8'))
	});

	playbook.on('close', function(code) {
		console.log('done.', code)
		done(null, task, playbook);
	});
}