app.config(function ($stateProvider, $couchPotatoProvider) {
	$stateProvider.state('identities', {
		url: '/identities',
		templateUrl: '/view/abstract.html',
		abstract: true
	})
	.state('identities.add', {
		url: '/add',
		pageTitle: 'Add Identity',
		templateUrl: "/public/html/identity/add.html.html",
		controller: 'AddIdentityCtrl',
		resolve: {
			dummy: $couchPotatoProvider.resolve(['controllers/identity/add'])
		}
	})
	.state('identities.list', {
		url: '/all',
		pageTitle: 'Identities',
		templateUrl: "/public/html/identity/list.html",
		controller: 'IdentitiesCtrl',
		resolve: {
			dummy: $couchPotatoProvider.resolve(['controllers/identity/identities'])
		}
	})

	.state('identity', {
		abstract: true,
		url: '/identity/:identity_id',
		templateUrl: '/public/html/abstract.html',
		controller: ['$scope', 'identity', function ($scope, identity) {
			$scope.identity = identity;
		}],
		resolve: {
			identity: ['Identity', '$stateParams', '$q', '$state', function (Identity, $stateParams, $q, $state) {
				var deferred = $q.defer();

				var identity = new Identity($stateParams.identity_id)
				identity.get()
				.success(function (data, status) {
					identity.data = data;
					deferred.resolve(identity);
				})
				.error(function (data, status) {
					if (status == 404) {
						$state.transitionTo('homepage');
						return deferred.reject();
					}
				});

				return deferred.promise;
			}]
		}
	})

	.state('identity.view', {
		url: '/',
		controller: 'IdentityCtrl',
		templateUrl: '/public/html/identity/view.html',
		resolve: {
			dummy: $couchPotatoProvider.resolve(['controllers/identity/identity'])
		}
	})
});