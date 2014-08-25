require.config({
	paths: {
		angular: '../vendor/angular/angular.min',
		uiRouter: '../vendor/angular-ui-router/release/angular-ui-router.min',
		jquery: '../vendor/jquery/dist/jquery.min',
		moment: '../vendor/moment/moment',
		bootstrap: '../vendor/bootstrap/dist/js/bootstrap.min',
		couchPotato: '../vendor/angular-couch-potato/dist/angular-couch-potato'
	},
	shim: {
		angular: {
			exports: 'angular'
		},
		uiRouter: {
			deps: ['angular']
		},
		bootstrap: ['jquery']
	}
});

require([
	'jquery',
	'angular',
	'couchPotato',
	'uiRouter',
	'app',
	'routes/routes'
], function($, angular) {
	var $html = angular.element(document.getElementsByTagName('html')[0]);
	
	require(['bootstrap'], function () {});
	
	angular.element().ready(function() {
		angular.bootstrap($html, ['semaphore'])
	});
});