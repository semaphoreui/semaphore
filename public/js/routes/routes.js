define([
	'app',
	'socketio',
	'services/user',
	'routes/playbooks',
	'routes/identities',
	'services/playbooks'
], function(app, io) {
	app.config(function($stateProvider, $urlRouterProvider, $locationProvider, $couchPotatoProvider) {
		$locationProvider.html5Mode({
			enabled: true,
			requireBase: false
		})
		
		$urlRouterProvider.otherwise('');
		
		$stateProvider
		.state('homepage', {
			url: '/',
			pageTitle: 'Homepage',
			templateUrl: "/view/homepage"
		})
		
		.state('logout', {
			url: '/logout',
			pageTitle: 'Log Out',
			controller: function($scope) {
				window.location = "/logout";
			}
		})
	})
	.run(function($rootScope, $state, $stateParams, $http, user, playbooks) {
		$rootScope.$state = $state
		$rootScope.$stateParams = $stateParams

		user.getUser(function() {});
		playbooks.getPlaybooks(function() {});
	})
})
