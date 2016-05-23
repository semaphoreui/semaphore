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
    abstract: true,
    templateUrl: '/tpl/abstract.html'
  })
	.state('users.list', {
    url: '',
    pageTitle: 'Users',
    templateUrl: '/tpl/users/list.html',
    controller: 'UsersCtrl',
    resolve: {
      $d: $couchPotatoProvider.resolve(['controllers/users'])
    }
  })
  .state('users.user', {
    url: '/:user_id',
    pageTitle: 'User',
    templateUrl: '/tpl/users/user.html',
    controller: 'UserCtrl',
    resolve: {
      $d: $couchPotatoProvider.resolve(['controllers/user']),
      user: ['$http', '$stateParams', function ($http, $stateParams) {
      	return $http.get('/users/' + $stateParams.user_id);
      }]
    }
  })
  .state('admin', {
  	url: '/admin',
  	pageTitle: 'System Info',
  	templateUrl: '/tpl/admin.html',
  	controller: 'AdminCtrl',
  	resolve: {
  		$d: $couchPotatoProvider.resolve(['controllers/admin'])
  	}
  })
  .state('user', {
  	url: '/user',
  	pageTitle: 'User',
  	templateUrl: '/tpl/users/user.html',
  	controller: 'UserCtrl',
  	resolve: {
  		$d: $couchPotatoProvider.resolve(['controllers/user']),
  		user: ['$http', function ($http) {
  			return $http.get('/user');
  		}]
  	}
  });
});

app.run(function($rootScope, $state, $stateParams, $http) {
	$rootScope.$state = $state;
	$rootScope.$stateParams = $stateParams;
});