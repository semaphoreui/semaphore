define([
	'angular',
	'couchPotato'
], function(angular, couchPotato, ga) {
	var app = angular.module('semaphore', ['scs.couch-potato', 'ui.router']);

	couchPotato.configureApp(app);

	app.run(function($rootScope, $window, $couchPotato) {
		app.lazy = $couchPotato;
		
		$rootScope.$on('$locationChangeStart', function(event, newUrl, oldUrl){
			if (newUrl.match(/\&no_router/)) {
				event.preventDefault();
				$window.location.href = newUrl.replace(/\&no_router/, '');
			}
		});

		$rootScope.$on('$stateChangeStart', function (event, toState, toParams, fromState, fromParams) { 
			if (toState.pageTitle) {
				$rootScope.pageTitle = "Loading " + toState.pageTitle;
			} else {
				$rootScope.pageTitle = "Loading..";
			}
		});
		$rootScope.$on('$stateChangeSuccess', function (event, toState, toParams, fromState, fromParams){
			if (toState.pageTitle) {
				$rootScope.pageTitle = toState.pageTitle;
			} else {
				$rootScope.pageTitle = "Semaphore Page";
			}
		})
	});
	
	return app;
})