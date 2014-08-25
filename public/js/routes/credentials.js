define([
	'app',
	'factories/playbook'
], function(app) {
	app.config(function($stateProvider, $couchPotatoProvider) {
		$stateProvider

		.state('credentials', {
			url: '/credentials',
			templateUrl: '/view/abstract',
			abstract: true
		})
		.state('credentials.add', {
			url: '/add',
			pageTitle: 'Add Credential',
			templateUrl: "/view/credential/add",
			controller: 'AddCredentialCtrl',
			resolve: {
				dummy: $couchPotatoProvider.resolve(['controllers/credential/add'])
			}
		})
		.state('credentials.list', {
			url: '/all',
			pageTitle: 'Credentials',
			templateUrl: "/view/credential/list",
			controller: 'CredentialsCtrl',
			resolve: {
				dummy: $couchPotatoProvider.resolve(['controllers/credential/credentials'])
			}
		})

		.state('credential', {
			abstract: true,
			url: '/credential/:credential_id',
			templateUrl: '/view/abstract',
			controller: function ($scope, credential) {
				$scope.credential = credential;
			},
			resolve: {
				credential: function (Credential, $stateParams, $q, $state) {
					var deferred = $q.defer();

					var credential = new Credential($stateParams.credential_id)
					credential.get()
					.success(function (data, status) {
						credential.data = data;
						deferred.resolve(credential);
					})
					.error(function (data, status) {
						if (status == 404) {
							$state.transitionTo('homepage');
							return deferred.reject();
						}
					});

					return deferred.promise;
				}
			}
		})

		.state('credential.view', {
			url: '/',
			controller: 'CredentialCtrl',
			templateUrl: '/view/credential/view',
			resolve: {
				dummy: $couchPotatoProvider.resolve(['controllers/credential/credential'])
			}
		})
	})
})