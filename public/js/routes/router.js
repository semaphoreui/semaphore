app.config(function ($stateProvider, $urlRouterProvider, $locationProvider, $couchPotatoProvider) {
	$locationProvider.html5Mode({
		enabled: true,
		requireBase: false
	});

	$urlRouterProvider.otherwise('/');

	$stateProvider
	.state('dashboard', {
		url: '/',
		pageTitle: 'Dashboard',
		templateUrl: '/tpl/dashboard.html',
		controller: 'DashboardCtrl',
		resolve: {
			$d: $couchPotatoProvider.resolveDependencies(['controllers/dashboard'])
		}
	});
});

app.run(function($rootScope, $state, $stateParams, $http) {
	$rootScope.$state = $state;
	$rootScope.$stateParams = $stateParams;
});