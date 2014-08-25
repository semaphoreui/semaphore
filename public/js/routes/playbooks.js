define([
	'app',
	'factories/playbook'
], function(app) {
	app.config(function($stateProvider, $couchPotatoProvider) {
		$stateProvider
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
			templateUrl: '/view/playbook/view',
			controller: function ($scope, playbook) {
				$scope.playbook = playbook;
			},
			resolve: {
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
			url: '/',
			controller: 'PlaybookCtrl',
			resolve: {
				dummy: $couchPotatoProvider.resolve(['controllers/playbook/playbook',
						'controllers/host/hosts',
						'controllers/job/jobs',
						'controllers/task/tasks'])
			},
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
	})
})