app.config(function($stateProvider, $couchPotatoProvider) {
	$stateProvider.state('playbooks', {
		url: '/playbooks',
		pageTitle: 'Playbooks',
		templateUrl: '/public/html/playbook/list.html',
		controller: 'PlaybooksCtrl',
		resolve: {
			dummy: $couchPotatoProvider.resolve(['controllers/playbook/list'])
		}
	})
	.state('addPlaybook', {
		url: '/add',
		pageTitle: 'Add Playbook',
		templateUrl: "/public/html/playbook/add.html",
		controller: 'AddPlaybookCtrl',
		resolve: {
			dummy: $couchPotatoProvider.resolve(['controllers/playbook/add'])
		}
	})

	.state('playbook', {
		abstract: true,
		url: '/playbook/:playbook_id',
		controller: 'PlaybookCtrl',
		templateUrl: '/public/html/playbook/view.html',
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
		templateUrl: "/public/html/playbook/add.html",
		controller: 'EditPlaybookCtrl',
		resolve: {
			dummy: $couchPotatoProvider.resolve(['controllers/playbook/edit'])
		}
	})

	.state('playbook.tasks', {
		url: '/tasks',
		templateUrl: "/public/html/playbook/tasks.html",
		controller: 'PlaybookTasksCtrl',
		resolve: {
			dummy: $couchPotatoProvider.resolve(['controllers/playbook/tasks'])
		}
	})
	.state('playbook.jobs', {
		url: '/jobs',
		templateUrl: "/public/html/playbook/jobs.html",
		controller: 'PlaybookJobsCtrl',
		resolve: {
			dummy: $couchPotatoProvider.resolve(['controllers/playbook/jobs'])
		}
	})
	.state('playbook.hosts', {
		url: '/hosts',
		templateUrl: "/public/html/playbook/hosts.html",
		controller: 'PlaybookHostsCtrl',
		resolve: {
			dummy: $couchPotatoProvider.resolve(['controllers/playbook/hosts'])
		}
	})
});