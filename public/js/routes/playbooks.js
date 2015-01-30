define([
	'app',
	'factories/playbook'
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
				dummy: $couchPotatoProvider.resolve(['controllers/playbook/playbook',
						'controllers/host/hosts',
						'controllers/job/jobs',
						'controllers/task/tasks']),
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

		.state('playbook.view', {
			url: '',
			views: {
				tasks: {
					templateUrl: '/view/task/tasks',
					controller: 'TasksCtrl'
				},
				jobs: {
					templateUrl: '/view/job/jobs',
					controller: 'JobsCtrl'
				},
				hosts: {
					templateUrl: '/view/host/hosts',
					controller: 'HostsCtrl'
				}
			}
		})

		.state('playbook.edit', {
			url: '/edit',
			templateUrl: "/view/playbook/add",
			controller: 'EditPlaybookCtrl',
			resolve: {
				dummy: $couchPotatoProvider.resolve(['controllers/playbook/edit'])
			},
			views: {
				tasks: {
					templateUrl: '/view/playbook/add',
					controller: 'EditPlaybookCtrl'
				}
			}
		})
	})
})