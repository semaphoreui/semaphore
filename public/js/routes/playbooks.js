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
			templateUrl: '/view/abstract',
			controller: function ($scope, playbook) {
				$scope.playbook = playbook;
			},
			resolve: {
				playbook: function (Playbook, $stateParams, $q) {
					var deferred = $q.defer();

					var playbook = new Playbook($stateParams.playbook_id, function (err, errStatus) {
						deferred.resolve(playbook);
					});

					return deferred.promise;
				}
			}
		})

		.state('playbook.view', {
			url: '/',
			controller: 'PlaybookCtrl',
			templateUrl: '/view/playbook/view',
			resolve: {
				dummy: $couchPotatoProvider.resolve(['controllers/playbook/playbook'])
			}
		})
	})
})