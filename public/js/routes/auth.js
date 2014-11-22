define([
	'app'
], function(app) {
	app.config(function($stateProvider, $urlRouterProvider, $locationProvider, $couchPotatoProvider) {
		$locationProvider.html5Mode({
			enabled: true,
			requireBase: false
		})


		$urlRouterProvider.otherwise('/');

		$stateProvider.state('login', {
			url: '/',
			pageTitle: "Sign In",
			templateUrl: "/view/auth/login",
			controller: "SignInCtrl",
			resolve: {
				dummy: $couchPotatoProvider.resolveDependencies(['controllers/auth/login'])
			}
		})
	});
});
