define([
	'app',
	'factories/playbook',
	'factories/job'
], function(app) {
	app.config(function($stateProvider, $couchPotatoProvider) {
		$stateProvider

		.state('playbooks', {
			url: '/playbooks',
			pageTitle: 'Playbooks',
			templateUrl: '/view/playbook/list',
			controller: 'PlaybooksCtrl',
			resolve: {
				dummy: $couchPotatoProvider.resolve(['controllers/playbook/list'])
			}
		})
		.state('addPlaybook', {
			url: '/add',
			pageTitle: 'Add Playbook',
			templateUrl: "/view/playbook/add",
			controller: 'AddPlaybookCtrl',
			resolve: {
				dummy: $couchPotatoProvider.resolve(['controllers/playbook/add'])
			}
		})

		.state('playbook', {
			abstract: true,
			url: '/playbook/:playbook_id',
			controller: 'PlaybookCtrl',
			templateUrl: '/view/playbook/view',
			resolve: {
				dummy: $couchPotatoProvider.resolve(['controllers/playbook/playbook']),
				playbook: function (Playbook, $stateParams, $q, $state) {
					var deferred = $q.defer();

					var playbook = new Playbook($stateParams.playbook_id, function (err, errStatus) {
						if (err && errStatus == 404) {
							$state.transitionTo('homepage');
							return deferred.reject();
						}

						deferred.resolve(playbook);
					});

					return deferred.promise;
				}
			}
		})

		.state('playbook.edit', {
			url: '/edit',
			templateUrl: "/view/playbook/add",
			controller: 'EditPlaybookCtrl',
			resolve: {
				dummy: $couchPotatoProvider.resolve(['controllers/playbook/edit'])
			}
		})

		.state('playbook.tasks', {
			url: '/tasks',
			templateUrl: "/view/playbook/tasks",
			controller: 'PlaybookTasksCtrl',
			resolve: {
				dummy: $couchPotatoProvider.resolve(['controllers/playbook/tasks'])
			}
		})
		.state('playbook.jobs', {
			url: '/jobs',
			templateUrl: "/view/playbook/jobs",
			controller: 'PlaybookJobsCtrl',
			resolve: {
				dummy: $couchPotatoProvider.resolve(['controllers/playbook/jobs'])
			}
		})
		.state('playbook.hosts', {
			url: '/hosts',
			templateUrl: "/view/playbook/hosts",
			controller: 'PlaybookHostsCtrl',
			resolve: {
				dummy: $couchPotatoProvider.resolve(['controllers/playbook/hosts'])
			}
		})
		.state('playbook.job', {
			url: '/job/:job_id',
			templateUrl: "/view/playbook/job",
			controller: 'PlaybookJobCtrl',
			resolve:{
				dummy: $couchPotatoProvider.resolve(['controllers/playbook/job']),
				job: function (Job, playbook, $stateParams, $q, $state) {
					var deferred = $q.defer();

					var job = new Job(playbook,$stateParams.job_id, function (err, errStatus) {
						if (err && errStatus == 404) {
							$state.transitionTo('homepage');
							return deferred.reject();
						}

						deferred.resolve(job);
					});

					return deferred.promise;
				}
			}
		})
		.state('playbook.newjob', {
			url: '/job/new',
			templateUrl: "/view/playbook/job",
			controller: 'PlaybookJobCtrl',
			resolve:{
				dummy: $couchPotatoProvider.resolve(['controllers/playbook/job'])
			}
		})
	})
})