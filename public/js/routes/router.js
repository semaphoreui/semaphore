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
	})
	.state('users', {
    url: '/users',
    pageTitle: 'Users',
    templateUrl: "/tpl/users/list.html",
    controller: 'UsersCtrl',
    resolve: {
      $d: $couchPotatoProvider.resolve(['controllers/users'])
    }
  });
});

app.run(function($rootScope, $state, $stateParams, $http) {
	$rootScope.$state = $state;
	$rootScope.$stateParams = $stateParams;
});