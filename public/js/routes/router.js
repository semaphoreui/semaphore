app.config(function ($stateProvider, $urlRouterProvider, $locationProvider, $couchPotatoProvider) {
	$locationProvider.html5Mode({
		enabled: true,
		requireBase: false
	});

	$urlRouterProvider.otherwise('/');

	$stateProvider
	.state('homepage', {
		url: '/',
		pageTitle: 'Homepage',
		templateUrl: "/public/html/homepage.html"
	})

	.state('logout', {
		url: '/logout',
		pageTitle: 'Log Out',
		controller: function ($scope) {
			window.location = "/logout";
		}
	})
});

app.run(function($rootScope, $state, $stateParams, $http) {
	$rootScope.$state = $state;
	$rootScope.$stateParams = $stateParams;
});