app.config(function ($stateProvider, $urlRouterProvider, $locationProvider, $couchPotatoProvider) {
	$stateProvider.state('login', {
		url: '/',
		pageTitle: "Sign In",
		templateUrl: "/public/html/auth/login.html",
		controller: "SignInCtrl",
		resolve: {
			dummy: $couchPotatoProvider.resolveDependencies(['controllers/auth/login'])
		}
	})
});