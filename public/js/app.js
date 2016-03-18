var app = angular.module('semaphore', ['scs.couch-potato', 'ui.router', 'ui.bootstrap', 'angular-loading-bar']);

couchPotato.configureApp(app);

app.run(['$rootScope', '$window', '$couchPotato', '$injector', '$state', '$http', function ($rootScope, $window, $couchPotato, $injector, $state, $http) {
	app.lazy = $couchPotato;

	$rootScope.$on('$stateChangeStart', function (event, toState, toParams, fromState, fromParams) {
		if (toState.pageTitle) {
			$rootScope.pageTitle = "Loading " + toState.pageTitle;
		} else {
			$rootScope.pageTitle = "Loading..";
		}
	});

	$rootScope.$on('$stateChangeSuccess', function (event, toState, toParams, fromState, fromParams) {
		$rootScope.previousState = {
			name: fromState.name,
			params: fromParams
		}

		if (toState.pageTitle) {
			$rootScope.pageTitle = toState.pageTitle;
		} else {
			$rootScope.pageTitle = "Ansible-Semaphore Page";
		}
	});

	$rootScope.refreshUser = function () {
		$rootScope.user = null;
		$rootScope.loggedIn = false;

		$http.get('/api/user')
		.then(function (user) {
			$rootScope.user = user;
			$rootScope.loggedIn = true;
		}, function () {
			$state.go('login');
		});
	}

	$rootScope.refreshUser();
}]);