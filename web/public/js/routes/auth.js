app.config(function ($stateProvider, $urlRouterProvider, $locationProvider, $couchPotatoProvider) {
	$stateProvider.state('auth', {
		url: '/auth',
		abstract: true,
		templateUrl: '/tpl/abstract.html'
	})
	.state('auth.login', {
		url: '/login',
		pageTitle: "Sign In",
		templateUrl: '/tpl/auth/login.html',
		controller: "SignInCtrl",
		resolve: {
			$d: $couchPotatoProvider.resolveDependencies(['controllers/login'])
		}
	})
	.state('auth.logout', {
		url: '/logout',
		public: true,
		templateUrl: '/tpl/auth/logout.html',
		controller: ['$http', '$rootScope', '$state', function ($http, $rootScope, $state) {
			$http.post('/auth/logout').then(function () {
				$rootScope.refreshUser();
				$state.go('auth.login');
			});
		}]
	});
});